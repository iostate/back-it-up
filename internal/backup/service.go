package backup

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type Service struct {
	dockerSvc DockerService
}

type DockerService interface {
	VerifyContainer(containerName string) error
	Exec(containerName string, command []string) ([]byte, error)
}

func NewService(dockerSvc DockerService) *Service {
	return &Service{
		dockerSvc: dockerSvc,
	}
}

// Backup performs a PostgreSQL backup and compresses it to tar.gz
func (s *Service) Backup(cfg Config) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("%s_%s.sql.gz",
		cfg.DatabaseName,
		cfg.Timestamp.Format("2006_01_02_15_04_05"))
	outputPath := filepath.Join(cfg.OutputDir, filename)

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Execute pg_dump via docker exec
	cmd := exec.Command("docker", "exec", cfg.ContainerName,
		"pg_dump", "-U", cfg.DatabaseUser, cfg.DatabaseName)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start pg_dump: %w", err)
	}

	// Copy and compress the output
	if _, err := io.Copy(gzWriter, stdout); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("pg_dump failed: %w", err)
	}

	return outputPath, nil
}
