// Package launchpad enables MIDI communication with a Novation Launchpad.
package launchpad

import "github.com/scgolang/midi"

// Brightness values.
const (
	Off uint8 = iota
	Low
	Medium
	Full
)

// MessageSize is the number of bytes in a launchpad MIDI message.
const MessageSize = 3

// Launchpad represents a device with an input and output MIDI stream.
type Launchpad struct {
	*midi.Device
}

// Button represents a button on the Launchpad.
type Button [2]int

type Color struct {
	Green uint8
	Red   uint8
}

// Open opens a connection Launchpad and initializes an input and output
// stream to the currently connected device.
// The deviceID is a system-specific string.
//
// On linux try
//     amidi -l
//
// On mac try using https://github.com/briansorahan/coremidi
//     coremidi -l
//
func Open(deviceID string) (*Launchpad, error) {
	l := &Launchpad{
		Device: &midi.Device{
			Name: deviceID,
		},
	}
	if err := l.Open(); err != nil {
		return nil, err
	}
	return l, nil
}

// Close closes the connection to the launchpad.
func (l *Launchpad) Close() error {
	return l.Device.Close()
}

// Light lights the button at x,y with the given greend and red values.
// x and y are [0, 8], g and r are [0, 3]
// Note that x=8 corresponds to the round scene buttons on the right side of the device,
// and y=8 corresponds to the round automap buttons on the top of the device.
func (l *Launchpad) Light(x, y uint8, color Color) error {
	var (
		note     = uint8(x + 16*y)
		velocity = uint8(16*color.Green + color.Red + 8 + 4)
	)
	if y >= 8 {
		return l.lightAutomap(x, velocity)
	}
	_, err := l.Write([]byte{0x90, note, velocity})
	return err
}

// lightAutomap lights the top row of buttons.
func (l *Launchpad) lightAutomap(x uint8, velocity uint8) error {
	_, err := l.Write([]byte{176, x + 104, velocity})
	return err
}

// ReadButton reads a certain number of button preses from a launchpad.
func (l *Launchpad) ReadButton(n int) ([]Button, error) {
	buf := make([]byte, n*MessageSize)
	if _, err := l.Read(buf); err != nil {
		return nil, err
	}
	buttons := []Button{}
	for i := 0; i < n; i++ {
	}
	return buttons, nil
}

// Reset turns off all the lights on the launchpad.
func (l *Launchpad) Reset() error {
	_, err := l.Write([]byte{0xb0, 0, 0})
	return err
}
