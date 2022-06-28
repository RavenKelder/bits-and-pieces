package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	nethttp "net/http"
	"os"
)

// NewDebuggingTransport Creates a new transport that logs to the specified
// writer on any request that uses this transport. WARNING: This will write
// out the request headers, including any authentication secrets.
//
//   httpTransport := NewDebuggingTransport(
//     http.DefaultTransport,
//     os.Stdout,
//   )
//
//   client := http.Client{Transport: httpTransport}
func NewDebuggingTransport(t nethttp.RoundTripper, writer io.Writer) *DebuggingTransport {
	if writer == nil {
		writer = os.Stdout
	}
	return &DebuggingTransport{
		RoundTripper: t,
		writer:       writer,
	}

}

type DebuggingTransport struct {
	nethttp.RoundTripper
	writer io.Writer
}

// RoundTrip performs a normal http request, and writes the result to the writer.
// The Method, URL, Header and Body are logged.
func (t *DebuggingTransport) RoundTrip(req *nethttp.Request) (*nethttp.Response, error) {
	// Marshal the header and body using indentation for readability.
	header, err := json.MarshalIndent(req.Header, "  ", "  ")
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

		bodyBytesIndented, err := json.MarshalIndent(bodyStruct, "  ", "  ")
		if err != nil {
			return nil, err
		}

		body = string(bodyBytesIndented)

		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create the output string, and output to the writer.
	out := fmt.Sprintf(outputFormat, req.Method, req.URL.String(), string(header), body)
	_, err = t.writer.Write([]byte(out))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to write to writer: %w",
			err,
		)
	}

	return t.RoundTripper.RoundTrip(req)
}

var outputFormat = `
Method: %v
URL: %v
Header:
  %v
Body:
  %v

`
