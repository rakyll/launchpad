package launchpad

import "github.com/rakyll/portmidi"

// ScrollingTextBuilder is used to build and display an scrolling text on the Launchpad.
type ScrollingTextBuilder struct {
	seq          []byte
	outputStream *portmidi.Stream
}

// Text will return a scrolling text builder whether you can build and
// perform an text with the given color which will be scrolled on the launchpad.
func (l *Launchpad) Text(g int, r int) *ScrollingTextBuilder {
	return l.text(g, r, false)
}

// TextLoop will return a scrolling text builder whether you can build and
// perform an text with the given color which will be scrolled endless on the launchpad.
// If you want to stop an text loop you have to build and execute an empty textLoop!
func (l *Launchpad) TextLoop(g int, r int) *ScrollingTextBuilder {
	return l.text(g, r, true)
}

func (l *Launchpad) text(g int, r int, loop bool) *ScrollingTextBuilder {
	color := byte(16*g + r + 8 + 4)
	if loop {
		color += 64
	}

	return &ScrollingTextBuilder{
		seq:          []byte{0xF0, 0x00, 0x20, 0x29, 0x09, color},
		outputStream: l.outputStream,
	}
}

// Add adds a text snipped with a given speed to the builder.
// The speed can be a value from 1-7. The text must be ASCII
// characters! Otherwise the result could be weired.
func (s *ScrollingTextBuilder) Add(speed byte, text string) *ScrollingTextBuilder {
	if speed > 7 {
		speed = 7
	} else if speed < 1 {
		speed = 1
	}

	s.seq = append(s.seq, speed)
	s.seq = append(s.seq, []byte(text)...)

	return s
}

// Perform sends the pre-built scrolling text to the launchpad.
func (s *ScrollingTextBuilder) Perform() error {
	s.seq = append(s.seq, 0xF7)

	// the syntax of the scrolling text message:
	// F0 00 20 29 09 <colour> <text inclusive speed ...> F7
	return s.outputStream.WriteSysExBytes(portmidi.Time(), s.seq)
}
