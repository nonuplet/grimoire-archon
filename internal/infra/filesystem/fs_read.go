package filesystem

import (
	"fmt"
	"os"
	"time"
)

// Stat はファイルが存在するかを確認します。
func (f *FileSystem) Stat(path string) (os.FileInfo, error) {
	path, err := f.getAbsolutePath(path)
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
func (f *FileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	path, err := f.getAbsolutePath(path)
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
func (f *FileSystem) ReadFile(path string) ([]byte, error) {
	path, err := f.getAbsolutePath(path)
	if err != nil {
		return nil, fmt.Errorf("絶対パスの取得に失敗しました: %w", err)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ファイル %s の読み込みに失敗しました: %w", file, err)
	}

	return file, nil
}

// GetTimestamp タイムスタンプの取得
func (f *FileSystem) GetTimestamp() string {
	return time.Now().Format("20060102_150405")
}
