package config

// RuntimeEnv ゲームの実行環境
type RuntimeEnv string

const (
	// RuntimeEnvNative はアプリケーションがlinux向けバイナリであることを示します。
	RuntimeEnvNative RuntimeEnv = "native"
	// RuntimeEnvWine はアプリケーションをwineで動かすことを示します。
	RuntimeEnvWine RuntimeEnv = "wine"
	// RuntimeEnvProton はアプリケーションをprotonで動かすことを示します。
	RuntimeEnvProton RuntimeEnv = "proton"
)

// Config yaml全体の構造
type Config struct {
	Games  map[string]*GameConfig `yaml:"games,omitempty"`
	Archon *ArchonConfig          `yaml:"archon"`
}

// ArchonConfig Archonの構成
type ArchonConfig struct {
	BackupDir   string `yaml:"backup_dir"`
	AppdataDir  string `yaml:"appdata_dir,omitempty"`
	DocumentDir string `yaml:"document_dir,omitempty"`
}

// GameConfig ゲームのコンフィグ
type GameConfig struct {
	Run           *RunConfig          `yaml:"run,omitempty"`
	Steam         *SteamConfig        `yaml:"steam,omitempty"`
	BackupTargets *BackupTargetConfig `yaml:"backup_targets,omitempty"`
	RuntimeEnv    RuntimeEnv          `yaml:"runtime_env,omitempty"`
	Name          string              `yaml:"name"`
	InstallDir    string              `yaml:"install_dir"`
	ServerConfig  string              `yaml:"server_config,omitempty"`
}

// BackupTargetConfig バックアップ対象の構成
type BackupTargetConfig struct {
	InstallDir         []string `yaml:"install_dir,omitempty"`
	UserHome           []string `yaml:"user_home,omitempty"`
	WinAppdataLocal    []string `yaml:"appdata_local,omitempty"`
	WinAppdataLocalLow []string `yaml:"appdata_locallow,omitempty"`
	WinAppdataRoaming  []string `yaml:"appdata_roaming,omitempty"`
	WinDocuments       []string `yaml:"win_documents,omitempty"`
	Absolute           []string `yaml:"absolute,omitempty"`
}

// IsEmpty は全てのターゲットリストが空である場合に true を返します。
func (bt *BackupTargetConfig) IsEmpty() bool {
	if bt == nil {
		return true
	}

	return len(bt.InstallDir) == 0 &&
		len(bt.UserHome) == 0 &&
		len(bt.WinAppdataLocal) == 0 &&
		len(bt.WinAppdataLocalLow) == 0 &&
		len(bt.WinAppdataRoaming) == 0 &&
		len(bt.WinDocuments) == 0 &&
		len(bt.Absolute) == 0
}

// RunConfig ゲームの実行構成
type RunConfig struct {
	Command string   `yaml:"command"`
	Envs    []string `yaml:"envs,omitempty"`
}

// SteamConfig ゲームのSteam関連情報
type SteamConfig struct {
	Platform string `yaml:"platform,omitempty"`
	AppID    string `yaml:"app_id"`
}
