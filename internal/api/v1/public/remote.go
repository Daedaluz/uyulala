package public

import (
	"sync"
	"uyulala/internal/api"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type Room struct {
	roomID  string
	clients map[*websocket.Conn]bool
	mutex   sync.RWMutex
}

func (r *Room) addClient(conn *websocket.Conn) int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if len(r.clients) == 2 {
		return -1
	}
	r.clients[conn] = true
	return len(r.clients)
}

func (r *Room) removeClient(conn *websocket.Conn) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.clients, conn)
	if len(r.clients) == 0 {
		roomsMutex.Lock()
		defer roomsMutex.Unlock()
		delete(rooms, r.roomID)
	} else if len(r.clients) == 1 {
		for conn := range r.clients {
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"waiting"}`))
		}
	}
}

func (r *Room) broadcast(src *websocket.Conn, msg []byte) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for conn := range r.clients {
		if conn != src {
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

var (
	roomsMutex = sync.Mutex{}
	rooms      = map[string]*Room{}
)

func getRoom(roomID string) *Room {
	newRoom := &Room{
		roomID:  roomID,
		clients: map[*websocket.Conn]bool{},
		mutex:   sync.RWMutex{},
	}
	roomsMutex.Lock()
	defer roomsMutex.Unlock()
	room, ok := rooms[roomID]
	if ok {
		return room
	} else {
		rooms[roomID] = newRoom
		return newRoom
	}
}

func remoteHandler(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		api.AbortError(ctx, 500, "internal_error", "Failed to upgrade", err)
		return
	}
	defer conn.Close()

	roomID := ctx.Param("id")
	room := getRoom(roomID)
	x := room.addClient(conn)
	defer room.removeClient(conn)

	switch x {
	case -1:
		_ = conn.WriteJSON(gin.H{
			"event": "busy",
		})
		conn.Close()
		return
	case 1:
		room.broadcast(nil, []byte(`{"event":"waiting"}`))
	case 2:
		room.broadcast(nil, []byte(`{"event":"ready"}`))
	}

	for {
		t, msg, err := conn.ReadMessage()
		if t != websocket.TextMessage || err != nil {
			break
		}
		room.broadcast(conn, msg)
	}
}
