# launchpad [![GoDoc](https://godoc.org/github.com/scgolang/launchpad?status.svg)](https://godoc.org/github.com/scgolang/launchpad)
A package allows you to talk to your Novation Launchpad in Go. Light buttons
or read your touches.

~~~ sh
go get github.com/scgolang/launchpad
~~~

## Usage
Initialize a new Launchpad. If there are no currently connected Launchpad
device, initialization will fail with an error. You can fake a device by
creating an input and output MIDI device and name them as Launchpad.
~~~ go
if pad, err = launchpad.Open(); err != nil {
    log.Error("error while initializing launchpad")
}

// turn off all of the lights
pad.Reset()
~~~

### Light buttons

~~~ go
pad.Light(0, 0, 3, 0) // lights the bottom left button with bright green
~~~

The coordinate system is illustrated below.
~~~
+--------- arrow keys -----------+  +--- mode keys ---+
{0, 8} {1, 8} {2, 8} {3, 8} {4, 8} {5, 8} {6, 8} {7, 8} | ableton
----------------------------------------------------------------
{0, 0} {1, 0} {2, 0} {3, 0} {4, 0} {5, 0} {6, 0} {7, 0} | {8, 0} vol
----------------------------------------------------------------
{0, 1} {1, 1} {2, 1} {3, 1} {4, 1} {5, 1} {6, 1} {7, 1} | {8, 1} pan
----------------------------------------------------------------
{0, 2} {1, 2} {2, 2} {3, 2} {4, 2} {5, 2} {6, 2} {7, 2} | {8, 2} sndA
----------------------------------------------------------------
{0, 3} {1, 3} {2, 3} {3, 3} {4, 3} {5, 3} {6, 3} {7, 3} | {8, 3} sndB
----------------------------------------------------------------
{0, 4} {1, 4} {2, 4} {3, 4} {4, 4} {5, 4} {6, 4} {7, 4} | {8, 4} stop
----------------------------------------------------------------
{0, 5} {1, 5} {2, 5} {3, 5} {4, 5} {5, 5} {6, 5} {7, 5} | {8, 5} trk on
----------------------------------------------------------------
{0, 6} {1, 6} {2, 6} {3, 6} {4, 6} {5, 6} {6, 6} {7, 6} | {8, 6} solo
----------------------------------------------------------------
{0, 7} {1, 7} {2, 7} {3, 7} {4, 7} {5, 7} {6, 7} {7, 7} | {8, 7} arm
----------------------------------------------------------------
~~~

### Read/listen for touches

~~~ go
hits, err := pad.Read() // reads at most 64 hits
for _, hit := range hits {
    log.Printf("touch at (%d, %d)", hit.X, hit.Y)
}

// or alternatively you can listen for hits
ch := pad.Listen()
hit := <-ch
~~~

### Close
Close your input and output streams once you're done. Likely to be called
on graceful termination.
~~~ go
pad.Close()
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
