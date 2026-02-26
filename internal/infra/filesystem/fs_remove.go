package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// ClearDirectoryContents ディレクトリ内のすべてのファイルを削除
func (f *FileSystem) ClearDirectoryContents(path string) error {
	path, err := f.getAbsolutePath(path)
	if err != nil {
		return fmt.Errorf("絶対パスの取得: %w", err)
	}

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

// RemoveAll は指定されたパスを削除します。
func (f *FileSystem) RemoveAll(path string) error {
	path, err := f.getAbsolutePath(path)
	if err != nil {
		return fmt.Errorf("%s のパス取得に失敗しました: %w", path, err)
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("%s の削除に失敗しました: %w", path, err)
	}
	return nil
}
