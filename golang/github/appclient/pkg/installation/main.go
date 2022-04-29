package installation

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"github.com/sirupsen/logrus"
)

// Client wraps the "github.com/google/go-github".Client to function with
// multiple Github App installations to allow increased rate limit by distributing
// API calls between Apps.
type Client struct {
	baseRoundTripper http.RoundTripper
	baseClient       *http.Client
	manager          *githubRoundTripperManager
	log              logrus.FieldLogger
	*github.Client
}

type ClientOptions struct {
	BaseClient       *http.Client
	BaseRoundTripper http.RoundTripper
	Logger           logrus.FieldLogger
	MinimumRateLimit int
}

func NewClient(opts *ClientOptions) *Client {
	m := &githubRoundTripperManager{
		roundTrippers:    []*githubRoundTripper{},
		MinimumRateLimit: opts.MinimumRateLimit,
	}
	c := *opts.BaseClient
	c.Transport = m
	return &Client{
		baseRoundTripper: opts.BaseRoundTripper,
		baseClient:       &c,
		manager:          m,
		log:              opts.Logger,
		Client:           github.NewClient(&c),
	}
}

func (c *Client) Register(
	ctx context.Context,
	appID int64,
	installationID int64,
	privateKey *rsa.PrivateKey,
) error {
	if c.manager.contains(installationID) {
		return fmt.Errorf("installation ID %d is already registered", installationID)
	}

	appRoundTripper := ghinstallation.NewAppsTransportFromPrivateKey(
		c.baseRoundTripper, appID, privateKey,
	)
	installationRoundTripper := ghinstallation.NewFromAppsTransport(
		appRoundTripper, installationID,
	)
	tempBaseClient := *c.baseClient
	tempBaseClient.Transport = installationRoundTripper
	tempClient := github.NewClient(&tempBaseClient)
	rateLimits, _, err := tempClient.RateLimits(ctx)
	if err != nil {
		return fmt.Errorf("failed to get rate limit: %w", err)
	}

	c.manager.roundTrippers = append(c.manager.roundTrippers, &githubRoundTripper{
		RoundTripper: installationRoundTripper,
		githubRoundTripperInfo: githubRoundTripperInfo{
			AppID:              appID,
			InstallationID:     installationID,
			RateLimitRemaining: rateLimits.GetCore().Remaining,
			RatelimitReset:     time.Unix(rateLimits.Core.Reset.Unix(), 0),
		},
	})

	return nil
}

func (c *Client) Status() []githubRoundTripperInfo {
	roundTripperInfo := []githubRoundTripperInfo{}
	for _, roundTripper := range c.manager.roundTrippers {
		info := roundTripper.githubRoundTripperInfo
		roundTripperInfo = append(roundTripperInfo, info)
	}

	return roundTripperInfo
}

type githubRoundTripper struct {
	http.RoundTripper
	githubRoundTripperInfo
}

type githubRoundTripperInfo struct {
	AppID              int64
	InstallationID     int64
	RateLimitRemaining int
	RatelimitReset     time.Time
}

type githubRoundTripperManager struct {
	currentRoundTripper *githubRoundTripper
	roundTrippers       []*githubRoundTripper
	MinimumRateLimit    int
}

func (m *githubRoundTripperManager) next() (*githubRoundTripper, error) {
	if m.currentRoundTripper != nil &&
		(m.currentRoundTripper.RateLimitRemaining > m.MinimumRateLimit ||
			time.Since(m.currentRoundTripper.RatelimitReset) > 0) {
		return m.currentRoundTripper, nil
	}

	for _, r := range m.roundTrippers {
		if r.RateLimitRemaining > m.MinimumRateLimit ||
			time.Since(r.RatelimitReset) > 0 {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no remaining installations with rate limit remaining")
}

func (m *githubRoundTripperManager) contains(installationID int64) bool {
	for _, roundTripper := range m.roundTrippers {
		if roundTripper.InstallationID == installationID {
			return true
		}
	}
	return false
}

func (m *githubRoundTripperManager) RoundTrip(req *http.Request) (*http.Response, error) {
	r, err := m.next()
	if err != nil {
		return nil, err
	}
	res, err := r.RoundTrip(req)
	if err != nil {
		return res, err
	}

	remaining := res.Header.Get("X-Ratelimit-Remaining")
	remainingInt, err := strconv.Atoi(remaining)
	if err != nil {
		return res, err
	}

	reset := res.Header.Get("X-Ratelimit-Reset")
	resetInt, err := strconv.Atoi(reset)
	if err != nil {
		return res, err
	}
	resetTime := time.Unix(int64(resetInt), 0)

	if resetTime.Sub(r.RatelimitReset) > 0 ||
		(r.RateLimitRemaining > remainingInt &&
			resetTime.Sub(r.RatelimitReset) <= 0) {
		r.RateLimitRemaining = remainingInt
		r.RatelimitReset = resetTime
	}
	return res, nil
}
