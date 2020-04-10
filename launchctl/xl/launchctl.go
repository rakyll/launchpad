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

// Package xl provides interfaces to talk to Novation Launch Control XL via MIDI in and out.
package xl

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rakyll/portmidi"
)

const (
	MIDI_Status_Note_Off       = 0x80
	MIDI_Status_Note_On        = 0x90
	MIDI_Status_Control_Change = 0xb0
	MIDI_Status_Code_Mask      = 0xf0
	MIDI_Channel_Mask          = 0x0f

	MaxEventsPerPoll = 1024
	ReadBufferDepth  = 16
	PollingPeriod    = 10 * time.Millisecond
	NumChannels      = 16
	NumControls      = 6*8 + 4 + 4
)

type Value uint8

// LaunchControl represents a device with an input and output MIDI stream.
type LaunchControl struct {
	inputStream  *portmidi.Stream
	outputStream *portmidi.Stream

	lock sync.Mutex

	value [NumChannels][NumControls]Value
}

// Open opens a connection to the XL and initializes an input and
// output stream to the currently connected device. If there are no
// devices connected, it returns an error.
func Open() (*LaunchControl, error) {
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
	lc := &LaunchControl{inputStream: inStream, outputStream: outStream}
	for ch := 0; ch < NumChannels; ch++ {
		for cc := 0; cc < NumControls; cc++ {
			lc.value[ch][cc] = 128
		}
	}
	return lc, nil
}

// Start begins listening for updates.
func (l *LaunchControl) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()
	ch := make(chan []portmidi.Event, ReadBufferDepth)

	go func() {
		l.Buffer(0, 0)
		l.Reset(0, 0+1+1<<4)
		l.Buffer(0, 1)
		l.Reset(0, 0+0+0<<4)
		for {
			// @@@
			time.Sleep(time.Second / 2)
			l.Flash(0, true)
			time.Sleep(time.Second / 2)
			l.Flash(0, false)
		}

	}()
	go func() {
		for {
			// return when canceled
			select {
			case <-ctx.Done():
				return
			default:
			}
			// TODO: Is there a portmidi or libusb function that lets us poll?
			time.Sleep(PollingPeriod)

			evts, err := l.inputStream.Read(MaxEventsPerPoll)
			if err != nil {
				fmt.Println("MIDI error", err)
				cancel()
				return
			}
			if len(evts) != 0 {
				ch <- evts
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evts := <-ch:
				for _, evt := range evts {
					l.event(evt)
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
	}
}

func (l *LaunchControl) event(evt portmidi.Event) {
	l.lock.Lock()
	defer l.lock.Unlock()

	midiChannel := Value(evt.Status & MIDI_Channel_Mask)

	switch evt.Status & MIDI_Status_Code_Mask {
	case MIDI_Status_Control_Change:
		//fmt.Println("CC Chan", midiChannel, evt.Data1, evt.Data2)
		l.controlChange(midiChannel, Value(evt.Data1), Value(evt.Data2))
	case MIDI_Status_Note_On, MIDI_Status_Note_Off:
		//fmt.Println("CC Note", midiChannel, evt.Data1, evt.Data2)
		l.noteChange(midiChannel, Value(evt.Data1), Value(evt.Data2))
	}

}

func (l *LaunchControl) controlChange(midiChannel, data1, data2 Value) {
	switch {
	case 13 <= data1 && data1 <= 20: // 0ffset 0
		l.value[midiChannel][data1-13+0] = data2

	case 29 <= data1 && data1 <= 36: // Offset 8
		l.value[midiChannel][data1-29+8] = data2

	case 49 <= data1 && data1 <= 56: // Offset 16
		l.value[midiChannel][data1-49+16] = data2

	case 77 <= data1 && data1 <= 84: // Offset 24
		l.value[midiChannel][data1-77+24] = data2

	case 104 <= data1 && data1 <= 107: // Offset 48
		l.value[midiChannel][data1-104+48] = data2
	}
}

func (l *LaunchControl) noteChange(midiChannel, data1, data2 Value) {
	switch {
	case 41 <= data1 && data1 <= 44: // 0ffset 32
		l.value[midiChannel][data1-41+32] = data2

	case 57 <= data1 && data1 <= 60: // Offset 36
		l.value[midiChannel][data1-57+36] = data2

	case 73 <= data1 && data1 <= 76: // Offset 40
		l.value[midiChannel][data1-73+40] = data2

	case 89 <= data1 && data1 <= 92: // Offset 44
		l.value[midiChannel][data1-89+44] = data2

	case 105 <= data1 && data1 <= 108: // Offset 52
		l.value[midiChannel][data1-105+52] = data2
	}
}

// Reset sends a "light all lights" SysEx command with color value.
func (l *LaunchControl) Reset(tmpl, color int) error {

	data := []byte{0xf0, 0x00, 0x20, 0x29, 0x02, 0x11, 0x78, byte(tmpl)}

	for i := 0; i < 48; i++ {
		data = append(data, byte(i), byte(color))
	}

	data = append(data, 0xf7)

	return l.outputStream.WriteSysExBytes(portmidi.Time(), data)
}

func (l *LaunchControl) Buffer(tmpl, b int) error {
	var data int64
	if b == 0 {
		data = 0x21 + 0x8
	} else {
		data = 0x24 + 0x8
	}
	return l.outputStream.WriteShort(0xb0+int64(tmpl), 0, data)
}

func (l *LaunchControl) Flash(tmpl int, on bool) error {
	var data int64
	if on {
		data = 0x28 // @@@
	} else {
		data = 0x20 // @@@
	}
	return l.outputStream.WriteShort(0xb0+int64(tmpl), 0, data)
}

func (l *LaunchControl) Close() error {
	l.inputStream.Close()
	l.outputStream.Close()
	return nil
}

// discovers the currently connected LaunchControl device
// as a MIDI device.
func discover() (input portmidi.DeviceID, output portmidi.DeviceID, err error) {
	in := -1
	out := -1
	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.Info(portmidi.DeviceID(i))
		if info.Name == "Launch Control XL" {
			if info.IsInputAvailable {
				in = i
			}
			if info.IsOutputAvailable {
				out = i
			}
		}
	}
	if in == -1 || out == -1 {
		err = errors.New("launchctl: no launch control xl is connected")
	} else {
		input = portmidi.DeviceID(in)
		output = portmidi.DeviceID(out)
	}
	return
}
