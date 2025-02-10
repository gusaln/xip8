package xip8

import (
	"log"

	"github.com/pkg/term"
)

// KeyboardLayout represents the a keyboard layout
//
// Note that the codes are not in ascending sequence, but rather in the layout
// of the original COSMAC VIP keypad.
// Keycodes should be given in the following order
//
// # 1 2 3 C
//
// # 4 5 6 D
//
// # 7 8 9 E
//
// # A 0 B F
type KeyboardLayout = [16]rune
type KeyboardLookupMap = map[rune]byte

// CosmacVipKeyboardLayout
//
// # 1 2 3 C
//
// # 4 5 6 D
//
// # 7 8 9 E
//
// # A 0 B F
var CosmacVipKeyboardLayout = KeyboardLayout{
	'1', '2', '3', 'C',
	'4', '5', '6', 'D',
	'7', '8', '9', 'E',
	'A', '0', 'B', 'F',
}

// CosmacVipInQwertyKeyboardLayout
//
// # 1 2 3 4
//
// # Q W E R
//
// # A S D F
//
// # Z X C V
var CosmacVipInQwertyKeyboardLayout = KeyboardLayout{
	'1', '2', '3', '4',
	'Q', 'W', 'E', 'R',
	'A', 'S', 'D', 'F',
	'Z', 'X', 'C', 'V',
}

// SimpleKeyboardLayout is an alias for CosmacVipKeyboardLayout
// Simple as in "every key correspond to its keypad equivalent"
var SimpleKeyboardLayout = CosmacVipKeyboardLayout

// DefaultKeyboardLayout is CosmacVipInQwertyKeyboardLayout
var DefaultKeyboardLayout = CosmacVipInQwertyKeyboardLayout

func LookupMap(layout KeyboardLayout) KeyboardLookupMap {
	m := map[rune]byte{
		layout[0]:  0x1,
		layout[1]:  0x2,
		layout[2]:  0x3,
		layout[3]:  0xc,
		layout[4]:  0x4,
		layout[5]:  0x5,
		layout[6]:  0x6,
		layout[7]:  0xd,
		layout[8]:  0x7,
		layout[9]:  0x8,
		layout[10]: 0x9,
		layout[11]: 0xe,
		layout[12]: 0xa,
		layout[13]: 0x0,
		layout[14]: 0xb,
		layout[15]: 0xf,
	}

	return m
}

type KeyboardState [16]bool

// Keyboard interface
type Keyboard interface {
	// Boot initializes the component
	Boot() error
	IsPressed(k byte) bool
	// GetPressed blocks the execution and waits for a key to be pressed
	GetPressed() (k byte, pressed bool)
}

type InMemoryKeyboard struct {
	State uint16
}

func NewInMemoryKeyboard() *InMemoryKeyboard {
	return &InMemoryKeyboard{
		State: 0,
	}
}

// Boot implements Keyboard.
func (kb *InMemoryKeyboard) Boot() error {
	return nil
}

// GetPressed implements Keyboard.
func (kb InMemoryKeyboard) GetPressed() (byte, bool) {
	for i := byte(0); i < 16; i++ {
		if kb.IsPressed(i) {
			return i, true
		}
	}

	return 0, false
}

const (
	KeyMask = 1 << 15
)

// IsPressed implements Keyboard.
func (kb *InMemoryKeyboard) IsPressed(k byte) bool {
	if k > 15 {
		return false
	}
	return (kb.State & (KeyMask >> k)) > 0
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
