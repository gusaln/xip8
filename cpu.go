package xip8

import (
	"crypto/rand"
	"fmt"
	"time"
)

var ErrStackUnderflow = fmt.Errorf("stack underflow: try to pop an empty stack")
var ErrStackOverflow = fmt.Errorf("stack overflow: try to push to a full stack")

const startOfProgram = 0x200
const startOfEtiProgram = 0x600

const ROM_MEMORY_SIZE = 4096

// Chip-8 CPU
type Cpu struct {
	Memory [ROM_MEMORY_SIZE]byte
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

	Display  Display
	Keyboard Keyboard
}

func NewCpu(display Display, keyboard Keyboard) *Cpu {
	return &Cpu{
		Memory: [4096]byte{
			// 0
			0xF0,
			0x90,
			0x90,
			0x90,
			0xF0,
			// 1
			0x20,
			0x60,
			0x20,
			0x20,
			0x70,
			// 2
			0xF0,
			0x10,
			0xF0,
			0x80,
			0xF0,
			// 3
			0xF0,
			0x10,
			0xF0,
			0x10,
			0xF0,
			// 4
			0x90,
			0x90,
			0xF0,
			0x10,
			0x10,
			// 5
			0xF0,
			0x80,
			0xF0,
			0x10,
			0xF0,
			// 6
			0xF0,
			0x80,
			0xF0,
			0x90,
			0xF0,
			// 7
			0xF0,
			0x10,
			0x20,
			0x40,
			0x40,
			// 8
			0xF0,
			0x90,
			0xF0,
			0x90,
			0xF0,
			// 9
			0xF0,
			0x90,
			0xF0,
			0x10,
			0xF0,
			// A
			0xF0,
			0x90,
			0xF0,
			0x90,
			0x90,
			// B
			0xE0,
			0x90,
			0xE0,
			0x90,
			0xE0,
			// C
			0xF0,
			0x80,
			0x80,
			0x80,
			0xF0,
			// D
			0xE0,
			0x90,
			0x90,
			0x90,
			0xE0,
			// E
			0xF0,
			0x80,
			0xF0,
			0x80,
			0xF0,
			// F
			0xF0,
			0x80,
			0xF0,
			0x80,
			0x80,
		},
		V:        [16]byte{},
		I:        0,
		Dt:       0,
		St:       0,
		Pc:       0,
		Sp:       0,
		Stack:    [16]uint16{},
		Display:  display,
		Keyboard: keyboard,
	}
}

func (cpu *Cpu) LoopAtSpeed(speedInHz int) error {
	var last, timerLast time.Time

	step := time.Second / time.Duration(speedInHz)
	timerStep := time.Second / time.Duration(60)
	for {
		last = time.Now()
		if err := cpu.RunNext(); err != nil {
			return err
		}

		if time.Since(timerLast) > timerStep {
			cpu.Dt = min(cpu.Dt-1, 0)
			cpu.St = min(cpu.St-1, 0)
			timerLast = time.Now()
		}

		if cpu.Pc*2 >= ROM_MEMORY_SIZE {
			return nil
		}

		time.Sleep(max(step-time.Since(last), 0))
	}
}

func (cpu *Cpu) Loop() error {
	return cpu.LoopAtSpeed(60)
}

func (cpu Cpu) IsSoundTimerActive() bool {
	return cpu.St > 0
}

func (cpu Cpu) IsDelayTimerActive() bool {
	return cpu.Dt > 0
}

