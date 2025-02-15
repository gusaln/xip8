package xip8

import (
	"errors"
	"fmt"
	"time"
)

var ErrCpuIsNotBooted = errors.New("the CPU has not been booted properly")

type ErrOpCodeUnknown struct {
	OpCode uint16
	Pc     uint16
}

func (err ErrOpCodeUnknown) Error() string {
	return fmt.Sprintf("unknown opcode=%X at PC=%d", err.OpCode, err.Pc)
}

var ErrStackUnderflow = errors.New("stack underflow: try to pop an empty stack")
var ErrStackOverflow = errors.New("stack overflow: try to push to a full stack")

// MachineRoutineInterpreter interpretes
type MachineRoutineInterpreter func(opCode uint16, cpu *Cpu) error

type QuirkFlag = byte

const (
	FlagQuirkVfReset          QuirkFlag = 0b00001
	FlagQuirkMemoryMovesIndex QuirkFlag = 0b00010
	FlagQuirkClipping         QuirkFlag = 0b00100
	FlagQuirkShiftWithVy      QuirkFlag = 0b01000
	FlagQuirkJumpUsesVx       QuirkFlag = 0b10000
)

const (
	Chip8Quirks QuirkFlag = FlagQuirkVfReset | FlagQuirkShiftWithVy | FlagQuirkMemoryMovesIndex
)

const (
	DefaultSpeed          uint = 500
	MaxSpeed              uint = 700
	MinSpeed              uint = 5
	DefaultCyclesPerFrame uint = 30
)

// Chip-8 CPU
type Cpu struct {
	Memory *Memory
	// V 8-bit registers
	V [16]byte
	// I 16-bit register (12-bit usable)
	I uint16
	// Delay timer register
	Dt byte
	// Sound timer register
	St byte
	// Program counter
	Pc uint16
	// Stack pointer
	Sp byte
	// Stack
	Stack [16]uint16

	cycles uint
	frames uint

	speedInHz      uint
	step           time.Duration
	CyclesPerFrame uint

	quirks QuirkFlag

	ScreenSettings ScreenSettings
	screen         []byte
	isScreenDirty  bool

	Display  Display
	Keyboard Keyboard
	Buzzer   Buzzer

	MachineRoutineInterpreter MachineRoutineInterpreter

	isBooted       bool
	isPaused       bool
	waitingForKey  bool
	keyDstRegister uint16
	lastError      error

	// Hooks that run before every frame
	beforeFrameHooks []Hook
	// Hooks that run before every cycle
	beforeCycleHooks []Hook
	// Hooks that run after every cycle
	afterCycleHooks []Hook
	// Hooks that run after every frame
	afterFrameHooks []Hook
	// Hooks that run after an error
	errorHooks []Hook
}

// CpuConfig
type CpuConfig struct {
	// Defaults to the
	Memory *Memory
	// Defaults to SmallScreen
	ScreenSettings ScreenSettings
	// Defaults to nothing
	Quirks QuirkFlag
	// Defaults to DummyDisplay
	Display Display
	// Defaults to InMemoryKeyboard
	Keyboard Keyboard
	// Defaults to DummyBuzzer
	Buzzer Buzzer

	CyclesPerFrame uint
}
type CpuConfigCb func(config *CpuConfig)

func NewCpu(configs ...CpuConfigCb) *Cpu {
	config := &CpuConfig{
		Memory:         NewMemory(),
		ScreenSettings: SmallScreen,
		Quirks:         Chip8Quirks,
		Display:        NewDummyDisplay(),
		Keyboard:       NewInMemoryKeyboard(),
		Buzzer:         NewDummyBuzzer(),
		CyclesPerFrame: DefaultCyclesPerFrame,
	}
	for _, cb := range configs {
		cb(config)
	}

	return &Cpu{
		Memory: config.Memory,

		V:     [16]byte{},
		I:     0,
		Dt:    0,
		St:    0,
		Pc:    0,
		Sp:    0,
		Stack: [16]uint16{},

		speedInHz:      DefaultSpeed,
		step:           time.Second / time.Duration(DefaultSpeed),
		CyclesPerFrame: config.CyclesPerFrame,

		quirks: config.Quirks,

		ScreenSettings: config.ScreenSettings,
		screen:         newScreen(config.ScreenSettings.Width, config.ScreenSettings.Height),
		isScreenDirty:  false,

		Display:  config.Display,
		Keyboard: config.Keyboard,
		Buzzer:   config.Buzzer,

		MachineRoutineInterpreter: nil,

		isBooted:       false,
		isPaused:       false,
		waitingForKey:  false,
		keyDstRegister: 0,
		lastError:      nil,

		beforeFrameHooks: make([]Hook, 0),
		beforeCycleHooks: make([]Hook, 0),
		afterCycleHooks:  make([]Hook, 0),
		afterFrameHooks:  make([]Hook, 0),
		errorHooks:       make([]Hook, 0),
	}
}

func (cpu Cpu) IsRunning() bool {
	return !cpu.isPaused
}

func (cpu Cpu) IsSoundTimerActive() bool {
	return cpu.St > 0
}

