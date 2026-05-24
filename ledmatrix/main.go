package main

import (
	"fmt"
	"log"

	"go.bug.st/serial"
)

const (
	LEFT  = "/dev/ttyACM0"
	RIGHT = "/dev/ttyACM1"
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

var testImage = [][]byte{
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

	n, err := l.Port.Write(cmd)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sent %v bytes\n", n)
}

func (l *LEDMatrix) flush() {
	cmd := []byte{0x32, 0xAC, CMD_FLUSH_COLS}
	l.Port.Write(cmd)
}

func (l *LEDMatrix) writeFrame(frame *Frame) {

	for index, col := range *frame {
		l.writeColumn(col, index)
	}

	l.flush()

}

func (l *LEDMatrix) showTest() {

	var testFrame = transpose(testImage)
	brightness := 255
	testFrame.scaleBrightness(byte(brightness))

	l.writeFrame(&testFrame)
}

func main() {
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
	port, err := serial.Open(LEFT, mode)

	if err != nil {
		log.Fatal(err)
	}

	matrix := &LEDMatrix{Port: port}

	matrix.showTest()

}
