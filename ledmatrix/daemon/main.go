package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.bug.st/serial"
)

const (
	LEFT  = "/dev/ttyACM1"
	RIGHT = "/dev/ttyACM0"
)

const (
	WIDTH  = 9
	HEIGHT = 34
)

const (
	CMD_STAGE_COL  = 0x07
	CMD_FLUSH_COLS = 0x08
)

const BAUD_RATE = 115200

const MIN_BRIGHTNESS = 5

var TERMINATE chan os.Signal = make(chan os.Signal)
var TERMINATION_COMPLETE chan any = make(chan any)

// serial data is written by collumn, so the rows and collumns r kinda backwards
type Frame [][]byte

func (f Frame) scaleBrightness(brightness byte) Frame {

	for i, col := range f {
		for j, val := range col {
			scaled_val := int(val * brightness)
			new_val := byte(max(0, min(255, scaled_val)))
			f[i][j] = new_val
		}
	}

	return f
}

func transpose(m [][]byte) Frame {
	transposed := make([][]byte, len(m[0]))

	for i := range transposed {
		transposed[i] = make([]byte, len(m))

	}

	for i := range m {
		for j := range m[i] {
			transposed[j][i] = m[i][j]
		}
	}

	return transposed
}

type LEDMatrix struct {
	Port serial.Port
}

func (l *LEDMatrix) writeColumn(data []byte, column_id int) {
	cmd := []byte{0x32, 0xAC, CMD_STAGE_COL, byte(column_id)}

	cmd = append(cmd, data...)

	_, err := l.Port.Write(cmd)

	if err != nil {
		log.Fatal(err)
	}

}

func (l *LEDMatrix) flush() {
	cmd := []byte{0x32, 0xAC, CMD_FLUSH_COLS}
	l.Port.Write(cmd)
}

var (
	SizeError = errors.New("Invalid Frame Size")
)

func checkFrame(frame [][]byte) error {
	if !(len(frame) == HEIGHT && len(frame[0]) == WIDTH) {
		return SizeError
	}

	for i, row := range frame {
		for j, val := range row {
			if val != 0 && val < 5 {
				slog.Warn("Brightness value below threshold, raising ...")

				frame[i][j] = 5
			}
		}
	}

	return nil

}

func (l *LEDMatrix) writeFrame(frame *Frame) error {

	if err := checkFrame(*frame); err != nil {
		return err
	}

	transposed := transpose(*frame)

	for index, col := range transposed {
		l.writeColumn(col, index)
	}

	l.flush()

	return nil

}

func (l *LEDMatrix) showTest() {

	// var testFrame = transpose(testImage)
	// brightness := 255
	// testFrame.scaleBrightness(byte(brightness))

	var testImage = Frame{
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 1, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 0},
		{0, 0, 0, 0, 0, 1, 1, 0, 0},
		{0, 1, 1, 0, 1, 1, 0, 0, 0},
		{0, 0, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	testImage.scaleBrightness(5)

	l.writeFrame(&testImage)
}

func main() {

	// gracefully shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		sig := <-c

		go func() {
			time.Sleep(3 * time.Second)
			println("Somebody wouldn't gracefully stop...\nExiting anyway...\n")
			time.Sleep(1 * time.Second)
			os.Exit(1)
		}()

		terminations := 0
		// make sure everybody who wants to get this signal gets it
		for {
			select {
			case TERMINATE <- sig:
				terminations++

			default:
				// 'goto considered harmful' -Dijkstra 1968
				goto echos
			}
		}

	echos:

		for range terminations {
			<-TERMINATION_COMPLETE
		}

		os.Exit(0)

	}()

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}

	mode := &serial.Mode{
		BaudRate: 115200,
	}
	left, err := serial.Open(LEFT, mode)
	right, err := serial.Open(RIGHT, mode)

	if err != nil {
		log.Fatal(err)
	}

	matrix := &LEDMatrix{Port: left}
	matrix1 := &LEDMatrix{Port: right}

	go listen_on_socket()

	daemon := NewDaemon(matrix, matrix1)

	daemon.startDaemon()

	select {}

}
