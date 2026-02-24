package storage

import (
	"fmt"
	"os"
)

// GetInfo はファイルが存在するかを確認します。
func GetInfo(path string) (os.FileInfo, error) {
	path, err := GetAbsolutePath(path)
	if err != nil {
		return nil, fmt.Errorf("絶対パスの取得に失敗しました: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%s が見つかりません", path)
	}

	return info, nil
}

// ReadDir は指定されたパスのディレクトリエントリを読み取ります。
func ReadDir(path string) ([]os.DirEntry, error) {
	path, err := GetAbsolutePath(path)
	if err != nil {
		return nil, fmt.Errorf("絶対パスの取得に失敗しました: %w", err)
	}

	entry, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("ディレクトリ '%s' の読み取りに失敗しました: %w", path, err)
	}
	return entry, nil
}

// ReadFile は指定されたパスのファイルを読み込みます。
func ReadFile(path string) ([]byte, error) {
	path, err := GetAbsolutePath(path)
	if err != nil {
		return nil, fmt.Errorf("絶対パスの取得に失敗しました: %w", err)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ファイル %s の読み込みに失敗しました: \n", file, err)
	}

	return file, nil
}
