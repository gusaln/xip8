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

	ScreenPixelSizeLarge = 16
	ScreenPixelSizeSmall = 12
	ScreenWidth          = 64 * ScreenPixelSizeLarge
	ScreenHeight         = 32 * ScreenPixelSizeLarge
	ScreenMargin         = 10
	ScreenPositionX      = 0
	ScreenPositionY      = ToolbarHeight + 1

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

type App struct {
	// The underlying console
	*xip8.InMemoryKeyboard
	// The underlying console
	Cpu *xip8.Cpu
	// Speed factor
	// Speed in Hz is max(1, speedFactor * 10)
	speedFactor float32
	// Unpacked screen representation
	screen     []byte
	pixelScale int32

	keyboardLayout    xip8.KeyboardLayout
	keyboardLookupMap map[ScanCode]byte

	useDebugger bool

	loadedProgramPath string

	// Window width and height
	winW, winH int

	loadBtn, startBtn, stopBtn, stepBtn, restBtn bool

	lastMessage      string
	lastMessageColor rl.Color
}

func speedFactorToHz(s float32) uint {
	return max(1, uint(s)*10)
}

func hzToSpeedFactor(hz uint) float32 {
	return float32(hz / 10)
}

// AppConfig
type AppConfig struct {
	// The initial speed
	Speed          uint
	UseDebugger    bool
	CyclesPerFrame uint
}
type AppConfigCb func(config *AppConfig)

func NewApp(configs ...AppConfigCb) *App {
	config := &AppConfig{
		Speed:          xip8.DefaultSpeed,
		UseDebugger:    false,
		CyclesPerFrame: xip8.DefaultCyclesPerFrame,
	}
	for _, cb := range configs {
		cb(config)
	}

	app := &App{
		InMemoryKeyboard:  xip8.NewInMemoryKeyboard(),
		Cpu:               nil,
		speedFactor:       hzToSpeedFactor(config.Speed),
		keyboardLayout:    xip8.DefaultKeyboardLayout,
		keyboardLookupMap: map[ScanCode]byte{},
		useDebugger:       config.UseDebugger,
	}

	app.Cpu = xip8.NewCpu(func(config *xip8.CpuConfig) {
		config.Display = app
		config.Keyboard = app
		config.Buzzer = app
	})
	app.screen = make([]byte, app.Cpu.ScreenSettings.Width*app.Cpu.ScreenSettings.Height)

	app.updateKeyboardLookupMap()
	app.updateWindowSize()

	if app.useDebugger {
		slog.Info("Using debugger")
	}

	return app
}

// Run initializes the console and the UI loop
func (app *App) Run(autostart bool) {
	go func(cpu *xip8.Cpu, autostart bool) {
		slog.Info("starting CPU loop on pause")
		cpu.Boot()
		if !app.hasProgramLoaded() || !autostart {
			cpu.Stop()
		}
		if err := cpu.Loop(); err != nil {
			app.showMessage(err.Error(), MessageError)
			slog.Error("Error booting CPU", slog.Any("error", err))
		}
	}(app.Cpu, autostart)

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
		app.drawDebugger()
		app.drawScreen()
		app.drawToolbar()

		rl.EndDrawing()
	}
}

func (app *App) Load(path string) {
	program, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Error loading program", slog.String("path", path), slog.Any("error", err))
		return
	}

	if err := app.Cpu.LoadProgram(program); err != nil {
		slog.Error("Error loading program", slog.String("path", path), slog.Any("error", err))
		return
	}

	app.loadedProgramPath = path
	slog.Info("Program loaded", slog.String("path", path))
	app.showMessage(fmt.Sprintf("Program '%s' loaded", app.loadedProgramPath), MessageInfo)

	app.Cpu.Start()
}

// Play implements xip8.Buzzer.
func (app *App) Play() {
}

// Stop implements xip8.Buzzer.
func (app *App) Stop() {
}

