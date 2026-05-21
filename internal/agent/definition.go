package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// LoadPrompt 加载指定名称的 prompt 模板文件内容。
func LoadPrompt(name string) (string, error) {
	// 尝试从可执行文件同级 prompts/ 目录加载
	exe, _ := os.Executable()
	paths := []string{
		filepath.Join(filepath.Dir(exe), "prompts", name),
		filepath.Join("prompts", name),
	}
	// 开发时从源码目录加载
	if _, file, _, ok := runtime.Caller(0); ok {
		paths = append(paths, filepath.Join(filepath.Dir(file), "..", "..", "prompts", name))
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("prompt %s not found", name)
}
