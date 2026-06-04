package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/my.sock")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.Write([]byte("hello from client\n"))

	time.Sleep(1 * time.Second)

	conn.Write([]byte("also from client\n"))

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	fmt.Printf("server said: %s\n", buf[:n])
}
