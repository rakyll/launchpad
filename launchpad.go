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

package launchpad

import (
	"github.com/rakyll/portmidi"
)

type Launchpad struct {
	output *portmidi.Stream
}

func New(deviceId portmidi.DeviceId) (pad *Launchpad, err error) {
	var stream *portmidi.Stream
	if stream, err = portmidi.NewOutputStream(deviceId, 1024, 0); err != nil {
		return nil, err
	}
	return &Launchpad{output: stream}, nil
}

func (l *Launchpad) Light(x, y, g, r int) error {
	note := int64(x + 16*(7-y))
	velocity := int64(16*g + r + 8 + 4)
	return l.output.WriteShort(0x90, note, velocity)
}

func (l *Launchpad) Reset() error {
	return l.output.WriteShort(0xb0, 0, 0)
}

func (l *Launchpad) Cleanup() error {
	return l.output.Close()
}