func (cpu *Cpu) RunNext() error {
	var opCode uint16
	opCode |= uint16(cpu.Memory[cpu.Pc*2+0]) << 8
	opCode |= uint16(cpu.Memory[cpu.Pc*2+1]) << 0

	switch opCode & 0xF000 {
	case 0x0000:
		if opCode == 0x00E0 {
			// CLS :: Clear the display.
			cpu.Display.Clear()
			cpu.Pc++
		} else if opCode == 0x00EE {
			// RET :: Return from a subroutine.
			if cpu.Sp == 0 {
				return ErrStackUnderflow
			}
			cpu.Sp--
			cpu.Pc = cpu.Stack[cpu.Sp]

		} else {
			// SYS :: Jump to a machine code routine at nnn.
			// Jump to a machine code routine at nnn.
			// This instruction is only used on the old computers on which Chip-8 was originally implemented. It is ignored by modern interpreters.
			cpu.Pc++
		}

	case 0x1000:
		// JP addr :: Jump to location nnn.
		cpu.Pc = uint16(opCode & 0x0111)

	case 0x2000:
		// CALL addr :: Call subroutine at nnn.
		if cpu.Sp > 15 {
			return ErrStackOverflow
		}
		cpu.Stack[cpu.Sp] = cpu.Pc
		cpu.Sp++

		cpu.Pc = uint16(opCode & 0x0111)

	case 0x3000:
		// SE Vx, byte :: Skip next instruction if Vx = kk.
		register := opCode & 0x0100
		kk := byte(opCode & 0x00FF)
		if cpu.V[register] == kk {
			cpu.Pc += 2
		} else {
			cpu.Pc++
		}

	case 0x4000:
		// SNE Vx, byte :: Skip next instruction if Vx != kk.
		register := opCode & 0x0100
		kk := byte(opCode & 0x00FF)
		if cpu.V[register] != kk {
			cpu.Pc += 2
		} else {
			cpu.Pc++
		}

	case 0x5000:
		// SE Vx, Vy :: Skip next instruction if Vx = Vy.
		x := opCode & 0x0100
		y := opCode & 0x0010
		if cpu.V[x] == cpu.V[y] {
			cpu.Pc += 2
		} else {
			cpu.Pc++
		}

	case 0x6000:
		// LD Vx, byte :: Set Vx = kk.
		x := opCode & 0x0100
		kk := byte(opCode & 0x00FF)
		cpu.V[x] = kk
		cpu.Pc++

	case 0x7000:
		// ADD Vx, byte :: Set Vx = Vx + kk.
		x := opCode & 0x0100
		kk := byte(opCode & 0x00FF)
		cpu.V[x] = cpu.V[x] + kk
		cpu.Pc++

	case 0x8000:
		// Inter-register operations

		x := opCode & 0x0100
		y := opCode & 0x0010

		switch opCode & 0x000F {
		case 0x0000:
			// LD Vx, Vy :: Set Vx = Vy.
			cpu.V[x] = cpu.V[y]

		case 0x0001:
			// OR Vx, Vy :: Set Vx = Vx OR Vy.
			cpu.V[x] = cpu.V[x] | cpu.V[y]

		case 0x0002:
			// AND Vx, Vy :: Set Vx = Vx AND Vy.
			cpu.V[x] = cpu.V[x] & cpu.V[y]

		case 0x0003:
			// XOR Vx, Vy :: Set Vx = Vx XOR Vy.
			cpu.V[x] = cpu.V[x] ^ cpu.V[y]

		case 0x0004:
			// ADD Vx, Vy :: Set Vx = Vx + Vy, set VF = carry.
			r := uint16(cpu.V[x]) + uint16(cpu.V[y])
			cpu.V[x] = byte(r & 0x0011)
			cpu.V[0xF] = byte(r >> 8)

		case 0x0005:
			// SUB Vx, Vy :: Set Vx = Vx - Vy, set VF = NOT borrow.
			cpu.V[0xF] = byte(bool2int(cpu.V[x] > cpu.V[y]))
			cpu.V[x] = cpu.V[x] - cpu.V[y]

		case 0x0006:
			// SHR Vx {, Vy} :: Set Vx = Vx SHR 1.
			cpu.V[0xF] = cpu.V[x] & 0x01
			cpu.V[x] = cpu.V[x] >> 1

		case 0x0007:
			// SUBN Vx, Vy :: Set Vx = Vy - Vx, set VF = NOT borrow.
			cpu.V[0xF] = byte(bool2int(cpu.V[y] > cpu.V[x]))
			cpu.V[x] = cpu.V[y] - cpu.V[x]

		case 0x000E:
			// SHL Vx {, Vy} :: Set Vx = Vx SHL 1.
			cpu.V[0xF] = (cpu.V[x] & 0x40) >> 7
			cpu.V[x] = cpu.V[x] << 1
		}
		cpu.Pc++

	case 0x9000:
		// SNE Vx, Vy :: Skip next instruction if Vx != Vy.
		x := opCode & 0x0100
		y := opCode & 0x0010
		if cpu.V[x] != cpu.V[y] {
			cpu.Pc += 2
		} else {
			cpu.Pc++
		}

	case 0xA000:
		// LD I, addr :: Set I = nnn.
		cpu.I = opCode & 0x0FFF
		cpu.Pc++

	case 0xB000:
		// JP V0, addr :: Jump to location nnn + V0.
		cpu.Pc = uint16(cpu.V[0]) + (opCode & 0x0FFF)

	case 0xC000:
		// RND Vx, byte :: Set Vx = random byte AND kk.
		x := opCode & 0x0100
		kk := byte(opCode & 0x00FF)

		buff := [1]byte{}
		n, err := rand.Read(buff[:])
		if n != 1 || err != nil {
			return err
		}

		cpu.V[x] = buff[0] & kk
		cpu.Pc++

	case 0xD000:
		// DRW Vx, Vy, nibble :: Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
		// The interpreter reads n bytes from memory, starting at the address stored in I. These bytes are then displayed as sprites on screen at coordinates (Vx, Vy). Sprites are XORed onto the existing screen. If this causes any pixels to be erased, VF is set to 1, otherwise it is set to 0. If the sprite is positioned so part of it is outside the coordinates of the display, it wraps around to the opposite side of the screen.
		x := byte(opCode & 0x0100)
		y := byte(opCode & 0x0010)
		n := opCode & 0x000F
		collision := false
		for i := uint16(0); i <= n; i++ {
			collision = cpu.Display.Display(x, y, cpu.Memory[cpu.I+i]) || collision
		}
		cpu.V[0xF] = byte(bool2int(collision))
		cpu.Pc++

	case 0xE000:
		// Skip if ...
		x := opCode & 0x0100

		switch opCode & 0x00FF {
		case 0x009E:
			// SKP Vx :: Skip next instruction if key with the value of Vx is pressed.
			if cpu.Keyboard.IsPressed(cpu.V[x]) {
				cpu.Pc += 2
			} else {
				cpu.Pc++
			}
		case 0x00A1:
			// SKNP Vx :: Skip next instruction if key with the value of Vx is not pressed.
			if !cpu.Keyboard.IsPressed(cpu.V[x]) {
				cpu.Pc += 2
			} else {
				cpu.Pc++
			}
		}

	case 0xF000:
		// other operations
		x := opCode & 0x0100

		switch opCode & 0x00FF {
		case 0x0007:
			// LD Vx, DT :: Set Vx = delay timer value.
			cpu.V[x] = cpu.Dt
		case 0x000A:
			// LD Vx, K :: Wait for a key press, store the value of the key in Vx.
			// TODO
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
			// TODO
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
		}

		cpu.Pc++
	}

	return nil
}

func bool2int(b bool) int {
	if b {
		return 1
	}

	return 0
}
