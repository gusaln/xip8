package xip8

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
