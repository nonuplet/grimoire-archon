// Package cmd コマンド
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "archon",
	Short: "ゲームサーバ管理用CLIツール",
	Long:  "ゲームサーバのインストール、セーブデータバックアップなどの操作を提供するCLIツールです。",
}

// Execute execute
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}
