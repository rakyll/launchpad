// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package mk3 provides interfaces to talk to Novation Launchpad MK3 via MIDI in and out.
// Tested with a Launchpad Mini MK3
package mk3

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rakyll/portmidi"
)

// Launchpad represents a device with an input and output MIDI stream.
type Launchpad struct {
	inputStream  *portmidi.Stream
	outputStream *portmidi.Stream
}

// Hit represents physical touches to Launchpad buttons.
type Hit struct {
	X int
	Y int
}

// Open opens a connection Launchpad and initializes an input and output
// stream to the currently connected device. If there are no
// devices are connected, it returns an error.
func Open() (*Launchpad, error) {
	input, output, err := discover()
	if err != nil {
		return nil, err
	}

	var inStream, outStream *portmidi.Stream
	if inStream, err = portmidi.NewInputStream(input, 1024); err != nil {
		return nil, err
	}
	if outStream, err = portmidi.NewOutputStream(output, 1024, 0); err != nil {
		return nil, err
	}
	// Switch to the session mode.
	outStream.WriteSysExBytes(portmidi.Time(), []byte{0xf0, 0x00, 0x20, 0x29, 0x02, 0x18, 0x22, 0x00, 0xf7})
	return &Launchpad{inputStream: inStream, outputStream: outStream}, nil
}

// Listen listens the input stream for hits.
func (l *Launchpad) Listen() <-chan Hit {
	ch := make(chan Hit)
	go func(pad *Launchpad, ch chan Hit) {
		for {

			// sleep for a while before the new polling tick,
			// otherwise operation is too intensive and blocking
			time.Sleep(10 * time.Millisecond)
			hits, err := pad.Read()
			if err != nil {
				continue
			}
			for i := range hits {
				ch <- hits[i]
			}
		}
	}(l, ch)
	return ch
}

// Read reads hits from the input stream. It returns max 64 hits for each read.
func (l *Launchpad) Read() (hits []Hit, err error) {
	var evts []portmidi.Event
	if evts, err = l.inputStream.Read(1024); err != nil {
		return
	}
	for _, evt := range evts {
		if evt.Data2 > 0 {
			var x, y int64

			x = evt.Data1%10 - 1
			y = (8 - (evt.Data1-x)/10)

			hits = append(hits, Hit{X: int(x), Y: int(y)})
		}
	}
	return
}

// Light lights the button at x,y with the given red, green, and blue values.
// All available colors are documented and visualized at Launchpad's Programmers Guide.
func (l *Launchpad) Light(x, y, red int, green int, blue int) error {
	// TODO(jbd): Support top row.
	led := int64((8-y)*10 + x + 1)
	//return l.outputStream.WriteShort(0x90, led, int64(color))
	return l.outputStream.WriteSysExBytes(portmidi.Time(), []byte{0xF0, 0x00, 0x20, 0x29, 0x02, 0x0D, 0x03, 0x03, byte(led), byte(red), byte(green), byte(blue), 0xF7})
}

func (l *Launchpad) FlashLight(x int, y int, onColor int, offColor int) error {
	// TODO(jbd): Support top row.
	led := int64((8-y)*10 + x + 1)
	return l.outputStream.WriteSysExBytes(portmidi.Time(), []byte{0xF0, 0x00, 0x20, 0x29, 0x02, 0x0D, 0x03, 0x01, byte(led), byte(onColor), byte(offColor), 0xF7})
}

// Reset turns off all buttons.
func (l *Launchpad) Reset() error {
	// Sends a "light all ligts" SysEx command with 0 color.
	return l.outputStream.WriteSysExBytes(portmidi.Time(), []byte{0xf0, 0x00, 0x20, 0x29, 0x02, 0x18, 0x0e, 0x00, 0xf7})
}

// Set Programmers Mode.
func (l *Launchpad) Program() error {
	// Set programmer mode using a SysEx command.
	return l.outputStream.WriteSysExBytes(portmidi.Time(), []byte{0xf0, 0x00, 0x20, 0x29, 0x02, 0x0d, 0x0e, 0x01, 0xf7})
}

func (l *Launchpad) Close() error {
	l.inputStream.Close()
	l.outputStream.Close()
	return nil
}

// discovers the currently connected Launchpad device
// as a MIDI device.
func discover() (input portmidi.DeviceID, output portmidi.DeviceID, err error) {
	in := -1
	out := -1
	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		fmt.Printf("%s\n", info.Name)
		if strings.Contains(info.Name, "Launchpad Mini MK3") {
			if info.IsInputAvailable {
				in = i
			}
			if info.IsOutputAvailable {
				out = i
			}
		}
	}
	if in == -1 || out == -1 {
		err = errors.New("launchpad: no launchpad is connected")
	} else {
		input = portmidi.DeviceID(in)
		output = portmidi.DeviceID(out)
	}
	return
}
