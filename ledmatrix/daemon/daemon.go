package main

import (
	"time"

	"github.com/dq1Mango/projects/ledmatrix/ipc"
)

const RefreshDelay = 10 * time.Second

var ActionChan chan ipc.Action = make(chan ipc.Action)

type Daemon struct {
	Left, Right *LEDMatrix

	Mode         ipc.Mode
	CurrentFrame Frame
	Stop         chan any
	Frames       chan *Frame
}

func NewDaemon(left, right *LEDMatrix) *Daemon {
	daemon := &Daemon{
		Mode:         ipc.Nothing,
		CurrentFrame: *EmptyFrame(),
		Stop:         make(chan any),
		Frames:       make(chan *Frame),

		Left: left, Right: right,
	}

	// gracefull shutdown by clearing the 'screen'
	go func() {
		<-TERMINATE
		empty := *EmptyFrame()
		left.writeFrame(&empty)
		right.writeFrame(&empty)
		TERMINATION_COMPLETE <- ""
	}()

	return daemon

}

func (s *Daemon) WriteFrame() error {
	println("writing new frames")

	if err := s.Left.writeFrame(&s.CurrentFrame); err != nil {
		return err
	}
	if err := s.Right.writeFrame(&s.CurrentFrame); err != nil {
		return err
	}

	return nil
}

func (s Daemon) showBattery() {
	refreshTimer := time.NewTimer(RefreshDelay)

	percentage := getBatteryPercentage()
	frame := makeBatteryFrame(percentage)
	s.Frames <- &frame

	for {
		select {
		case <-refreshTimer.C:
			percentage := getBatteryPercentage()
			frame := makeBatteryFrame(percentage)
			s.Frames <- &frame

		case <-s.Stop:
			return
		}
	}

}

func (s *Daemon) SetMode(mode ipc.Mode) {
	select {
	case s.Stop <- "":
	default:
	}

	switch mode {
	case ipc.Nothing:
		println("clearing screen")
		frame := *EmptyFrame()
		s.Frames <- &frame

	case ipc.Battery:
		go s.showBattery()

	case ipc.Stars:
		go s.Stars()

	}
}

func (s *Daemon) startDaemon() {

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

		case f := <-s.Frames:
			println("new frame")
			s.CurrentFrame = *f
			s.WriteFrame()

		case <-refreshTimer.C:
			s.WriteFrame()
		}

		refreshTimer = time.NewTimer(RefreshDelay)

	}

}
