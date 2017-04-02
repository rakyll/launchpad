package launchpad_test

import (
	"context"
	"testing"
	"time"

	"github.com/scgolang/launchpad"
	"github.com/scgolang/syncosc"
)

func TestSequencer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		if err := seq.Main(ctx); err != nil && err != context.DeadlineExceeded {
			t.Fatal(err)
		}
		close(done)
	}()
	time.Sleep(20 * time.Second)
	seq.SetMode(launchpad.ModeMutes)

	time.Sleep(20 * time.Second)
	seq.SetMode(launchpad.ModePattern)

	time.Sleep(20 * time.Second)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

// mockSyncConnector starts a ticker that pulses at 120 bpm.
func mockSyncConnector(ctx context.Context, slave syncosc.Slave, host string) error {
	go func() {
		const tempo = float32(120)

		for count := int32(0); true; count++ {
			select {
			// Multiply by 6 for sixteenth notes.
			case <-time.NewTicker(syncosc.GetPulseDuration(tempo) * 6).C:
				_ = slave.Pulse(syncosc.Pulse{
					Count: count,
					Tempo: tempo,
				})
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
