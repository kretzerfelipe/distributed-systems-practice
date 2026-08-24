package main

import (
	"fmt"
	"net"
	"time"
)

const ConnectTimeout = 10 * time.Second

func Listen(port int) (net.Conn, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	fmt.Printf("Listening on %s. Waiting for a peer...\n", listener.Addr())

	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Connect(host string, port int) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Connecting to %s...\n", addr)

	conn, err := net.DialTimeout("tcp", addr, ConnectTimeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
