package filesystem

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ZipBomb検出用: zipファイルのバッファサイズ 10MB
	bufSize int64 = 10 * 1024 * 1024
	// ZipBomb検出用: zipファイルの最大展開サイズ 1TB
	maxDecompressLimit uint64 = 1 << 40
)

// IsZipFile は指定したファイルがzipファイルかどうか判定します。
func (f *FileSystem) IsZipFile(path string) bool {
	path, err := f.getAbsolutePath(path)
	if err != nil {
		return false
	}

	// 拡張子のチェック
	if path[len(path)-4:] != ".zip" {
		return false
	}

	// ファイルを開いてマジックバイトの確認
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "zipファイル %s のクローズに失敗しました: %v\n", path, err)
		}
	}(file)

	buf := make([]byte, 4)
	if _, err := file.Read(buf); err != nil {
		return false
	}

	// zipのマジックバイトは 50 4b 03 04
	if len(buf) < 4 {
		return false
	}
	return buf[0] == 0x50 && buf[1] == 0x4B && buf[2] == 0x03 && buf[3] == 0x04
}

// Zip はdirの内容をzipFilePathに圧縮します。
func (f *FileSystem) Zip(dir, zipFilePath string) error {
	dir, err := f.getAbsolutePath(dir)
	if err != nil {
		return fmt.Errorf("ディレクトリパスの取得: %w", err)
	}
	zipFilePath, err = f.getAbsolutePath(zipFilePath)
	if err != nil {
		return fmt.Errorf("zipファイルパスの取得: %w", err)
	}

	file, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("zipファイル %s の作成に失敗しました: %w", zipFilePath, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "zipファイル %s のクローズに失敗しました: %v\n", zipFilePath, err)
		}
	}(file)

	zw := zip.NewWriter(file)
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

// Unzip は zipFilePath を dstDir に展開します。
func (f *FileSystem) Unzip(zipFilePath, dstDir string) error {
	// 絶対パスへ変換
	zipFilePath, err := f.getAbsolutePath(zipFilePath)
	if err != nil {
		return fmt.Errorf("zipファイルパスの取得エラー: %w", err)
	}
	dstDir, err = f.getAbsolutePath(dstDir)
	if err != nil {
		return fmt.Errorf("展開先ディレクトリパスの取得エラー: %w", err)
	}

	// ZIPファイルを開く
	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("ZIPファイルを開けませんでした: %w", err)
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "zipファイル %s のクローズに失敗しました: %v\n", zipFilePath, err)
		}
	}(r)

	// 展開先ディレクトリが存在しない場合は作成する
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("展開先ディレクトリの作成に失敗しました: %w", err)
	}

	// ZIP内の各ファイル・ディレクトリを順番に処理
	for _, file := range r.File {
		if err := extractAndWriteFile(dstDir, file); err != nil {
			return err
		}
	}

	return nil
}

// extractAndWriteFile は単一のファイルを安全に解凍・書き出しします
// ※ループ内で defer を安全に実行するために関数を分離しています
func extractAndWriteFile(dstDir string, file *zip.File) error {
	// 展開先のフルパスを構築
	fpath := filepath.Join(dstDir, filepath.Clean(file.Name))

	// Zip Slip脆弱性対策：展開先パスが指定ディレクトリ内にあるか確認
	if !strings.HasPrefix(fpath, filepath.Clean(dstDir)+string(os.PathSeparator)) {
		return fmt.Errorf("不正なファイルパスを検出しました (Zip Slip対策): %s", fpath)
	}

	// Zip Bomb対策: 異常な圧縮率のデータを弾く
	if isSuspiciousRatio(file) {
		return fmt.Errorf("圧縮率が異常なファイルを検出しました。zip bombの可能性があります: %s", file.Name)
	}

	// エントリがディレクトリの場合は作成して終了
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
			return fmt.Errorf("展開先ディレクトリの作成に失敗しました: %w", err)
		}
		return nil
	}

	// ファイルを配置する親ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
		return fmt.Errorf("展開先ディレクトリの作成に失敗しました: %w", err)
	}

	// ZIP内のファイルを開く
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("zip内のファイルを開けませんでした: %w", err)
	}
	defer func(rc io.ReadCloser) {
		rcErr := rc.Close()
		if rcErr != nil {
			fmt.Fprintf(os.Stderr, "zip内ファイルのクローズに失敗しました: %v\n", rcErr)
		}
	}(rc)

	// Zip Bomb対策2: サイズの上限を設定
	if file.UncompressedSize64 > maxDecompressLimit {
		return fmt.Errorf("圧縮率が異常なファイルを検出しました。zip bombの可能性があります: %s", file.Name)
	}

	// nolint:gosec // G115 size already checked
	limitedReader := io.LimitReader(rc, int64(file.UncompressedSize64)+bufSize)

	// 展開先のファイルを作成・オープン
	// f.Mode() で元のファイルの権限を引き継ぐ
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("展開先ファイル(%s)を開けませんでした: %w", fpath, err)
	}
	defer func(outFile *os.File) {
		outFileErr := outFile.Close()
		if outFileErr != nil {
			fmt.Fprintf(os.Stderr, "展開先ファイルのクローズに失敗しました: %v\n", outFileErr)
		}
	}(outFile)

	// ファイルの中身をコピー
	if _, err := io.Copy(outFile, limitedReader); err != nil {
		return fmt.Errorf("ファイルのコピーに失敗しました (%s): %w", file.Name, err)
	}

	return nil
}

const maxCompressionRatio = 100

// isSuspiciousRatio は圧縮率が異常かどうかを判定します
func isSuspiciousRatio(f *zip.File) bool {
	if f.CompressedSize64 == 0 {
		return false // 0除算を避ける
	}
	ratio := float64(f.UncompressedSize64) / float64(f.CompressedSize64)
	return ratio > maxCompressionRatio
}
