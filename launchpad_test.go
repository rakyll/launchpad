package launchpad_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/scgolang/launchpad"
)

var lp *launchpad.Launchpad

func TestMain(m *testing.M) {
	var err error
	lp, err = launchpad.Open()
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
	_ = lp.Reset()

	hits, err := lp.Hits()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Please hit the bottom scene launch button")

	hit := <-hits

	if expected, got := uint8(8), hit.X; expected != got {
		fmt.Printf("expected x=%d, got x=%d\n", expected, got)
		t.Fail()
	}
	if expected, got := uint8(7), hit.Y; expected != got {
		fmt.Printf("expected y=%d, got y=%d\n", expected, got)
		t.Fail()
	}
	if !t.Failed() {
		fmt.Println("Great!")
	}

	_ = lp.Reset()
}

// TestLight flashes the launchpad for a short time.
func TestLight(t *testing.T) {
	_ = lp.Reset()

	x, y := uint8(0), uint8(8)

	if err := lp.Light(x, y, launchpad.Color{
		Green: launchpad.Full,
		Red:   launchpad.Off,
	}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	if err := lp.Light(x, y, launchpad.Color{
		Green: launchpad.Off,
		Red:   launchpad.Full,
	}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	if err := lp.Light(x, y, launchpad.Color{
		Green: launchpad.Full,
		Red:   launchpad.Full,
	}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	_ = lp.Reset()
}
