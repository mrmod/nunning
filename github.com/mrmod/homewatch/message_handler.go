package main

import "log"

type SyslogMessageHandler struct {
	Messages    chan *SyslogMessage
	Events      chan *IndexedEvent
	IndexEvents chan string
	VideoEvents chan string
	control     chan int
}

func NewSyslogMessageHandler() SyslogMessageHandler {
	return SyslogMessageHandler{
		make(chan *SyslogMessage, 1),
		make(chan *IndexedEvent, 1),
		make(chan string, 1),
		make(chan string, 1),
		make(chan int, 1),
	}
}
func (s SyslogMessageHandler) Run() {
	defer close(s.Messages)
	defer close(s.Events)
	go func() {
		if flagDebug {
			log.Printf("Message handler starter")
		}
		for message := range s.Messages {
			if flagVerbose {
				log.Printf("SyslogMessage recieved")
			}
			switch messageType := message.MessageType(); messageType {
			case SftpRenameMessageType:
				renameMessage := message.RenameMessage()
				if flagDebug {
					log.Printf("Renamed %s to %s", renameMessage.Src, renameMessage.Dest)
				}
				if isIndexFilePath(renameMessage.Dest) {
					s.IndexEvents <- renameMessage.Dest
				}
				if isVideoFilePath(renameMessage.Dest) {
					s.VideoEvents <- renameMessage.Dest
				}

			default:
				if flagVerbose {
					log.Printf("Uknown message: %s", message.Message)
				}
			}
		}
	}()
	<-s.control
}
