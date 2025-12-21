package backup

import (
	"compress/gzip"
	"crypto/md5"
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

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start pg_dump: %w", err)
	}

	// Copy and compress the output
	if _, err := io.Copy(gzWriter, stdout); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	// Read any error output
	stderrOutput, _ := io.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		if len(stderrOutput) > 0 {
			return "", fmt.Errorf("pg_dump failed: %w\nError output: %s", err, string(stderrOutput))
		}
		return "", fmt.Errorf("pg_dump failed: %w", err)
	}

	return outputPath, nil
}

// Restore restores a PostgreSQL backup from a compressed file
func (s *Service) Restore(cfg RestoreConfig) error {
	// Verify container exists
	if err := s.dockerSvc.VerifyContainer(cfg.ContainerName); err != nil {
		return fmt.Errorf("container verification failed: %w", err)
	}

	// Open backup file
	backupFile, err := os.Open(cfg.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer backupFile.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Drop existing database if requested
	if cfg.DropExisting {
		dropCmd := exec.Command("docker", "exec", cfg.ContainerName,
			"psql", "-U", cfg.DatabaseUser, "-d", "template1", "-c",
			fmt.Sprintf("DROP DATABASE IF EXISTS %s;", cfg.DatabaseName))
		if output, err := dropCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to drop database: %w\nOutput: %s", err, string(output))
		}
	}

	// Create database
	createCmd := exec.Command("docker", "exec", cfg.ContainerName,
		"psql", "-U", cfg.DatabaseUser, "-d", "template1", "-c",
		fmt.Sprintf("CREATE DATABASE %s;", cfg.DatabaseName))
	if output, err := createCmd.CombinedOutput(); err != nil {
		// Ignore error if database already exists
		if !cfg.DropExisting {
			fmt.Printf("Warning: Database may already exist: %s\n", string(output))
		} else {
			return fmt.Errorf("failed to create database: %w\nOutput: %s", err, string(output))
		}
	}

	// Restore via psql
	restoreCmd := exec.Command("docker", "exec", "-i", cfg.ContainerName,
		"psql", "-U", cfg.DatabaseUser, "-d", cfg.DatabaseName)

	stdin, err := restoreCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stderr, err := restoreCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := restoreCmd.Start(); err != nil {
		return fmt.Errorf("failed to start restore: %w", err)
	}

	// Copy decompressed backup to psql
	if _, err := io.Copy(stdin, gzReader); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}
	stdin.Close()

	// Read any error output
	stderrOutput, _ := io.ReadAll(stderr)

	if err := restoreCmd.Wait(); err != nil {
		if len(stderrOutput) > 0 {
			return fmt.Errorf("restore failed: %w\nError output: %s", err, string(stderrOutput))
		}
		return fmt.Errorf("restore failed: %w", err)
	}

	return nil
}

// Verify compares two databases to ensure they contain the same data
func (s *Service) Verify(cfg VerifyConfig) (bool, error) {
	// Verify both containers exist
	if err := s.dockerSvc.VerifyContainer(cfg.SourceContainer); err != nil {
		return false, fmt.Errorf("source container verification failed: %w", err)
	}
	if err := s.dockerSvc.VerifyContainer(cfg.TargetContainer); err != nil {
		return false, fmt.Errorf("target container verification failed: %w", err)
	}

	// Get checksums of both databases
	sourceChecksum, err := s.getDatabaseChecksum(cfg.SourceContainer, cfg.DatabaseName, cfg.DatabaseUser)
	if err != nil {
		return false, fmt.Errorf("failed to get source checksum: %w", err)
	}

	targetChecksum, err := s.getDatabaseChecksum(cfg.TargetContainer, cfg.DatabaseName, cfg.DatabaseUser)
	if err != nil {
		return false, fmt.Errorf("failed to get target checksum: %w", err)
	}

	return sourceChecksum == targetChecksum, nil
}

// getDatabaseChecksum generates a checksum of the database contents
func (s *Service) getDatabaseChecksum(containerName, dbName, dbUser string) (string, error) {
	cmd := exec.Command("docker", "exec", containerName,
		"pg_dump", "-U", dbUser, "--data-only", "--inserts", dbName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to dump database for checksum: %w\nOutput: %s", err, string(output))
	}

	hash := md5.Sum(output)
	return fmt.Sprintf("%x", hash), nil
}
