package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/guslan/xip8/gui"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})))
}

func main() {
	flag.Parse()

	app := gui.NewConsoleApp()

	if flag.NArg() > 0 {
		app.Load(flag.Arg(0))
	}

	app.Run()

	// rl.CloseWindow()
}
