// Package cmd コマンド
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "My CLI tool",
}

// Execute execute
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		return
	}
}
