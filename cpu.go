package xip8

import (
	"crypto/rand"
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

	speedInHz  uint
	timerEvery uint
	cycles     uint
	frames     uint

	Display        Display
	CyclesPerFrame uint
	Keyboard       Keyboard
	Buzzer         Buzzer

	MachineRoutineInterpreter MachineRoutineInterpreter

	isBooted  bool
	isPaused  bool
	lastError error

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

	renderCh chan interface{}
}

var speedInHzDefault uint = 30
var cyclesPerFrameDefault uint = 30

func NewCpu(memory *Memory, display Display, keyboard Keyboard, buzzer Buzzer) *Cpu {
	return &Cpu{
		Memory: memory,

		V:     [16]byte{},
		I:     0,
		Dt:    0,
		St:    0,
		Pc:    0,
		Sp:    0,
		Stack: [16]uint16{},

		speedInHz: speedInHzDefault,

		Display:        display,
		CyclesPerFrame: cyclesPerFrameDefault,
		Keyboard:       keyboard,
		Buzzer:         buzzer,

		MachineRoutineInterpreter: nil,

		isBooted:  false,
		isPaused:  false,
		lastError: nil,

		beforeFrameHooks: make([]Hook, 0),
		beforeCycleHooks: make([]Hook, 0),
		afterCycleHooks:  make([]Hook, 0),
		afterFrameHooks:  make([]Hook, 0),
		errorHooks:       make([]Hook, 0),

		renderCh: make(chan interface{}),
	}
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
}

func (cpu Cpu) Cycles() uint {
	return cpu.cycles
}

func (cpu Cpu) Frames() uint {
	return cpu.frames
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

	go func(cpu *Cpu) {
		for range cpu.renderCh {
			cpu.Display.Render()
		}
	}(cpu)

	return nil
}

// LoadProgram loads the program into memory and sets the PC to the start-of-program address
func (cpu *Cpu) LoadProgram(program []byte) error {
	cpu.Reset()
	return cpu.Memory.LoadProgram(program)
}

func (cpu *Cpu) Reset() {
	cpu.Pc = startOfProgram
	cpu.frames = 0
	cpu.cycles = 0
	cpu.lastError = nil

	cpu.Display.Clear()
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

	step := time.Second / time.Duration(cpu.speedInHz)
	// timerStep := time.Second / time.Duration(60)
	for {
		if cpu.isPaused {
			time.Sleep(max(step-time.Since(last), 0))
			last = time.Now()
			continue
		}

		if done, err := cpu.nextFrame(); err != nil {
			return err
		} else if done {
			return nil
		}

		time.Sleep(max(step-time.Since(last), 0))
		last = time.Now()
	}
}

func (cpu *Cpu) SingleFrame() error {
	if !cpu.isBooted {
		return ErrCpuIsNotBooted
	}

	if cpu.lastError != nil {
		return cpu.lastError
	}

	if _, err := cpu.nextFrame(); err != nil {
		return err
	}

	return nil
}

func (cpu *Cpu) nextFrame() (bool, error) {
	cpu.runBeforeFrameHooks()

	for i := 0; i < int(cpu.CyclesPerFrame); i++ {
		cpu.runBeforeCycleHooks()
		if err := cpu.nextCycle(); err != nil {
			cpu.runErrorHooks()
			cpu.lastError = err
			return false, err
		}
		cpu.cycles++
		cpu.runAfterCycleHooks()

		if cpu.Pc >= MEMORY_SIZE {
			return true, nil
		}
	}

	cpu.Dt = min(cpu.Dt-1, 0)
	cpu.St = min(cpu.St-1, 0)

	if cpu.St > 0 {
		cpu.Buzzer.Play()
	}

	cpu.renderCh <- nil
	cpu.frames++

	cpu.runAfterFrameHooks()

	return false, nil
}

func (cpu *Cpu) nextCycle() error {
	var opCode uint16
	opCode |= uint16(cpu.Memory[cpu.Pc+0]) << 8
	opCode |= uint16(cpu.Memory[cpu.Pc+1]) << 0
	cpu.Pc += 2

	return cpu.execute(opCode)
}

