package main

import (
	"context"
	"log"

	launchctl "github.com/jmacd/launchmidi/launchctl/xl"
)

func main() {
	ctl, err := launchctl.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchctl: %v", err)
	}
	defer ctl.Close()

	ctx := context.Background()

	ctl.Start(ctx)

	// 	for {
	// 		if err := ctl.Buffer(0, 0); err != nil {
	// 			log.Panic("Reset failed: ", err)
	// 		}

	// 		time.Sleep(10 * time.Millisecond)

	// 		if err := ctl.Buffer(0, 1); err != nil {
	// 			log.Panic("Reset failed: ", err)
	// 		}
	// 		time.Sleep(10 * time.Millisecond)

	// 	}

	select {}
}
