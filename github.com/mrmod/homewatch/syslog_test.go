package main

import (
	"io"
	"os"
	"testing"
)

func TestMessageHandler(t *testing.T) {
	flagVerbose = true
	flagDebug = true
	fp, err := os.Open(renameMessage)
	if err != nil {
		t.Fatalf("Unable to open test data: %s", err)
	}
	data, err := io.ReadAll(fp)
	if err != nil {
		t.Fatalf("Unable to read test data: %s", err)
	}

	syslogMessage := NewSyslogMessage(data)

	if messageType := syslogMessage.MessageType(); messageType != SftpRenameMessageType {
		t.Fatalf("Expected SftpRenameMessageType, got %v", messageType)
	}
}
