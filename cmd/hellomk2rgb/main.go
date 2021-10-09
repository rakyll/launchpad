package main

import (
	"log"
	"time"

	"github.com/IanLewis/launchpad/mk2"
	// "github.com/rakyll/launchpad/mk2"
)

func main() {
	pad, err := mk2.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchpad: %v", err)
	}
	defer pad.Close()

	// render cycles through the pure red, green, and blue palettes.
	var palette int
	render := func(i, j int) {
		pad.Clear()
		// Turn on all the buttons in sequence.
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				c := mk2.Color{}
				switch palette {
				case 0:
					c.R = j + (i * 8)
				case 1:
					c.G = j + (i * 8)
				case 2:
					c.B = j + (i * 8)
				}
				pad.LightRGB(i, j, c)
				time.Sleep(20 * time.Millisecond)
			}
		}
		palette = (palette + 1) % 3
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
