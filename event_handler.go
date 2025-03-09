package main

import (
	"fmt"
	"log"
	"time"
)

type IndexEventHandler struct {
	ConsolidationInterval time.Duration
	Datapoints            []Datapoint
	events                chan *IndexedEvent
	publisher             Publisher
	listenerControl       chan int
	publisherControl      chan int
	isLocked              bool
	datapointMetrics      map[string]DatapointMetric
	FileUploader          *S3FileUploader
}

func NewIndexEventHandler(interval string, events chan *IndexedEvent, publisher Publisher) *IndexEventHandler {
	d, err := time.ParseDuration(interval)
	if err != nil {
		log.Printf("invalid interval %s, defaulting to 5m: %s", interval, err)
		d = time.Minute * 5
	}
	return &IndexEventHandler{
		ConsolidationInterval: d,
		publisher:             publisher,
		events:                events,
		listenerControl:       make(chan int, 1),
		publisherControl:      make(chan int, 1),
		datapointMetrics:      map[string]DatapointMetric{},
	}
}
func (h *IndexEventHandler) Lock() {
	h.isLocked = true
	if flagVerbose {
		log.Printf("Locked")
	}
}
func (h *IndexEventHandler) Unlock() {
	h.isLocked = false
	if flagVerbose {
		log.Printf("Unlocked")
	}
}
func (h *IndexEventHandler) Stop() {
	h.listenerControl <- 1
	h.publisherControl <- 1

}
func (h *IndexEventHandler) Publisher() {
	log.Printf("EventHandler Publisher started")
	go func() {
		for {
			time.Sleep(h.ConsolidationInterval)
			// Spin waiting on lock to release
			for h.isLocked {
				time.Sleep(time.Millisecond * 5)
			}
			h.Lock()
			if flagDebug {
				log.Printf("Consolidating %d datapoints", len(h.Datapoints))
			}
			for _, datapoint := range h.Datapoints {
				dp, exists := h.datapointMetrics[datapoint.Source]
				if exists {
					dp.Count += datapoint.Count
					if datapoint.Detection != nil {
						dp.Detections = append(dp.Detections, *datapoint.Detection)
					}
					h.datapointMetrics[datapoint.Source] = dp
				} else {
					detections := []RuleDetection{}
					if datapoint.Detection != nil {
						detections = append(detections, *datapoint.Detection)
					}
					h.datapointMetrics[datapoint.Source] = DatapointMetric{datapoint, detections}
				}
			}
			// Memory of datapoint presence
			h.Datapoints = []Datapoint{}
			if flagVerbose {
				log.Printf("Cleared datapoints")
			}
			h.Unlock()
			if err := h.publisher.Publish(flattenMetrics(h.datapointMetrics)); err != nil {
				log.Printf("Error while publishing: %s", err)
			}
			for source, datapoint := range h.datapointMetrics {
				datapoint.Count = 0
				datapoint.Detections = []RuleDetection{}
				h.datapointMetrics[source] = datapoint
			}
		}
	}()
	<-h.publisherControl
}

func flattenMetrics(metricMap map[string]DatapointMetric) []DatapointMetric {
	var datapointMetrics []DatapointMetric

	for _, datapointMetric := range metricMap {
		datapointMetrics = append(datapointMetrics, datapointMetric)
	}
	return datapointMetrics
}

func (h *IndexEventHandler) Listen() {
	log.Printf("EventHandler Listener started")

	go func() {
		for event := range h.events {

			newDatapoints := CreateDatapoints(event)

			for h.isLocked {
				time.Sleep(time.Millisecond * 50)
			}

			h.Lock()
			h.Datapoints = append(h.Datapoints, newDatapoints...)
			h.Unlock()
		}
	}()
	<-h.listenerControl
}

type Datapoint struct {
	Source    string         `json:"source"`
	Count     int            `json:"count"`
	Detection *RuleDetection `json:"detection"`
}
type DatapointMetric struct {
	Datapoint
	Detections []RuleDetection `json:"detections"`
}

func CreateDatapoints(event *IndexedEvent) []Datapoint {
	var datapoints []Datapoint

	for _, e := range event.Events {
		rd := e.RuleDetection()
		datapoints = append(datapoints, Datapoint{
			Source:    mkSource(event, e, rd),
			Count:     1,
			Detection: rd,
		})
	}
	return datapoints
}

func mkSource(ie *IndexedEvent, e Event, rd *RuleDetection) string {
	if ie == nil {
		return "EventMissing"
	}
	source := fmt.Sprintf("%s:%s:%s", ie.Source, e.Action, e.Name)
	if rd == nil {
		return source
	}

	return fmt.Sprintf("%s:%s", source, rd.Object.ObjectType)
}
