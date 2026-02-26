package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nonuplet/grimoire-archon/internal/domain"
)

// RestoreUsecase restoreのユースケース
type RestoreUsecase struct {
	archonCfg *domain.ArchonConfig
	gameCfg   *domain.GameConfig
	snapshot  Snapshot
	fs        FileSystem
}

// NewRestoreUsecase restoreのユースケースを作成
func NewRestoreUsecase(archonCfg *domain.ArchonConfig, gameCfg *domain.GameConfig, snapshot Snapshot, fs FileSystem) *RestoreUsecase {
	return &RestoreUsecase{
		archonCfg: archonCfg,
		gameCfg:   gameCfg,
		snapshot:  snapshot,
		fs:        fs,
	}
}

// Execute restoreの実行
func (u *RestoreUsecase) Execute(zipPath string) error {
	if _, err := u.fs.Stat(zipPath); err != nil {
		return fmt.Errorf("アーカイブファイル %s が見つかりません。", zipPath)
	}

	if err := u.checkPreRestore(zipPath); err != nil {
		return err
	}

	if err := u.snapshot.CheckAndCreateSnapshotDir(); err != nil {
		return fmt.Errorf("展開用ディレクトリの作成に失敗しました: %w", err)
	}

	if err := u.restoreSnapshot(zipPath); err != nil {
		return err
	}

	return nil
}

// checkPreRestore restore前チェック
func (u *RestoreUsecase) checkPreRestore(zipPath string) error {
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

	if !u.fs.IsZipFile(zipPath) {
		return fmt.Errorf("指定したファイル %s はzipファイルではありません", zipPath)
	}

	return nil
}

func (u *RestoreUsecase) restoreSnapshot(zipPath string) error {
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

	// unzipしたあとのディレクトリ指定に使う
	archiveName := strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))
	archiveDir := filepath.Join(tmpDir, archiveName)

	// <backup_dir>/<game_name>/tmp/ に展開
	fmt.Printf("zipファイル '%s' を展開しています... \n", filepath.Join(u.archonCfg.BackupDir, u.gameCfg.Name))
	if err := u.fs.Unzip(zipPath, tmpDir); err != nil {
		return fmt.Errorf("zipファイルの展開に失敗しました: %w", err)
	}

	// リストア
	if err := u.snapshot.RestoreFromTmp(archiveDir); err != nil {
		return fmt.Errorf("リストアに失敗しました: %w", err)
	}

	return nil
}
