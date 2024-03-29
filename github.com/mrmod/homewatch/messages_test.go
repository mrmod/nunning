package main

import (
	"io"
	"os"
	"testing"
)

var (
	renameMessage        = "testdata/internalSftp.rename.syslogMessage"
	sessionClosedMessage = "testdata/internalSftp.sessionClosed.syslogMessage"
	putMessage           = "sftp.put.message"
	unknownCommand       = "sftp.unknownCommand.message"
)

func TestUnmarshalSessionClosedMessage(t *testing.T) {
	fp, err := os.Open(sessionClosedMessage)
	if err != nil {
		t.Fail()
	}
	defer fp.Close()

	b, err := io.ReadAll(fp)
	if err != nil {
		t.Fail()
	}
	m := SyslogMessage{}
	err = m.UnmarshalText(b)
	if err != nil {
		t.Fatalf("Expected to decode the messsage, got %s", err)
	}
	if cmd := m.Command; cmd != "session" {
		t.Fatalf("Expected 'session', got %s", cmd)
	}
}

func TestUnmarshalSyslogMessageText(t *testing.T) {
	fp, err := os.Open(renameMessage)
	if err != nil {
		t.Fail()
	}
	defer fp.Close()

	b, err := io.ReadAll(fp)
	if err != nil {
		t.Fail()
	}
	m := SyslogMessage{}
	err = m.UnmarshalText(b)
	if err != nil {
		t.Fatalf("Expected to decode the messsage, got %s", err)
	}

	// TODO: The timezone interpretation may be different on build systems
	// s := m.Timestamp.Format("Jan 2, 2006 15:04:05 MST")
	// expectation := "Mar 2, 2023 08:13:51 PST"
	// if s != expectation {
	// 	t.Fatalf("Expected %s, got %s", expectation, s)
	// }

	if c := m.Code; c != "190" {
		t.Fatalf("Expected '190', got %s", c)
	}

	if cmd := m.Command; cmd != "posix-rename" && cmd != "rename" {
		t.Fatalf("Expected 'posix-rename' or 'rename', got %s", cmd)
	}
	if action := m.Action; len(action) == 0 {
		t.Fatalf("Expected a non-zero action length")
	}
}

func TestParseRenameMessage(t *testing.T) {
	fp, err := os.Open(renameMessage)
	if err != nil {
		t.Fail()
	}
	defer fp.Close()

	b, err := io.ReadAll(fp)
	if err != nil {
		t.Fail()
	}
	m := SyslogMessage{}
	err = m.UnmarshalText(b)
	if err != nil {
		t.Fatalf("Expected to decode the messsage, got %s", err)
	}

	rm := m.RenameMessage()
	if rm == nil {
		t.Fatalf("Expected a RenameMessage, got %v", rm)
	}
	srcExpect := "\\\"/home/cameras/SomeTestCamera/2000-01-01/001/dav/00/00.35.41-00.35.41[M][0@0][0].dav_\\"
	if rm.Src != srcExpect {
		t.Fatalf("Expected '%s', got %s", srcExpect, rm.Src)
	}
	destExpect := "\\\"/home/cameras/SomeTestCamera/2000-01-01/001/dav/00/00.35.43-00.42.22[M][0@0][0].dav\\"
	if rm.Dest != destExpect {
		t.Fatalf("Expected '%s', got %s", destExpect, rm.Dest)
	}
}

func xTestParsePutMessage(t *testing.T) {
	fp, err := os.Open(putMessage)
	if err != nil {
		t.Fail()
	}
	defer fp.Close()

	b, err := io.ReadAll(fp)
	if err != nil {
		t.Fail()
	}
	m := SyslogMessage{}
	err = m.UnmarshalText(b)
	if err != nil {
		t.Fatalf("Expected to decode the messsage, got %s", err)
	}

	pm := m.PutMessage()

	if pm == nil {
		t.Fatalf("Expected to decode a PutMessage")
	}
	if pm.Mode != "0644" {
		t.Fatalf("Expected mode '0644', got %s", pm.Mode)
	}
	expectFilename := "/mnt/VideoUploads/./badData.json"
	if pm.Filename != expectFilename {
		t.Fatalf("Expected filename %s, got %s", expectFilename, pm.Filename)
	}
	if len(pm.Flags) != 3 {
		t.Fatalf("Expected 3 flags, got %#v", pm.Flags)
	}

}

func xTestNegativetests(t *testing.T) {
	parseMessage := func(messageFile string) *SyslogMessage {
		fp, err := os.Open(messageFile)
		if err != nil {
			t.Logf("Unable to open %s", messageFile)
			return nil
		}
		defer fp.Close()

		b, err := io.ReadAll(fp)
		if err != nil {
			t.Logf("Unable to read %s", messageFile)
			return nil
		}
		m := SyslogMessage{}
		err = m.UnmarshalText(b)
		if err != nil {
			t.Logf("Unable to parse %s", messageFile)
			return nil
		}
		return &m
	}
	m := parseMessage(putMessage)
	if m == nil {
		t.FailNow()
	}
	if m.Command != PutCmd {
		t.Fatalf("Expected PutCommand, got %s", m.Command)
	}
	if rm := m.RenameMessage(); rm != nil {
		t.Fatalf("Expected no rename message")
	}

	m = parseMessage(renameMessage)
	if m == nil {
		t.FailNow()
	}
	if m.Command != PosixRenameCmd {
		t.Fatalf("Expected PosixRenameCmd, got %s", m.Command)
	}
	if pm := m.PutMessage(); pm != nil {
		t.Fatalf("Expected no put message")
	}
	m = parseMessage(unknownCommand)
	if m.Command != "unknown-command" {
		t.Fatalf("Unexpected command %s", m.Command)
	}
	if m == nil {
		t.Fatalf("Expected no message, got %#v", m)
	}
	if m.PutMessage() != nil || m.RenameMessage() != nil {
		t.Fatalf("Expected no decodable action")
	}

}
