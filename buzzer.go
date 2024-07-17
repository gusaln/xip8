package xip8

type Buzzer interface {
	Play()
	Stop()
}

type DummyBuzzer struct {
	IsPlaying bool
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
