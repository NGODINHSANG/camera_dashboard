package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"camera-dashboard-backend/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ChatClient struct {
	UserID int64
	conn   *websocket.Conn
	send   chan []byte
}

type ChatHub struct {
	clients    map[int64]*ChatClient // userID → client
	mu         sync.RWMutex
	register   chan *ChatClient
	unregister chan *ChatClient
	broadcast  chan *hubMessage
}

type hubMessage struct {
	targetUserIDs []int64
	event         models.WSEvent
}

var GlobalChatHub = newChatHub()

func newChatHub() *ChatHub {
	return &ChatHub{
		clients:    make(map[int64]*ChatClient),
		register:   make(chan *ChatClient, 64),
		unregister: make(chan *ChatClient, 64),
		broadcast:  make(chan *hubMessage, 256),
	}
}

func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("[ChatHub] User %d connected (%d online)", client.UserID, len(h.clients))
			// Thông báo user online
			h.BroadcastToAll(models.WSEvent{
				Type:    models.WSEventUserOnline,
				Payload: map[string]int64{"userId": client.UserID},
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if h.clients[client.UserID] == client {
				delete(h.clients, client.UserID)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[ChatHub] User %d disconnected (%d online)", client.UserID, len(h.clients))
			h.BroadcastToAll(models.WSEvent{
				Type:    models.WSEventUserOffline,
				Payload: map[string]int64{"userId": client.UserID},
			})

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg.event)
			if err != nil {
				continue
			}
			h.mu.RLock()
			for _, uid := range msg.targetUserIDs {
				if c, ok := h.clients[uid]; ok {
					select {
					case c.send <- data:
					default:
						// buffer đầy → bỏ qua
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUsers gửi event đến danh sách user cụ thể
func (h *ChatHub) SendToUsers(userIDs []int64, event models.WSEvent) {
	h.broadcast <- &hubMessage{targetUserIDs: userIDs, event: event}
}

// BroadcastToAll gửi đến tất cả user đang online
func (h *ChatHub) BroadcastToAll(event models.WSEvent) {
	h.mu.RLock()
	ids := make([]int64, 0, len(h.clients))
	for uid := range h.clients {
		ids = append(ids, uid)
	}
	h.mu.RUnlock()
	h.broadcast <- &hubMessage{targetUserIDs: ids, event: event}
}

// OnlineUserIDs trả về danh sách user đang online
func (h *ChatHub) OnlineUserIDs() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]int64, 0, len(h.clients))
	for uid := range h.clients {
		ids = append(ids, uid)
	}
	return ids
}

// ServeWS nâng cấp HTTP lên WebSocket và gắn client vào hub
func (h *ChatHub) ServeWS(w http.ResponseWriter, r *http.Request, userID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ChatHub] Upgrade error: %v", err)
		return
	}

	client := &ChatClient{
		UserID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump(h)
	go client.readPump(h)
}

func (c *ChatClient) writePump(h *ChatHub) {
	defer func() {
		c.conn.Close()
	}()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func (c *ChatClient) readPump(h *ChatHub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// Xử lý typing event từ client
		var event models.WSEvent
		if err := json.Unmarshal(msg, &event); err == nil {
			if event.Type == models.WSEventUserTyping {
				// forward cho các member khác trong conversation (handled by chat handler)
				_ = event
			}
		}
	}
}
