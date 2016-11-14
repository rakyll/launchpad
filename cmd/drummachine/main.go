package main

import (
	"flag"
	"log"
	"time"

	"github.com/rakyll/launchpad"
	"github.com/rakyll/portmidi"
)

const (
	NumberofHGrids = 8
	NumberofVGrids = 8
)

var (
	pad         *launchpad.Launchpad
	out         *portmidi.Stream
	grid        [][]bool
	instruments []int
)

var flagOutput = flag.Int("o", 0, "The output MIDI device ID")

func main() {
	flag.Parse()

	var err error
	if err = portmidi.Initialize(); err != nil {
		log.Fatal("error while initializing portmidi", err)
	}

	if out, err = portmidi.NewOutputStream(portmidi.DeviceID(*flagOutput), 1024, 0); err != nil {
		log.Fatal("error while initializing connection to midi out stream")
	}

	if pad, err = launchpad.Open(); err != nil {
		log.Fatal("error while initializing connection to launchpad", err)
	}

	grid = newGrid()
	instruments = []int{
		70, // maracas
		38, // snare 1
		40, // snare 2
		51, // ride cymbal
		42, // closed hi-hat
		58, // vibra slap
		46, // open hi-hat
		81} // open triangle

	// load an initial drum pattern
	// hi-hats
	grid[0][4] = true
	grid[2][4] = true
	grid[4][4] = true
	grid[6][4] = true
	// snares
	grid[5][3] = true
	grid[5][2] = true
	grid[5][1] = true
	// bells
	grid[6][7] = true
	grid[7][7] = true
	grid[5][6] = true

	// clear
	pad.Reset()

	// listen button toggles
	ch := pad.Listen()
	go func() {
		for {
			hit := <-ch
			log.Println("drum toggled at", hit)
			if hit.Y == -1 || hit.X > 7 {
				// a controller button is hit
				continue
			}
			grid[hit.X][hit.Y] = !grid[hit.X][hit.Y]
			if !grid[hit.X][hit.Y] {
				// turn off immediately
				pad.Light(hit.X, hit.Y, 0, 0)
			}
		}
	}()

	for {
		for x := 0; x < NumberofHGrids; x++ {
			tick(x)
		}
	}
}

func tick(x int) {
	for y := 0; y < NumberofVGrids; y++ {
		pad.Light((x-1+NumberofHGrids)%NumberofHGrids, y, 0, 0)
		pad.Light(x, y, 2, 2)
	}
	drawAndPlay(x)
	time.Sleep(time.Millisecond * 250)
}

func drawAndPlay(x int) {
	for x1 := 0; x1 < NumberofHGrids; x1++ {
		for y := 0; y < NumberofVGrids; y++ {
			if grid[x1][y] {
				if x == x1 {
					//out.WriteShort(int64(0x90+9), int64(instruments[y]), 100)
				}
				pad.Light(x1, y, 3, 0)
			}
		}
	}
}

func newGrid() [][]bool {
	grid := make([][]bool, NumberofHGrids)
	for i := range grid {
		grid[i] = make([]bool, NumberofVGrids)
	}
	return grid
}
