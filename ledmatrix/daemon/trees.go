package main

import (
	"math"
	"math/rand/v2"
	"time"
)

const (
	GROWTH_CHANCE = 0.01
	FIRE_CHANCE   = 1e-4
)

type Tree int

const (
	Empty Tree = iota
	Grown
	Burning
)

var Tree_Brightness = map[Tree]byte{
	Empty:   0,
	Grown:   20,
	Burning: 255,
}

type Forest [][]Tree

func NewForest() *Forest {

	lattice := make(Forest, HEIGHT)

	for i := range lattice {
		lattice[i] = make([]Tree, WIDTH)
	}

	return &lattice
}

func (f *Forest) SpreadFire() bool {

	forest := *f

	fire := false

	for i, row := range *f {
		for j, tree := range row {
			if tree == Burning {
				forest[i][j] = Empty
				fire = true

				if i+1 < HEIGHT && forest[i+1][j] == Grown {
					forest[i+1][j] = Burning
				}
				if i-1 >= 0 && forest[i-1][j] == Grown {
					forest[i-1][j] = Burning
				}
				if j+1 < WIDTH && forest[i][j+1] == Grown {
					forest[i][j+1] = Burning
				}
				if j-1 >= 0 && forest[i][j-1] == Grown {
					forest[i][j-1] = Burning
				}
			}
		}
	}

	return fire
}

func (f *Forest) Tick() {
	forest := *f

	fire := f.SpreadFire()
	_ = fire
	//
	// if fire {
	// 	return
	// }

	for i, row := range forest {
		for j, tree := range row {
			if tree == Empty {
				if rand.Float64() < GROWTH_CHANCE {
					forest[i][j] = Grown
				}
			} else if tree == Grown {
				if rand.Float64() < FIRE_CHANCE {
					println("ts lit bruh")
					forest[i][j] = Burning
				}
			}
		}
	}
}

func (f *Forest) ProduceFrame() *Frame {
	frame := *EmptyFrame()

	for i, row := range *f {
		for j, tree := range row {
			frame[i][j] = Tree_Brightness[tree]
		}
	}

	return &frame

}

func (d *Daemon) ForestFire() {

	const fps = 10.0

	forest := NewForest()

	frame := forest.ProduceFrame()
	d.Frames <- frame

	var delay int = int(math.Round(1.0 / fps))
	duration := time.Duration(delay * int(time.Second))
	// var delay = int(d)
	refresh := time.NewTimer(duration)

	for {
		select {
		case <-refresh.C:
			forest.Tick()
			frame := forest.ProduceFrame()
			d.Frames <- frame

			refresh = time.NewTimer(duration)

		case <-d.Stop:
			return
		}
	}
}
