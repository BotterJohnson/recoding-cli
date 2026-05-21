package render

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/botter/recoding-cli/internal/provider"
	"github.com/charmbracelet/glamour"
)

const maxCodeLines = 10

// Stream 消费流式事件，等待完成后折叠代码块并渲染输出。
func Stream(ch <-chan provider.StreamEvent) error {
	var buf strings.Builder
	firstToken := true

	stopSpinner := make(chan struct{})
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-stopSpinner:
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s 等待模型响应...", frames[i%len(frames)])
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	charCount := 0
	for ev := range ch {
		switch ev.Type {
		case provider.EventTextDelta:
			if firstToken {
				close(stopSpinner)
				fmt.Fprintf(os.Stderr, "✍️  生成中")
				firstToken = false
			}
			buf.WriteString(ev.Text)
			charCount += len(ev.Text)
			if charCount/200 > (charCount-len(ev.Text))/200 {
				fmt.Fprintf(os.Stderr, ".")
			}
		case provider.EventError:
			if firstToken {
				close(stopSpinner)
			}
			return fmt.Errorf("stream error: %w", ev.Error)
		case provider.EventDone:
			if firstToken {
				close(stopSpinner)
			}
		}
	}

	fmt.Fprintf(os.Stderr, " 完成\n\n")

	content := buf.String()
	if content == "" {
		return nil
	}

	collapsed := collapseCodeBlocks(content)

	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(100))
	if err != nil {
		fmt.Println(collapsed)
		return nil
	}
	out, err := r.Render(collapsed)
	if err != nil {
		fmt.Println(collapsed)
		return nil
	}
	fmt.Print(out)
	return nil
}

// collapseCodeBlocks 将超长代码块截断为前 N 行 + 省略提示。
func collapseCodeBlocks(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")
	inCode := false
	codeLineCount := 0
	skippedLines := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检测代码块开始（```xxx）
		if !inCode && strings.HasPrefix(trimmed, "```") {
			inCode = true
			codeLineCount = 0
			skippedLines = 0
			result.WriteString(line + "\n")
			continue
		}

		// 检测代码块结束（单独的 ```）
		if inCode && trimmed == "```" {
			if skippedLines > 0 {
				result.WriteString(fmt.Sprintf("// ... 省略 %d 行\n", skippedLines))
			}
			inCode = false
			result.WriteString(line + "\n")
			continue
		}

		if inCode {
			codeLineCount++
			if codeLineCount <= maxCodeLines {
				result.WriteString(line + "\n")
			} else {
				skippedLines++
			}
		} else {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}
