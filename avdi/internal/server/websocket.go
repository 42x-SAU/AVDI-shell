package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// В разработке разрешаем все origin
		return true
	},
}

// Client представляет подключение WebSocket.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	// Подписки клиента (например, "agent:1", "task:2", "broadcast")
	subscriptions map[string]bool
}

// Hub управляет подключенными клиентами и рассылкой сообщений.
type Hub struct {
	// Зарегистрированные клиенты.
	clients map[*Client]bool

	// Входящие сообщения от клиентов.
	broadcast chan []byte

	// Запросы на регистрацию клиента.
	register chan *Client

	// Запросы на удаление клиента.
	unregister chan *Client

	// Подписки: канал -> множество клиентов
	subscriptions map[string]map[*Client]bool

	// Мьютекс для безопасного доступа к подпискам
	mu sync.RWMutex
}

// NewHub создаёт новый хаб.
func NewHub() *Hub {
	return &Hub{
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		subscriptions: make(map[string]map[*Client]bool),
	}
}

// Run запускает хаб в отдельной горутине.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				// Удалить клиента из всех подписок
				h.mu.Lock()
				for channel := range client.subscriptions {
					if subs, ok := h.subscriptions[channel]; ok {
						delete(subs, client)
						if len(subs) == 0 {
							delete(h.subscriptions, channel)
						}
					}
				}
				h.mu.Unlock()
			}
		case message := <-h.broadcast:
			// Отправить всем подключенным клиентам
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Subscribe добавляет клиента в подписку на канал.
func (h *Hub) Subscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subscriptions[channel]; !ok {
		h.subscriptions[channel] = make(map[*Client]bool)
	}
	h.subscriptions[channel][client] = true
	client.subscriptions[channel] = true
}

// Unsubscribe удаляет клиента из подписки на канал.
func (h *Hub) Unsubscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs, ok := h.subscriptions[channel]; ok {
		delete(subs, client)
		if len(subs) == 0 {
			delete(h.subscriptions, channel)
		}
	}
	delete(client.subscriptions, channel)
}

// Publish отправляет сообщение всем подписанным на канал клиентам.
func (h *Hub) Publish(channel string, message []byte) {
	h.mu.RLock()
	subs, ok := h.subscriptions[channel]
	if !ok {
		h.mu.RUnlock()
		return
	}
	// Копируем клиентов, чтобы не блокировать мьютекс во время отправки
	clients := make([]*Client, 0, len(subs))
	for client := range subs {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			// Если канал полный, закрываем соединение
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// readPump читает сообщения от клиента.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			break
		}
		// Обработка входящих сообщений (например, подписка/отписка)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if action, ok := msg["action"].(string); ok {
				switch action {
				case "subscribe":
					if channel, ok := msg["channel"].(string); ok {
						c.hub.Subscribe(c, channel)
					}
				case "unsubscribe":
					if channel, ok := msg["channel"].(string); ok {
						c.hub.Unsubscribe(c, channel)
					}
				}
			}
		}
	}
}

// writePump отправляет сообщения клиенту.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Хаб закрыл канал
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Добавить ожидающие сообщения в текущий writer
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWebSocket обрабатывает HTTP-запрос на upgrade до WebSocket.
func ServeWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}
	client.hub.register <- client

	// Запускаем горутины для чтения и записи
	go client.writePump()
	go client.readPump()
}

// Event представляет событие для отправки через WebSocket.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// PublishEvent отправляет событие в указанный канал.
func PublishEvent(hub *Hub, channel string, event Event) {
	message, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal event: %v", err)
		return
	}
	hub.Publish(channel, message)
}

// PublishBroadcast отправляет событие всем подключенным клиентам.
func PublishBroadcast(hub *Hub, event Event) {
	message, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal event: %v", err)
		return
	}
	hub.broadcast <- message
}