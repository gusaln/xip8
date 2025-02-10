package gui

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

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
	// Speed factor
	// Speed in Hz is speedFactor+1 * 5
	speedFactor float32
	// Unpacked screen representation
	screen []byte

	keyboardLayout    xip8.KeyboardLayout
	keyboardLookupMap map[ScanCode]byte

	// Window width and height
	winW, winH int

	// Toolbar
	loadBtn, startBtn, stopBtn, stepBtn, restBtn bool

	loadedProgramPath string

	lastMessage      string
	lastMessageColor rl.Color
}

func speedFactorToHz(s float32) uint {
	return uint((s + 1) * 5)
}

func hzToSpeedFactor(hz uint) float32 {
	return float32(hz)/5 - 1
}

func NewConsoleApp() *ConsoleApp {
	app := &ConsoleApp{
		InMemoryKeyboard:  xip8.NewInMemoryKeyboard(),
		Cpu:               nil,
		speedFactor:       hzToSpeedFactor(xip8.DefaultSpeed),
		keyboardLayout:    xip8.DefaultKeyboardLayout,
		keyboardLookupMap: map[ScanCode]byte{},
	}

	app.Cpu = xip8.NewCpu(xip8.NewMemory(), xip8.SmallScreen, app, app, app)
	app.screen = make([]byte, app.Cpu.ScreenSettings.Width*app.Cpu.ScreenSettings.Height)

	app.updateKeyboardLookupMap()
	app.updateWindowSize()

	return app
}

// Run initializes the console and the UI loop
func (app *ConsoleApp) Run() {
	go func(cpu *xip8.Cpu) {
		slog.Info("starting CPU loop on pause")
		cpu.Boot()
		cpu.Stop()
		if err := cpu.Loop(); err != nil {
			app.showMessage(err.Error(), MessageError)
			slog.Error("Error booting CPU", slog.Any("error", err))
		}
	}(app.Cpu)

	rl.InitWindow(int32(app.winW), int32(app.winH), "xip8")
	defer rl.CloseWindow()

	app.loadStyles()
	rl.SetTargetFPS(60)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		app.handleFileLoad()
		app.handleActions()
		app.handleKeyPress()
		app.updateCpuSpeed()

		// Sections get rendered from bottom to the top because otherwise so that dropdown menus not
		app.drawMessageBar()
		app.drawScreen()
		app.drawToolbar()

		rl.EndDrawing()
	}
}

func (app *ConsoleApp) Load(path string) {
	program, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Error loading program", slog.String("path", path), slog.Any("error", err))
		return
	}

	if err = app.Cpu.LoadProgram(program); err != nil {
		slog.Error("Error loading program", slog.String("path", path), slog.Any("error", err))
		return
	}

	// app.Cpu.Memory[0x1FF] = 1

	app.loadedProgramPath = path
	slog.Info("Program loaded", slog.String("path", path))
	app.showMessage(fmt.Sprintf("Program '%s' loaded", app.loadedProgramPath), MessageInfo)

	app.Cpu.Start()
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

func (app *ConsoleApp) updateKeyboardLookupMap() {
	runeToConsoleKey := xip8.LookupMap(app.keyboardLayout)
	// app.keyboardMap = map[ScanCode]byte{}
	for r, k := range runeToConsoleKey {
		app.keyboardLookupMap[runeToKey[r]] = k
	}
}

func (app *ConsoleApp) loadStyles() {
	slog.Info("Loading styles")
	gui.LoadStyleFromMemory(resources.StyleRgs)
}

func (app *ConsoleApp) handleFileLoad() {
	if rl.IsFileDropped() {
		files := rl.LoadDroppedFiles()
		defer rl.UnloadDroppedFiles()

		slog.Info("Files were dropped", "files", strings.Join(files, ","))

		app.Load(files[0])
	}
}

func (app ConsoleApp) hasProgramLoaded() bool {
	return len(app.loadedProgramPath) > 0
}

func (app *ConsoleApp) handleActions() {
	if app.startBtn {
		if app.hasProgramLoaded() {
			app.Cpu.Start()
			slog.Info("Starting the console")
		} else {
			app.showMessage("There is no program loaded", MessageError)
		}
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
		app.Cpu.LoopOnce()
		slog.Info("Running a single frame")
	}
}

func (app *ConsoleApp) handleKeyPress() {
	for scanCode, key := range app.keyboardLookupMap {
		if rl.IsKeyDown(scanCode) {
			app.InMemoryKeyboard.State |= (0b1000000000000000 >> key)
			// fmt.Printf("keyboard pressed %016b\n", app.InMemoryKeyboard.State)
		} else {
			app.InMemoryKeyboard.State &= ^(0b1000000000000000 >> key)
		}
	}
}

func (app *ConsoleApp) updateCpuSpeed() {
	app.Cpu.SetSpeedInHz(speedFactorToHz(app.speedFactor))
}

// var active int32 = 3
// var dropdownOpen bool = false
// var speeds = []int{
// 	700, 600, 500, 100, 60, 30,
// }

const (
	MinSpeed = float32(xip8.MinSpeed/5) - 1
	MaxSpeed = float32(xip8.MaxSpeed/5) - 1
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

	app.startBtn = gui.Button(
		rl.NewRectangle(ToolbarGap+ToolbarBtnOffset*0, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_PLAYER_PLAY, "Start"),
	)
	app.stopBtn = gui.Button(
		rl.NewRectangle(ToolbarGap+ToolbarBtnOffset*1, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_PLAYER_STOP, "Stop"),
	)
	app.stepBtn = gui.Button(
		rl.NewRectangle(ToolbarGap+ToolbarBtnOffset*2, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_PLAYER_NEXT, "Step"),
	)
	app.restBtn = gui.Button(
		rl.NewRectangle(ToolbarGap+ToolbarBtnOffset*3, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
		gui.IconText(gui.ICON_ROTATE, "Reset"),
	)

	if app.Cpu.IsRunning() {
		gui.Label(
			rl.NewRectangle(ToolbarGap+ToolbarBtnOffset*4, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
			"Running",
		)
	} else {
		gui.Label(
			rl.NewRectangle(ToolbarGap+ToolbarBtnOffset*4, ToolbarGap, ToolbarBtnWidth, ToolbarBtnHeight),
			"Stopped",
		)
	}

	gui.Label(
		rl.NewRectangle(float32(app.winW)-ToolbarGap-150, 26, 50, 20),
		fmt.Sprintf("%d Hz", speedFactorToHz(app.speedFactor)),
	)

	if gui.Button(
		rl.NewRectangle(float32(app.winW)-ToolbarGap-150+50, 26, 50, 20),
		gui.IconText(gui.ICON_ROTATE, ""),
	) {
		app.speedFactor = hzToSpeedFactor(xip8.DefaultSpeed)
	}

	app.speedFactor = gui.Slider(
		rl.NewRectangle(float32(app.winW)-ToolbarGap-150, ToolbarGap, 100, 20),
		"1 Hz", "700 Hz",
		app.speedFactor,
		MinSpeed,
		MaxSpeed,
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
