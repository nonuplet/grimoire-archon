package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// ClearDirectoryContents ディレクトリ内のすべてのファイルを削除
func ClearDirectoryContents(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("ディレクトリの読み取りに失敗しました: %w", err)
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return fmt.Errorf("ファイルの削除に失敗しました: %w", err)
		}
	}
	return nil
}
