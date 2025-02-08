package web

import (
	"github.com/gorilla/websocket"
	"github.com/guslan/xip8"
)

// Boot implements Display.
func (server *Server) Boot() error {
	return nil
}

func (server *Server) setWs(conn *websocket.Conn) error {
	server.wsMutex.Lock()
	server.socket = conn
	defer server.wsMutex.Unlock()

	return nil
}

func (server *Server) unsetWs() error {
	server.wsMutex.Lock()
	server.socket = nil
	defer server.wsMutex.Unlock()

	return nil
}

// Render implements Display.
func (server *Server) Render(screen xip8.Screen, settings xip8.ScreenSettings) error {
	if server.socket != nil {
		server.wsMutex.RLock()
		err := server.socket.WriteMessage(websocket.BinaryMessage, screen)
		server.wsMutex.RUnlock()
		if err != nil {
			return err
		}
	}

	return nil
}
