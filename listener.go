package httpatos

import (
	"fmt"
	"io"
	"net"
)

type onReceiveCallback func(rawRequest []byte) (rawResponse []byte)

type listener struct {
	Port      int
	OnReceive onReceiveCallback
}

func newListener(port int, onReceive onReceiveCallback) *listener {
	return &listener{
		Port:      port,
		OnReceive: onReceive,
	}
}

func (l *listener) handleConnection(conn net.Conn) {
	defer conn.Close()

	// max size: 5MB
	buf := make([]byte, 5_000_000)

	n, err := conn.Read(buf)

	if err == io.EOF {
		// sem mais dados para ler
		return
	}

	if err != nil {
		fmt.Printf("Failed to read connection: %v\n", err)
		return
	}

	request := buf[:n]
	response := l.OnReceive(request)

	conn.Write(response)
}

func (l *listener) listen() {
	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", l.Port))
	if err != nil {
		panic(err)
	}

	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept client connection: %v\n", err)
			continue
		}

		go l.handleConnection(conn)
	}
}
