package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	UnknownSource = "UnknownSource"
)

type Event struct {
	Action string
	Data   json.RawMessage
	Index  int
	Name   string
}

// XYPoing: X,Y
type XYPoint []int

// BoundingBox: X,Y, W, H
type BoundingBox []int

// RGBA: R,G, B,Alpha
type RGBA []int
type BaseEventData struct {
	Action  string
	Name    string
	EventID int
	GroupID int
}
type RuleDetection struct {
	BaseEventData
	Class        string
	CountInGroup int
	DetectRegion []XYPoint
	Object       ObjectDetection
	PTS          float64
	RuleID       int
	Track        interface{}
	UTC, UTCMS   int64
}
type ObjectDetection struct {
	Action       string
	BoundingBox  BoundingBox
	BrandYear    int
	CarLogoIndex int
	Category     string
	Center       XYPoint
	Confidence   float64
	MainColor    RGBA
	ObjectID     int
	ObjectType   string
	RelativeID   int
	Speed        int
	SubBrand     int
	Text         string
}

type AudioEncoding map[string]interface{}
type VideoEncoding map[string]interface{}
type Encoding struct {
	AudioEnable bool
	VideoEnable bool
	Audio       AudioEncoding
	Video       VideoEncoding
}

// TextUnmarshaler
type Frame struct {
	Type          string
	FrameNumber   int64
	StartOffsetMs int64
	EndOffsetMs   int64
	IsFirstFrame  bool
}

type IndexedEvent struct {
	Source    string
	Events    []Event
	Encodings []Encoding
	Frames    []Frame
}

func tryRemove(filepath string) {

	_, err := os.Stat(filepath)
	if err != nil {
		log.Printf("Unable to stat %s:\n\t%s", filepath, err)
		if flagDebug {
			debugFilepath(filepath)
		}
	}
	err = os.Remove(filepath)
	if os.IsNotExist(err) {
		log.Printf("Unable to remove %s:\n\t%s", filepath, err)
		return
	}
	if err != nil {
		log.Printf("Failed to remove %s:\n\t%s", filepath, err)
	} else {
		log.Printf("Removed %s", filepath)
	}
}
func NewIndexedEvent(filepath string, cleanup bool) *IndexedEvent {
	event, err := ReadIndex(filepath)
	event.Source = GetSourceFromPath(filepath)

	defer func() {
		if cleanup {
			tryRemove(filepath)
		}
	}()

	if err != nil {
		log.Printf("Error reading %s: %s", filepath, err)
		return nil
	}

	return &event
}
func ReadIndex(filepath string) (IndexedEvent, error) {
	var event IndexedEvent
	fp, err := os.Open(filepath)
	_, filename := path.Split(filepath)
	if err != nil {
		log.Printf("Unable to read indexed event %s: %s", filename, err)
		return event, err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)

	return demuxIndex(reader)
}

/* demuxIndex: Locate the Frames, Encodings, and Events in a .idx reader
 */
func demuxIndex(reader *bufio.Reader) (IndexedEvent, error) {
	event := IndexedEvent{}

	data, err := reader.ReadBytes('\n')
	for err != io.EOF {
		line := bytes.SplitN(data, []byte("="), 2)
		lineType := string(line[0])
		lineData := line[1]
		data, err = reader.ReadBytes('\n')

		switch {
		case lineType == "Frame":
			frame := &Frame{}

			decodeError := frame.UnmarshalText(lineData)
			if decodeError != nil {
				return event, decodeError
			}
			event.Frames = append(event.Frames, *frame)

		case lineType == "Event":
			v := &Event{}
			decodeError := json.Unmarshal(lineData, v)
			if decodeError != nil {
				return event, decodeError
			}

			event.Events = append(event.Events, *v)
		case lineType == "EncodeFormat":
			v := &Encoding{}
			decodeError := json.Unmarshal(lineData, v)
			if decodeError != nil {
				return event, decodeError
			}
			event.Encodings = append(event.Encodings, *v)

			if len(event.Encodings) > 1 {
				log.Printf("Warning: Found multiple encodings")
			}
		default:
			log.Printf("Unknown line type: %s", lineType)
		}
	}
	return event, nil
}

func (frame *Frame) UnmarshalText(b []byte) error {
	line := strings.ReplaceAll(string(b), "\"", "")
	fields := strings.Split(line, ",")

	frame.Type = fields[0]
	if frameNumber, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
		frame.FrameNumber = int64(frameNumber)
	}
	if startOffsetMs, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
		frame.StartOffsetMs = startOffsetMs
	}
	if endOffsetMs, err := strconv.ParseInt(fields[3], 10, 64); err == nil {
		frame.EndOffsetMs = endOffsetMs
	}
	if v, err := strconv.ParseBool(fields[len(fields)-1]); err == nil {
		frame.IsFirstFrame = v
	}
	return nil
}

func (e Event) RuleDetection() *RuleDetection {
	var ruleDetection RuleDetection
	err := json.Unmarshal([]byte(e.Data), &ruleDetection)
	if err == nil {
		return &ruleDetection
	}
	log.Printf("Unable to decode rule detection: %s", err)
	return nil
}

func isIndexFilePath(filepath string) bool {
	return strings.HasSuffix(filepath, ".idx")
}

func isVideoFilePath(filepath string) bool {
	return strings.HasSuffix(filepath, ".dav")
}

func GetSourceFromPath(filepath string) string {
	s := strings.Split(filepath, string(os.PathSeparator))
	if len(s)-6 >= 0 {
		return s[len(s)-6]
	}
	return UnknownSource
}
