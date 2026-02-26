package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nonuplet/grimoire-archon/internal/appversion"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョンを表示",
	Long:  "Archonのバージョン情報を表示します。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version: %s\n", appversion.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
