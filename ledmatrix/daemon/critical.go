package main

import (
	"math"
	"time"
)

const CRITICAL = 4

// sand per tick
const SPT = 20

var Brightnesses = map[Sand]byte{
	0: 0,
	1: 10,
	2: 30,
	3: 60,
	4: 255,
}

type Sand int

type Lattice [][]Sand

type SandPile struct {
	current Lattice
	next    Lattice
}

func NewSandPile() *SandPile {

	return &SandPile{
		current: *NewLattice(),
		next:    *NewLattice(),
	}
}

func NewLattice() *Lattice {

	lattice := make(Lattice, HEIGHT)

	for i := range lattice {
		lattice[i] = make([]Sand, WIDTH)
	}

	return &lattice
}

func (s *SandPile) PropogateAvalanches() bool {

	avalanche := false

	for i, row := range s.current {
		for j, sand := range row {
			if sand >= 4 {
				s.next[i][j] = 0
				avalanche = true

				if i+1 < HEIGHT {
					s.next[i+1][j]++
				}
				if i-1 >= 0 {
					s.next[i-1][j]++
				}
				if j+1 < WIDTH {
					s.next[i][j+1]++
				}
				if j-1 >= 0 {
					s.next[i][j-1]++
				}
			} else {
				s.next[i][j] += sand

			}
		}
	}

	s.current = s.next
	s.next = *NewLattice()

	return avalanche

}

func (s *SandPile) Tick() {
	avalanches := s.PropogateAvalanches()

	if avalanches {
		return
	}
	s.current[HEIGHT/4][WIDTH/2]++
	s.current[HEIGHT*3/4][WIDTH/2]++

	// s.current[HEIGHT/6+1][WIDTH/2]++
	// s.current[HEIGHT/2][WIDTH/2]++
	// s.current[HEIGHT*5/6][WIDTH/2]++

}

func (s *SandPile) ProduceFrame() *Frame {
	frame := *EmptyFrame()

	for i, row := range s.current {
		for j, sand := range row {
			frame[i][j] = Brightnesses[sand]
		}
	}

	return &frame

}

func (d *Daemon) Criticalilty() {

	const fps = 10

	sandpile := NewSandPile()

	frame := sandpile.ProduceFrame()
	d.Frames <- frame

	var delay int = int(math.Round(1.0 / fps))
	duration := time.Duration(delay * int(time.Second))
	// var delay = int(d)
	refresh := time.NewTimer(duration)

	for {
		select {
		case <-refresh.C:
			sandpile.Tick()
			frame := sandpile.ProduceFrame()
			d.Frames <- frame

			refresh = time.NewTimer(duration)

		case <-d.Stop:
			return
		}
	}
}
