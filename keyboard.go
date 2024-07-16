package xip8

type KeyboardState [16]bool

type Keyboard interface {
	IsPressed(k byte) bool
	Get() KeyboardState
}
