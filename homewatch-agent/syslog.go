package main

import (
	"log"
	"net"
)

const (
	stopSyslogServer = iota
)

type SyslogServer struct {
	// BindAddress: IP:Port to bind to
	BindAddress string
	// DatagramSize: Size of message buffer
	DatagramSize int
	control      chan int
}

func NewSyslogServer(bindAddress string) SyslogServer {
	return SyslogServer{
		BindAddress:  bindAddress,
		DatagramSize: 64 * 1024,
		control:      make(chan int, 1),
	}
}
func (s SyslogServer) Stop() {
	s.control <- stopSyslogServer
}
func (s SyslogServer) Serve(stream chan *SyslogMessage) {

	go func() {
		addr, err := net.ResolveUDPAddr("udp", s.BindAddress)
		if err != nil {
			log.Fatalf("Unable to resolve address %s: %s", s.BindAddress, err)
		}

		listener, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatalf("Unable to bind listener to %s: %s", addr, err)
		}
		listener.SetReadBuffer(s.DatagramSize)

		log.Printf("Started listener on %s", listener.LocalAddr())
		for {
			data := make([]byte, s.DatagramSize)
			// byteCount, connectionAddress, err := listener.ReadFrom(data)
			byteCount, err := listener.Read(data)
			if err != nil {
				log.Printf("Failed to read: %s", err)
				continue
			}
			if byteCount > 0 {
				message := NewSyslogMessage(data[0:byteCount])
				if flagVerbose {
					log.Printf("Dispatching message to stream %v", message)
				}
				stream <- message
			}
		}
	}()
	<-s.control
	log.Printf("Quitting camera event streamer")
}
