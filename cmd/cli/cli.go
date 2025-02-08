/*
 *   Copyright (c) 2024 Gustavo Lopez <git.gustavolopez.xyz@gmail.com>
 *   All rights reserved.
 */
package main

import (
	"flag"
	"log"
	"os"

	xip8 "github.com/guslan/xip8"
)

func main() {
	speedPtr := flag.Uint("speed", 30, "specify the speed of the chip in Hz (default: 30)")

	flag.Parse()

	mem := xip8.NewMemory()
	kb := xip8.NewTerminalKeyboard()
	b := xip8.NewDummyBuzzer()
	d := xip8.NewTerminalDisplay()

	cpu := xip8.NewCpu(mem, xip8.SmallScreen, d, kb, b)
	if flag.NArg() < 1 {
		log.Fatalln("must provide the path to a rom as an argument")
	}

	program, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}
	cpu.LoadProgram(program)

	if err := cpu.Boot(); err != nil {
		log.Fatalln(err)
	}

	if err := cpu.LoopAtSpeed(*speedPtr); err != nil {
		log.Fatalln(err)
	}
}
