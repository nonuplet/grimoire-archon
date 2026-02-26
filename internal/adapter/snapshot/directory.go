package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nonuplet/grimoire-archon/internal/domain"
	"github.com/nonuplet/grimoire-archon/internal/infra/cli"
	"github.com/nonuplet/grimoire-archon/internal/infra/filesystem"
)

// CheckAndCreateSnapshotDir バックアップ先ディレクトリの存在確認と作成
func CheckAndCreateSnapshotDir(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig) error {
	snapshotPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	if _, err := filesystem.GetInfo(snapshotPath); os.IsNotExist(err) {
		// Ask
		ok, askErr := cli.AskYesNo(os.Stdin, fmt.Sprintf("バックアップ用ディレクトリ '%s' が存在しません。作成しますか?", snapshotPath), true)
		if askErr != nil {
			return fmt.Errorf("バックアップ用ディレクトリの作成確認に失敗しました: %w", askErr)
		}

		if !ok {
			return fmt.Errorf("バックアップ用ディレクトリの作成がキャンセルされました。")
		}

		// 処理
		if err := filesystem.MkdirAll(snapshotPath, 0o755); err != nil {
			return fmt.Errorf("バックアップ用ディレクトリの作成に失敗しました: %w", err)
		}

		fmt.Printf("バックアップ用ディレクトリ '%s' を作成しました。\n", snapshotPath)
	}

	return nil
}
