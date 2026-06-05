package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/dq1Mango/projects/ledmatrix/ipc"
)

func main() {
	conn, err := net.Dial("unix", ipc.SOCK)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	args := ""
	for i, arg := range os.Args[1:] {
		args += arg

		if i < len(os.Args[1:])-1 {
			args += " "
		} else {
			args += "\n"
		}
	}

	fmt.Println(args)

	conn.Write([]byte(args))

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()

		fmt.Println(line)

		if line == "done" {
			return
		}
	}

	fmt.Println("bye!")

}