func (app *App) updateWindowSize() {
	app.winW = ScreenWidth
	app.winH = ScreenHeight + ToolbarHeight + MessageBarHeigh
	if app.Cpu.ScreenSettings.Width == xip8.SmallScreen.Width {
		app.pixelScale = ScreenPixelSizeLarge
	} else {
		app.pixelScale = ScreenPixelSizeSmall
	}

	if app.useDebugger {
		app.winH += 2*DebuggerRegisterMargin + 8*DebuggerRegisterHeight
	}
	slog.Info("Updating window size", slog.Int("width", app.winW), slog.Int("height", app.winH))
}

func (app *App) updateKeyboardLookupMap() {
	runeToConsoleKey := xip8.LookupMap(app.keyboardLayout)
	// app.keyboardMap = map[ScanCode]byte{}
	for r, k := range runeToConsoleKey {
		app.keyboardLookupMap[runeToKey[r]] = k
	}
}

func (app *App) loadStyles() {
	slog.Info("Loading styles")
	gui.LoadStyleFromMemory(resources.StyleRgs)
	// gui.SetFont(rl.LoadFontFromMemory(".ttf", resources.FiraMonoTtf, 14, resources.FiraMonoCodePoints))
}

func (app *App) handleFileLoad() {
	if rl.IsFileDropped() {
		files := rl.LoadDroppedFiles()
		defer rl.UnloadDroppedFiles()

		slog.Info("Files were dropped", "files", strings.Join(files, ","))

		app.Load(files[0])
	}
}

func (app App) hasProgramLoaded() bool {
	return len(app.loadedProgramPath) > 0
}

func (app *App) handleActions() {
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

func (app *App) handleKeyPress() {
	for scanCode, key := range app.keyboardLookupMap {
		if rl.IsKeyDown(scanCode) {
			app.InMemoryKeyboard.State |= (0b1000000000000000 >> key)
			// fmt.Printf("keyboard pressed %016b\n", app.InMemoryKeyboard.State)
		} else {
			app.InMemoryKeyboard.State &= ^(0b1000000000000000 >> key)
		}
	}
}

func (app *App) updateCpuSpeed() {
	app.Cpu.SetSpeedInHz(speedFactorToHz(app.speedFactor))
}

// var active int32 = 3
// var dropdownOpen bool = false
// var speeds = []int{
// 	700, 600, 500, 100, 60, 30,
// }

const (
	minSpeed = 0
	maxSpeed = 70
)

func (app *App) drawToolbar() {
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
		minSpeed,
		maxSpeed,
	)

}

var t int

func (app *App) drawScreen() {
	for y := 0; y < app.Cpu.ScreenSettings.Height; y++ {
		for x := 0; x < app.Cpu.ScreenSettings.Width; x++ {
			t = y*app.Cpu.ScreenSettings.Width + x

			if app.screen[t] > 0 {
				rl.DrawRectangle(
					ScreenPositionX+app.pixelScale*int32(x),
					ScreenPositionY+app.pixelScale*int32(y),
					app.pixelScale,
					app.pixelScale,
					ScreenPixelColor)
			} else {
				rl.DrawRectangle(
					ScreenPositionX+app.pixelScale*int32(x),
					ScreenPositionY+app.pixelScale*int32(y),
					app.pixelScale,
					app.pixelScale,
					ScreenBgColor)
			}
		}
	}
}

const (
	DebuggerRegisterPosX   = 0
	DebuggerRegisterPosY   = (ToolbarHeight + ScreenHeight + 1)
	DebuggerRegisterMargin = 5
	DebuggerRegisterWidth  = 50
	DebuggerRegisterHeight = 14

	DebuggerRegisterColGap   = 15
	DebuggerRegisterCol1PosX = DebuggerRegisterMargin
	DebuggerRegisterCol1PosY = DebuggerRegisterPosY + DebuggerRegisterMargin

	DebuggerRegisterCol2PosX = DebuggerRegisterCol1PosX + DebuggerRegisterWidth + DebuggerRegisterColGap
	DebuggerRegisterCol2PosY = DebuggerRegisterPosY + DebuggerRegisterMargin

	DebuggerRegisterCol3PosX = DebuggerRegisterCol2PosX + DebuggerRegisterWidth + DebuggerRegisterColGap
	DebuggerRegisterCol3PosY = DebuggerRegisterPosY + DebuggerRegisterMargin

	DebuggerRegisterCol4PosX = DebuggerRegisterCol3PosX + 2*DebuggerRegisterWidth + DebuggerRegisterColGap
	DebuggerRegisterCol4PosY = DebuggerRegisterPosY + DebuggerRegisterMargin

	DebuggerRegisterCol5PosX = DebuggerRegisterCol4PosX + 2*DebuggerRegisterWidth + DebuggerRegisterColGap
	DebuggerRegisterCol5PosY = DebuggerRegisterPosY + DebuggerRegisterMargin
)

