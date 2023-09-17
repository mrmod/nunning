package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

/*
	Entry

A single journald journal entry
*/
type Entry struct {
	BootId              string `json:"_boot_id"`
	Cursor              string `json:"__cursor"`
	Hostname            string `json:"_hostname"`
	Message             string `json:"message"`
	MonotonicTimestamp  string `json:"__monotonic_timestamp"`
	PID                 string `json:"_pid"`
	RealtimeTimestamp   string `json:"__realtime_timestamp"`
	StreamId            string `json:"_stream_id"`
	SystemdInvocationId string `json:"_systemd_invocation_id"`
	SyslogFacility      string `json:"syslog_facility"`
	UID                 string `json:"_uid"`
}

/* LoadJournal
 * Load a journal file into a slice of Entry structs
 */
func LoadJournal(filename string) []Entry {
	journalFile, err := os.Open(filename)

	if err != nil {
		panic(err)
	}

	defer journalFile.Close()
	log.Printf("Loading journal of %d bytes", func() int64 {
		s, _ := journalFile.Stat()
		return s.Size()
	}())
	decoder := json.NewDecoder(journalFile)
	var decodeError error
	journalEntries := []Entry{}
	for {
		entry := Entry{}
		err := decoder.Decode(&entry)
		if err == io.EOF {
			break
		}
		if decodeError != nil && decodeError != io.EOF {
			log.Fatal(decodeError)
		}
		// log.Printf("Adding %#v", entry)
		journalEntries = append(journalEntries, entry)
	}

	return journalEntries
}
