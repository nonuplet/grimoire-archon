package infra

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// SteamCmd steamcmdの操作
type SteamCmd struct{}

// Update 対象アプリのインストール/アップデート
func (s *SteamCmd) Update(ctx context.Context, appID int, installDir, platform string) error {
	// steamcmd コマンドのチェック
	if _, err := exec.LookPath("steamcmd"); err != nil {
		return fmt.Errorf("steamcmdコマンドが見つかりません: %w", err)
	}

	// 引数の構築
	var args []string
	if platform != "" {
		args = append(args, "+@sSteamCmdForcePlatformType", platform)
	}
	args = append(args,
		"+force_install_dir", installDir,
		"+login", "anonymous",
		"+app_update", fmt.Sprint(rune(appID)),
		"+quit",
	)

	// 実行
	cmd := exec.CommandContext(ctx, "steamcmd", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("steamcmdを使ったアップデートに失敗しました: %w", err)
	}

	return nil
}
