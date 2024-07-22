package xip8

import (
	"errors"
	"log"

	"github.com/pkg/term"
)

type KeyboardState [16]bool

type Keyboard interface {
	// Boot initializes the component
	Boot() error
	IsPressed(k byte) bool
	// WaitForKey blocks the execution and waits for a key to be pressed
	WaitForKey() (k byte, err error)
}

type InMemoryKeyboard struct {
	State uint16
}

func NewInMemoryKeyboard() *InMemoryKeyboard {
	return &InMemoryKeyboard{
		State: 0,
	}
}

func (kb *InMemoryKeyboard) Boot() error {
	return nil
}

func (kb InMemoryKeyboard) getPressed() (byte, bool) {
	for i := byte(0); i < 8; i++ {
		if kb.IsPressed(i) {
			return i, true
		}
	}

	return 0, false
}

func (kb *InMemoryKeyboard) IsPressed(k byte) bool {
	if k > 15 {
		return false
	}
	return (kb.State & (1 >> k)) > 0
}

func (kb *InMemoryKeyboard) WaitForKey() (byte, error) {
	k, ok := kb.getPressed()
	if ok {
		return k, nil
	}

	return 0, errors.New("keyboard does not support blocking execution")
}

// func (kb InMemoryKeyboard) Get() KeyboardState {
// 	return kb.State
// }

// func (kb *InMemoryKeyboard) Press(k byte) {
// 	if k > 15 {
// 		return
// 	}

// 	kb.State[k] = true
// }

// func (kb *InMemoryKeyboard) Release(k byte) {
// 	if k > 15 {
// 		return
// 	}

// 	kb.State[k] = false
// }

type TerminalKeyboard struct {
	KeyMap map[rune]byte
}

func NewTerminalKeyboard() *TerminalKeyboard {
	return &TerminalKeyboard{
		KeyMap: map[rune]byte{},
	}
}

func (kb *TerminalKeyboard) Boot() error {
	return nil
}

func (kb *TerminalKeyboard) IsPressed(k byte) bool {
	return false
}

func (kb *TerminalKeyboard) WaitForKey() (byte, error) {
	for {
		t, _ := term.Open("/dev/tty")

		err := term.RawMode(t)
		if err != nil {
			log.Fatal(err)
		}

		var read int
		buf := make([]byte, 3)
		read, err = t.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		t.Restore()
		t.Close()

		if read == 1 && buf[0] < 16 {
			return buf[0], nil
		}
	}
}
