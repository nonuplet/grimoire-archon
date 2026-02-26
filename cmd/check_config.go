package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nonuplet/grimoire-archon/internal/usecase"
)

// checkConfigCmd checkConfigコマンドの生成
var checkConfigCmd = &cobra.Command{
	Use:   "check-config",
	Short: "コンフィグのチェックを行います。",
	Long: `現在読み込んでいるコンフィグのチェックを行います。
指定したコンフィグファイルをチェックしたい場合は、-c か --config を使って指定してください。
例外的として、ゲームのRunコマンドのチェックは行いません。
`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		checkConfigUsecase := usecase.NewCheckConfigUsecase(&cfg, fs, cliUtil)

		fmt.Println("コンフィグをチェックします...")
		if err := checkConfigUsecase.Execute(); err != nil {
			fmt.Println("エラーが見つかりました！")
		} else {
			fmt.Println("チェックが完了しました。エラーはありません。")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkConfigCmd)
}
