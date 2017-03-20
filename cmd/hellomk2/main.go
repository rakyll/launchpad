package main

import (
	"log"
	"time"

	"github.com/rakyll/launchpad/mk2"
)

func main() {
	pad, err := mk2.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchpad: %v", err)
	}
	defer pad.Close()

	var color int
	render := func(i, j int) {
		pad.Reset()
		// Turn all buttons to bright red.
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				pad.Light(i, j, color)
				time.Sleep(20 * time.Millisecond)
				color = (color + 1) % 128
			}
		}
	}

	render(0, 0)
	ch := pad.Listen()
	for {
		hit := <-ch
		log.Printf("Button pressed at <x=%d, y=%d>", hit.X, hit.Y)
		// Re-render the color palette again.
		render(hit.X, hit.Y)
	}

}
