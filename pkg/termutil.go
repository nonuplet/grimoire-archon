package pkg

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AskYesNo ユーザーにYes/Noの選択を促す
func AskYesNo(r io.Reader, question string, defaultYes bool) (bool, error) {
	reader := bufio.NewReader(r)
	if defaultYes {
		fmt.Printf("%s [Y/n]: ", question)
	} else {
		fmt.Printf("%s [y/N]: ", question)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("askでエラーが発生しました: %w", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))

	if defaultYes {
		return input == "" || input == "y" || input == "yes", nil
	}
	return !(input == "" || input == `n` || input == "no"), nil
}

// CopyFileOrDir ファイルまたはディレクトリをコピー
func CopyFileOrDir(src, dst string) error {
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
	return copyFile(dst, src, info.Mode())
}

// copyFile 単一ファイルのコピー
func copyFile(dst, src string, mode os.FileMode) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("コピー元 %s が開けません: %w", src, err)
	}
	defer func(source *os.File) {
		srcErr := source.Close()
		if srcErr != nil {
			fmt.Fprintf(os.Stderr, "ファイルのクローズに失敗しました: %v", err)
			return
		}
	}(source)

	// 出力先の親ディレクトリが存在しない場合はエラー
	if _, parentErr := os.Stat(filepath.Dir(dst)); os.IsNotExist(parentErr) {
		return fmt.Errorf("出力先の親ディレクトリが存在しません: %s", filepath.Dir(dst))
	}

	// 元のファイルと同じ権限で作成
	destination, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("ファイルが開けませんでした: %w", err)
	}
	defer func(destination *os.File) {
		dstErr := destination.Close()
		if dstErr != nil {
			fmt.Fprintf(os.Stderr, "%s をクローズできませんでした", err)
		}
	}(destination)

	_, err = io.Copy(destination, source)
	return fmt.Errorf("ファイルのコピーに失敗しました: %w", err)
}

// GetTimestamp タイムスタンプの取得
func GetTimestamp() string {
	return time.Now().Format("20060102_150405")
}

// Zip ディレクトリを圧縮
func Zip(src, dst string) error {
	zipfile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("ZIPファイルの作成に失敗しました: %w", err)
	}
	defer func(zipfile *os.File) {
		err := zipfile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ZIPファイルのクローズに失敗しました: %v", err)
		}
	}(zipfile)

	w := zip.NewWriter(zipfile)
	defer func(w *zip.Writer) {
		err := w.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ZIPファイルのクローズに失敗しました: %v", err)
		}
	}(w)

	if err := w.AddFS(os.DirFS(src)); err != nil {
		return fmt.Errorf("zip圧縮時、ディレクトリの追加に失敗しました: %w", err)
	}

	return nil
}
