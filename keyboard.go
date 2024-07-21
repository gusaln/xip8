package xip8

type KeyboardState [16]bool

type Keyboard interface {
	// Boot initializes the component
	Boot() error
	IsPressed(k byte) bool
}

type DummyKeyboard struct {
	State KeyboardState
}

// Boot implements Keyboard.
func (kb *DummyKeyboard) Boot() error {
	return nil
}

func NewDummyKeyboard() *DummyKeyboard {
	return &DummyKeyboard{
		State: [16]bool{},
	}
}

func (kb *DummyKeyboard) IsPressed(k byte) bool {
	if k > 15 {
		return false
	}
	return kb.State[k]
}

func (kb DummyKeyboard) Get() KeyboardState {
	return kb.State
}

func (kb *DummyKeyboard) Press(k byte) {
	if k > 15 {
		return
	}

	kb.State[k] = true
}

func (kb *DummyKeyboard) Release(k byte) {
	if k > 15 {
		return
	}

	kb.State[k] = false
}