func (cpu *Cpu) execute(opCode uint16) error {
	switch opCode & 0xF000 {
	case 0x0000:
		switch opCode {
		case 0x00E0:
			// CLS :: Clear the display.
			cpu.Display.Clear()

		case 0x00EE:
			// RET :: Return from a subroutine.
			if cpu.Sp == 0 {
				return ErrStackUnderflow
			}
			cpu.Sp--
			cpu.Pc = cpu.Stack[cpu.Sp]

		default:
			// SYS :: Jump to a machine code routine at nnn.
			// Jump to a machine code routine at nnn.
			// This instruction is only used on the old computers on which Chip-8 was originally implemented. It is ignored by modern interpreters.
			if cpu.MachineRoutineInterpreter != nil {
				cpu.MachineRoutineInterpreter(opCode, cpu)
			}
		}

	case 0x1000:
		// JP addr :: Jump to location nnn.
		cpu.Pc = uint16(opCode & 0x0FFF)

	case 0x2000:
		// CALL addr :: Call subroutine at nnn.
		if cpu.Sp > 15 {
			return ErrStackOverflow
		}
		cpu.Stack[cpu.Sp] = cpu.Pc
		cpu.Sp++

		cpu.Pc = uint16(opCode & 0x0FFF)

	case 0x3000:
		// SE Vx, byte :: Skip next instruction if Vx = kk.
		x := (opCode & 0x0F00) >> 8
		kk := byte(opCode & 0x00FF)
		if cpu.V[x] == kk {
			cpu.Pc += 2
		}

	case 0x4000:
		// SNE Vx, byte :: Skip next instruction if Vx != kk.
		x := (opCode & 0x0F00) >> 8
		kk := byte(opCode & 0x00FF)
		if cpu.V[x] != kk {
			cpu.Pc += 2
		}

	case 0x5000:
		// SE Vx, Vy :: Skip next instruction if Vx = Vy.
		x := (opCode & 0x0F00) >> 8
		y := (opCode & 0x00F0) >> 4
		if cpu.V[x] == cpu.V[y] {
			cpu.Pc += 2
		}

	case 0x6000:
		// LD Vx, byte :: Set Vx = kk.
		x := (opCode & 0x0F00) >> 8
		kk := byte(opCode & 0x00FF)
		cpu.V[x] = kk

	case 0x7000:
		// ADD Vx, byte :: Set Vx = Vx + kk.
		x := (opCode & 0x0F00) >> 8
		kk := byte(opCode & 0x00FF)
		cpu.V[x] = cpu.V[x] + kk

	case 0x8000:
		// Inter-register operations

		x := (opCode & 0x0F00) >> 8
		y := (opCode & 0x00F0) >> 4

		switch opCode & 0x000F {
		case 0x0000:
			// LD Vx, Vy :: Set Vx = Vy.
			cpu.V[x] = cpu.V[y]

		case 0x0001:
			// OR Vx, Vy :: Set Vx = Vx OR Vy.
			cpu.V[x] |= cpu.V[y]

		case 0x0002:
			// AND Vx, Vy :: Set Vx = Vx AND Vy.
			cpu.V[x] &= cpu.V[y]

		case 0x0003:
			// XOR Vx, Vy :: Set Vx = Vx XOR Vy.
			cpu.V[x] ^= cpu.V[y]

		case 0x0004:
			// ADD Vx, Vy :: Set Vx = Vx + Vy, set VF = carry.
			r := uint16(cpu.V[x]) + uint16(cpu.V[y])
			cpu.V[x] = byte(r & 0x00FF)
			cpu.V[0xF] = byte(r >> 8)

		case 0x0005:
			// SUB Vx, Vy :: Set Vx = Vx - Vy, set VF = NOT borrow.
			carry := cpu.V[x] >= cpu.V[y]
			cpu.V[x] = cpu.V[x] - cpu.V[y]
			cpu.V[0xF] = bool2byte(carry)

		case 0x0006:
			// SHR Vx {, Vy} :: Set Vx = Vx SHR 1.
			carry := cpu.V[x] & 0b00000001
			cpu.V[x] = cpu.V[x] >> 1
			cpu.V[0xF] = carry

		case 0x0007:
			// SUBN Vx, Vy :: Set Vx = Vy - Vx, set VF = NOT borrow.
			carry := cpu.V[y] >= cpu.V[x]
			cpu.V[x] = cpu.V[y] - cpu.V[x]
			cpu.V[0xF] = bool2byte(carry)

		case 0x000E:
			// SHL Vx {, Vy} :: Set Vx = Vx SHL 1.
			carry := (cpu.V[x] & 0b10000000) >> 7
			cpu.V[x] = cpu.V[x] << 1
			cpu.V[0xF] = carry
		}

	case 0x9000:
		// SNE Vx, Vy :: Skip next instruction if Vx != Vy.
		x := (opCode & 0x0F00) >> 8
		y := (opCode & 0x00F0) >> 4
		if cpu.V[x] != cpu.V[y] {
			cpu.Pc += 2
		}

	case 0xA000:
		// LD I, addr :: Set I = nnn.
		cpu.I = opCode & 0x0FFF

	case 0xB000:
		// JP V0, addr :: Jump to location nnn + V0.
		cpu.Pc = uint16(cpu.V[0]) + (opCode & 0x0FFF)

	case 0xC000:
		// RND Vx, byte :: Set Vx = random byte AND kk.
		x := (opCode & 0x0F00) >> 8
		kk := byte(opCode & 0x00FF)

		buff := [1]byte{}
		n, err := rand.Read(buff[:])
		if n != 1 || err != nil {
			return err
		}

		cpu.V[x] = buff[0] & kk

	case 0xD000:
		// DRW Vx, Vy, nibble :: Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
		// The interpreter reads n bytes from memory, starting at the address stored in I. These bytes are then
		// displayed as sprites on screen at coordinates (Vx, Vy). Sprites are XORed onto the existing screen.
		// If this causes any pixels to be erased, VF is set to 1, otherwise it is set to 0. If the sprite is
		// positioned so part of it is outside the coordinates of the display, it wraps around to the opposite side of
		// the screen.
		x := byte((opCode & 0x0F00) >> 8)
		y := byte((opCode & 0x00F0) >> 4)
		n := byte(opCode & 0x000F)
		for i := byte(0); i < n; i++ {
			cpu.V[0xF] = bool2byte(cpu.Display.Display(cpu.V[x], cpu.V[y]+i, cpu.Memory[cpu.I+uint16(i)]))
		}

	case 0xE000:
		// Skip if ...
		x := (opCode & 0x0F00) >> 8

		switch opCode & 0x00FF {
		case 0x009E:
			// SKP Vx :: Skip next instruction if key with the value of Vx is pressed.
			if cpu.Keyboard.IsPressed(cpu.V[x]) {
				cpu.Pc += 2
			}
		case 0x00A1:
			// SKNP Vx :: Skip next instruction if key with the value of Vx is not pressed.
			if !cpu.Keyboard.IsPressed(cpu.V[x]) {
				cpu.Pc += 2
			}
		}

	case 0xF000:
		// other operations
		x := (opCode & 0x0F00) >> 8

		switch opCode & 0x00FF {
		case 0x0007:
			// LD Vx, DT :: Set Vx = delay timer value.
			cpu.V[x] = cpu.Dt
		case 0x000A:
			// LD Vx, K :: Wait for a key press, store the value of the key in Vx.
			k, err := cpu.Keyboard.WaitForKey()
			if err != nil {
				return err
			}
			cpu.V[x] = k

		case 0x0015:
			// LD DT, Vx :: Set delay timer = Vx.
			cpu.Dt = cpu.V[x]
		case 0x0018:
			// LD ST, Vx :: Set sound timer = Vx.
			cpu.St = cpu.V[x]
		case 0x001E:
			// ADD I, Vx :: Set I = I + Vx.
			cpu.I = cpu.I + uint16(cpu.V[x])
		case 0x0029:
			// LD F, Vx :: Set I = location of sprite for digit Vx.
			cpu.I = uint16(cpu.V[x]) * 5
		case 0x0033:
			// LD B, Vx :: Store BCD representation of Vx in memory locations I, I+1, and I+2.
			top := x / 100
			middle := (x - top) / 10
			bottom := x - top - middle
			cpu.Memory[cpu.I+0] = byte(top)
			cpu.Memory[cpu.I+1] = byte(middle)
			cpu.Memory[cpu.I+2] = byte(bottom)
		case 0x0055:
			// LD [I], Vx :: Store registers V0 through Vx in memory starting at location I.
			for i := uint16(0); i <= x; i++ {
				cpu.Memory[cpu.I+i] = cpu.V[i]
			}
		case 0x0065:
			// LD Vx, [I] :: Read registers V0 through Vx from memory starting at location I.
			for i := uint16(0); i <= x; i++ {
				cpu.V[i] = cpu.Memory[cpu.I+i]
			}
		default:
			return ErrOpCodeUnknown{
				OpCode: opCode,
				Pc:     cpu.Pc,
			}
		}

	default:
		return ErrOpCodeUnknown{
			OpCode: opCode,
			Pc:     cpu.Pc,
		}
	}

	return nil
}

func bool2byte(b bool) byte {
	if b {
		return 1
	}

	return 0
}
