package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/guslan/xip8"
	"github.com/guslan/xip8/gui"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
}

func main() {
	autostart := flag.Bool("start", false, "Starts the console automatically if there is a program loaded (defaults = false).")
	debug := flag.Bool("debug", false, "Show debug information for the console (defaults = false).")
	initialSpeed := flag.Uint("speed", xip8.DefaultSpeed, fmt.Sprintf("The starting speed of the CPU in Hz. It has to be in the range [5, 700] (defaults = %d).", xip8.DefaultSpeed))
	cyclesPerFrame := flag.Uint("xframes", xip8.DefaultCyclesPerFrame, fmt.Sprintf("The number of cycles that run between each frame (defaults = %d).", xip8.DefaultCyclesPerFrame))

	flag.Parse()

	var app *gui.App

	app = gui.NewApp(func(config *gui.AppConfig) {
		config.Speed = max(*initialSpeed, 5)
		config.UseDebugger = *debug
		config.CyclesPerFrame = *cyclesPerFrame
	})

	if flag.NArg() > 0 {
		app.Load(flag.Arg(0))
	}

	app.Run(*autostart)

	// rl.CloseWindow()
}
