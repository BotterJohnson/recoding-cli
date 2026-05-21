package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	flagModel   string
	flagVerbose bool
	flagDebug   bool
)

var rootCmd = &cobra.Command{
	Use:   "recoding-cli",
	Short: "AI 编程助手",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagModel, "model", "m", "", "指定模型")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "调试模式")
}

// Execute 执行根命令。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
