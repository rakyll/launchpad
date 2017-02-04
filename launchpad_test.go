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

func TestHit(t *testing.T) {
	lp.Reset()

	hits, err := lp.Hits()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Please hit the bottom scene launch button")

	hit := <-hits

	if expected, got := 8, hit.X; expected != got {
		fmt.Printf("expected x=%d, got x=%d\n", expected, got)
		t.Fail()
	}
	if expected, got := 7, hit.Y; expected != got {
		fmt.Printf("expected y=%d, got y=%d\n", expected, got)
		t.Fail()
	}
	if !t.Failed() {
		fmt.Println("Great!")
	}

	lp.Reset()
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
