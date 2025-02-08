package xip8_test

import (
	"testing"

	"github.com/guslan/xip8"
)

func runNCycles(cpu *xip8.Cpu, program []byte, n int) error {
	cpu.CyclesPerFrame = 1

	if err := cpu.LoadProgram(program); err != nil {
		return err
	}

	if err := cpu.Boot(); err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		err := cpu.SingleFrame()
		if err != nil {
			return err
		}
	}

	return nil
}

func expectVxEqVy(t *testing.T, cpu *xip8.Cpu, x, y byte) {
	if cpu.V[x] != cpu.V[y] {
		t.Fatalf(`cpu.V[%x] = %x and cpu.V[%x] = %x, expected them to be equal`, x, cpu.V[x], y, cpu.V[y])
	}
}

func assertVxEq(t *testing.T, msg string, cpu *xip8.Cpu, x, kk byte) {
	if cpu.V[x] != kk {
		t.Fatalf(`%s: cpu.V[%x] = %x, expected %x`, msg, x, cpu.V[x], kk)
	}
}

// TestProgramLoading loads a program that jumps to the las address to exit immediately
func TestProgramLoading(t *testing.T) {
	mem := xip8.NewMemory()
	kb := xip8.NewInMemoryKeyboard()
	b := xip8.NewDummyBuzzer()
	d := xip8.DummyDisplay{}

	cpu := xip8.NewCpu(mem, xip8.SmallScreen, d, kb, b)

	program := []byte{
		// move to the last address
		0x1F, 0xFE,
	}
	if err := runNCycles(cpu, program, 2); err != nil {
		t.Fatalf(`Loop() returned an error %v`, err)
	}

	expectedPc := uint16(4096)
	if cpu.Pc != expectedPc {
		t.Fatalf(`cpu.Pc = %d, expected for %d`, cpu.Pc, expectedPc)
	}
}

// TestConstantSetInstructions
func TestConstantSetInstructions(t *testing.T) {
	mem := xip8.NewMemory()
	kb := xip8.NewInMemoryKeyboard()
	b := xip8.NewDummyBuzzer()
	d := xip8.DummyDisplay{}

	cpu := xip8.NewCpu(mem, xip8.SmallScreen, d, kb, b)
	cpu.CyclesPerFrame = 1

	program := []byte{
		// set v0 to 128
		0x60, 128,
		// set v1 to 16
		0x61, 16,
		// set v2 to 1
		0x62, 1,
		// add to v2 4
		0x72, 4,
		// move to the last address
		0x1F, 0xFE,
	}
	if err := runNCycles(cpu, program, 5); err != nil {
		t.Fatalf(`Loop() returned an error %v`, err)
	}
	var want int
	var get int

	want = 128
	get = int(cpu.V[0])
	if get != want {
		t.Fatalf(`cpu.V[0] = %x, expected for %x`, get, want)
	}

	want = 16
	get = int(cpu.V[1])
	if get != want {
		t.Fatalf(`cpu.V[1] = %x, expected for %x`, get, want)
	}

	want = 5
	get = int(cpu.V[2])
	if get != want {
		t.Fatalf(`cpu.V[2] = %x, expected for %x`, get, want)
	}
}

// TestSimpleSkips loads a program that jumps to the las address to exit immediately
func TestSimpleSkips(t *testing.T) {
	mem := xip8.NewMemory()
	kb := xip8.NewInMemoryKeyboard()
	b := xip8.NewDummyBuzzer()
	d := xip8.DummyDisplay{}

	cpu := xip8.NewCpu(mem, xip8.SmallScreen, d, kb, b)

	program := []byte{
		// set v0 to 128
		0x60, 128,
		// set v1 to 16
		0x61, 16,
		// set v2 to 128
		0x62, 128,

		// if v0 == 128, do not set v3 to 1
		0x30, 128,
		0x63, 1,

		// if v0 == 16, do not set vA to 1
		0x30, 16,
		0x6A, 1,

		// if v0 != 128, do not set v4 to 1
		0x40, 128,
		0x64, 1,

		// if v0 != 16, do not set v3 to 1
		0x40, 16,
		0x6B, 1,

		// if v0 == v1, do not set v5 to 1
		0x50, 0x10,
		0x65, 1,

		// if v0 == v2, do not set v6 to 1
		0x50, 0x20,
		0x66, 1,

		// move to the last address
		0x1F, 0xFE,
	}
	if err := runNCycles(cpu, program, 12); err != nil {
		t.Fatalf(`Loop() returned an error %v`, err)
	}

	assertVxEq(t, "SE Vx kk true", cpu, 0x3, 0x0)
	assertVxEq(t, "SE Vx kk false", cpu, 0xA, 0x1)
	assertVxEq(t, "SNE Vx kk true", cpu, 0xB, 0x0)
	assertVxEq(t, "SNE Vx kk false", cpu, 0x4, 0x1)
	assertVxEq(t, "SE Vx V2 true", cpu, 0x6, 0x0)
	assertVxEq(t, "SE Vx V1 false", cpu, 0x5, 0x1)
}
