package playerHandler

import (
	"github.com/gorilla/websocket"
	"sync"
)

// Represents a player
// Server uses these objects in main map to keep track of player status and send messages
type Player struct {
	Name     string
	PlayerId string
	GameId   string
	Ready    bool
	Alive	 bool
	Ws       *websocket.Conn
	WsLock   *sync.RWMutex
}