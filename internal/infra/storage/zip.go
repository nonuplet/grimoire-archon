package storage

import (
	"archive/zip"
	"fmt"
	"os"
)

// ZipDir はdirの内容をzipFilePathに圧縮します。
func ZipDir(dir, zipFilePath string) error {
	dir, err := GetAbsolutePath(dir)
	if err != nil {
		return fmt.Errorf("ディレクトリパスの取得: %w", err)
	}
	zipFilePath, err = GetAbsolutePath(zipFilePath)
	if err != nil {
		return fmt.Errorf("zipファイルパスの取得: %w", err)
	}

	f, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("zipファイル %s の作成に失敗しました: %w", zipFilePath, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "zipファイル %s のクローズに失敗しました: %v\n", zipFilePath, err)
		}
	}(f)

	zw := zip.NewWriter(f)
	defer func(zw *zip.Writer) {
		err := zw.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "zipファイル %s のクローズに失敗しました: %v\n", zipFilePath, err)
		}
	}(zw)

	if err := zw.AddFS(os.DirFS(dir)); err != nil {
		return fmt.Errorf("add fs %s: %w", dir, err)
	}
	return nil
}
