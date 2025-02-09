package main

import (
	"fmt"
	"log/slog"
	"os"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/guslan/xip8"
	"github.com/guslan/xip8/resources"
)

const (
	ToolbarGap       = 5
	ToolbarBtnWidth  = 80
	ToolbarBtnHeight = 40
	ToolbarHeight    = 50
	ToolbarBtnOffset = ToolbarBtnWidth + ToolbarGap

	ScreenMargin    = 10
	ScreenHeight    = 400
	ScreenPixelSize = 15
	ScreenPositionX = 0
	ScreenPositionY = ToolbarHeight + 1

	MessageBarGap   = 5
	MessageBarHeigh = 30
)

var ScreenBgColor = rl.Gold
var ScreenPixelColor = rl.Yellow
var MessageBarBgColor = rl.DarkGray
var MessageBarInfoColor = rl.SkyBlue
var MessageBarSuccessColor = rl.Lime
var MessageBarWarningColor = rl.Gold
var MessageBarErrorColor = rl.Red

type MessageType byte

const (
	MessageInfo MessageType = iota
	MessageSuccess
	MessageWarning
	MessageError
)

type ConsoleApp struct {
	// The underlying console
	*xip8.InMemoryKeyboard
	// The underlying console
	Cpu *xip8.Cpu
	// Speed in Hz
	speed float32
	// Unpacked screen representation
	screen []byte
	// screenSettings xip8.ScreenSettings

	// Window width and height
	winW, winH int

	// Toolbar
	loadBtn, startBtn, stopBtn, stepBtn, restBtn bool

	loadedProgramPath string

	lastMessage      string
	lastMessageColor rl.Color
}

func newConsoleApp() *ConsoleApp {
	app := &ConsoleApp{
		InMemoryKeyboard: xip8.NewInMemoryKeyboard(),
		Cpu:              nil,
		speed:            float32(xip8.DefaultSpeed),
	}

	app.Cpu = xip8.NewCpu(xip8.NewMemory(), xip8.SmallScreen, app, app, app)
	app.screen = make([]byte, app.Cpu.ScreenSettings.Width*app.Cpu.ScreenSettings.Height)
	app.updateWindowSize()

	return app
}

func (app *ConsoleApp) loadStyles() {
	slog.Info("Loading styles")
	gui.LoadStyleFromMemory(resources.StyleRgs)
}

func (app *ConsoleApp) bootCpu() {
	go func(cpu *xip8.Cpu) {
		slog.Info("starting CPU loop on pause")
		cpu.Boot()
		cpu.Stop()
		if err := cpu.Loop(); err != nil {
			app.showMessage(err.Error(), MessageError)
			slog.Error("Error booting CPU", slog.Any("error", err))
		}
	}(app.Cpu)
}

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

// Play implements xip8.Buzzer.
func (app *ConsoleApp) Play() {
}

// Stop implements xip8.Buzzer.
func (app *ConsoleApp) Stop() {
}

func (app *ConsoleApp) updateWindowSize() {
	app.winW = app.Cpu.ScreenSettings.Width * ScreenPixelSize
	app.winH = app.Cpu.ScreenSettings.Height*ScreenPixelSize + ToolbarHeight + MessageBarHeigh
	slog.Info("Updating window size", slog.Int("width", app.winW), slog.Int("height", app.winH))
}

func (app *ConsoleApp) load(path string) {
	program, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Error loading program", slog.String("path", path), slog.Any("error", err))
		return
	}

	if err = app.Cpu.LoadProgram(program); err != nil {
		slog.Error("Error loading program", slog.String("path", path), slog.Any("error", err))
		return
	}

	app.loadedProgramPath = path
	slog.Info("Program loaded", slog.String("path", path))
	app.showMessage(fmt.Sprintf("Program '%s' loaded", app.loadedProgramPath), MessageInfo)
}

func (app *ConsoleApp) handleFileLoad() {
	// if app.loadBtn {
	// } else if rl.IsFileDropped() {
	// 	files := rl.LoadDroppedFiles()

	// 	app.load(files[0])

	// 	rl.UnloadDroppedFiles()
	// }
	if rl.IsFileDropped() {
		files := rl.LoadDroppedFiles()
		defer rl.UnloadDroppedFiles()

		// slog.Info("Files were dropped", "files", strings.Join(files, ","))

		app.load(files[0])
	}
}

