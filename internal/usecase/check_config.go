package usecase

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// CheckConfigUsecase コンフィグのチェックを行うユースケース
type CheckConfigUsecase struct {
	cfg *domain.Config
	fs  FileSystem
	cli Cli
}

// NewCheckConfigUsecase CheckConfigUsecaseのインスタンスを生成する
func NewCheckConfigUsecase(cfg *domain.Config, fs FileSystem, cli Cli) *CheckConfigUsecase {
	return &CheckConfigUsecase{
		cfg: cfg,
		fs:  fs,
		cli: cli,
	}
}

// Execute コンフィグのチェックを実行する
func (u *CheckConfigUsecase) Execute() error {
	// Config のチェック
	res := u.checkConfig()
	if res != "" {
		fmt.Println("全体のコンフィグにエラーが見つかりました。")
		fmt.Fprint(os.Stderr, res)
		return fmt.Errorf("全体のコンフィグにエラーが見つかりました。")
	}

	// ArchonConfigのチェック
	res = u.checkArchonConfig(u.cfg.Archon)
	if res != "" {
		fmt.Println("archon設定にエラーが見つかりました。")
		fmt.Fprint(os.Stderr, res)
		return fmt.Errorf("archon設定にエラーが見つかりました。")
	}

	// ゲームのチェック
	if result := u.checkAllGameConfigs(u.cfg.Games); result {
		fmt.Println("ゲーム設定にエラーが見つかりました。")
		return fmt.Errorf("ゲーム設定にエラーが見つかりました。")
	}

	return nil
}

// checkConfig 大本の Config 型のチェック
func (u *CheckConfigUsecase) checkConfig() string {
	baseMsg := "全体のエラー: "
	var sb strings.Builder

	if u.cfg == nil {
		u.cli.Writeln(&sb, baseMsg, "コンフィグファイルが空です！")
		return sb.String()
	}

	if u.cfg.Archon == nil {
		u.cli.Writeln(&sb, baseMsg, "archon 設定が見つかりません。")
	}

	if len(u.cfg.Games) == 0 {
		u.cli.Writeln(&sb, baseMsg, "ゲーム設定が見つかりません。")
	}

	return sb.String()
}

// checkArchonConfig ArchonConfigのチェック
func (u *CheckConfigUsecase) checkArchonConfig(archonCfg *domain.ArchonConfig) string {
	baseMsg := "archon: "
	var sb strings.Builder

	if archonCfg.BackupDir == "" {
		u.cli.Writeln(&sb, baseMsg, "バックアップディレクトリが設定されていません。")
		return sb.String()
	}
	if _, err := u.fs.Stat(archonCfg.BackupDir); err != nil {
		u.cli.Writeln(&sb, baseMsg, "バックアップディレクトリ ", archonCfg.BackupDir, " は見つかりません。backup/restoreコマンドを実行すると、自動作成されます。")
	}

	if _, err := u.fs.Stat(archonCfg.AppdataDir); err != nil {
		u.cli.Writeln(&sb, baseMsg, "Appdataディレクトリ ", archonCfg.AppdataDir, " は指定されていますが、見つかりません。")
	}

	if _, err := u.fs.Stat(archonCfg.DocumentDir); err != nil {
		u.cli.Writeln(&sb, baseMsg, "Documentディレクトリ ", archonCfg.DocumentDir, " は指定されていますが、見つかりません。")
	}

	return sb.String()
}

// checkAllGameConfigs 全ゲームのコンフィグのチェック
func (u *CheckConfigUsecase) checkAllGameConfigs(games map[string]*domain.GameConfig) bool {
	isError := false

	for game, gameCfg := range games {
		fmt.Printf("%s ... ", game)
		if res := u.checkGameConfig(game, gameCfg); res != "" {
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
func (u *CheckConfigUsecase) checkGameConfig(game string, gameCfg *domain.GameConfig) string {
	baseMsg := fmt.Sprintf("game: %s: ", game)
	var sb strings.Builder

	// install_dir
	if gameCfg.InstallDir == "" {
		u.cli.Writeln(&sb, baseMsg, "インストールディレクトリが設定されていません。")
		return sb.String()
	}
	if _, err := u.fs.Stat(gameCfg.InstallDir); err != nil {
		u.cli.Writeln(&sb, baseMsg, "インストールディレクトリ ", gameCfg.InstallDir, " は見つかりません。")
		return sb.String()
	}

	// Linux向けチェック
	if runtime.GOOS == "linux" {
		winBinary := gameCfg.Steam.Platform == "windows"
		runNative := gameCfg.RuntimeEnv == "" || gameCfg.RuntimeEnv == domain.RuntimeEnvNative

		if winBinary && runNative {
			u.cli.Writeln(&sb, baseMsg, "LinuxでWindowsゲームを動かす場合、runtime_env に wine, proton のどちらかを指定してください。")
		}

		if !winBinary && !runNative {
			u.cli.Writeln(&sb, baseMsg, "wineまたはprotonが指定されていますが、ダウンロードするファイルはLinux用になっています。steam.platform を windows に変更してください。")
		}
	}

	// TODO: 将来的にファイルチェックも行う

	return sb.String()
}
