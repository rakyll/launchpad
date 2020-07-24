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
	defer pad.Close()

	pad.Clear()

	// Send Text-Loop
	pad.Text(3, 0).
		Add(7, "Hello ").
		Add(1, "World!").
		Perform()

	ch := pad.Listen()
	for {
		hit := <-ch

		if hit.IsScrollTextEndMarker() {
			log.Printf("Scrolling text is ended now.")
		}
	}
}