func (app *ConsoleApp) handleActions() {
	if app.startBtn {
		app.Cpu.Start()
		slog.Info("Starting the console")
	}
	if app.stopBtn {
		app.Cpu.Stop()
		slog.Info("Stopping the console")
	}
	if app.restBtn {
		app.Cpu.Reset()
		slog.Info("Resetting the program to the beginning")
	}
	if app.stepBtn {
		app.Cpu.SingleFrame()
		slog.Info("Running a single frame")
	}
}

func (app *ConsoleApp) updateCpuSpeed() {
	app.Cpu.SetSpeedInHz(uint(app.speed))
}

var active int32 = 3
var dropdownOpen bool = false
var speeds = []int{
	700, 600, 500, 100, 60, 30,
}

const (
	MinSpeed = float32(xip8.MinSpeed)
	MaxSpeed = float32(xip8.MaxSpeed)
)

func (app *ConsoleApp) drawToolbar() {
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), ToolbarHeight, rl.Gray)

	// if gui.DropdownBox(
	// 	rl.NewRectangle(ToolbarGap, ToolbarGap, 100, ToolbarBtnHeight),
	// 	"Speed;700;600;500;100;60;30",
	// 	&active,
	// 	dropdownOpen,
	// ) {
	// 	dropdownOpen = !dropdownOpen
	// 	if active == 0 {
	// 		// dropdownOpen = !dropdownOpen
	// 	} else {
	// 		slog.Info("speed selected", "active", active, "speed", speeds[active-1])
	// 	}
	// }

	gui.Label(
		rl.NewRectangle(ToolbarGap, 26, 50, 20),
		fmt.Sprintf("%.0f Hz", app.speed),
	)

	if gui.Button(
		rl.NewRectangle(ToolbarGap+50, 26, 50, 20),
		gui.IconText(gui.ICON_ROTATE, ""),
	) {
		app.speed = float32(xip8.DefaultSpeed)
	}

	app.speed = gui.Slider(
		rl.NewRectangle(ToolbarGap*6, ToolbarGap, 100, 20),
		"1 Hz", "700 Hz",
		app.speed,
		MinSpeed,
		MaxSpeed,
	)

	app.startBtn = gui.Button(
		rl.NewRectangle(float32(app.winW)-4*ToolbarBtnOffset, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_PLAYER_PLAY, "Start"),
	)
	app.stopBtn = gui.Button(
		rl.NewRectangle(float32(app.winW)-3*ToolbarBtnOffset, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_PLAYER_STOP, "Stop"),
	)
	app.stepBtn = gui.Button(
		rl.NewRectangle(float32(app.winW)-2*ToolbarBtnOffset, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_PLAYER_NEXT, "Step"),
	)
	app.restBtn = gui.Button(
		rl.NewRectangle(float32(app.winW)-1*ToolbarBtnOffset, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_ROTATE, "Reset"),
	)
}

var t int

func (app *ConsoleApp) drawScreen() {
	for y := 0; y < app.Cpu.ScreenSettings.Height; y++ {
		for x := 0; x < app.Cpu.ScreenSettings.Width; x++ {
			t = y*app.Cpu.ScreenSettings.Width + x

			if app.screen[t] > 0 {
				rl.DrawRectangle(
					ScreenPositionX+ScreenPixelSize*int32(x),
					ScreenPositionY+ScreenPixelSize*int32(y),
					ScreenPixelSize,
					ScreenPixelSize,
					ScreenPixelColor)
			} else {
				rl.DrawRectangle(
					ScreenPositionX+ScreenPixelSize*int32(x),
					ScreenPositionY+ScreenPixelSize*int32(y),
					ScreenPixelSize,
					ScreenPixelSize,
					ScreenBgColor)
			}
		}
	}
}

func (app *ConsoleApp) showMessage(msg string, mType MessageType) {
	app.lastMessage = msg
	switch mType {
	case MessageInfo:
		app.lastMessageColor = MessageBarInfoColor

	case MessageSuccess:
		app.lastMessageColor = MessageBarSuccessColor

	case MessageWarning:
		app.lastMessageColor = MessageBarWarningColor

	case MessageError:
		app.lastMessageColor = MessageBarErrorColor
	}
}

func (app *ConsoleApp) drawMessageBar() {
	rl.DrawRectangle(
		0,
		int32(app.winH)-MessageBarHeigh,
		int32(app.winW),
		MessageBarHeigh,
		MessageBarBgColor,
	)

	rl.DrawText(
		app.lastMessage,
		MessageBarGap,
		int32(app.winH)-MessageBarHeigh+MessageBarGap,
		16,
		app.lastMessageColor,
	)
}
