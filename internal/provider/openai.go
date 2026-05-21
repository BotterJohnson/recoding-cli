package provider

import (
	"context"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider 是 OpenAI 兼容的 Provider 实现。
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider 创建 OpenAI 兼容的 Provider。
// baseURL 为空时使用默认 OpenAI endpoint。
func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &OpenAIProvider{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}
}

// ChatStream 流式调用 LLM，返回事件 channel。
func (p *OpenAIProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	msgs := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    msgs,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("provider stream: %w", err)
	}

	ch := make(chan StreamEvent)
	go func() {
		defer close(ch)
		defer stream.Close()
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				ch <- StreamEvent{Type: EventDone}
				return
			}
			if err != nil {
				ch <- StreamEvent{Type: EventError, Error: err}
				return
			}
			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					ch <- StreamEvent{Type: EventTextDelta, Text: delta}
				}
			}
		}
	}()
	return ch, nil
}

// Chat 非流式调用 LLM。
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	msgs := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    msgs,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("provider chat: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("provider chat: empty response")
	}
	return &ChatResponse{Content: resp.Choices[0].Message.Content}, nil
}

// ModelMeta 返回模型元信息。
func (p *OpenAIProvider) ModelMeta() *ModelMeta {
	return &ModelMeta{
		MaxContextTokens: 128000,
		MaxOutputTokens:  4096,
		SupportsStream:   true,
	}
}
