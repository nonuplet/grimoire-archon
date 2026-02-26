package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	snapshot2 "github.com/nonuplet/grimoire-archon/internal/adapter/snapshot"
	"github.com/nonuplet/grimoire-archon/internal/app"
	"github.com/nonuplet/grimoire-archon/internal/domain"
	"github.com/nonuplet/grimoire-archon/internal/infra/filesystem"
)

// BackupUsecase backupのユースケース
type BackupUsecase struct{}

// NewBackupUsecase backupユースケースの生成
func NewBackupUsecase() *BackupUsecase {
	return &BackupUsecase{}
}

// Execute backupの実行
func (u *BackupUsecase) Execute(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig) error {
	// 必要なコンフィグの情報があるかチェック
	if err := u.checkPreBackup(archonCfg, gameCfg); err != nil {
		return err
	}

	// バックアップディレクトリの存在確認と作成
	if err := snapshot2.CheckAndCreateSnapshotDir(archonCfg, gameCfg); err != nil {
		return fmt.Errorf("バックアップディレクトリ作成に失敗しました: %w", err)
	}

	if err := u.createSnapshot(archonCfg, gameCfg); err != nil {
		return err
	}

	return nil
}

// checkPreBackup backupの処理前チェック
func (u *BackupUsecase) checkPreBackup(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig) error {
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

// createSnapshot バックアップ処理の実行
func (u *BackupUsecase) createSnapshot(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig) error {
	// バックアップ先パスの設定
	snapshotPath := filepath.Join(archonCfg.BackupDir, gameCfg.Name)

	// tmpフォルダを作成 作業が終わったら成功しても失敗しても消す
	tmpDir := filepath.Join(snapshotPath, "tmp")
	if err := filesystem.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗しました: %w", err)
	}
	defer func(path string) {
		err := filesystem.RemoveAll(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "一時ディレクトリの削除に失敗しました: %v", err)
		}
	}(tmpDir)

	// tmpフォルダ以下に <gamename>_<timestamp> ディレクトリを作成 ここにデータを格納する
	archiveName := fmt.Sprintf("%s_%s", gameCfg.Name, filesystem.GetTimestamp())
	archiveDir := filepath.Join(tmpDir, archiveName)

	entries, err := snapshot2.CopyToTmp(archiveDir, archonCfg, gameCfg)
	if err != nil {
		return fmt.Errorf("バックアップファイルのコピーに失敗しました: %w", err)
	}

	// metadata.yamlの構築と保存
	meta := &domain.Metadata{
		Version:     domain.MetaVersion,
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
	zipPath := filepath.Join(snapshotPath, fmt.Sprintf("%s.zip", archiveName))
	if err := filesystem.ZipDir(tmpDir, zipPath); err != nil {
		return fmt.Errorf("バックアップの圧縮に失敗しました: %w", err)
	}

	return nil
}
