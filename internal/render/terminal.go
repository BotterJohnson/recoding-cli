package render

import (
	"fmt"
	"io"

	"github.com/charmbracelet/glamour"
)

// Markdown 将 markdown 文本渲染为带语法高亮的终端输出。
func Markdown(w io.Writer, content string) error {
	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		// fallback: 直接输出原文
		fmt.Fprint(w, content)
		return nil
	}
	out, err := r.Render(content)
	if err != nil {
		fmt.Fprint(w, content)
		return nil
	}
	fmt.Fprint(w, out)
	return nil
}
