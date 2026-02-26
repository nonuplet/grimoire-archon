package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/nonuplet/grimoire-archon/internal/appversion"
	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// BackupUsecase backupのユースケース
type BackupUsecase struct {
	archonCfg *domain.ArchonConfig
	gameCfg   *domain.GameConfig
	snapshot  Snapshot
	fs        FileSystem
	cli       Cli
}

// NewBackupUsecase backupユースケースの生成
// nolint:lll // 初期化なので
func NewBackupUsecase(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, snapshot Snapshot, fs FileSystem, cli Cli) *BackupUsecase {
	return &BackupUsecase{
		archonCfg: archonCfg,
		gameCfg:   gameCfg,
		snapshot:  snapshot,
		fs:        fs,
		cli:       cli,
	}
}

// Execute backupの実行
func (u *BackupUsecase) Execute() error {
	// 必要なコンフィグの情報があるかチェック
	if err := u.checkPreBackup(); err != nil {
		return err
	}

	// バックアップディレクトリの存在確認と作成
	if err := u.snapshot.CheckAndCreateSnapshotDir(); err != nil {
		return fmt.Errorf("バックアップディレクトリ作成に失敗しました: %w", err)
	}

	if err := u.createSnapshot(); err != nil {
		return err
	}

	return nil
}

// checkPreBackup backupの処理前チェック
func (u *BackupUsecase) checkPreBackup() error {
	if u.archonCfg == nil {
		return fmt.Errorf("archonのコンフィグが定義されていません。")
	}
	if u.archonCfg.BackupDir == "" {
		return fmt.Errorf("バックアップ先が設定されていません。")
	}
	if u.gameCfg.InstallDir == "" {
		return fmt.Errorf("%s にインストール先が設定されていません。", u.gameCfg.Name)
	}

	// バックアップ指定したファイルの数を確認
	if u.gameCfg.BackupTargets.IsEmpty() {
		return fmt.Errorf("%s にバックアップの対象が指定されていません。", u.gameCfg.Name)
	}

	return nil
}

// createSnapshot バックアップ処理の実行
func (u *BackupUsecase) createSnapshot() error {
	// バックアップ先パスの設定
	snapshotPath := filepath.Join(u.archonCfg.BackupDir, u.gameCfg.Name)

	// tmpフォルダを作成 作業が終わったら成功しても失敗しても消す
	tmpDir := filepath.Join(snapshotPath, "tmp")
	if err := u.fs.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("一時ディレクトリの作成に失敗しました: %w", err)
	}
	defer func(path string) {
		err := u.fs.RemoveAll(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "一時ディレクトリの削除に失敗しました: %v", err)
		}
	}(tmpDir)

	// tmpフォルダ以下に <gamename>_<timestamp> ディレクトリを作成 ここにデータを格納する
	archiveName := fmt.Sprintf("%s_%s", u.gameCfg.Name, u.fs.GetTimestamp())
	archiveDir := filepath.Join(tmpDir, archiveName)

	entries, err := u.snapshot.CopyToTmp(archiveDir)
	if err != nil {
		return fmt.Errorf("バックアップファイルのコピーに失敗しました: %w", err)
	}

	// metadata.yamlの構築と保存
	meta := &domain.Metadata{
		Version:     domain.MetaVersion,
		Name:        u.gameCfg.Name,
		CreatedAt:   time.Now(),
		ToolVersion: appversion.Version(),
		Os:          runtime.GOOS,
		Files:       entries,
	}
	if err := u.snapshot.SaveMetaData(filepath.Join(archiveDir, "metadata.yaml"), meta); err != nil {
		return fmt.Errorf("metadata.yamlの保存に失敗しました: %w", err)
	}

	// zipにする
	zipPath := filepath.Join(snapshotPath, fmt.Sprintf("%s.zip", archiveName))
	if err := u.fs.Zip(tmpDir, zipPath); err != nil {
		return fmt.Errorf("バックアップの圧縮に失敗しました: %w", err)
	}

	return nil
}