func (app *App) drawDebugger() {
	if !app.useDebugger {
		return
	}

	for i, vi := range app.Cpu.V[:8] {
		gui.Label(rl.NewRectangle(
			DebuggerRegisterCol1PosX, DebuggerRegisterCol1PosY+DebuggerRegisterHeight*float32(i),
			DebuggerRegisterWidth, DebuggerRegisterHeight),
			fmt.Sprintf("V%X 0x%02X", i, vi))
	}

	for i, vi := range app.Cpu.V[8:] {
		gui.Label(rl.NewRectangle(
			DebuggerRegisterCol2PosX, DebuggerRegisterCol2PosY+DebuggerRegisterHeight*float32(i),
			DebuggerRegisterWidth, DebuggerRegisterHeight),
			fmt.Sprintf("V%X 0x%02X", i+8, vi))
	}

	gui.Label(rl.NewRectangle(
		DebuggerRegisterCol3PosX, DebuggerRegisterCol3PosY+DebuggerRegisterHeight*0,
		DebuggerRegisterWidth*2, DebuggerRegisterHeight),
		fmt.Sprintf("PC 0x%04X", app.Cpu.Pc))

	gui.Label(rl.NewRectangle(
		DebuggerRegisterCol3PosX, DebuggerRegisterCol3PosY+DebuggerRegisterHeight*1,
		DebuggerRegisterWidth*2, DebuggerRegisterHeight),
		fmt.Sprintf("SP 0x%02X", app.Cpu.Sp))

	gui.Label(rl.NewRectangle(
		DebuggerRegisterCol3PosX, DebuggerRegisterCol3PosY+DebuggerRegisterHeight*2,
		DebuggerRegisterWidth*2, DebuggerRegisterHeight),
		fmt.Sprintf("I 0x%04X", app.Cpu.I))

	gui.Label(rl.NewRectangle(
		DebuggerRegisterCol3PosX, DebuggerRegisterCol3PosY+DebuggerRegisterHeight*3,
		DebuggerRegisterWidth, DebuggerRegisterHeight),
		fmt.Sprintf("DT 0x%02X", app.Cpu.Dt))
	gui.Label(rl.NewRectangle(
		DebuggerRegisterCol3PosX+DebuggerRegisterMargin+DebuggerRegisterWidth, DebuggerRegisterCol3PosY+DebuggerRegisterHeight*3,
		DebuggerRegisterWidth, DebuggerRegisterHeight),
		fmt.Sprintf("ST 0x%02X", app.Cpu.St))

	for i, si := range app.Cpu.Stack[:8] {
		gui.Label(rl.NewRectangle(
			DebuggerRegisterCol4PosX, DebuggerRegisterCol4PosY+DebuggerRegisterHeight*float32(i),
			DebuggerRegisterWidth*2, DebuggerRegisterHeight),
			fmt.Sprintf("S[%X] 0x%04X", i, si))
	}

	for i, si := range app.Cpu.Stack[8:] {
		gui.Label(rl.NewRectangle(
			DebuggerRegisterCol5PosX, DebuggerRegisterCol5PosY+DebuggerRegisterHeight*float32(i),
			DebuggerRegisterWidth*2, DebuggerRegisterHeight),
			fmt.Sprintf("S[%X] 0x%04X", i+8, si))
	}
}

func (app *App) showMessage(msg string, mType MessageType) {
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

func (app *App) drawMessageBar() {
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
