package main

import (
	"fmt"
	"time"
)

const FPS = 1
const BRIGHTNESS = 100

type Star struct {
	MaxIntensity     int
	CurrentIntensity int
	TTL              int
	Change           int
}

func NewStar(max_i, ttl int) *Star {
	return &Star{
		MaxIntensity:     max_i,
		CurrentIntensity: 3,
		TTL:              ttl,
		Change:           1,
	}
}

type Sky [][]*Star

func NewSky() *Sky {
	sky := make(Sky, HEIGHT)

	for i := range sky {
		sky[i] = make([]*Star, WIDTH)
	}

	sky[HEIGHT/2][WIDTH/2] = NewStar(3, 100)

	return &sky
}

func (s *Sky) Tick() {

	for i, row := range *s {
		for j := range row {
			if star := row[j]; star != nil {
				star.TTL -= 1
				if star.TTL == 0 {
					(*s)[i][j] = nil

				}

				if star.CurrentIntensity+star.Change > star.MaxIntensity {
					star.Change *= -1
				} else if star.CurrentIntensity+star.Change < 0 {
					star.Change *= -1
				}

				star.CurrentIntensity += star.Change
			}
		}
	}
}

func (s *Sky) ProduceFrame() *Frame {
	frame := *EmptyFrame()

	for i, row := range *s {
		for j := range row {
			if star := row[j]; star != nil {
				for n := range star.CurrentIntensity {
					if i+n < HEIGHT {
						frame[i+n][j] = BRIGHTNESS
					}
					if i-n > 0 {
						frame[i-n][j] = BRIGHTNESS
					}

					if j+n < WIDTH {
						frame[i][j+n] = BRIGHTNESS
					}
					if j-n > 0 {
						frame[i][j-n] = BRIGHTNESS
					}
				}
			}
		}
	}

	return &frame

}

// type Sky struct {
// 	[][]
// }

func (d *Daemon) Stars() {
	sky := NewSky()
	frame := sky.ProduceFrame()
	d.Frames <- frame

	refresh := time.NewTimer(FPS * time.Second)

	for {
		select {
		case <-refresh.C:
			sky.Tick()
			frame := sky.ProduceFrame()
			d.Frames <- frame

			refresh = time.NewTimer(FPS * time.Second)

		case <-d.Stop:
			return
		}
	}
}
