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
	dateDecoder = regexp.MustCompile(`^<(?P<code>\d+)>(?P<month>[A-Z][a-z]*)\ *(?P<dom>\d{1,2})\ (?P<hms>\d{1,2}:\d{1,2}:\d{1,2})\ *(?P<rest>.*$)`)
	bodyDecoder = regexp.MustCompile(`(?P<logHost>\w+)\ *(?P<service>[\w-]*)\[(?P<pid>\d+)\]:\ *(?P<cmd>[\w-]*)\ *(?P<action>.*$)`)
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

	zone, _ := time.Now().Zone()
	template := fmt.Sprintf("$month $dom, %d $hms %s", time.Now().Local().Year(), zone)

	output := []byte{}

	date := dateDecoder.ExpandString(output, template, logMessage, matches[0])

	t, err := time.Parse("Jan 2, 2006 15:04:05 MST", string(date))
	if err != nil {
		log.Printf("Failed to parse %s: %s", string(date), err)
	}

	return &t
}

func (m *SyslogMessage) RenameMessage() *RenameMessage {
	if m.Command != PosixRenameCmd && m.Command != RenameCmd {
		return nil
	}

	parts := strings.Split(m.Action, " ")

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
	logMessage := string(b)
	matches := dateDecoder.FindAllStringSubmatchIndex(logMessage, -1)
	if len(matches) != 1 {
		if flagVerbose {
			log.Printf("No matches for %s", logMessage)
		}
		return fmt.Errorf("invalid log message structure")
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
