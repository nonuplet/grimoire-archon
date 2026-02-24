package usecase

import "context"

// SteamCmd steamcmdのインターフェース
type SteamCmd interface {
	Check() error
	Update(ctx context.Context, appID string, installDir, platform string) error
}
