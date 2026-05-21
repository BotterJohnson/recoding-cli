package agent

import (
	"context"
	"fmt"

	"github.com/botter/recoding-cli/internal/provider"
)

// Runtime 是 Agent 执行引擎。
type Runtime struct {
	provider provider.Provider
}

// NewRuntime 创建 Agent Runtime。
func NewRuntime(p provider.Provider) *Runtime {
	return &Runtime{provider: p}
}

// RunStream 执行单轮流式对话，返回事件 channel。
func (r *Runtime) RunStream(ctx context.Context, systemPrompt, userPrompt string) (<-chan provider.StreamEvent, error) {
	req := &provider.ChatRequest{
		Messages: []provider.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   4096,
	}
	ch, err := r.provider.ChatStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("agent run: %w", err)
	}
	return ch, nil
}
