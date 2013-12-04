# launchpad
A package allows you to talk to your Novation Launchpad in Go. Light buttons
or read your touches.

~~~ sh
go get github.com/rakyll/launchpad
~~~

## Usage
Initialize a new Launchpad. If there are no currently connected Launchpad
device, initialization will fail with an error. You can fake a device by
creating an input and output MIDI device and name them as Launchpad.
~~~ go
if pad, err = launchpad.New(); err != nil {
    log.Error("error while initializing launchpad")
}

// turn off all of the lights, before we begin
pad.Reset()
~~~

### Light buttons

~~~ go
pad.Light(0, 0, 3, 0) // lights the bottom left button with bright green
~~~

### Read/listen touches

~~~ go
hits, err := pad.Read() // reads at most 64 hits
for _, hit := range hits {
    log.Printf("touch at (%d, %d)", hit.X, hit.Y)
}

// or alternatively you can listen hits
ch := pad.Listen()
hit := <-ch
~~~

### Cleanup
Cleanup your input and output streams once you're done. Likely to be called
on graceful termination.
~~~ go
pad.Cleanup()
~~~

## Demo: Light your touchs

![A demo](https://googledrive.com/host/0ByfSjdPVs9MZbkhjeUhMYzRTeEE/demo.gif)

A simple program to light every touch:

~~~ go
pad, _ := launchpad.New()
pad.Reset()

ch := pad.Listen()
for {
	select {
	case hit := <-ch:
		pad.Light(hit.X, hit.Y, 3, 3)
	}
}
~~~
    
## License
    Copyright 2013 Google Inc. All Rights Reserved.
    
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
    
         http://www.apache.org/licenses/LICENSE-2.0
    
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
