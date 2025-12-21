package backup

import "time"

type Config struct {
	ContainerName string
	DatabaseName  string
	DatabaseUser  string
	OutputDir     string
	Timestamp     time.Time
}
