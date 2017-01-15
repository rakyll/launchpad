// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package launchpad provides interfaces to talk to
// Novation Launchpads via MIDI in and out.
package launchpad

import (
	"errors"
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
	if err := portmidi.Initialize(); err != nil {
		return nil, err
	}

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
			if evt.Status == 176 {
				// top row button
				x = evt.Data1 - 104
				y = 8
			} else {
				x = evt.Data1 % 16
				y = (evt.Data1 - x) / 16
			}
			hits = append(hits, Hit{X: int(x), Y: int(y)})
		}
	}
	return
}

// Light lights the button at x,y with the given greend and red values.
// x and y are [0, 8], g and r are [0, 3]
// Note that x=8 corresponds to the round scene buttons on the right side of the device,
// and y=8 corresponds to the round automap buttons on the top of the device.
func (l *Launchpad) Light(x, y, g, r int) error {
	note := int64(x + 16*y)
	velocity := int64(16*g + r + 8 + 4)
	if y >= 8 {
		return l.lightAutomap(x, velocity)
	}
	return l.outputStream.WriteShort(0x90, note, velocity)
}

// lightAutomap lights the top row of buttons.
func (l *Launchpad) lightAutomap(x int, velocity int64) error {
	return l.outputStream.WriteShort(176, int64(x+104), velocity)
}

func (l *Launchpad) Reset() error {
	return l.outputStream.WriteShort(0xb0, 0, 0)
}

func (l *Launchpad) Close() error {
	portmidi.Terminate()
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
		if strings.Contains(info.Name, "Launchpad") {
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
