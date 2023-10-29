package main

import (
	"bytes"
	"net/http"
)

/*
An example of replacing the Transport on the `http.DefaultClient` to
allow for tesing requests and responses
*/

func doRequest(httpUrl string) (*http.Response, error) {

	req, err := http.NewRequest("GET", httpUrl, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func sendRequest(httpUrl, data string) (*http.Response, error) {
	req, err := http.NewRequest("POST", httpUrl, bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
