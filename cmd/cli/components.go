package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/guslan/xip8"
)

const C_ESC byte = '\x1b'
const C_CTRL_START byte = '['
const C_CTRL_END byte = '\\'

var C_ERASE = [...]byte{
	// CSI
	C_ESC, C_CTRL_START,
	// 2 = clear all screen and move cursor to start
	'2',
	// command
	'J',
}

type Terminal struct {
	buf             []byte
	size            int
	terminal        io.Writer
	OnChar, OffChar string

	kbLayout xip8.KeyboardLayout
}

func NewTerminal() *Terminal {
	return NewTerminalWithOutput(os.Stdout)
}

func NewTerminalWithOutput(out io.Writer) *Terminal {
	return &Terminal{
		// Display
		buf:      make([]byte, 1024),
		size:     0,
		terminal: out,
		OnChar:   "##",
		OffChar:  "  ",

		// Keyboard
		kbLayout: xip8.CosmacVipInQwertyKeyboardLayout,
	}
}

func (t *Terminal) Boot() error {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	// t.clear()
	// t.flush()

	return nil
}

func (t *Terminal) Render(screen xip8.Screen, settings xip8.ScreenSettings) error {
	minSize := settings.Width*settings.Height*2 + settings.Height + 128
	if len(t.buf) < minSize {
		t.buf = make([]byte, minSize)
	}

	t.setCursor(0, 0)
	t.clear()
	// buff = append(buff, ESC, '[', '0', 'J')
	for i, b := range screen {
		for bitJ := range 8 {
			bit := b & (1 << (7 - byte(bitJ)))

			if bit > 0 {
				t.writeString(t.OnChar)
			} else {
				t.writeString(t.OffChar)
			}
		}

		if ((i+1)*8)%settings.Width == 0 {
			t.writeAll([]byte{'|', '\n'})
		}
	}

	return t.flush()
}

func (t *Terminal) setCursor(x, y byte) int {
	return t.writeAll([]byte{
		// CSI
		C_ESC, C_CTRL_START,
		// arguments
		x,
		';',
		y,
		// command
		'H',
	})
}

func (t *Terminal) clear() int {
	return t.writeAll(C_ERASE[:])
}

func (t *Terminal) carriageReturn() int {
	return t.writeSingle('\r')
}

func (t *Terminal) write(c ...byte) int {
	return t.writeAll(c)
}

func (t *Terminal) writeString(s string) int {
	return t.writeAll([]byte(s))
}

func (t *Terminal) writeAll(c []byte) int {
	for _, ci := range c {
		t.buf[t.size] = ci
		t.size += 1
	}
	return t.size
}

func (t *Terminal) writeSingle(c byte) int {
	t.buf[t.size] = c
	t.size += 1
	return t.size
}

func (t *Terminal) flush() error {
	_, err := t.terminal.Write(t.buf[:t.size])
	t.size = 0

	return err
}

// IsPressed implements Keyboard.
func (t Terminal) IsPressed(k byte) bool {
	return false
}

// GetPressed implements Keyboard.
func (t Terminal) GetPressed() (byte, bool) {
	lookupMap := xip8.LookupMap(t.kbLayout)

	// // disable input buffering
	// exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// // do not display entered characters on the screen
	// exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	// // display entered characters on the screen
	// defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	var pressed byte = 0
	buf := make([]byte, 64)
	for {
		// if err := ctx.Err(); err != nil {
		// 	break
		// }

		n, err := os.Stdin.Read(buf)
		if err != nil {
			panic(err)
		}

		if n == 1 {
			if pressedKey, mapped := lookupMap[rune(buf[0])]; mapped {
				pressed = pressedKey
				break
			}
		}
	}

	return pressed, false
}

// SetKeyMap implements Keyboard.
func (t Terminal) SetKeyMap(l xip8.KeyboardLayout) {
	t.kbLayout = l
}
