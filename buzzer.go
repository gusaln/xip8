package xip8

type Buzzer interface {
	// Boot initializes the component
	Boot() error
	Play()
	Stop()
}

type DummyBuzzer struct {
	IsPlaying bool
}

// Boot implements Buzzer.
func (b *DummyBuzzer) Boot() error {
	return nil
}

func NewDummyBuzzer() *DummyBuzzer {
	return &DummyBuzzer{
		IsPlaying: false,
	}
}

// Play implements Buzzer.
func (b *DummyBuzzer) Play() {
	b.IsPlaying = true
}

// Stop implements Buzzer
func (b *DummyBuzzer) Stop() {
	b.IsPlaying = false
}
