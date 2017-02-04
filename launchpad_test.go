package launchpad_test

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/scgolang/launchpad"
)

var lp *launchpad.Launchpad

func TestMain(m *testing.M) {
	var deviceID string
	flag.StringVar(&deviceID, "device", "hw:0,0,0", "System-specific MIDI device ID")
	flag.Parse()

	var err error
	lp, err = launchpad.Open(deviceID)
	if err != nil {
		fmt.Printf("error initializing launchpad: %s\n", err)
		os.Exit(1)
	}
	code := m.Run()

	_ = lp.Reset()
	_ = lp.Close()

	os.Exit(code)
}

// TestLight flashes the launchpad for a short time.
func TestLight(t *testing.T) {
	lp.Reset()

	x, y := uint8(0), uint8(8)

	lp.Light(x, y, launchpad.Color{
		Green: launchpad.Full,
		Red:   launchpad.Off,
	})
	time.Sleep(500 * time.Millisecond)

	lp.Light(x, y, launchpad.Color{
		Green: launchpad.Off,
		Red:   launchpad.Full,
	})
	time.Sleep(500 * time.Millisecond)

	lp.Light(x, y, launchpad.Color{
		Green: launchpad.Full,
		Red:   launchpad.Full,
	})
	time.Sleep(500 * time.Millisecond)

	lp.Reset()
}
