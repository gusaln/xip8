package xip8

import "crypto/rand"

var x uint16
var y uint16
var n byte
var kk byte
var nnn uint16

func (cpu *Cpu) executeInstruction(opCode uint16) error {
	x = (opCode & 0x0F00) >> 8
	y = (opCode & 0x00F0) >> 4
	n = byte(opCode & 0x000F)
	kk = byte(opCode & 0x00FF)
	nnn = (opCode & 0x0FFF)

	switch opCode & 0xF000 {
	case 0x0000:
		switch opCode {
		case 0x00E0:
			// CLS :: Clear the display.
			cpu.clearScreen()

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
		cpu.Pc = nnn

	case 0x2000:
		// CALL addr :: Call subroutine at nnn.
		if cpu.Sp > 15 {
			return ErrStackOverflow
		}
		cpu.Stack[cpu.Sp] = cpu.Pc
		cpu.Sp++

		cpu.Pc = nnn

	case 0x3000:
		// SE Vx, byte :: Skip next instruction if Vx = kk.
		// kk := byte(opCode & 0x00FF)
		if cpu.V[x] == kk {
			cpu.Pc += 2
		}

	case 0x4000:
		// SNE Vx, byte :: Skip next instruction if Vx != kk.
		// kk := byte(opCode & 0x00FF)
		if cpu.V[x] != kk {
			cpu.Pc += 2
		}

	case 0x5000:
		// SE Vx, Vy :: Skip next instruction if Vx = Vy.
		if cpu.V[x] == cpu.V[y] {
			cpu.Pc += 2
		}

	case 0x6000:
		// LD Vx, byte :: Set Vx = kk.
		// kk := byte(opCode & 0x00FF)
		cpu.V[x] = kk

	case 0x7000:
		// ADD Vx, byte :: Set Vx = Vx + kk.
		// kk := byte(opCode & 0x00FF)
		cpu.V[x] = cpu.V[x] + kk

	case 0x8000:
		// Inter-register operations

		switch opCode & 0x000F {
		case 0x0000:
			// LD Vx, Vy :: Set Vx = Vy.
			cpu.V[x] = cpu.V[y]

		case 0x0001:
			// OR Vx, Vy :: Set Vx = Vx OR Vy.
			if (cpu.quirks & FlagQuirkVfReset) > 0 {
				cpu.V[0xF] = 0
			}
			cpu.V[x] |= cpu.V[y]

		case 0x0002:
			// AND Vx, Vy :: Set Vx = Vx AND Vy.
			if (cpu.quirks & FlagQuirkVfReset) > 0 {
				cpu.V[0xF] = 0
			}
			cpu.V[x] &= cpu.V[y]

		case 0x0003:
			// XOR Vx, Vy :: Set Vx = Vx XOR Vy.
			if (cpu.quirks & FlagQuirkVfReset) > 0 {
				cpu.V[0xF] = 0
			}
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
			if (cpu.quirks & FlagQuirkShiftWithVy) > 0 {
				cpu.V[x] = cpu.V[y]
			}
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
			if (cpu.quirks & FlagQuirkShiftWithVy) > 0 {
				cpu.V[x] = cpu.V[y]
			}
			carry := (cpu.V[x] & 0b10000000) >> 7
			cpu.V[x] = cpu.V[x] << 1
			cpu.V[0xF] = carry
		}

	case 0x9000:
		// SNE Vx, Vy :: Skip next instruction if Vx != Vy.
		if cpu.V[x] != cpu.V[y] {
			cpu.Pc += 2
		}

	case 0xA000:
		// LD I, addr :: Set I = nnn.
		cpu.I = nnn

	case 0xB000:
		// JP V0, addr :: Jump to location nnn + V0 or xnn + Vx .
		if (cpu.quirks & FlagQuirkJumpUsesVx) > 0 {
			cpu.Pc = uint16(cpu.V[x]) + nnn
		} else {
			cpu.Pc = uint16(cpu.V[0]) + nnn
		}

	case 0xC000:
		// RND Vx, byte :: Set Vx = random byte AND kk.
		// kk := byte(opCode & 0x00FF)

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
		cpu.V[0xF] = 0
		for i := byte(0); i < n; i++ {
			cpu.V[0xF] |= bool2byte(cpu.displayToScreen(cpu.V[x], cpu.V[y]+i, cpu.Memory[cpu.I+uint16(i)]))
		}

	case 0xE000:
		// Skip if ...

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

		switch opCode & 0x00FF {
		case 0x0007:
			// LD Vx, DT :: Set Vx = delay timer value.
			cpu.V[x] = cpu.Dt
		case 0x000A:
			// LD Vx, K :: Wait for a key press, store the value of the key in Vx.
			cpu.waitingForKey = true
			cpu.keyDstRegister = x

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
			if (cpu.quirks & FlagQuirkMemoryMovesIndex) > 0 {
				cpu.I += x + 1
			}
		case 0x0065:
			// LD Vx, [I] :: Read registers V0 through Vx from memory starting at location I.
			for i := uint16(0); i <= x; i++ {
				cpu.V[i] = cpu.Memory[cpu.I+i]
			}
			if (cpu.quirks & FlagQuirkMemoryMovesIndex) > 0 {
				cpu.I += x + 1
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
