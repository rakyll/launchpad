package main

import (
	"log"

	"github.com/rakyll/launchpad/mk2"
)

func main() {
	pad, err := mk2.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchpad: %v", err)
	}
	defer pad.Close()

	// Turn all buttons to bright red.
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			pad.Light(i, j, 72)
		}
	}

	ch := pad.Listen()
	for {
		hit := <-ch
		log.Printf("Button pressed at <x=%d, y=%d>", hit.X, hit.Y)
		// Turn to green.
	}
}
