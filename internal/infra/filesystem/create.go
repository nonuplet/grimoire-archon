package filesystem

import (
	"fmt"
	"os"
)

// MkdirAll は指定されたパスにディレクトリを作成します。
func MkdirAll(path string, perm os.FileMode) error {
	path, err := GetAbsolutePath(path)
	if err != nil {
		return fmt.Errorf("ディレクトリ (%s) のパス取得に失敗しました: %w", path, err)
	}

	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("ディレクトリ (%s) の作成に失敗しました: %w", path, err)
	}

	return nil
}

// WriteFile は指定されたパスにファイルを作成し、データを書き込みます。
func WriteFile(path string, data []byte, perm os.FileMode) error {
	path, err := GetAbsolutePath(path)
	if err != nil {
		return fmt.Errorf("ファイル (%s) のパス取得に失敗しました: %w", path, err)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("ファイル (%s) の書き込みに失敗しました: %w", path, err)
	}

	return nil
}
