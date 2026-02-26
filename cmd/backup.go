package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nonuplet/grimoire-archon/internal/adapter/snapshot"
	"github.com/nonuplet/grimoire-archon/internal/usecase"
)

// backupCmd backupコマンドの生成
var backupCmd = &cobra.Command{
	Use:   "backup <name>",
	Short: "指定したゲームのバックアップを取ります。",
	Long: `指定したゲームのバックアップを取ります。
引数で .archon.yaml のコンフィグで指定したゲーム名を渡してください。
保存先はコンフィグで指定した backup_dir 以下に、ゲームの name でディレクトリが作成されます。
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		game, ok := cfg.Games[name]
		if !ok {
			return fmt.Errorf("%s は設定されていません。コンフィグを確認してください", name)
		}

		snap := snapshot.NewSnapshot(cfg.Archon, game, fs, cliUtil)
		backupUsecase := usecase.NewBackupUsecase(cfg.Archon, game, snap, fs, cliUtil)

		fmt.Printf("%s のバックアップを取得します...\n", name)

		if err := backupUsecase.Execute(); err != nil {
			return fmt.Errorf("%s のバックアップに失敗しました : %w", name, err)
		}

		fmt.Printf("%s のバックアップに成功しました。\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
