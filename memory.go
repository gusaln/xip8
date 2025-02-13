package xip8

import (
	"errors"
	"fmt"
	"strings"
)

var ErrProgramDoesNotFitIntoMemory = errors.New("the program does not fit into memory")

const startOfProgram = 0x200
const startOfEtiProgram = 0x600

const MEMORY_SIZE = 4096

type Memory [MEMORY_SIZE]byte

func newEmptyMemory() *Memory {
	m := Memory([MEMORY_SIZE]byte{})
	return &m
}

// NewMemory creates an empty memory of 4096 bytes
func NewMemory() *Memory {
	m := newEmptyMemory()

	return m
}

func (mem Memory) Clone() *Memory {
	m := NewMemory()

	copy(m[:], mem[:])

	return m
}

func (mem Memory) String() string {
	sb := strings.Builder{}

	sb.WriteString("[ ")
	for _, b := range mem[:startOfProgram] {
		sb.WriteString(fmt.Sprintf("%X ", b))
	}
	sb.WriteString("]\n")
	sb.WriteString("[ ")
	for _, b := range mem[startOfProgram:] {
		sb.WriteString(fmt.Sprintf("%X ", b))
	}
	sb.WriteString("]")

	return sb.String()
}

func (mem Memory) IsEqual(other Memory) bool {
	yes := true
	for i, b := range mem {
		if b != other[i] {
			yes = false
			break
		}
	}

	return yes
}

// LoadProgram loads the program at the appropriate location
func (mem *Memory) LoadProgram(program []byte) error {
	loadCharactersInto(mem)

	if len(program) > MEMORY_SIZE-startOfProgram {
		return ErrProgramDoesNotFitIntoMemory
	}

	copy(mem[startOfProgram:], program)

	return nil
}

func loadCharactersInto(mem *Memory) {
	copy(mem[:], []byte{
		// 0
		0xF0, 0x90, 0x90, 0x90, 0xF0,
		// 1
		0x20, 0x60, 0x20, 0x20, 0x70,
		// 2
		0xF0, 0x10, 0xF0, 0x80, 0xF0,
		// 3
		0xF0, 0x10, 0xF0, 0x10, 0xF0,
		// 4
		0x90, 0x90, 0xF0, 0x10, 0x10,
		// 5
		0xF0, 0x80, 0xF0, 0x10, 0xF0,
		// 6
		0xF0, 0x80, 0xF0, 0x90, 0xF0,
		// 7
		0xF0, 0x10, 0x20, 0x40, 0x40,
		// 8
		0xF0, 0x90, 0xF0, 0x90, 0xF0,
		// 9
		0xF0, 0x90, 0xF0, 0x10, 0xF0,
		// A
		0xF0, 0x90, 0xF0, 0x90, 0x90,
		// B
		0xE0, 0x90, 0xE0, 0x90, 0xE0,
		// C
		0xF0, 0x80, 0x80, 0x80, 0xF0,
		// D
		0xE0, 0x90, 0x90, 0x90, 0xE0,
		// E
		0xF0, 0x80, 0xF0, 0x80, 0xF0,
		// F
		0xF0, 0x80, 0xF0, 0x80, 0x80})
}
