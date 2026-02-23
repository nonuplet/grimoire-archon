package usecase

import "context"

// SteamCmd steamcmdのインターフェース
type SteamCmd interface {
	Update(ctx context.Context, appID int, installDir, platform string) error
}
