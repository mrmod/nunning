package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type DebugPublisher struct{}
type HttpPublisher struct {
	Url           string
	Authorization string
}

type Publisher interface {
	Publish([]DatapointMetric) error
}

func (p DebugPublisher) Publish(metrics []DatapointMetric) error {
	for _, metric := range metrics {
		log.Printf("Publishing %s with count %d", metric.Source, metric.Count)
	}
	return nil
}

func (p HttpPublisher) Publish(metrics []DatapointMetric) error {
	log.Printf("HttpPublisher: Publishing to %s", p.Url)
	if len(p.Url) == 0 {
		return fmt.Errorf("invalid url: %s", p.Url)
	}

	body := &bytes.Buffer{}
	json.NewEncoder(body).Encode(metrics)
	log.Printf("HttpPublisher: Publishing %d bytes", body.Len())
	request, err := http.NewRequest(http.MethodPut, p.Url, body)
	request.Header.Add("Authorization", p.Authorization)

	if err != nil {
		log.Printf("Error building HttpPublisher request to %s: %s", p.Url, err)
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("Error publishing metrics to %s: %s", p.Url, err)
		return err
	}
	if response.StatusCode != 200 {
		defer response.Body.Close()
		b, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("failed to publish metrics to %s: [%d] %s", p.Url, response.StatusCode, string(b))
	}
	if flagDebug {
		log.Printf("Published %d metrics to %s", len(metrics), p.Url)
	}
	return nil
}
