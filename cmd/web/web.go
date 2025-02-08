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
	"github.com/guslan/xip8/web"
)

func main() {
	port := flag.Int("port", 9999, "The port of the server (default = 9999)")
	speed := flag.Int("speed", 1, "Speed in cycles per second (default = 1)")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalln("must provide the path to a rom as an argument")
	}

	// var speed uint = 30

	program, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}

	mem := xip8.NewMemory()
	server := web.NewServer(mem, func(config *web.ServerConfig) {
		config.UseDebugger = true
	})

	server.Speed(*speed)
	server.LoadProgram(program)
	if err := server.Listen(*port); err != nil {
		log.Fatalln(err)
	}
}
