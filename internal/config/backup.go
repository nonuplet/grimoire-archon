package config

// BackupData バックアップデータのyaml構造
type BackupData struct {
	Files []BackupConfig `yaml:"files"`
}

// BackupConfig バックアップと元のパスの相対関係
type BackupConfig struct {
	OriginalPath string `yaml:"original_path"`
	BackupPath   string `yaml:"backup_path"`
}