func (cpu Cpu) IsDelayTimerActive() bool {
	return cpu.Dt > 0
}

func (cpu Cpu) SpeedInHz() uint {
	return cpu.speedInHz
}

func (cpu *Cpu) SetSpeedInHz(inHz uint) {
	cpu.speedInHz = inHz
	cpu.step = time.Second / time.Duration(inHz)
}

func (cpu Cpu) Cycles() uint {
	return cpu.cycles
}

func (cpu Cpu) Frames() uint {
	return cpu.frames
}

func (cpu *Cpu) SetQuirks(q QuirkFlag) {
	cpu.quirks = q
}

// Boot initializes all the components
// If the CPU was already booted, this method is a noop
func (cpu *Cpu) Boot() error {
	if cpu.isBooted {
		return nil
	}

	if err := cpu.Display.Boot(); err != nil {
		return err
	}

	if err := cpu.Keyboard.Boot(); err != nil {
		return err
	}

	if err := cpu.Buzzer.Boot(); err != nil {
		return err
	}

	cpu.isBooted = true

	return nil
}

// LoadProgram loads the program into memory and sets the PC to the start-of-program address
func (cpu *Cpu) LoadProgram(program []byte) error {
	cpu.Reset()
	return cpu.Memory.LoadProgram(program)
}

func (cpu *Cpu) Reset() {
	cpu.V = [16]byte{}
	cpu.I = 0
	cpu.Dt = 0
	cpu.St = 0
	cpu.Pc = startOfProgram
	cpu.Sp = 0
	cpu.Stack = [16]uint16{}

	cpu.frames = 0
	cpu.cycles = 0
	cpu.waitingForKey = false
	cpu.lastError = nil

	cpu.clearScreen()
	cpu.Display.Render(cpu.screen, cpu.ScreenSettings)
}

// Loop sets the speed an starts the loop
func (cpu *Cpu) LoopAtSpeed(speedInHz uint) error {
	cpu.SetSpeedInHz(speedInHz)
	return cpu.Loop()
}

// Loop starts the loop at the current speed
func (cpu *Cpu) Loop() error {
	if !cpu.isBooted {
		return ErrCpuIsNotBooted
	}

	if cpu.lastError != nil {
		return cpu.lastError
	}

	var last time.Time

	for {
		if done, err := cpu.runNextCycle(); err != nil {
			return err
		} else if done {
			return nil
		}

		// Prevent the CPU from running faster than expected
		time.Sleep(max(cpu.step-time.Since(last), 0))
		last = time.Now()
	}
}

// LoopOnce runs a single cycle bypassing the pause state
func (cpu *Cpu) LoopOnce() error {
	if !cpu.isBooted {
		return ErrCpuIsNotBooted
	}

	if cpu.lastError != nil {
		return cpu.lastError
	}

	prev := cpu.isPaused
	cpu.isPaused = false
	defer func(cpu *Cpu, prev bool) {
		cpu.isPaused = prev
	}(cpu, prev)

	if _, err := cpu.runNextCycle(); err != nil {
		return err
	}

	return nil
}

func (cpu *Cpu) runNextCycle() (bool, error) {
	cpu.runBeforeFrameHooks()

	if cpu.isPaused {
		return false, nil
	}

	if cpu.waitingForKey {
		if k, pressed := cpu.Keyboard.GetPressed(); pressed {
			cpu.V[cpu.keyDstRegister] = k
			cpu.waitingForKey = false
		}
		for cpu.Keyboard.IsPressed(cpu.V[cpu.keyDstRegister]) {

		}
	} else {
		// for i := 0; i < int(cpu.CyclesPerFrame); i++ {
		cpu.runBeforeCycleHooks()
		if err := cpu.executeNextInstruction(); err != nil {
			cpu.runErrorHooks()
			cpu.lastError = err
			return false, err
		}
		cpu.cycles++
		cpu.runAfterCycleHooks()

		if cpu.Pc >= MEMORY_SIZE {
			return true, nil
		}
		// }
	}

	cpu.Dt = min(cpu.Dt-1, 0)
	cpu.St = min(cpu.St-1, 0)

	if cpu.St > 0 {
		cpu.Buzzer.Play()
	}

	if cpu.cycles%cpu.CyclesPerFrame == 0 {
		// if cpu.isScreenDirty {
		cpu.isScreenDirty = false
		if err := cpu.Display.Render(cpu.screen, cpu.ScreenSettings); err != nil {
			cpu.runErrorHooks()
			cpu.lastError = err
			return false, err
		}
		// }

		cpu.frames++
	}

	cpu.runAfterFrameHooks()

	return false, nil
}

func (cpu *Cpu) executeNextInstruction() error {
	var opCode uint16
	opCode |= uint16(cpu.Memory[cpu.Pc+0]) << 8
	opCode |= uint16(cpu.Memory[cpu.Pc+1]) << 0
	cpu.Pc += 2

	return cpu.executeInstruction(opCode)
}

func bool2byte(b bool) byte {
	if b {
		return 1
	}

	return 0
}
