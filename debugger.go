package xip8

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{} // use default options

func (d *HttpDebugger) Listen(port int) error {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/debugger", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close()
		for {
			running := true
			for running {
				select {
				case cpu := <-d.send:
					err = conn.WriteMessage(websocket.BinaryMessage, d.formatAsEvent(cpu))
					if err != nil {
						log.Fatalln("error write:", err)
						running = false
					}

				case <-r.Context().Done():
					running = false
				}
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
	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		d.Cpu.Stop()
		d.Cpu.Reset()
	})
	http.HandleFunc("/step", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		d.Cpu.SingleFrame()
		d.send <- *d.Cpu
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

func (d HttpDebugger) formatAsEvent(cpu Cpu) []byte {
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

	return buf
}
