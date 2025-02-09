package main

import (
	"flag"
	"log/slog"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
}

func main() {
	flag.Parse()

	app := newConsoleApp()

	if flag.NArg() > 0 {
		app.load(flag.Arg(0))
	}

	rl.InitWindow(int32(app.winW), int32(app.winH), "xip8")
	defer rl.CloseWindow()

	app.loadStyles()
	app.bootCpu()
	rl.SetTargetFPS(30)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		app.handleFileLoad()
		app.handleActions()
		app.updateCpuSpeed()

		app.drawMessageBar()
		app.drawScreen()
		app.drawToolbar()

		rl.EndDrawing()
	}

	// rl.CloseWindow()
}
