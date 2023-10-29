package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// playbackTripper is a http.RoundTripper that replays the same response
// for every request.
type playbackTripper struct {
	response   []byte
	statusCode int
}

// inputTripper is a http.RoundTripper that replays the same response
// for every request
// and validates the request body
type inputTripper struct {
	playbackTripper
	request []byte
}

// RoundTrip implements http.RoundTripper.
// It returns the same response for every request.
func (p playbackTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: p.statusCode,
		Body:       io.NopCloser(bytes.NewBuffer(p.response)),
	}, nil
}

// RoundTrip implements http.RoundTripper.
// It validates the request body and returns the same response for every request.
// It returns `200` when the request body matches the expected request body
func (i inputTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	inputRequest, _ := io.ReadAll(r.Body)
	if string(inputRequest) != string(i.request) {
		return (&playbackTripper{
			response:   i.response,
			statusCode: i.playbackTripper.statusCode,
		}).RoundTrip(r)
	}

	return (&playbackTripper{response: i.response, statusCode: http.StatusOK}).RoundTrip(r)
}

func TestDoRequestWithInput(t *testing.T) {
	defaultTransport := http.DefaultClient.Transport
	defer func() {
		http.DefaultClient.Transport = defaultTransport
	}()
	expectedResponse := []byte("RecordedResponsePlayback")
	expectedRequest := []byte("ExpectedRequest")
	http.DefaultClient.Transport = inputTripper{
		request: expectedRequest,
		playbackTripper: playbackTripper{
			response:   expectedResponse,
			statusCode: http.StatusBadRequest,
		},
	}

	response, err := sendRequest("http://example.com", string(expectedRequest)+"cat")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected %d : BadRequest, got %d", http.StatusBadRequest, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(expectedResponse) {
		t.Fatalf("Expected %s, got %s", expectedResponse, body)
	}
}

func TestDoRequestWithPlayback(t *testing.T) {
	defaultTransport := http.DefaultClient.Transport
	defer func() {
		http.DefaultClient.Transport = defaultTransport
	}()

	expectedResponse := []byte("RecordedResponsePlayback")
	http.DefaultClient.Transport = playbackTripper{
		response:   expectedResponse,
		statusCode: http.StatusOK,
	}

	response, err := doRequest("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(expectedResponse) {
		t.Fatalf("Expected %s, got %s", expectedResponse, body)
	}
}

func TestDoRequest(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "TestResponse")
			}),
	)

	// http.DefaultClient.Transport = server

	defer server.Close()

	resp, err := doRequest(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "TestResponse\n" {
		t.Fatalf("Expected %s, got %s", "TestResponse", string(body))
	}
}
