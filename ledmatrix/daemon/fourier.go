package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/argusdusty/gofft"
)

type Fourier struct {
}

func PickMonitor() string {
	return "alsa_output.usb-Framework_Audio_Expansion_Card-00.analog-stereo.monitor"
}

func parseS16LEStereo(buf []byte) (left []complex128, right []complex128) {
	// each frame = 4 bytes (left sample + right sample)
	left = make([]complex128, len(buf)/2)
	right = make([]complex128, len(buf)/2)

	for i := 0; i+3 < len(buf); i += 4 {
		l := int16(binary.LittleEndian.Uint16(buf[i:]))
		r := int16(binary.LittleEndian.Uint16(buf[i+2:]))

		// normalize to [-1.0, 1.0] if you need floating point
		leftF := float64(l) / 32768.0
		rightF := float64(r) / 32768.0

		left[i/4] = complex(leftF, 0)
		right[i/4+1] = complex(rightF, 0)
	}

	return left, right
}

func makeFFTUsefull(fft []complex128) []float64 {

	output := make([]float64, len(fft))

	for i, val := range fft {
		output[i] = math.Sqrt(real(val)*real(val) + imag(val)*imag(val))
	}

	return output

}

// of the form y = a * e ^ (-x * b)
func ExpDecay(a, b, x float64) float64 {
	return a * math.Exp(-x*b)
}

func BrightnessGradient(start, stop int, pct float64) byte {
	b := -math.Log(float64(stop) / float64(start))

	return byte(int(ExpDecay(float64(start), b, pct)))
}

func FourierFrame(bins []float64) *Frame {
	frame := *EmptyFrame()
	// var brightness byte = 120

	for col, bin := range bins {
		height := int((HEIGHT/2-1)*bin) + 1

		for h := range height {
			brightness := BrightnessGradient(100, 5, float64(h)/(HEIGHT/2))
			// frame[HEIGHT/2-h][col] = brightness - brightness/17*byte(h)
			// frame[HEIGHT/2+h][col] = brightness - brightness/17*byte(h)
			frame[HEIGHT/2-h][col] = brightness
			frame[HEIGHT/2+h][col] = brightness
		}
	}

	return &frame
}

func PickXBins(fourier []float64, x int) []float64 {
	quotient := len(fourier) / x
	output := make([]float64, x)

	for i := range x {
		output[i] = fourier[i*quotient]
	}

	return output
}

func NormalizeFloat64(arr []float64) {
	most := slices.Max(arr)

	for i := range arr {
		arr[i] /= most
	}
}

func (f *Fourier) Start(frames chan<- *Frame, stop chan any) error {
	monitor := PickMonitor()
	sample_rate := 44100

	cmd := exec.Command("parec",
		"--device="+monitor, // or e.g. "alsa_output.pci-0000_00_1f.3.analog-stereo.monitor"
		"--format=s16le",
		fmt.Sprintf("--rate=%d", sample_rate),
		"--channels=2")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// ensure cleanup on exit
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	const fps float64 = 1000
	const FFT_SIZE = 4096
	const channels = 2

	// skips := sample_rate/FFT_SIZE*channels/fps - 1

	refresh := time.NewTicker(time.Second / time.Duration(fps))
	defer refresh.Stop()

	buf := make([]byte, FFT_SIZE*channels)
	for {

		select {
		case <-stop:
			fmt.Println("fourier is shutting down...")
			return nil
		default:
		}

		n, err := stdout.Read(buf)

		// println("read da buffa")
		// if err == io.EOF {
		// 	println("bork")
		// 	break
		// }
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "read %d bytes of audio\n", n)

		// select {
		// case <-refresh.C:
		// 	println("real tick")
		//
		// default:
		// 	println("no tick")
		// 	continue
		//
		// }

		complicated, _ := parseS16LEStereo(buf)

		err = gofft.FFT(complicated)
		if err != nil {
			panic(err)
		}

		usefull := makeFFTUsefull(complicated[:len(complicated)/2])
		// usefull = usefull[:len(usefull)/2]

		selected := PickXBins(usefull, WIDTH)
		NormalizeFloat64(selected)

		// fmt.Println(selected)

		frame := FourierFrame(selected)

		frames <- frame

		// buf[:n] is raw interleaved s16le PCM samples
		// do something with buf[:n]...
	}
}
