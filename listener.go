package main

import (
	"fmt"
	"net"
)

type onReceiveCallback func(rawRequest []byte) (rawResponse []byte)

type Listener struct {
	Port      int
	OnReceive onReceiveCallback
}

func NewListener(port int, onReceive onReceiveCallback) *Listener {
	return &Listener{
		Port:      port,
		OnReceive: onReceive,
	}
}

func (l *Listener) handleConnection(conn net.Conn) {
	defer conn.Close()

	// max size: 5MB
	buf := make([]byte, 5_000_000)

	n, err := conn.Read(buf)

	if err != nil {
		fmt.Printf("Failed to read connection: %w\n", err)
		return
	}

	request := buf[:n]
	response := l.OnReceive(request)

	conn.Write(response)
}

func (l *Listener) Listen() {
	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", l.Port))
	if err != nil {
		panic(err)
	}

	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept client connection: %w\n", err)
			continue
		}

		go l.handleConnection(conn)
	}
}
