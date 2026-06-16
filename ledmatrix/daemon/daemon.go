package main

import (
	"log/slog"
	"time"

	"github.com/dq1Mango/projects/ledmatrix/ipc"
)

const RefreshDelay = 10 * time.Second

var ActionChan chan ipc.Action = make(chan ipc.Action)
var StopDaemon = make(chan bool)

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

	go func() {
		<-TERMINATE
		println("trying to stop mode")
		daemon.Stop <- true
		println("mode stopped")
		StopDaemon <- true
		println("daemon stopped")
		empty := *EmptyFrame()
		left.writeFrame(&empty)
		right.writeFrame(&empty)

		time.Sleep(100 * time.Millisecond)

		TERMINATION_COMPLETE <- ""

	}()

	return daemon

}

func (s *Daemon) WriteFrame() error {

	if err := s.Left.writeFrame(&s.CurrentFrame); err != nil {
		return err
	}
	if err := s.Right.writeFrame(&s.CurrentFrame); err != nil {
		return err
	}

	return nil
}

func (s *Daemon) showBattery() {
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

func (s *Daemon) Nothing() {
	println("clearing screen")
	frame := *EmptyFrame()
	s.Frames <- &frame
	<-s.Stop
}

type Module interface {
	Start(chan *Frame)
}

func (s *Daemon) SetMode(mode ipc.Mode) {
	// select {
	// case s.Stop <- "":
	// default:
	// }

	// this is kind of scary... i wanted to do a non blocking send,
	// but u cant guarentee the running mode is gunna be listening,
	// and if it was 1-buffered, there would need to be a response,
	// currently nothing can fail, but i suspect that might change...
	s.Stop <- ""

	var err error

	switch mode {
	case ipc.Nothing:
		go s.Nothing()

	case ipc.Battery:
		go s.showBattery()

	case ipc.Stars:
		go s.Stars()

	case ipc.Sand:
		go s.Criticalilty()

	case ipc.Forest:
		go s.ForestFire()

	case ipc.Fourier:
		fourier := &Fourier{}
		err = fourier.Start(s.Frames, s.Stop)

	default:
		slog.Warn("Should not happen")

	}

	if err != nil {
		go s.Nothing()
	}
}

func (s *Daemon) startDaemon() {

	go s.Nothing()

	refreshTimer := time.NewTimer(RefreshDelay)

	for {
		select {

		case action := <-ActionChan:
			println("got action which needs to be renamed")

			switch a := action.(type) {
			case *ipc.SetMode:
				go s.SetMode(a.Mode)

			default:

			}

		case f := <-s.Frames:
			s.CurrentFrame = *f
			s.WriteFrame()

		case <-refreshTimer.C:
			s.WriteFrame()

			// gracefull shutdown by clearing the 'screen'
		case <-StopDaemon:
			println("I stopped")
			return
		}

		refreshTimer = time.NewTimer(RefreshDelay)

	}

}
