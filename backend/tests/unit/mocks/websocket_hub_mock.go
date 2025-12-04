package mocks

import (
	"fmt"
	"sync"
)

// MockWebSocketHub WebSocket Hub Mock实现
type MockWebSocketHub struct {
	mu sync.RWMutex

	// 记录所有发送的消息
	Messages []SentMessage

	// 可配置的错误（按ticket）
	Errors map[string]error
}

// SentMessage 记录发送的消息
type SentMessage struct {
	Ticket     string
	MessageType string
	Data       interface{}
	Message    string
}

// NewMockWebSocketHub 创建新的MockWebSocketHub
func NewMockWebSocketHub() *MockWebSocketHub {
	return &MockWebSocketHub{
		Messages: make([]SentMessage, 0),
		Errors:   make(map[string]error),
	}
}

// SendMessage 实现HubInterface接口
func (m *MockWebSocketHub) SendMessage(ticket, messageType string, data interface{}, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否有配置的错误
	if err, ok := m.Errors[ticket]; ok {
		return err
	}

	// 记录消息
	m.Messages = append(m.Messages, SentMessage{
		Ticket:      ticket,
		MessageType: messageType,
		Data:        data,
		Message:     message,
	})

	return nil
}

// GetMessages 获取所有发送的消息
func (m *MockWebSocketHub) GetMessages() []SentMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SentMessage, len(m.Messages))
	copy(result, m.Messages)
	return result
}

// GetMessagesByTicket 获取指定ticket的消息
func (m *MockWebSocketHub) GetMessagesByTicket(ticket string) []SentMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []SentMessage
	for _, msg := range m.Messages {
		if msg.Ticket == ticket {
			result = append(result, msg)
		}
	}
	return result
}

// GetMessagesByType 获取指定类型的消息
func (m *MockWebSocketHub) GetMessagesByType(messageType string) []SentMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []SentMessage
	for _, msg := range m.Messages {
		if msg.MessageType == messageType {
			result = append(result, msg)
		}
	}
	return result
}

// HasMessage 检查是否发送了指定消息
func (m *MockWebSocketHub) HasMessage(ticket, messageType, message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, msg := range m.Messages {
		if msg.Ticket == ticket && msg.MessageType == messageType && msg.Message == message {
			return true
		}
	}
	return false
}

// SetError 设置指定ticket的错误
func (m *MockWebSocketHub) SetError(ticket string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Errors[ticket] = err
}

// ClearError 清除指定ticket的错误
func (m *MockWebSocketHub) ClearError(ticket string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Errors, ticket)
}

// Reset 重置Mock状态
func (m *MockWebSocketHub) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Messages = make([]SentMessage, 0)
	m.Errors = make(map[string]error)
}

// MessageCount 获取消息总数
func (m *MockWebSocketHub) MessageCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.Messages)
}

// String 返回Mock状态的字符串表示（用于调试）
func (m *MockWebSocketHub) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("MockWebSocketHub{Messages: %d, Errors: %d}", len(m.Messages), len(m.Errors))
}

