package launchpad

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scgolang/syncosc"
)

const (
	gridX    = 8
	gridY    = 8
	gridSize = gridX * gridY
)

// Colors used to light the pads.
// We use a different color for showing sequencer position versus
// showing which steps are turned on in a pattern.
var (
	posColor  = Color{Green: Medium}
	stepColor = Color{Red: Full}
)

// Mode is a
type Mode int

const (
	ModePattern Mode = iota
	ModeMutes
)

// Trig is a sequencer trigger.
// It provides the track that is being triggered as well
// as the value of the sequencer for that track.
type Trig struct {
	Track uint8
	Value uint8
}

// Trigger is a thing that can be triggered by a Sequencer.
// Use the Track method to get notified of track selection.
type Trigger interface {
	Track(track uint8) error
	Trig(step uint8, trigs []Trig) error
}

// Sequencer is a simple sequencer controlled by a Novation Launchpad.
type Sequencer struct {
	modeChan      chan Mode
	mutes         [gridSize]bool
	pad           *Launchpad
	prevStep      uint8
	step          uint8
	stepSkip      int
	syncConnector syncosc.ConnectorFunc
	syncHost      string
	tick          chan syncosc.Pulse
	track         uint8
	tracks        [gridSize][gridSize]uint8 // Track => Step => Value
	triggers      []Trigger
}

// NewSequencer creates a new sequencer.
func (l *Launchpad) NewSequencer(syncConnector syncosc.ConnectorFunc, syncHost string) *Sequencer {
	return &Sequencer{
		modeChan:      make(chan Mode, 1),
		pad:           l,
		syncConnector: syncConnector,
		syncHost:      syncHost,
		tick:          make(chan syncosc.Pulse),
	}
}

// AddTrigger adds the Trigger to the sequencer.
func (seq *Sequencer) AddTrigger(t Trigger) {
	seq.triggers = append(seq.triggers, t)
}

// advance advances the internal counter of the sequencer.
// It returns the true if the sequencer's internal counter has actually advanced
// and false otherwise.
func (seq *Sequencer) advance(step int32) bool {
	if seq.stepSkip <= 1 {
		// step increments the sequencer's counter directly.
		seq.prevStep = seq.step
		seq.step = uint8(step % gridSize)
		return true
	}
	// This step has no effect.
	if step%int32(seq.stepSkip) != 0 {
		return false
	}
	// step does not increment the sequencer's counter directly
	// we increment it based on the previous value
	seq.prevStep = seq.step
	seq.step = uint8((seq.step + 1) % gridSize)
	return true
}

// advanceLights advances the lights on the launchpad according to the
// internal counter of the sequencer.
func (seq *Sequencer) advanceLights() error {
	var (
		prevValue = seq.tracks[seq.track][seq.prevStep]
		prevHit   = stepToHit(seq.prevStep)
		hit       = stepToHit(seq.step)
	)
	if seq.prevStep == seq.step {
		// We just started the sequencer and the first pulse
		// is the beginning of the sequence.
		return seq.pad.Light(0, 0, posColor)
	}
	if err := seq.pad.Light(hit.X, hit.Y, posColor); err != nil {
		return err
	}
	if prevValue == 0 {
		if err := seq.pad.Light(prevHit.X, prevHit.Y, Color{}); err != nil {
			return err
		}
	} else {
		if err := seq.pad.Light(prevHit.X, prevHit.Y, stepColor); err != nil {
			return err
		}
	}
	return nil
}

// flashStepTriggers flashes the tracks that are triggered for the current step.
func (seq *Sequencer) flashStepTriggers() error {
	hits := []Hit{}
	for track, steps := range seq.tracks {
		if seq.mutes[track] {
			continue
		}
		if val := steps[seq.step]; val > 0 {
			hit := stepToHit(uint8(track))
			hits = append(hits, hit)
			if err := seq.pad.Light(hit.X, hit.Y, posColor); err != nil {
				return err
			}
		}
	}
	time.Sleep(40 * time.Millisecond)
	for _, hit := range hits {
		if err := seq.pad.Light(hit.X, hit.Y, Color{}); err != nil {
			return err
		}
	}
	return nil
}

