package xip8

import (
	"math"
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
	// TODO: handle mid-byte positions
	t := disp.toScreenCoord(x, y) / 8

	buf := disp.Screen[t]
	disp.Screen[t] = disp.Screen[t] ^ sprite

	// previous & ~current
	return (buf & (disp.Screen[t] ^ 0xFF)) > 0
}
