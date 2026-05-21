package provider

import "context"

// Provider 是 LLM 调用的统一接口。
type Provider interface {
	// ChatStream 流式对话，返回事件 channel。
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error)
	// Chat 非流式对话。
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	// ModelMeta 返回模型元信息。
	ModelMeta() *ModelMeta
}

// ChatRequest 是对话请求。
type ChatRequest struct {
	Messages    []Message
	Temperature float64
	MaxTokens   int
}

// Message 是对话消息。
type Message struct {
	Role    string
	Content string
}

// ChatResponse 是非流式对话响应。
type ChatResponse struct {
	Content string
}

// EventType 是流式事件类型。
type EventType int

const (
	EventTextDelta EventType = iota
	EventDone
	EventError
)

// StreamEvent 是流式事件。
type StreamEvent struct {
	Type  EventType
	Text  string
	Error error
}
