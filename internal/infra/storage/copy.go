package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CopyFileOrDir ファイルまたはディレクトリをコピー
func CopyFileOrDir(src, dst string) error {
	// 絶対パスに変換
	src, err := GetAbsolutePath(src)
	if err != nil {
		return fmt.Errorf("コピー元パスの取得: %w", err)
	}
	dst, err = GetAbsolutePath(dst)
	if err != nil {
		return fmt.Errorf("コピー先パスの取得: %w", err)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("コピー元 %s が取得できません", src)
	}

	// Directory
	if info.IsDir() {
		fmt.Printf("ディレクトリをコピー中: %s -> %s\n", src, dst)
		if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
			return fmt.Errorf("ディレクトリのコピーに失敗しました: %w", err)
		}
	}

	// File
	fmt.Printf("ファイルをコピー中: %s -> %s\n", src, dst)
	return copyFile(dst, src)
}

// CopyFile はファイル src を dst へコピーします。
// dst はコピー先のファイルパスです。
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("コピー元が取得できません (%s): %w", src, err)
	}
	defer func(in *os.File) {
		inErr := in.Close()
		if inErr != nil {
			fmt.Fprintf(os.Stderr, "コピー元のクローズに失敗しました %v", err)
		}
	}(in)

	if mkdirErr := os.MkdirAll(filepath.Dir(dst), 0o755); mkdirErr != nil {
		return fmt.Errorf("親ディレクトリの作成に失敗しました (%s): %w", filepath.Dir(dst), err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました (%s): %w", dst, err)
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "コピー先のクローズに失敗しました %v", err)
		}
	}(out)

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("ファイルのコピーに失敗しました (%s -> %s): %w", src, dst, err)
	}
	return nil
}

// GetTimestamp タイムスタンプの取得
func GetTimestamp() string {
	return time.Now().Format("20060102_150405")
}
