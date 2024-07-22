package xip8

import (
	"fmt"
	"net/http"
	"strings"
)

type HttpDebugger struct {
	Cpu           *Cpu
	CurrentOpCode uint16
	Cycle         int

	SendEvery int
	send      chan Cpu
}

// NewHttpDebugger creates a new debugger
// This method will pause the cpu, register the hooks, set the execution cpu.CyclesPerFrame to 1
func NewHttpDebugger(cpu *Cpu) *HttpDebugger {
	deb := HttpDebugger{
		Cpu:           cpu,
		CurrentOpCode: 0,
		Cycle:         0,
		SendEvery:     1,
		send:          make(chan Cpu),
	}

	cpu.AddBeforeFrameHook(deb.beforeFrame)
	cpu.AddBeforeCycleHook(deb.beforeCycle)
	cpu.AddAfterCycleHook(deb.afterCycle)
	cpu.AddAfterFrameHook(deb.afterFrame)
	cpu.CyclesPerFrame = 1

	cpu.Stop()

	return &deb
}

func (d *HttpDebugger) Listen(port int) error {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		running := true
		for running {
			select {
			case cpu := <-d.send:
				fmt.Fprintf(w, "data: %s\n\n", d.formatAsEvent(cpu))
				w.(http.Flusher).Flush()

			case <-r.Context().Done():
				running = false
			}
		}
	})
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		d.Cpu.Start()
	})
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		d.Cpu.Stop()
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (d *HttpDebugger) beforeFrame(cpu *Cpu) {
}

func (d *HttpDebugger) beforeCycle(cpu *Cpu) {
	d.CurrentOpCode = uint16(cpu.Memory[cpu.Pc+0]) << 8
	d.CurrentOpCode |= uint16(cpu.Memory[cpu.Pc+1]) << 0
}

func (d *HttpDebugger) afterCycle(cpu *Cpu) {
	if d.Cpu.cycles%uint(d.SendEvery) == 0 {
		d.send <- *cpu
	}
}

func (d *HttpDebugger) afterFrame(cpu *Cpu) {
}

func (d HttpDebugger) formatAsEvent(cpu Cpu) string {
	sb := strings.Builder{}

	sb.WriteRune(rune(d.CurrentOpCode))
	sb.WriteRune(rune(cpu.Pc))
	for _, b := range cpu.V {
		sb.WriteByte(b)
	}
	sb.WriteRune(rune(cpu.I))
	sb.WriteByte(byte(cpu.Sp))
	for _, b := range cpu.Stack {
		sb.WriteRune(rune(b))
	}
	sb.WriteByte(cpu.Dt)
	sb.WriteByte(cpu.St)

	return sb.String()
}
