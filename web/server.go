package web

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/guslan/xip8"
)

var signal = struct{}{}

type Server struct {
	*xip8.InMemoryKeyboard
	*xip8.DummyBuzzer

	cpu      *xip8.Cpu
	debugger *HttpDebugger

	socket  *websocket.Conn
	wsMutex sync.RWMutex

	renderCh chan struct{}
	keyCh    chan xip8.KeyboardState
}

type ServerConfig struct {
	ScreenSettings xip8.ScreenSettings
	UseDebugger    bool
}
type ServerConfigCb func(config *ServerConfig)

func NewServer(mem *xip8.Memory, configs ...ServerConfigCb) *Server {
	config := &ServerConfig{
		ScreenSettings: xip8.SmallScreen,
		UseDebugger:    false,
	}
	for _, cb := range configs {
		cb(config)
	}

	s := &Server{
		InMemoryKeyboard: xip8.NewInMemoryKeyboard(),
		DummyBuzzer:      xip8.NewDummyBuzzer(),

		cpu:      nil,
		debugger: nil,
		wsMutex:  sync.RWMutex{},

		renderCh: make(chan struct{}),
		keyCh:    make(chan xip8.KeyboardState),
	}

	s.cpu = xip8.NewCpu(mem, config.ScreenSettings, s, s, s.DummyBuzzer)
	if config.UseDebugger {
		s.debugger = NewHttpDebugger(s.cpu)
	}

	return s
}

// func (server *Server) WithDisplaySize(w, h int)  {
// 	server.InMemoryDisplay.
// }

// func NewServerWithDebugger(cpu *Cpu, configs ...ServerConfig) *Server {
// 	return NewServer(cpu, append(configs, func(server *Server) {
// 		server.addDebugger()
// 	})...)
// }

func (server *Server) Speed(s int) {
	server.cpu.SetSpeedInHz(uint(s))
}

func (server *Server) Listen(port int) error {
	if err := server.cpu.Boot(); err != nil {
		slog.Error(err.Error())
		log.Fatalln(err)
	}

	go func() {
		server.cpu.Stop()
		server.cpu.CyclesPerFrame = 1
		if err := server.cpu.Loop(); err != nil {
			log.Fatalln(err)
		}
	}()

	slog.Info("Listening on port", slog.Int("port", port))

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		slog.Info("Starting")
		server.cpu.Start()
	})
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		slog.Info("Stopping")
		server.cpu.Stop()
	})
	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		slog.Info("Stopping and resetting")
		server.cpu.Stop()
		server.cpu.Reset()
	})
	http.HandleFunc("/step", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Cache-Control", "no-cache")

		slog.Info("Single Frame")
		server.cpu.LoopOnce()
	})
	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close()

		slog.Info("Connecting to display")
		server.setWs(conn)
		defer server.unsetWs()

		running := true
		for running {
			select {
			case <-r.Context().Done():
				running = false
				server.unsetWs()
				slog.Info("Disconnecting from display")
			}
		}
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// LoadProgram loads the program into memory and sets the PC to the start-of-program address
func (server *Server) LoadProgram(program []byte) error {
	return server.cpu.LoadProgram(program)
}
