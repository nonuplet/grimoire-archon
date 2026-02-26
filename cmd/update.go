package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nonuplet/grimoire-archon/internal/infra/steamcmd"
	"github.com/nonuplet/grimoire-archon/internal/usecase"
)

// updateCmd updateコマンドの生成
var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "指定したゲームを更新します。",
	Long:  "指定したゲームを更新します。引数で .archon.yaml のコンフィグで指定したゲーム名を渡してください。",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		game, ok := cfg.Games[name]
		if !ok {
			return fmt.Errorf("%s は設定されていません。コンフィグを確認してください", name)
		}

		ctx := context.Background()
		steam := steamcmd.NewSteamCmd()
		updateUsecase := usecase.NewUpdateUsecase(game, steam, fs)

		fmt.Printf("%s を更新中...\n", name)

		if err := updateUsecase.Execute(ctx); err != nil {
			return fmt.Errorf("%s のアップデートに失敗しました : %w", name, err)
		}

		fmt.Printf("%s のアップデートに成功しました。\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
