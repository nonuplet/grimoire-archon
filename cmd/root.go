package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/nonuplet/grimoire-archon/internal/config"
)

var (
	cfgPath string
	cfg     config.Config
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

func init() {
	cobra.OnInitialize(initConfig)

	// -c, --config
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "コンフィグファイル (デフォルトで ./archon.yaml, なければ ~/.archon.yamlを読み込みます)")
}

// initConfig コンフィグファイルの読み込み
func initConfig() {
	// ロード先の決定
	targetPath := getConfigPath()
	if targetPath == "" {
		fmt.Fprint(os.Stderr, "コンフィグファイルが見つかりません。--configで指定するか、./.archon.yaml または ~/.archon.yaml に配置してください。\n", cfgPath)
		os.Exit(1)
	}

	// 読み込み処理
	// #nosec G304
	file, err := os.ReadFile(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "コンフィグファイルの読み込みに失敗しました: %s\n", err)
		os.Exit(1)
	}

	if err := yaml.Unmarshal(file, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "コンフィグファイルのパースに失敗しました: %s\n", err)
		os.Exit(1)
	}

	if cfg.Games == nil {
		fmt.Fprint(os.Stderr, "ゲーム設定が見つかりません。config.yaml に games セクションを追加してください。\n")
		os.Exit(1)
	}
}

func getConfigPath() string {
	// 1. フラグで指定された場合
	if cfgPath != "" {
		if _, err := os.Stat(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "指定されたコンフィグファイルが見つかりません: %s\n", cfgPath)
			os.Exit(1)
		}
		return cfgPath
	}

	// 2. カレントディレクトリ
	cwdPath := "./.archon.yaml"
	if _, err := os.Stat(cwdPath); err == nil {
		return cwdPath
	}

	// 3. ~/.archon.yaml の確認
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ホームディレクトリの取得に失敗しました: %s\n", err)
		os.Exit(1)
	}
	homePath := filepath.Join(home, "archon.yaml")
	if _, err := os.Stat(homePath); err == nil {
		return homePath
	}

	// 4. 何も見つからなければ空文字列を返す
	return ""
}
