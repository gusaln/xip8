package xip8

import (
	"math"
	"os"
)

// Display abstraction for a display
// Common display sizes are 64x32 and 128x64.
// Other uncommon sizes are 64x48 and 64x64.
type Display interface {
	// Boot initializes the component
	Boot() error
	// Clears the screen.
	Clear()
	// Display displays the sprite at location x, y
	// Sprites are XORed onto the existing screen.
	Display(x, y, sprite byte) bool
	// Size returns the size of the screen
	Size() int
	// Render
	Render() error
}

// InMemoryDisplay stores the information of the screen in a slice
// Useful for embedding and debugging
type InMemoryDisplay struct {
	Width, Height int
	Screen        []byte
}

func sizeInBytesOf(w, h int) int {
	return int(math.Ceil(float64(w*h) / 8.0))
}

// NewDefaultInMemoryDisplay creates an in-memory display of size 64x32
func NewDefaultInMemoryDisplay() *InMemoryDisplay {
	return NewInMemoryDisplay(64, 32)
}

func NewInMemoryDisplay(w, h int) *InMemoryDisplay {
	return &InMemoryDisplay{
		Width:  w,
		Height: h,
		Screen: make([]byte, sizeInBytesOf(w, h)),
	}
}

func (disp *InMemoryDisplay) Boot() error {
	return nil
}

func (disp *InMemoryDisplay) Clear() {
	disp.Screen = make([]byte, sizeInBytesOf(disp.Width, disp.Height))
}

func (disp *InMemoryDisplay) Size() int {
	return disp.Width * disp.Height
}

func (disp InMemoryDisplay) toScreenCoord(x, y byte) uint {
	x = x % byte(disp.Width)
	y = y % byte(disp.Height)

	return uint(y)*uint(disp.Width) + uint(x)
}

func (disp *InMemoryDisplay) Display(x, y, sprite byte) bool {
	tReal := disp.toScreenCoord(x, y)

	if tReal%8 == 0 {
		t := tReal / 8

		buf := disp.Screen[t]
		disp.Screen[t] = disp.Screen[t] ^ sprite

		// previous & ~current
		return (buf & (disp.Screen[t] ^ 0xFF)) > 0
	}

	tOffset := tReal % 8

	t := (tReal - tOffset) / 8

	firstBuf := disp.Screen[t]
	disp.Screen[t] = disp.Screen[t] ^ byte(sprite>>byte(tOffset))

	secondBuf := disp.Screen[t+1]
	disp.Screen[t+1] = disp.Screen[t+1] ^ byte(sprite<<byte(8-tOffset))

	// previous & ~current
	return ((firstBuf & (disp.Screen[t] ^ 0xFF)) > 0) || ((secondBuf & (disp.Screen[t+1] ^ 0xFF)) > 0)
}

func (disp *InMemoryDisplay) Render() error {
	return nil
}

const ESC = 0x1B

type TerminalDisplay struct {
	OnChar, OffChar string

	*InMemoryDisplay
}

func NewTerminalDisplay(w, h int) *TerminalDisplay {
	return &TerminalDisplay{
		OnChar:  "##",
		OffChar: "  ",

		InMemoryDisplay: NewInMemoryDisplay(w, h),
	}
}

// NewDefaultTerminalDisplay creates an terminal display of size 64x32
func NewDefaultTerminalDisplay() *TerminalDisplay {
	return NewTerminalDisplay(64, 32)
}

// Boot implements Display.
func (disp *TerminalDisplay) Boot() error {
	os.Stdout.Write([]byte{
		// Move cursor do start
		ESC, '[', '1', 'H',
		// clear the terminal
		ESC, '[', '0', 'J',
	})

	disp.Render()

	return nil
}

func (disp *TerminalDisplay) Clear() {
	disp.InMemoryDisplay.Clear()

	disp.Render()
}

func (disp *TerminalDisplay) Display(x, y, sprite byte) bool {
	collision := disp.InMemoryDisplay.Display(x, y, sprite)

	disp.Render()

	return collision
}

func (disp *TerminalDisplay) Render() error {
	buff := make([]byte, 0, disp.Size()*2+disp.Height+64)
	buff = append(buff, ESC, '[', '1', 'H')
	// buff = append(buff, ESC, '[', '0', 'J')
	for i, b := range disp.Screen {
		for bitJ := 0; bitJ < 8; bitJ++ {
			bit := b & (1 << (7 - byte(bitJ)))

			if bit > 0 {
				buff = append(buff, disp.OnChar...)
			} else {
				buff = append(buff, disp.OffChar...)
			}
		}

		if ((i+1)*8)%disp.Width == 0 {
			buff = append(buff, '|', '\n')
		}

	}

	os.Stdout.Write(buff)
	return nil
}
