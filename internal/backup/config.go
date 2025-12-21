package backup

import "time"

type Config struct {
	ContainerName string
	DatabaseName  string
	DatabaseUser  string
	OutputDir     string
	Timestamp     time.Time
}

type RestoreConfig struct {
	ContainerName string
	DatabaseName  string
	DatabaseUser  string
	BackupPath    string
	DropExisting  bool
}

type VerifyConfig struct {
	SourceContainer string
	TargetContainer string
	DatabaseName    string
	DatabaseUser    string
}
