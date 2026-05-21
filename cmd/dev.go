package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/botter/recoding-cli/internal/agent"
	"github.com/botter/recoding-cli/internal/config"
	"github.com/botter/recoding-cli/internal/provider"
	"github.com/botter/recoding-cli/internal/render"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev [需求描述]",
	Short: "根据需求生成代码",
	RunE:  runDev,
}

func init() {
	rootCmd.AddCommand(devCmd)
}

func runDev(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	if cfg.Provider.APIKey == "" {
		return fmt.Errorf("未配置 API Key，请设置环境变量 RECODING_API_KEY 或在 ~/.recoding/config.yaml 中配置")
	}
	if flagModel != "" {
		cfg.Provider.Model = flagModel
	}

	systemPrompt, err := agent.LoadPrompt("dev.tmpl")
	if err != nil {
		return fmt.Errorf("加载 prompt 失败: %w", err)
	}

	p := provider.NewOpenAIProvider(cfg.Provider.APIKey, cfg.Provider.BaseURL, cfg.Provider.Model)
	rt := agent.NewRuntime(p)

	if len(args) > 0 {
		return runOnce(rt, systemPrompt, args[0])
	}
	return runInteractive(rt, systemPrompt)
}

func runOnce(rt *agent.Runtime, systemPrompt, userPrompt string) error {
	ch, err := rt.RunStream(context.Background(), systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("调用模型失败: %w", err)
	}
	return render.Stream(ch)
}

func runInteractive(rt *agent.Runtime, systemPrompt string) error {
	fmt.Println("╭─ recoding-cli dev ─────────────────────────╮")
	fmt.Println("│ 输入需求描述，按回车发送。输入 /exit 退出。 │")
	fmt.Println("╰────────────────────────────────────────────╯")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "/exit" || input == "/quit" {
			fmt.Println("👋 再见！")
			return nil
		}

		ch, err := rt.RunStream(context.Background(), systemPrompt, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 错误: %v\n\n", err)
			continue
		}
		if err := render.Stream(ch); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 错误: %v\n\n", err)
		}
	}
	return nil
}
