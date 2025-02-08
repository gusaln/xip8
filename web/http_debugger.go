package web

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/guslan/xip8"
)

type HttpDebugger struct {
	Cpu           *xip8.Cpu
	CurrentOpCode uint16
	Cycle         int

	SendEvery int
	send      chan xip8.Cpu
}

// NewHttpDebugger creates a new debugger
// This method will pause the cpu, register the hooks, set the execution cpu.CyclesPerFrame to 1
func NewHttpDebugger(cpu *xip8.Cpu) *HttpDebugger {
	deb := HttpDebugger{
		Cpu:           cpu,
		CurrentOpCode: 0,
		Cycle:         0,
		SendEvery:     1,
		send:          make(chan xip8.Cpu),
	}

	deb.setupWs()

	cpu.AddBeforeFrameHook(deb.beforeFrame)
	cpu.AddBeforeCycleHook(deb.beforeCycle)
	cpu.AddAfterCycleHook(deb.afterCycle)
	cpu.AddAfterFrameHook(deb.afterFrame)
	cpu.CyclesPerFrame = 1

	cpu.Stop()

	return &deb
}

var upgrader = websocket.Upgrader{} // use default options

func (d *HttpDebugger) setupWs() {
	http.HandleFunc("/debugger", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Connecting  to debugger")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close()

		go func(cpu xip8.Cpu) {
			d.send <- cpu
		}(*d.Cpu)

		slog.Info("Listening for events")
		for {
			running := true
			for running {
				select {
				case cpu := <-d.send:
					err = conn.WriteMessage(websocket.BinaryMessage, d.formatAsEvent(cpu))
					if err != nil {
						slog.Error("Error writing debugger message")
						running = false
					}

				case <-r.Context().Done():
					running = false
				}
			}
		}
	})
}

func (d *HttpDebugger) beforeFrame(cpu *xip8.Cpu) {
}

func (d *HttpDebugger) beforeCycle(cpu *xip8.Cpu) {
	d.CurrentOpCode = uint16(cpu.Memory[cpu.Pc+0]) << 8
	d.CurrentOpCode |= uint16(cpu.Memory[cpu.Pc+1]) << 0
}

func (d *HttpDebugger) afterCycle(cpu *xip8.Cpu) {
	if d.Cpu.Cycles()%uint(d.SendEvery) == 0 {
		d.send <- *cpu
	}

	slog.Info("Cycle ran")
}

func (d *HttpDebugger) afterFrame(cpu *xip8.Cpu) {
	slog.Info("Frame ran")
}

func (d HttpDebugger) formatAsEvent(cpu xip8.Cpu) []byte {
	buf := make([]byte, 0, 64)

	buf = append(buf, byte((d.CurrentOpCode&0xFF00)>>8))
	buf = append(buf, byte((d.CurrentOpCode&0x00FF)>>0))

	buf = append(buf, byte((cpu.Pc&0xFF00)>>8))
	buf = append(buf, byte((cpu.Pc&0x00FF)>>0))
	for _, b := range cpu.V {
		buf = append(buf, b)
	}
	buf = append(buf, byte((cpu.I&0xFF00)>>8))
	buf = append(buf, byte((cpu.I&0x00FF)>>0))
	buf = append(buf, cpu.Sp)
	for _, b := range cpu.Stack {
		buf = append(buf, byte((b&0xFF00)>>8))
		buf = append(buf, byte((b&0x00FF)>>0))
	}
	buf = append(buf, cpu.Dt)
	buf = append(buf, cpu.St)
	// For some unknown reasons there is a single empty byte between the sound timer and following stuff
	buf = append(buf, byte(cpu.ScreenSettings.Width))
	buf = append(buf, byte(cpu.ScreenSettings.Height))

	return buf
}
