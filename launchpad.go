// Package launchpad enables MIDI communication with a Novation Launchpad.
package launchpad

import "github.com/scgolang/midi"

// Launchpad represents a device with an input and output MIDI stream.
type Launchpad struct {
	*midi.Device
}

// Button represents a button on the Launchpad.
type Button struct {
	X int
	Y int
}

// Open opens a connection Launchpad and initializes an input and output
// stream to the currently connected device.
// The deviceID is a system-specific string.
// On linux try
//     amidi -l
func Open(deviceID string) (*Launchpad, error) {
	d, err := midi.Open(deviceID)
	if err != nil {
		return nil, err
	}
	return &Launchpad{Device: d}, nil
}

// Light lights the button at x,y with the given greend and red values.
// x and y are [0, 8], g and r are [0, 3]
// Note that x=8 corresponds to the round scene buttons on the right side of the device,
// and y=8 corresponds to the round automap buttons on the top of the device.
func (l *Launchpad) Light(x, y, g, r uint8) error {
	note := uint8(x + 16*y)
	velocity := uint8(16*g + r + 8 + 4)
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

// Reset turns off all the lights on the launchpad.
func (l *Launchpad) Reset() error {
	_, err := l.Write([]byte{0xb0, 0, 0})
	return err
}
