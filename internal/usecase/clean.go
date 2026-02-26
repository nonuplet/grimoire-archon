package usecase

import (
	"fmt"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// CleanUsecase cleanのユースケース
type CleanUsecase struct {
	archonCfg *domain.ArchonConfig
	gameCfg   *domain.GameConfig
	fs        FileSystem
	cli       Cli
}

// NewCleanUsecase cleanのユースケースのインスタンスを生成する
func NewCleanUsecase(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, fs FileSystem, cli Cli) *CleanUsecase {
	return &CleanUsecase{
		archonCfg: archonCfg,
		gameCfg:   gameCfg,
		fs:        fs,
		cli:       cli,
	}
}

// Execute cleanの実行
func (u *CleanUsecase) Execute() error {
	// コンフィグのチェック
	if err := u.checkPreClean(); err != nil {
		return err
	}

	// ユーザに確認
	fmt.Printf("%s の削除処理を実行します...\n", u.gameCfg.Name)
	err := u.fs.ClearDirectoryContents(u.gameCfg.InstallDir)
	if err != nil {
		return fmt.Errorf("削除処理に失敗しました: %w", err)
	}

	return nil
}

// checkPreClean cleanの処理前チェック
func (u *CleanUsecase) checkPreClean() error {
	if u.archonCfg == nil {
		return fmt.Errorf("archonのコンフィグが定義されていません。")
	}
	if u.gameCfg.InstallDir == "" {
		return fmt.Errorf("インストールディレクトリが設定されていません。")
	}

	return nil
}
