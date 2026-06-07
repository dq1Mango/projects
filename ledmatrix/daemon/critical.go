package main

import (
	"math"
	"math/rand/v2"
	"time"
)

const CRITICAL = 4

// sand per tick
const SPT = 20

var Brightnesses = map[Sand]byte{
	0: 0,
	1: 10,
	2: 50,
	3: 200,
}

type Sand int

type Lattice [][]Sand

func NewLattice() *Lattice {

	lattice := make(Lattice, HEIGHT)

	for i := range lattice {
		lattice[i] = make([]Sand, WIDTH)
	}

	return &lattice
}

func (la *Lattice) Tick() {
	l := *la
	for range 1 { //SPT {
		i, j := rand.IntN(HEIGHT), rand.IntN(WIDTH)
		i, j = 5, 5
		l[i][j] += 1

		if l[i][j] >= CRITICAL {
			l[i][j] = 0
			// 	for l[i][j] > 0 {
			//
			// 	}
			if i+1 < HEIGHT {
				l[i+1][j]++
			}
			if i-1 >= 0 {
				l[i-1][j]++
			}
			if j+1 < WIDTH {
				l[i][j+1]++
			}
			if j-1 >= 0 {
				l[i][j-1]++
			}
		}
	}
}

	for range 1 { //SPT {
		i, j := rand.IntN(HEIGHT), rand.IntN(WIDTH)
		i, j = 5, 5
		l[i][j] += 1

		if l[i][j] >= CRITICAL {
			l[i][j] = 0
			// 	for l[i][j] > 0 {
			//
			// 	}
			if i+1 < HEIGHT {
				l[i+1][j]++
			}
			if i-1 >= 0 {
				l[i-1][j]++
			}
			if j+1 < WIDTH {
				l[i][j+1]++
			}
			if j-1 >= 0 {
				l[i][j-1]++
			}
		}
	}
}

func (l *Lattice) ProduceFrame() *Frame {
	frame := *EmptyFrame()

	for i, row := range *l {
		for j, sand := range row {
			frame[i][j] = Brightnesses[sand]
		}
	}

	return &frame

}

func (d *Daemon) Criticalilty() {

	const fps = 2

	lattice := NewLattice()

	frame := lattice.ProduceFrame()
	d.Frames <- frame

	var delay int = int(math.Round(1.0 / fps))
	duration := time.Duration(delay * int(time.Second))
	// var delay = int(d)
	refresh := time.NewTimer(duration)

	for {
		select {
		case <-refresh.C:
			lattice.Tick()
			frame := lattice.ProduceFrame()
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
