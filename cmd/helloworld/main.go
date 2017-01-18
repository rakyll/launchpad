package main

import (
	"log"

	"github.com/rakyll/launchpad"
)

func main() {
	pad, err := launchpad.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchpad: %v", err)
	}
	pad.Reset()

	// Set <0,0> to yellow.
	pad.Light(0, 0, 2, 2)

	ch := pad.Listen()
	for {
		hit := <-ch
		log.Printf("Button pressed at <x=%d, y=%d>", hit.X, hit.Y)
		// Turn to green.
		pad.Light(hit.X, hit.Y, 3, 0)
	}
}
