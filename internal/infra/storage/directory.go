package storage

import (
	"fmt"
	"os"
)

// CreateDirIfNotExist ディレクトリをチェックし、なければ作成
func CreateDirIfNotExist(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました (%s): %w", dirPath, err)
		}
	}
	return nil
}
