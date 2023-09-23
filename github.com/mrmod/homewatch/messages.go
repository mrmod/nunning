package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"time"
)

type SyslogMessage struct {
	Code      string
	Timestamp time.Time
	LogHost   string
	Service   string
	PID       string
	Action    string
	Message   string
	Command   string
}
type RenameMessage struct {
	SyslogMessage
	Src, Dest string
}
type PutMessage struct {
	SyslogMessage
	Flags    []string
	Mode     string
	Filename string
}

const (
	CloseDirCmd    = "closedir"
	CloseFileCmd   = "close"
	MkDirCmd       = "mkdir"
	OpenDirCmd     = "opendir"
	OpenFileCmd    = "open"
	PosixRenameCmd = "posix-rename"
	PutCmd         = "open"
	RenameCmd      = "rename"
	SentCmd        = "sent"
	SessionCmd     = "session"

	SftpRenameMessageType = iota
	SftpPutMessageType
	UnkonwnMessageType
)

var (
	dateDecoderV1 = regexp.MustCompile(`^<(?P<code>\d+)>(?P<month>[A-Z][a-z]*)\ *(?P<dom>\d{1,2})\ (?P<hms>\d{1,2}:\d{1,2}:\d{1,2})\ *(?P<rest>.*$)`)
	dataDecoderV2 = regexp.MustCompile(`^(?P<code>\d+) (?P<dateTime>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [+-]\d{4} UTC)\ *(?P<rest>.*$)`)
	// bodyDecoderV1 = regexp.MustCompile(`(?P<logHost>\w+)\ *(?P<service>[\w-]*)\[(?P<pid>\d+)\]:\ *(?P<cmd>[\w-]*)\ *(?P<action>.*$)`)
	bodyDecoder  = regexp.MustCompile(`^(?P<logHost>\w+)[\t\ ]*(?P<service>[\w-]*)[\t\ ]*(?P<pid>\d+)[\t\ ]*(?P<cmd>[\w-]*)\ *(?P<action>.*$)`)
	dateDecoders = []*regexp.Regexp{dateDecoderV1, dataDecoderV2}
	timeFormats  = []string{
		// dateDecoderV1
		"Jan 2, 2006 15:04:05 MST",
		// dateDecoderV2
		"2006-01-02 15:04:05 -0700 UTC",
	}
	timeStringers = []func() string{
		// dataDecoderV1
		func() string {
			zone, _ := time.Now().Zone()
			return fmt.Sprintf("$month $dom, %d $hms %s", time.Now().Local().Year(), zone)
		},
		// dateDecoderV2
		func() string {
			return "$dateTime"
		},
	}
)

func NewSyslogMessage(b []byte) *SyslogMessage {
	m := &SyslogMessage{}
	if err := m.UnmarshalText(b); err != nil {
		log.Printf("Error decoding message: %s", err)
		if flagDebug {
			log.Printf("Debug:%s", string(b))
		}
		return nil
	}
	return m
}

func extractTime(logMessage string, matches [][]int) *time.Time {

	output := []byte{}
	var t time.Time
	for i, decoder := range dateDecoders {
		template := timeStringers[i]()
		date := decoder.ExpandString(output, template, logMessage, matches[0])

		_t, err := time.Parse(timeFormats[i], string(date))
		if err != nil {
			if flagDebug {
				log.Printf("Failed to parse %s: %s", string(date), err)
			}

			continue
		}
		if flagVerbose {
			log.Printf("[%s] Parsed log message time from %s", string(date), logMessage)
		}
		t = _t
		break
	}

	return &t
}

func (m *SyslogMessage) RenameMessage() *RenameMessage {
	if m.Command != PosixRenameCmd && m.Command != RenameCmd {
		return nil
	}

	parts := strings.Split(m.Action, " ")

	if flagVerbose {
		log.Printf("Parts: %#v", parts)
	}
	return &RenameMessage{
		SyslogMessage: *m,
		Src:           strings.Trim(parts[1], "\""),
		Dest:          strings.Trim(parts[3], "\""),
	}
}

func (m *SyslogMessage) PutMessage() *PutMessage {
	if m.Command != PutCmd {
		return nil
	}
	parts := strings.Split(m.Action, " ")

	return &PutMessage{
		SyslogMessage: *m,
		Filename:      strings.Trim(parts[0], "\""),
		Flags:         strings.Split(parts[2], ","),
		Mode:          parts[len(parts)-1],
	}

}

func (m *SyslogMessage) UnmarshalText(b []byte) error {
	var (
		dateDecoder *regexp.Regexp
		matches     = [][]int{}
	)

	logMessage := string(b)

	for i, decoder := range dateDecoders {

		matches = decoder.FindAllStringSubmatchIndex(logMessage, -1)
		if len(matches) != 1 {
			if flagVerbose {
				log.Printf("No matches for decoder %d with %s", i, logMessage)
			}
			continue
		}
		if flagVerbose {
			log.Printf("Matched decoder %d", i)
		}
		dateDecoder = decoder
		break
	}
	messageTime := extractTime(logMessage, matches)

	m.Timestamp = *messageTime
	m.Code = string(dateDecoder.ExpandString([]byte{}, "$code", logMessage, matches[0]))
	m.Message = string(dateDecoder.ExpandString([]byte{}, "$rest", logMessage, matches[0]))

	if flagVerbose {
		log.Printf("Decoding: %s", m.Message)
	}
	bm := bodyDecoder.FindAllStringSubmatch(m.Message, -1)
	if flagVerbose {
		log.Printf("BM: %#v\n", bm)
	}
	if len(bm) != 1 {
		return fmt.Errorf("invalid log message body")
	}
	bodyMatches := bm[0]

	m.Command = bodyMatches[bodyDecoder.SubexpIndex("cmd")]
	m.LogHost = bodyMatches[bodyDecoder.SubexpIndex("logHost")]
	m.Service = bodyMatches[bodyDecoder.SubexpIndex("service")]
	m.PID = bodyMatches[bodyDecoder.SubexpIndex("pid")]
	m.Action = bodyMatches[bodyDecoder.SubexpIndex("action")]

	if flagVerbose {
		log.Printf("Message: %#v", m)
	}
	return nil
}

func (m *SyslogMessage) MessageType() int {
	if m.RenameMessage() != nil {
		return SftpRenameMessageType
	}
	if m.PutMessage() != nil {
		return SftpPutMessageType
	}
	return UnkonwnMessageType
}
