package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckAndCreateSnapshotDir バックアップ先ディレクトリの存在確認と作成
func (snap Snapshot) CheckAndCreateSnapshotDir() error {
	snapshotPath := filepath.Join(snap.archonCfg.BackupDir, snap.gameCfg.Name)

	if _, err := snap.fs.Stat(snapshotPath); os.IsNotExist(err) {
		// Ask
		ok, askErr := snap.cli.AskYesNo(os.Stdin, fmt.Sprintf("バックアップ用ディレクトリ '%s' が存在しません。作成しますか?", snapshotPath), true)
		if askErr != nil {
			return fmt.Errorf("バックアップ用ディレクトリの作成確認に失敗しました: %w", askErr)
		}

		if !ok {
			return fmt.Errorf("バックアップ用ディレクトリの作成がキャンセルされました。")
		}

		// 処理
		if err := snap.fs.MkdirAll(snapshotPath, 0o755); err != nil {
			return fmt.Errorf("バックアップ用ディレクトリの作成に失敗しました: %w", err)
		}

		fmt.Printf("バックアップ用ディレクトリ '%s' を作成しました。\n", snapshotPath)
	}

	return nil
}
