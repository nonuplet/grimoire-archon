package usecase

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/nonuplet/grimoire-archon/internal/domain"
	"github.com/nonuplet/grimoire-archon/internal/infra/cli"
	"github.com/nonuplet/grimoire-archon/internal/infra/filesystem"
)

// CheckConfigUsecase コンフィグのチェックを行うユースケース
type CheckConfigUsecase struct{}

// NewCheckConfigUsecase CheckConfigUsecaseのインスタンスを生成する
func NewCheckConfigUsecase() *CheckConfigUsecase {
	return &CheckConfigUsecase{}
}

// Execute コンフィグのチェックを実行する
func (u *CheckConfigUsecase) Execute(cfg *domain.Config) error {
	// Config のチェック
	res := checkConfig(cfg)
	if res != "" {
		fmt.Println("全体のコンフィグにエラーが見つかりました。")
		fmt.Fprint(os.Stderr, res)
		return fmt.Errorf("全体のコンフィグにエラーが見つかりました。")
	}

	// ArchonConfigのチェック
	res = checkArchonConfig(cfg.Archon)
	if res != "" {
		fmt.Println("archon設定にエラーが見つかりました。")
		fmt.Fprint(os.Stderr, res)
		return fmt.Errorf("archon設定にエラーが見つかりました。")
	}

	// ゲームのチェック
	if result := checkAllGameConfigs(cfg.Games); result {
		fmt.Println("ゲーム設定にエラーが見つかりました。")
		return fmt.Errorf("ゲーム設定にエラーが見つかりました。")
	}

	return nil
}

// checkConfig 大本の Config 型のチェック
func checkConfig(cfg *domain.Config) string {
	baseMsg := "全体のエラー: "
	var sb strings.Builder

	if cfg == nil {
		cli.Writeln(&sb, baseMsg, "コンフィグファイルが空です！")
		return sb.String()
	}

	if cfg.Archon == nil {
		cli.Writeln(&sb, baseMsg, "archon 設定が見つかりません。")
	}

	if len(cfg.Games) == 0 {
		cli.Writeln(&sb, baseMsg, "ゲーム設定が見つかりません。")
	}

	return sb.String()
}

// checkArchonConfig ArchonConfigのチェック
func checkArchonConfig(archonCfg *domain.ArchonConfig) string {
	baseMsg := "archon: "
	var sb strings.Builder

	if archonCfg.BackupDir == "" {
		cli.Writeln(&sb, baseMsg, "バックアップディレクトリが設定されていません。")
		return sb.String()
	}
	if _, err := filesystem.GetInfo(archonCfg.BackupDir); err != nil {
		cli.Writeln(&sb, baseMsg, "バックアップディレクトリ ", archonCfg.BackupDir, " は見つかりません。backup/restoreコマンドを実行すると、自動作成されます。")
	}

	if _, err := filesystem.GetInfo(archonCfg.AppdataDir); err != nil {
		cli.Writeln(&sb, baseMsg, "Appdataディレクトリ ", archonCfg.AppdataDir, " は指定されていますが、見つかりません。")
	}

	if _, err := filesystem.GetInfo(archonCfg.DocumentDir); err != nil {
		cli.Writeln(&sb, baseMsg, "Documentディレクトリ ", archonCfg.DocumentDir, " は指定されていますが、見つかりません。")
	}

	return sb.String()
}

// checkAllGameConfigs 全ゲームのコンフィグのチェック
func checkAllGameConfigs(games map[string]*domain.GameConfig) bool {
	isError := false

	for game, gameCfg := range games {
		fmt.Printf("%s ... ", game)
		if res := checkGameConfig(game, gameCfg); res != "" {
			fmt.Println("error!")
			fmt.Fprint(os.Stderr, res)
			isError = true
		} else {
			fmt.Println("OK.")
		}
	}

	return isError
}

// checkGameConfig ゲーム単体のコンフィグチェック
func checkGameConfig(game string, gameCfg *domain.GameConfig) string {
	baseMsg := fmt.Sprintf("game: %s: ", game)
	var sb strings.Builder

	// install_dir
	if gameCfg.InstallDir == "" {
		cli.Writeln(&sb, baseMsg, "インストールディレクトリが設定されていません。")
		return sb.String()
	}
	if _, err := filesystem.GetInfo(gameCfg.InstallDir); err != nil {
		cli.Writeln(&sb, baseMsg, "インストールディレクトリ ", gameCfg.InstallDir, " は見つかりません。")
		return sb.String()
	}

	// Linux向けチェック
	if runtime.GOOS == "linux" {
		winBinary := gameCfg.Steam.Platform == "windows"
		runNative := gameCfg.RuntimeEnv == "" || gameCfg.RuntimeEnv == domain.RuntimeEnvNative

		if winBinary && runNative {
			cli.Writeln(&sb, baseMsg, "LinuxでWindowsゲームを動かす場合、runtime_env に wine, proton のどちらかを指定してください。")
		}

		if !winBinary && !runNative {
			cli.Writeln(&sb, baseMsg, "wineまたはprotonが指定されていますが、ダウンロードするファイルはLinux用になっています。steam.platform を windows に変更してください。")
		}
	}

	// TODO: 将来的にファイルチェックも行う

	return sb.String()
}
