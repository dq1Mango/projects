package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/dq1Mango/projects/ledmatrix/ipc"
)

func listen_on_socket() {
	os.Remove(ipc.SOCK) // clean up any stale socket file

	listener, err := net.Listen("unix", ipc.SOCK)
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
	defer listener.Close()

	// clean up socket file on shutdown
	go func() {
		<-TERMINATE
		listener.Close()
		os.Remove(ipc.SOCK)
		TERMINATION_COMPLETE <- ""
	}()

	fmt.Println("Listening on", ipc.SOCK)

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

		action, err := parseCommand(line)

		if err != nil {

			writer.WriteString(err.Error())
			writer.WriteString("\n")
			writer.WriteString("done\n")
			writer.Flush()
			continue
		}

		writer.WriteString("Sending action...\n")
		writer.Flush()

		ActionChan <- action

		writer.WriteString("done\n")
		writer.Flush()

		// response := "echo: " + strings.ToUpper(line) + "\n"
		// writer.WriteString(response)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[%s] error: %v\n", addr, err)
	}
	fmt.Printf("[%s] disconnected\n", addr)
}

var BadUsage = errors.New("Unknown command\nSee 'ledmatrix --help' for usage\n")

func parseCommand(command string) (ipc.Action, error) {

	parts := strings.Split(command, " ")

	var action ipc.Action
	var err error

	switch parts[0] {
	case "mode":
		action, err = parseMode(parts[1])
	default:
		err = BadUsage
	}

	return action, err

}

var InvalidModeError = errors.New("Invalid mode name")

func parseMode(mode string) (ipc.Action, error) {
	modeNum, ok := ipc.ModeMap[mode]

	if !ok {
		return nil, InvalidModeError
	}

	message := &ipc.SetMode{Mode: modeNum}

	return message, nil

}
