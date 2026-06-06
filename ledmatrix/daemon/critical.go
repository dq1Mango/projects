package main

import (
	"math"
	"time"
)

const CRITICAL = 4

type Sand int

type Lattice [][]Sand

func NewLattice() *Lattice {

}

func (d *Daemon) Criticalilty() {

	frame := sky.ProduceFrame()
	d.Frames <- frame

	var delay int = int(math.Round(1.0 / FPS))
	duration := time.Duration(delay * int(time.Second))
	// var delay = int(d)
	refresh := time.NewTimer(duration)

	for {
		select {
		case <-refresh.C:
			sky.Tick()
			frame := sky.ProduceFrame()
			d.Frames <- frame

			refresh = time.NewTimer(duration)

		case <-d.Stop:
			println(
				"stopped",
			)
			return
		}
	}
}
