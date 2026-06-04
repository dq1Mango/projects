package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func listen_on_socket() {
	socketPath := "/tmp/my.sock"
	os.Remove(socketPath) // clean up any stale socket file

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
	defer listener.Close()

	// clean up socket file on shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		listener.Close()
		os.Remove(socketPath)
		os.Exit(0)
	}()

	fmt.Println("Listening on", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {

			if errors.Is(err, net.ErrClosed) {
				return
			}

			fmt.Println("accept error:", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	fmt.Printf("[%s] connected\n", addr)

	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("[%s] received: %s\n", addr, line)

		response := "echo: " + strings.ToUpper(line) + "\n"
		writer.WriteString(response)
		writer.Flush()
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[%s] error: %v\n", addr, err)
	}
	fmt.Printf("[%s] disconnected\n", addr)
}