// invokeTriggers invokes the sequencer's triggers for the provided step.
func (seq *Sequencer) invokeTriggers() error {
	trigs := []Trig{}

	for track, steps := range seq.tracks {
		if val := steps[seq.step]; val > 0 {
			trigs = append(trigs, Trig{
				Track: uint8(track),
				Value: val,
			})
		}
	}
	for _, trigger := range seq.triggers {
		if err := trigger.Trig(uint8(seq.step), trigs); err != nil {
			return err
		}
	}
	return nil
}

// invokeTriggersTrack invokes the Track method of all the Trigger's
// that have been added to the sequencer.
func (seq *Sequencer) invokeTriggersTrack() error {
	// Invoke all the trigs.
	for _, trig := range seq.triggers {
		if err := trig.Track(seq.track); err != nil {
			return err
		}
	}
	return nil
}

// lightCurrentTrack lights the track buttons based on the currently selected track.
func (seq *Sequencer) lightCurrentTrack() error {
	var (
		curX = seq.track % gridX
		curY = seq.track / gridY
	)
	if err := seq.pad.Light(curX, gridY, stepColor); err != nil {
		return err
	}
	return seq.pad.Light(gridX, curY, stepColor)
}

// lightMutes lights the mutes.
func (seq *Sequencer) lightMutes() error {
	for track, isMuted := range seq.mutes {
		if isMuted {
			hit := stepToHit(uint8(track))
			if err := seq.pad.Light(hit.X, hit.Y, stepColor); err != nil {
				return err
			}
		}
	}
	return nil
}

// lightTrackSteps lights all the steps of the current track.
func (seq *Sequencer) lightTrackSteps() error {
	for step, val := range seq.tracks[seq.track] {
		hit := stepToHit(uint8(step))

		if val > 0 {
			if err := seq.pad.Light(hit.X, hit.Y, stepColor); err != nil {
				return err
			}
			continue
		}
		if err := seq.pad.Light(hit.X, hit.Y, Color{}); err != nil {
			return err
		}
	}
	return nil
}

// loopMutes is an infinite loop that the sequencer uses when it is in "Mutes" mode.
// If the mode is changed then it will be returned with a nil error.
// The only other time this func returns is when there is an error.
func (seq *Sequencer) loopMutes(ctx context.Context, hits <-chan Hit) (Mode, error) {
	if err := seq.lightMutes(); err != nil {
		return 0, err
	}
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case hit := <-hits:
			if hit.Err != nil {
				return 0, hit.Err
			}
			if hit.X == gridX || hit.Y == gridY {
				if err := seq.setCurrentTrackFrom(hit); err != nil {
					return 0, err
				}
				if err := seq.pad.Reset(); err != nil {
					return 0, err
				}
				if err := seq.lightCurrentTrack(); err != nil {
					return 0, err
				}
				if err := seq.lightMutes(); err != nil {
					return 0, err
				}
				continue
			}
			if err := seq.toggleMuteFrom(hit); err != nil {
				return 0, err
			}
		case mode := <-seq.modeChan:
			return mode, nil
		case pulse := <-seq.tick:
			if advanced := seq.advance(pulse.Count); !advanced {
				continue
			}
			// Flash all the triggered tracks.
			if err := seq.flashStepTriggers(); err != nil {
				return 0, err
			}
			if err := seq.invokeTriggers(); err != nil {
				return 0, err
			}
		}
	}
}

// loopPattern is an infinite loop that the sequencer uses when it is in "Pattern" mode.
// If the mode is changed then it will be returned with a nil error.
// The only other time this func returns is when there is an error.
func (seq *Sequencer) loopPattern(ctx context.Context, hits <-chan Hit) (Mode, error) {
	if err := seq.lightTrackSteps(); err != nil {
		return 0, err
	}
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case hit := <-hits:
			if hit.Err != nil {
				return 0, hit.Err
			}
			if hit.X == gridX || hit.Y == gridY {
				if err := seq.setCurrentTrackFrom(hit); err != nil {
					return 0, err
				}
				if err := seq.selectPatternTrackFrom(hit); err != nil {
					return 0, err
				}
				continue
			}
			if err := seq.toggle(hit); err != nil {
				return 0, err
			}
		case mode := <-seq.modeChan:
			return mode, nil
		case pulse := <-seq.tick:
			if advanced := seq.advance(pulse.Count); !advanced {
				continue
			}
			if err := seq.advanceLights(); err != nil {
				return 0, err
			}
			if err := seq.invokeTriggers(); err != nil {
				return 0, err
			}
		}
	}
}

