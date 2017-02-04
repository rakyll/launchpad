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

// Launchpad represents a device with an input and output MIDI stream.
type Launchpad struct {
	*midi.Device
}

// Button represents a button on the Launchpad.
type Button [2]int

// Color represents the color of a launcpad button.
type Color struct {
	Green uint8
	Red   uint8
}

// Hit represents physical touches to Launchpad buttons.
type Hit struct {
	X uint8
	Y uint8
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

// Hits returns a channel that emits when the launchpad buttons are hit.
func (l *Launchpad) Hits() (<-chan Hit, error) {
	packets, err := l.Packets()
	if err != nil {
		return nil, err
	}
	hits := make(chan Hit)
	go relayPackets(packets, hits)
	return hits, nil
}

// Light lights the button at x,y with the given greend and red values.
// x and y are [0, 8], g and r are [0, 3]
// Note that x=8 corresponds to the round scene buttons on the right side of the device,
// and y=8 corresponds to the round automap buttons on the top of the device.
func (l *Launchpad) Light(x, y uint8, color Color) error {
	var (
		note     = x + 16*y
		velocity = 16*color.Green + color.Red + 8 + 4
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

// Reset turns off all the lights on the launchpad.
func (l *Launchpad) Reset() error {
	_, err := l.Write([]byte{0xb0, 0, 0})
	return err
}

// relayPackets turns packets into hits.
func relayPackets(packets <-chan midi.Packet, hits chan<- Hit) {
	for packet := range packets {
		if packet[2] == 0 {
			continue
		}
		var x, y uint8

		if packet[0] == 176 {
			// top row button
			x = packet[1] - 104
			y = 8
		} else if packet[0] == 144 {
			x = packet[1] % 16
			y = (packet[1] - x) / 16
		} else {
			continue
		}
		hits <- Hit{X: x, Y: y}
	}
}
