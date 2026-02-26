package usecase

import (
	"context"
	"fmt"
	"os"

	"github.com/nonuplet/grimoire-archon/internal/domain"
	"github.com/nonuplet/grimoire-archon/internal/infra/filesystem"
)

// UpdateUsecase updateのユースケース
type UpdateUsecase struct {
	steamCmd SteamCmd
}

// NewUpdateUsecase UpdateUsecaseのインスタンスを生成
func NewUpdateUsecase(steamCmd SteamCmd) *UpdateUsecase {
	return &UpdateUsecase{steamCmd}
}

// Execute 更新処理を実行
// TODO: 現在はSteam経由のダウンロード以外は対応していません Minecraft や Terraria 対応はそのうちやる
func (u *UpdateUsecase) Execute(ctx context.Context, gameCfg *domain.GameConfig) error {
	// 処理前チェック
	if err := u.checkPreUpdate(gameCfg); err != nil {
		return err
	}

	// gameCfg.InstallDir がなかった場合ディレクトリを作成
	if _, err := filesystem.GetInfo(gameCfg.InstallDir); os.IsNotExist(err) {
		fmt.Printf("インストール先のディレクトリ %s を作成しています...\n", gameCfg.InstallDir)
		if dirErr := filesystem.MkdirAll(gameCfg.InstallDir, 0o750); dirErr != nil {
			return fmt.Errorf("インストールディレクトリの作成に失敗しました: %w", dirErr)
		}
	} else if err != nil {
		return fmt.Errorf("インストールディレクトリの確認に失敗しました: %w", err)
	}

	if err := u.steamCmd.Update(ctx, gameCfg.Steam.AppID, gameCfg.InstallDir, gameCfg.Steam.Platform); err != nil {
		return fmt.Errorf("更新に失敗しました: %w", err)
	}
	return nil
}

// checkPreUpdate アップデート実行前のチェック
func (u *UpdateUsecase) checkPreUpdate(gameCfg *domain.GameConfig) error {
	if err := u.steamCmd.Check(); err != nil {
		return fmt.Errorf("アップデートに失敗しました: %w", err)
	}
	if gameCfg.InstallDir == "" {
		return fmt.Errorf("インストールディレクトリが設定されていません")
	}
	if gameCfg.Steam == nil {
		return fmt.Errorf("steam設定が見つかりません")
	}
	if gameCfg.Steam.AppID == "" {
		return fmt.Errorf("steamのappIDが設定されていません")
	}

	return nil
}
