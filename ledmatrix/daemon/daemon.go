package main

import (
	"time"

	"github.com/dq1Mango/projects/ledmatrix/ipc"
)

const RefreshDelay = 10 * time.Second

var ActionChan chan ipc.Action = make(chan ipc.Action)

type State struct {
	Left, Right *LEDMatrix

	Mode      ipc.Mode
	Frame     Frame
	Stop      chan any
	FrameChan chan *Frame
}

func NewState(left, right *LEDMatrix) *State {
	return &State{
		Mode:      ipc.Nothing,
		Frame:     *EmptyFrame(),
		Stop:      make(chan any),
		FrameChan: make(chan *Frame),

		Left: left, Right: right,
	}

}

func (s *State) WriteFrame() error {
	println("writing new frames")

	if err := s.Left.writeFrame(&s.Frame); err != nil {
		return err
	}
	if err := s.Right.writeFrame(&s.Frame); err != nil {
		return err
	}

	return nil
}

func (s State) showBattery() {
	refreshTimer := time.NewTimer(RefreshDelay)

	percentage := getBatteryPercentage()
	frame := makeBatteryFrame(percentage)
	s.FrameChan <- &frame

	for {
		select {
		case <-refreshTimer.C:
			percentage := getBatteryPercentage()
			frame := makeBatteryFrame(percentage)
			s.FrameChan <- &frame

		case <-s.Stop:
			return
		}
	}

}

func (s *State) SetMode(mode ipc.Mode) {
	select {
	case s.Stop <- "":
	default:
	}

	switch mode {
	case ipc.Nothing:
		println("clearing screen")
		frame := *EmptyFrame()
		s.FrameChan <- &frame

	case ipc.Battery:
		go s.showBattery()

	}
}

func (s *State) startDaemon() {

	refreshTimer := time.NewTimer(RefreshDelay)

	for {
		println("waiting for something to do")
		select {

		case action := <-ActionChan:
			println("got action which needs to be renamed")
			switch a := action.(type) {
			case *ipc.SetMode:
				go s.SetMode(a.Mode)

			default:

			}

		case f := <-s.FrameChan:
			println("new frame")
			s.Frame = *f
			s.WriteFrame()

		case <-refreshTimer.C:
			s.WriteFrame()
		}

		refreshTimer = time.NewTimer(RefreshDelay)

	}

}
