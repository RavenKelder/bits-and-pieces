package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	nethttp "net/http"
)

// DebuggingTransport acts as a middleware to log requests that are being made.
// Use by wrapping an existing http.Transport
//
//   httpTransport := &DebuggingTransport{
//	   RoundTripper: baseHttpTransport,
//   }
//
type DebuggingTransport struct {
	nethttp.RoundTripper
}

// RoundTrip performs a normal http request, and logs the results using package fmt.
// The Method, URL, Header and Body are logged.
func (t *DebuggingTransport) RoundTrip(req *nethttp.Request) (*nethttp.Response, error) {
	data, err := json.MarshalIndent(req.Header, "        ", "  ")
	if err != nil {
		return nil, err
	}

	var body string
	if req.Body != nil {
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		var bodyStruct interface{}

		err = json.Unmarshal(bodyBytes, &bodyStruct)
		if err != nil {
			return nil, err
		}

		bodyBytesIndented, err := json.MarshalIndent(bodyStruct, "", "  ")
		if err != nil {
			return nil, err
		}

		body = string(bodyBytesIndented)

		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	fmt.Printf("REQ %v %v\nHeader: %v\nBody: %v\n", req.Method, req.URL.String(), string(data), body)
	// Do not just do t.RoundTrip(req) since that makes this recursive.
	return t.RoundTripper.RoundTrip(req)
}
