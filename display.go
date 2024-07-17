package xip8

import "math"

type DisplayMode int

const (
	standard_64x32 DisplayMode = iota
	eti660_64x48
	eti660_64x64
	superChip48_128x64
)

// Display abstraction for a display
type Display interface {
	// Clears the screen.
	Clear()
	// Display displays the sprite at location x, y
	// Sprites are XORed onto the existing screen.
	Display(x, y, sprite byte) bool
}

func toScreenCoord(w, h, x, y byte) uint {
	if x > w-1 {
		x = x - w
	}

	if y > h-1 {
		y = y - h
	}

	return uint(y)*uint(h) + uint(x)
}

// InMemoryDisplay stores the information of the screen in a slice
// Useful for embedding and debugging
type InMemoryDisplay struct {
	W, H   byte
	Screen []byte
}

func NewDefaultInMemoryDisplay() *InMemoryDisplay {
	return NewInMemoryDisplay(64, 32)
}

func NewInMemoryDisplay(w, h byte) *InMemoryDisplay {
	return &InMemoryDisplay{
		W:      w,
		H:      h,
		Screen: make([]byte, int(math.Ceil(float64(w*h)/8.0))),
	}
}

func (disp *InMemoryDisplay) Clear() {
	disp.Screen = make([]byte, int(math.Ceil(float64(disp.W*disp.H)/8.0)))
}

func (disp *InMemoryDisplay) Display(x, y, sprite byte) bool {
	t := toScreenCoord(byte(disp.W), byte(disp.H), x, y)
	buf := disp.Screen[t]
	disp.Screen[t] = disp.Screen[t] ^ sprite

	// previous & ~current
	return (buf & (disp.Screen[t] ^ 0xFF)) > 0
}
