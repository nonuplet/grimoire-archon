package snapshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nonuplet/grimoire-archon/internal/config"
	"github.com/nonuplet/grimoire-archon/internal/infra/cli"
	"github.com/nonuplet/grimoire-archon/internal/infra/storage"
)

// CheckAndCreateSnapshotDir バックアップ先ディレクトリの存在確認と作成
func CheckAndCreateSnapshotDir(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	snapshotPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	if _, err := storage.GetInfo(snapshotPath); os.IsNotExist(err) {
		// Ask
		ok, askErr := cli.AskYesNo(os.Stdin, fmt.Sprintf("バックアップ用ディレクトリ '%s' が存在しません。作成しますか?", snapshotPath), true)
		if askErr != nil {
			return fmt.Errorf("バックアップ用ディレクトリの作成確認に失敗しました: %w", askErr)
		}

		if !ok {
			return fmt.Errorf("バックアップ用ディレクトリの作成がキャンセルされました。")
		}

		// 処理
		if err := storage.MkdirAll(snapshotPath, 0o755); err != nil {
			return fmt.Errorf("バックアップ用ディレクトリの作成に失敗しました: %w", err)
		}

		fmt.Printf("バックアップ用ディレクトリ '%s' を作成しました。\n", snapshotPath)
	}

	return nil
}
