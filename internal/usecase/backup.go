package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/nonuplet/grimoire-archon/internal/app"
	"github.com/nonuplet/grimoire-archon/internal/config"
	"github.com/nonuplet/grimoire-archon/internal/infra/cli"
	"github.com/nonuplet/grimoire-archon/internal/infra/storage"
	"github.com/nonuplet/grimoire-archon/internal/snapshot"
)

// BackupUsecase backupのユースケース
type BackupUsecase struct{}

// NewBackupUsecase backupユースケースの生成
func NewBackupUsecase() *BackupUsecase {
	return &BackupUsecase{}
}

// Execute backupの実行
func (u *BackupUsecase) Execute(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	// 必要なコンフィグの情報があるかチェック
	if err := u.checkPreBackup(archonCfg, gameCfg); err != nil {
		return err
	}

	// バックアップディレクトリの存在確認と作成
	if err := u.checkAndCreateSnapshotDir(archonCfg, gameCfg); err != nil {
		return err
	}

	if err := u.createSnapshot(archonCfg, gameCfg); err != nil {
		return err
	}

	return nil
}

// checkPreBackup backupの処理前チェック
func (u *BackupUsecase) checkPreBackup(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	if archonCfg == nil {
		return fmt.Errorf("archonのコンフィグが定義されていません。")
	}
	if archonCfg.BackupDir == "" {
		return fmt.Errorf("バックアップ先が設定されていません。")
	}
	if gameCfg.InstallDir == "" {
		return fmt.Errorf("%s にインストール先が設定されていません。", gameCfg.Name)
	}

	// バックアップ指定したファイルの数を確認
	if gameCfg.BackupTargets.IsEmpty() {
		return fmt.Errorf("%s にバックアップの対象が指定されていません。", gameCfg.Name)
	}

	return nil
}

// checkAndCreateSnapshotDir バックアップ先ディレクトリの存在確認と作成
func (u *BackupUsecase) checkAndCreateSnapshotDir(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
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

// createSnapshot バックアップ処理の実行
func (u *BackupUsecase) createSnapshot(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	// バックアップ先パスの設定
	snapshotPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	// tmpフォルダを作成 作業が終わったら成功しても失敗しても消す
	tmpDir := filepath.Join(snapshotPath, "tmp")

	// tmpフォルダ以下に <gamename>_<timestamp> ディレクトリを作成 ここにデータを格納する
	archiveDir := filepath.Join(tmpDir, fmt.Sprintf("%s_%s", gameCfg.Name, storage.GetTimestamp()))
	defer func(path string) {
		err := storage.RemoveAll(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "一時ディレクトリの削除に失敗しました: %v", err)
		}
	}(tmpDir)

	entries, err := snapshot.CopyToTmp(archiveDir, archonCfg, gameCfg)
	if err != nil {
		return fmt.Errorf("バックアップファイルのコピーに失敗しました: %w", err)
	}

	// metadata.yamlの構築と保存
	meta := &snapshot.Metadata{
		Version:     snapshot.MetaVersion,
		Name:        gameCfg.Name,
		CreatedAt:   time.Now(),
		ToolVersion: app.Version,
		Os:          runtime.GOOS,
		Files:       entries,
	}
	if err := meta.Save(filepath.Join(archiveDir, "metadata.yaml")); err != nil {
		return fmt.Errorf("metadata.yamlの保存に失敗しました: %w", err)
	}

	// zipにする
	timestamp := storage.GetTimestamp()
	zipPath := filepath.Join(snapshotPath, fmt.Sprintf("%s_%s.zip", gameCfg.Name, timestamp))
	if err := storage.ZipDir(tmpDir, zipPath); err != nil {
		return fmt.Errorf("バックアップの圧縮に失敗しました: %w", err)
	}

	return nil
}
