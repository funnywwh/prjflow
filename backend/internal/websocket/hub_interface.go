package websocket

// HubInterface WebSocket Hub接口
// 用于依赖注入和测试mock
type HubInterface interface {
	// SendMessage 发送消息到指定ticket的连接
	SendMessage(ticket, messageType string, data interface{}, message string) error
}

// 确保Hub实现了HubInterface接口
var _ HubInterface = (*Hub)(nil)

