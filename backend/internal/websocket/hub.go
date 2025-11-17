package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub 管理所有WebSocket连接
type Hub struct {
	// 注册的连接，key是ticket，value是连接
	connections map[string]*Connection
	// 互斥锁
	mu sync.RWMutex
}

// Connection 表示一个WebSocket连接
type Connection struct {
	// WebSocket连接
	ws *websocket.Conn
	// 发送消息的通道
	send chan []byte
	// ticket标识
	ticket string
}

// Message WebSocket消息结构
type Message struct {
	Type    string      `json:"type"`    // 消息类型：success, error, info
	Data    interface{} `json:"data"`    // 消息数据
	Ticket  string      `json:"ticket"`  // ticket标识
	Message string      `json:"message"` // 消息内容
}

var hub *Hub

func init() {
	hub = &Hub{
		connections: make(map[string]*Connection),
	}
}

// GetHub 获取全局Hub实例
func GetHub() *Hub {
	return hub
}

// Register 注册一个新的连接
func (h *Hub) Register(ticket string, conn *websocket.Conn) *Connection {
	h.mu.Lock()
	defer h.mu.Unlock()

	connection := &Connection{
		ws:     conn,
		send:   make(chan []byte, 256),
		ticket: ticket,
	}

	h.connections[ticket] = connection

	// 启动读写goroutine
	go connection.writePump()
	go connection.readPump()

	log.Printf("WebSocket连接已注册: ticket=%s", ticket)
	return connection
}

// Unregister 注销连接
func (h *Hub) Unregister(ticket string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.connections[ticket]; ok {
		close(conn.send)
		delete(h.connections, ticket)
		log.Printf("WebSocket连接已注销: ticket=%s", ticket)
	}
}

// SendMessage 向指定ticket发送消息
func (h *Hub) SendMessage(ticket string, msgType string, data interface{}, message string) error {
	h.mu.RLock()
	conn, ok := h.connections[ticket]
	h.mu.RUnlock()

	if !ok {
		return nil // 连接不存在，忽略
	}

	msg := Message{
		Type:    msgType,
		Data:    data,
		Ticket:  ticket,
		Message: message,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case conn.send <- msgBytes:
	default:
		// 发送失败，可能连接已关闭
		h.Unregister(ticket)
	}

	return nil
}

// readPump 从WebSocket读取消息
func (c *Connection) readPump() {
	defer func() {
		c.ws.Close()
		GetHub().Unregister(c.ticket)
	}()

	// 设置读取限制
	c.ws.SetReadLimit(512)
	// 设置ping处理器
	c.ws.SetPongHandler(func(string) error {
		return nil
	})

	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}
	}
}

// writePump 向WebSocket写入消息
func (c *Connection) writePump() {
	defer c.ws.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// 通道已关闭
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket写入错误: %v", err)
				return
			}
		}
	}
}

