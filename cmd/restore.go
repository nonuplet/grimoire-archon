package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nonuplet/grimoire-archon/internal/infra/storage"
	"github.com/nonuplet/grimoire-archon/internal/usecase"
)

// cleanCmd restoreコマンドの生成
var restoreCmd = &cobra.Command{
	Use:   "restore <name> <archive>",
	Short: "指定したゲームのバックアップを復元します。",
	Long: `指定したゲームのバックアップを復元します。
第一引数に .archon.yaml のコンフィグで指定したゲーム名を渡してください。
第二引数に元にするバックアップデータ(.zip)を指定してください。
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		zipPath := args[1]
		restoreUsecase := usecase.NewRestoreUsecase()

		game, ok := cfg.Games[name]
		if !ok {
			return fmt.Errorf("%s は設定されていません。コンフィグを確認してください", name)
		}
		if _, err := storage.GetInfo(zipPath); err != nil {
			return fmt.Errorf("アーカイブファイル %s が見つかりません。", zipPath)
		}

		fmt.Printf("%s の復元処理を行います...\n", name)

		if err := restoreUsecase.Execute(cfg.Archon, game, zipPath); err != nil {
			return fmt.Errorf("%s の復元に失敗しました : %w", name, err)
		}

		fmt.Printf("%s の復元に成功しました。\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
