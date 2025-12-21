package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/iostate/back-it-up/internal/backup"
	"github.com/iostate/back-it-up/internal/docker"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "backup":
		if err := runBackup(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	containerName := fs.String("container", "", "Docker container name (required)")
	fs.StringVar(containerName, "c", "", "Docker container name (shorthand)")
	outputDir := fs.String("output", "./backups", "Output directory for backup file")
	fs.StringVar(outputDir, "o", "./backups", "Output directory for backup file (shorthand)")
	dbName := fs.String("database", "postgres", "Database name")
	fs.StringVar(dbName, "d", "postgres", "Database name (shorthand)")
	dbUser := fs.String("user", "postgres", "Database user")
	fs.StringVar(dbUser, "u", "postgres", "Database user (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *containerName == "" {
		fmt.Fprintln(os.Stderr, "Error: --container flag is required")
		fs.Usage()
		return fmt.Errorf("missing required flag: --container")
	}

	// Initialize services
	dockerSvc := docker.NewService()
	backupSvc := backup.NewService(dockerSvc)

	// Verify container exists
	fmt.Printf("Verifying container '%s' exists...\n", *containerName)
	if err := dockerSvc.VerifyContainer(*containerName); err != nil {
		return fmt.Errorf("container verification failed: %w", err)
	}

	// Perform backup
	fmt.Println("Starting backup...")
	outputPath, err := backupSvc.Backup(backup.Config{
		ContainerName: *containerName,
		DatabaseName:  *dbName,
		DatabaseUser:  *dbUser,
		OutputDir:     *outputDir,
		Timestamp:     time.Now(),
	})
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("Backup completed successfully: %s\n", outputPath)
	return nil
}

func printUsage() {
	fmt.Println(`back-it-up - PostgreSQL database backup CLI tool

Usage:
  back-it-up <command> [flags]

Commands:
  backup      Backup a PostgreSQL database from a Docker container
  help        Show this help message

Backup Flags:
  -c, --container string   Docker container name (required)
  -d, --database string    Database name (default "postgres")
  -u, --user string        Database user (default "postgres")
  -o, --output string      Output directory for backup file (default "./backups")

Examples:
  back-it-up backup -c my-postgres-container
  back-it-up backup --container my-postgres-container --database mydb --user postgres
  back-it-up backup -c my-container -d mydb -o /backups`)
}
