package appversion

import (
	"runtime/debug"
	"sync"
)

// Version はアプリケーションのバージョンです。buildinfoから取得します
var Version = sync.OnceValue(func() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return buildInfo.Main.Version
})
