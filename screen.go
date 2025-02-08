package xip8

import (
	"math"
)

// Screen representation
type Screen []byte

// ScreenSettings for the console
// Common display sizes are 64x32 and 128x64.
// Other uncommon sizes are 64x48 and 64x64.
type ScreenSettings struct {
	Width, Height int
}

var SmallScreen = ScreenSettings{
	Width:  64,
	Height: 32,
}

func (cpu *Cpu) clearScreen() {
	cpu.screen = make([]byte, sizeInBytesOfScreen(cpu.ScreenSettings.Width, cpu.ScreenSettings.Height))
}

func sizeInBytesOfScreen(w, h int) int {
	return int(math.Ceil(float64(w*h) / 8.0))
}

func newScreen(w, h int) []byte {
	return make([]byte, sizeInBytesOfScreen(w, h))
}

func (cpu Cpu) toScreenCoord(x, y byte) uint {
	x = x % byte(cpu.ScreenSettings.Width)
	y = y % byte(cpu.ScreenSettings.Height)

	return uint(y)*uint(cpu.ScreenSettings.Width) + uint(x)
}

// displayToScreen displays the sprite at location x, y
// Sprites are XORed onto the existing screen.
// Returns whether there was a collision or not.
func (cpu *Cpu) displayToScreen(x, y, sprite byte) bool {
	tReal := cpu.toScreenCoord(x, y)

	// We are drawing to an aligned position
	if tReal%8 == 0 {
		t := tReal / 8

		buf := cpu.screen[t]
		cpu.screen[t] = cpu.screen[t] ^ sprite

		// previous & ~current
		return (buf & (cpu.screen[t] ^ 0xFF)) > 0
	}

	// Not an aligned position.
	tOffset := tReal % 8

	t := (tReal - tOffset) / 8

	firstBuf := cpu.screen[t]
	cpu.screen[t] = cpu.screen[t] ^ byte(sprite>>byte(tOffset))

	secondBuf := cpu.screen[t+1]
	cpu.screen[t+1] = cpu.screen[t+1] ^ byte(sprite<<byte(8-tOffset))

	// previous & ~current
	return ((firstBuf & (cpu.screen[t] ^ 0xFF)) > 0) || ((secondBuf & (cpu.screen[t+1] ^ 0xFF)) > 0)
}
