package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	netHttp "net/http"
	"os"

	"github.com/RavenKelder/bits-and-pieces/golang/http"

	"github.com/hashicorp/go-cleanhttp"
)

func main() {
	// Create basic transport and wrap with the debugging transport.
	baseTransport := cleanhttp.DefaultTransport()

	t := http.NewDebuggingTransport(
		baseTransport, os.Stdout,
	)

	// Attach transport to an http client.
	client := cleanhttp.DefaultClient()
	client.Transport = t

	// Create a JSON body as a buffer.
	body := bytes.NewBufferString(`{"my_field": "my_value"}`)

	// Make a new request.
	request, err := netHttp.NewRequest(
		"GET",
		"https://localhost:12345",
		io.NopCloser(body),
	)
	if err != nil {
		log.Fatalf("Unable to create new request: %v", err)
	}

	request.Header.Add("My-Header", "header_value")
	request.Header.Add("My-Header", "header_value-2")
	request.Header.Add("My-Header-2", "header_value-3")

	// Perform the request.
	res, err := client.Do(request)
	if err != nil {
		log.Fatalf("Unable to perform request: %v", err)
	}

	// Log the result.
	log.Printf("Received status %s", res.Status)

	if res.Body != nil {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}

		log.Printf("Received body: %s", string(bodyBytes))
	}
}
