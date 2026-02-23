package usecase

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/nonuplet/grimoire-archon/internal/config"
)

// RestoreUsecase restoreのユースケース
type RestoreUsecase struct{}

// NewRestoreUsecase restoreユースケースのインスタンスを生成
func NewRestoreUsecase() *RestoreUsecase {
	return &RestoreUsecase{}
}

// Execute バックアップファイルからデータを復元する
func (u *RestoreUsecase) Execute(archonCfg *config.ArchonConfig, gameCfg *config.GameConfig) error {
	err := restore(zipFile)
	if err != nil {
		return fmt.Errorf("バックアップからのリストア中にエラーが発生しました: %w", err)
	}
	return nil
}

// restore バックアップファイルからデータを復元する
func restore(zipFile string, gameCfg *config.GameConfig) error {
	// ZIPを開く
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("zipファイルを開けませんでした: %w", err)
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "zipファイルのクローズに失敗しました")
		}
	}(r)

	// metadata.yaml を読み込む
	var yamlData []byte
	for _, f := range r.File {
		if f.Name == "metadata.yaml" {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("metadata.yamlを開けませんでした: %w", err)
			}
			yamlData, err = io.ReadAll(rc)
			yamlErr := rc.Close()
			if yamlErr != nil {
				return fmt.Errorf("metadata.yamlのクローズに失敗しました: %w", yamlErr)
			}
			if err != nil {
				return fmt.Errorf("metadata.yamlを読み込みませんでした: %w", err)
			}
			break
		}
	}
	if yamlData == nil {
		return fmt.Errorf("metadata.yamlがバックアップファイル内に見つかりませんでした。")
	}

	// yamlのパースとチェック
	var backupData config.BackupData
	if err := yaml.Unmarshal(yamlData, &backupData); err != nil {
		return fmt.Errorf("metadata.yamlのパースに失敗しました: %w", err)
	}
	if len(backupData.Files) < 1 {
		return fmt.Errorf("metadata.yamlにバックアップの対応表が見つかりません。")
	}

	for _, entry := range backupData.Files {
	}

	// BackupPath -> OriginalPath のマップを作成
	pathMap := make(map[string]string, len(backupData.Files))
	for _, entry := range backupData.Files {
		pathMap[entry.BackupPath] = entry.OriginalPath
	}

	// ZIP内のファイルを復元
	for _, f := range r.File {
		if f.Name == "metadata.yaml" {
			continue
		}

		originalPath, ok := pathMap[f.Name]
		if !ok {
			return fmt.Errorf("対応するファイルがzipファイル内に見つかりません: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(originalPath, f.Mode()); err != nil {
				return fmt.Errorf("ディレクトリ %s を作成できませんでした: %w", originalPath, err)
			}
			continue
		}

		// 親ディレクトリを作成
		if err := os.MkdirAll(filepath.Dir(originalPath), 0o755); err != nil {
			return fmt.Errorf("%s の親ディレクトリを作成できませんでした: %w", originalPath, err)
		}

		// ファイルを書き出す
		if err := extractFile(f, originalPath); err != nil {
			return fmt.Errorf("ファイル %s の展開に失敗しました: %w", originalPath, err)
		}
	}

	return nil
}

func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func(rc io.ReadCloser) {
		err := rc.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s のクローズに失敗しました", f.Name)
		}
	}(rc)

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s のクローズに失敗しました", f.Name)
		}
	}(out)

	_, err = io.Copy(out, rc)
	return err
}
