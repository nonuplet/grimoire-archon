package usecase

import (
	"context"
	"fmt"
	"os"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// UpdateUsecase updateのユースケース
type UpdateUsecase struct {
	gameCfg  *domain.GameConfig
	steamCmd SteamCmd
	fs       FileSystem
}

// NewUpdateUsecase UpdateUsecaseのインスタンスを生成
func NewUpdateUsecase(gameCfg *domain.GameConfig, steamCmd SteamCmd, fs FileSystem) *UpdateUsecase {
	return &UpdateUsecase{
		gameCfg:  gameCfg,
		steamCmd: steamCmd,
		fs:       fs,
	}
}

// Execute 更新処理を実行
// TODO: 現在はSteam経由のダウンロード以外は対応していません Minecraft や Terraria 対応はそのうちやる
func (u *UpdateUsecase) Execute(ctx context.Context) error {
	// 処理前チェック
	if err := u.checkPreUpdate(); err != nil {
		return err
	}

	// u.gameCfg.InstallDir がなかった場合ディレクトリを作成
	if _, err := u.fs.Stat(u.gameCfg.InstallDir); os.IsNotExist(err) {
		fmt.Printf("インストール先のディレクトリ %s を作成しています...\n", u.gameCfg.InstallDir)
		if dirErr := u.fs.MkdirAll(u.gameCfg.InstallDir, 0o750); dirErr != nil {
			return fmt.Errorf("インストールディレクトリの作成に失敗しました: %w", dirErr)
		}
	} else if err != nil {
		return fmt.Errorf("インストールディレクトリの確認に失敗しました: %w", err)
	}

	if err := u.steamCmd.Update(ctx, u.gameCfg.Steam.AppID, u.gameCfg.InstallDir, u.gameCfg.Steam.Platform); err != nil {
		return fmt.Errorf("更新に失敗しました: %w", err)
	}
	return nil
}

// checkPreUpdate アップデート実行前のチェック
func (u *UpdateUsecase) checkPreUpdate() error {
	if err := u.steamCmd.Check(); err != nil {
		return fmt.Errorf("アップデートに失敗しました: %w", err)
	}
	if u.gameCfg.InstallDir == "" {
		return fmt.Errorf("インストールディレクトリが設定されていません")
	}
	if u.gameCfg.Steam == nil {
		return fmt.Errorf("steam設定が見つかりません")
	}
	if u.gameCfg.Steam.AppID == "" {
		return fmt.Errorf("steamのappIDが設定されていません")
	}

	return nil
}