// Main is the main loop of the sequencer.
// It loops forever on input from the launchpad.
// If ctx is cancelled it returns the ctx.Err().
func (seq *Sequencer) Main(ctx context.Context) error {
	hits, err := seq.pad.Hits()
	if err != nil {
		return err
	}
	// This func could block forever.
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		if err := seq.syncConnector(ctx, seq, seq.syncHost); err != nil {
			cancel()
			fmt.Fprintf(os.Stderr, "connecting to sync source: %s", err.Error())
		}
	}()
	if err := seq.lightCurrentTrack(); err != nil {
		return err
	}
	loop := seq.loopPattern

	for {
		if err := seq.pad.Reset(); err != nil {
			return err
		}
		if err := seq.lightCurrentTrack(); err != nil {
			return err
		}
		mode, err := loop(ctx, hits)
		if err != nil {
			return err
		}
		switch mode {
		case ModePattern:
			loop = seq.loopPattern
		case ModeMutes:
			loop = seq.loopMutes
		default:
			return errors.Errorf("unrecognized mode: %d", mode)
		}
	}
}

// Pulse receives pulses from oscsync.
func (seq *Sequencer) Pulse(pulse syncosc.Pulse) error {
	seq.tick <- pulse
	return nil
}

// ReadFrom reads the sequencer's state from an io.Reader.
// TODO
func (seq *Sequencer) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}

// selectPatternTrackFrom selects a track from the provided hit.
// This func is used in pattern mode.
func (seq *Sequencer) selectPatternTrackFrom(hit Hit) error {
	// Reset the launchpad.
	if err := seq.pad.Reset(); err != nil {
		return errors.Wrap(err, "resetting launchpad")
	}
	// Light the current track.
	if err := seq.lightCurrentTrack(); err != nil {
		return err
	}
	// Light all the steps of the current track.
	return seq.lightTrackSteps()
}

func (seq *Sequencer) setCurrentTrackFrom(hit Hit) error {
	if hit.Y == gridY {
		// We hit the top row.
		curX := seq.track % gridX

		if curX == hit.X {
			return nil // Nothing to do.
		}
		// Set the current track.
		seq.track = hit.X + seq.track - curX
	} else if hit.X == gridX {
		// Hit the column on the right side of the device.
		curY := seq.track / gridY

		if curY == hit.Y {
			return nil // Nothing to do.
		}
		// Set the current track.
		seq.track = (hit.Y * gridY) + (seq.track % gridX)
	} else {
		return errors.New("hit is not for track selection")
	}
	return nil
}

// SetMode sets the display mode of the sequencer.
func (seq *Sequencer) SetMode(mode Mode) error {
	seq.modeChan <- mode
	return nil
}

// SetResolution sets the clock resolution for the sequencer.
// This is set as a human-readable note resolution, e.g. 16th or 32nd.
func (seq *Sequencer) SetResolution(resolution string) error {
	res, ok := resolutionMap[resolution]
	if !ok {
		return errors.Errorf("unrecognized resolution: %s", resolution)
	}
	seq.stepSkip = res
	return nil
}

// toggle toggles the button that has been hit.
func (seq *Sequencer) toggle(hit Hit) error {
	var (
		step = hitToStep(hit)
		val  = seq.tracks[seq.track][step]
	)
	if val == 0 {
		seq.tracks[seq.track][step] = 1
		return seq.pad.Light(hit.X, hit.Y, stepColor)
	}
	seq.tracks[seq.track][step] = 0
	return seq.pad.Light(hit.X, hit.Y, Color{})
}

// toggleMuteFrom toggles the state of a mute from a hit on the launchpad.
func (seq *Sequencer) toggleMuteFrom(hit Hit) error {
	track := hitToStep(hit)
	if seq.mutes[track] {
		seq.mutes[track] = false
		return seq.pad.Light(hit.X, hit.Y, Color{})
	}
	seq.mutes[track] = true
	return seq.pad.Light(hit.X, hit.Y, stepColor)
}

// WriteTo writes the current sequencer data to w.
// TODO
func (seq *Sequencer) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func hitToStep(hit Hit) uint8 {
	return (8 * hit.Y) + hit.X
}

func stepToHit(step uint8) Hit {
	return Hit{
		X: step % gridX,
		Y: step / 8,
	}
}

var resolutionMap = map[string]int{
	"16th": 6,
	"32nd": 3,
	"96th": 1,
}

type loopFunc func(ctx context.Context, hits <-chan Hit) error
