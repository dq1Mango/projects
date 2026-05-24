package main

import (
	"fmt"
	"log"
	"time"

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
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
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

var testFrame = transpose(testImage).scaleBrightness(20)

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

func send_column(
	column_id byte,
	values []float64,
	serial_port serial.Port,
	brightness_scale float64,
) {
	cmd := []byte{0x32, 0xAC, CMD_STAGE_COL, column_id}

	for _, val := range values {
		scaled_val := int(val * brightness_scale)
		val := byte(max(0, min(255, scaled_val)))
		cmd = append(cmd, val)
	}

	n, err := serial_port.Write(cmd)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %v bytes\n", n)

	send_flush(serial_port)
}

func send_flush(serial_port serial.Port) {
	cmd := []byte{0x32, 0xAC, CMD_FLUSH_COLS}
	serial_port.Write(cmd)
}

func (l *LEDMatrix) showTest() {
	fmt.Println(testFrame)
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

	data := make([]float64, HEIGHT)

	for i := range data {
		data[i] = 1
	}

	send_column(1, data, port, 10)

	time.Sleep(time.Second)

	matrix := &LEDMatrix{Port: port}

	matrix.showTest()

}
