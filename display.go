package xip8

import (
	"io"
	"os"
)

// Display abstraction for a display
type Display interface {
	// Boot initializes the component
	Boot() error
	// Render
	Render(Screen, ScreenSettings) error
}

// DummyDisplay is a display that does nothing
type DummyDisplay struct {
}

func NewDummyDisplay() *DummyDisplay {
	return &DummyDisplay{}
}

func (d DummyDisplay) Boot() error {
	return nil
}

func (d DummyDisplay) Render(screen Screen, settings ScreenSettings) error {
	return nil
}

const ESC = 0x1B

type TerminalDisplay struct {
	terminal        io.Writer
	OnChar, OffChar string
}

func NewTerminalDisplay() *TerminalDisplay {
	return NewTerminalDisplayWithOutput(os.Stdout)
}

func NewTerminalDisplayWithOutput(out io.Writer) *TerminalDisplay {
	return &TerminalDisplay{
		terminal: out,
		OnChar:   "##",
		OffChar:  "  ",
	}
}

// Boot implements Display.
func (disp *TerminalDisplay) Boot() error {
	os.Stdout.Write([]byte{
		// Move cursor do start
		ESC, '[', '1', 'H',
		// clear the terminal
		ESC, '[', '0', 'J',
	})

	return nil
}

func (disp *TerminalDisplay) Render(screen Screen, settings ScreenSettings) error {
	buff := make([]byte, 0, settings.Width*settings.Height*2+settings.Height+64)
	buff = append(buff, ESC, '[', '1', 'H')
	// buff = append(buff, ESC, '[', '0', 'J')
	for i, b := range screen {
		for bitJ := 0; bitJ < 8; bitJ++ {
			bit := b & (1 << (7 - byte(bitJ)))

			if bit > 0 {
				buff = append(buff, disp.OnChar...)
			} else {
				buff = append(buff, disp.OffChar...)
			}
		}

		if ((i+1)*8)%settings.Width == 0 {
			buff = append(buff, '|', '\n')
		}

	}

	_, err := disp.terminal.Write(buff)
	return err
}
