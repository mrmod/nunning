package main

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

var (
	eventWithDetection = "eventWithDetection.json"
)

func TestEventJSONSerialization(t *testing.T) {
	var event Event
	b, err := os.ReadFile(eventWithDetection)
	if err != nil {
		t.Fatalf("Expected to read event but failed %s", err)
		t.FailNow()
	}
	err = json.Unmarshal(b, &event)
	if err != nil {
		t.Fatalf("Expected to unmarshal event but failed %s", err)
		t.FailNow()
	}
	_, err = json.Marshal(event)
	if err != nil {
		t.Fatalf("Expected to JSON serialized object, got %s", err)
	}
}

func TestDatapointSerialization(t *testing.T) {
	datapoint := Datapoint{
		Source:    "source",
		Count:     1,
		Detection: nil,
	}
	b, err := json.Marshal(datapoint)
	if err != nil {
		t.Fatalf("Expected to serialize datapoint, got %s", err)
	}
	s := string(b)

	matched, _ := regexp.MatchString(`null`, s)

	if !matched {
		t.Fatalf("Expected Detection to be nil, got %s", s)
	}
}
