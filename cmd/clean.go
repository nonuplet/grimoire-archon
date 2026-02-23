package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nonuplet/grimoire-archon/internal/usecase"
)

// cleanCmd cleanコマンドの生成
var cleanCmd = &cobra.Command{
	Use:   "clean <name>",
	Short: "指定したゲームを削除します。",
	Long: `指定したゲームを削除します。
引数で .archon.yaml のコンフィグで指定したゲーム名を渡してください。
バックアップを取っていない場合、エラーチェックが入ります。強制的に削除することも可能です。
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cleanUsecase := usecase.NewCleanUsecase()

		game, ok := cfg.Games[name]
		if !ok {
			return fmt.Errorf("%s は設定されていません。コンフィグを確認してください", name)
		}

		fmt.Printf("%s の削除処理を行います...\n", name)

		if err := cleanUsecase.Execute(cfg.Archon, game); err != nil {
			return fmt.Errorf("%s の削除に失敗しました : %w", name, err)
		}

		fmt.Printf("%s の削除に成功しました。\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
