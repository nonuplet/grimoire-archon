package config

// Config yaml全体の構造
type Config struct {
	Games  map[string]*GameConfig `yaml:"games,omitempty"`
	Archon ArchonConfig           `yaml:"archon"`
}

// ArchonConfig Archonの構成
type ArchonConfig struct {
	BackupDir string `yaml:"backup_dir"`
}

// GameConfig ゲームのコンフィグ
type GameConfig struct {
	Run          *RunConfig   `yaml:"run,omitempty"`
	Steam        *SteamConfig `yaml:"steam,omitempty"`
	Name         string       `yaml:"name"`
	InstallDir   string       `yaml:"install_dir"`
	ServerConfig string       `yaml:"server_config,omitempty"`
	BackupFiles  []string     `yaml:"backup_files,omitempty"`
}

// RunConfig ゲームの実行構成
type RunConfig struct {
	Command string   `yaml:"command"`
	Envs    []string `yaml:"envs,omitempty"`
}

// SteamConfig ゲームのSteam関連情報
type SteamConfig struct {
	Platform string `yaml:"platform,omitempty"`
	AppID    int    `yaml:"app_id"`
}
