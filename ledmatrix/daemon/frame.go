package main

func EmptyFrame() *Frame {
	frame := make(Frame, HEIGHT)

	for i := range frame {
		frame[i] = make([]byte, WIDTH)
	}

	return &frame
}
