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
	noTermPtr := flag.Bool("noterm", false, "turn off the terminal display of the emulator")
	debugPtr := flag.Bool("debug", false, "turn on the debugger for the emulator")
	portPtr := flag.Int("port", 9999, "specify the port of the debugger")

	flag.Parse()

	mem := xip8.NewMemory()
	kb := xip8.NewTerminalKeyboard()
	b := xip8.NewDummyBuzzer()
	var d xip8.Display
	if *noTermPtr {
		d = xip8.NewDefaultInMemoryDisplay()
	} else {
		d = xip8.NewDefaultTerminalDisplay()
	}

	cpu := xip8.NewCpu(mem, d, kb, b)
	if flag.NArg() < 1 {
		log.Fatalln("must provide the path to a rom as an argument")
	}

	var speed uint = 30
	if *debugPtr {
		speed = 1
		deb := xip8.NewHttpDebugger(cpu)
		go func(deb *xip8.HttpDebugger, port int) {
			log.Println("server listening on port ", port)

			if err := deb.Listen(port); err != nil {
				log.Fatalln(err)
			}
		}(deb, *portPtr)
	}

	program, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}
	cpu.LoadProgram(program)

	if err := cpu.Boot(); err != nil {
		log.Fatalln(err)
	}

	if err := cpu.LoopAtSpeed(speed); err != nil {
		log.Fatalln(err)
	}
}
