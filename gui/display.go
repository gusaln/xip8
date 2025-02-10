package gui

import (
	"github.com/guslan/xip8"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var ScreenBgColor = rl.Gold
var ScreenPixelColor = rl.Yellow

// Boot implements xip8.Display.
func (app *ConsoleApp) Boot() error {
	return nil
}

// Render implements xip8.Display.
func (app *ConsoleApp) Render(screen xip8.Screen, settings xip8.ScreenSettings) error {
	// if len(app.screen) < len(screen) {
	// 	app.screen = make(xip8.Screen, settings.Width*settings.Height)
	// }

	for i, t := 0, 0; t < settings.Width*settings.Height; i, t = i+1, t+8 {
		app.screen[t+0] = (screen[i] >> 7) & 0b1
		app.screen[t+1] = (screen[i] >> 6) & 0b1
		app.screen[t+2] = (screen[i] >> 5) & 0b1
		app.screen[t+3] = (screen[i] >> 4) & 0b1
		app.screen[t+4] = (screen[i] >> 3) & 0b1
		app.screen[t+5] = (screen[i] >> 2) & 0b1
		app.screen[t+6] = (screen[i] >> 1) & 0b1
		app.screen[t+7] = (screen[i] >> 0) & 0b1
	}

	return nil
}
