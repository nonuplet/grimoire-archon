package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version Application version
var Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョンを表示",
	Long:  "Archonのバージョン情報を表示します。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version: %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
