package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/iostate/back-it-up/internal/backup"
	"github.com/iostate/back-it-up/internal/docker"
)

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

func runRestore(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	containerName := fs.String("container", "", "Docker container name (required)")
	fs.StringVar(containerName, "c", "", "Docker container name (shorthand)")
	backupPath := fs.String("file", "", "Backup file path (required)")
	fs.StringVar(backupPath, "f", "", "Backup file path (shorthand)")
	dbName := fs.String("database", "postgres", "Database name")
	fs.StringVar(dbName, "d", "postgres", "Database name (shorthand)")
	dbUser := fs.String("user", "postgres", "Database user")
	fs.StringVar(dbUser, "u", "postgres", "Database user (shorthand)")
	dropExisting := fs.Bool("drop", false, "Drop existing database before restore")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *containerName == "" || *backupPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --container and --file flags are required")
		fs.Usage()
		return fmt.Errorf("missing required flags")
	}

	// Initialize services
	dockerSvc := docker.NewService()
	backupSvc := backup.NewService(dockerSvc)

	// Perform restore
	fmt.Printf("Restoring backup to container '%s'...\n", *containerName)
	if err := backupSvc.Restore(backup.RestoreConfig{
		ContainerName: *containerName,
		DatabaseName:  *dbName,
		DatabaseUser:  *dbUser,
		BackupPath:    *backupPath,
		DropExisting:  *dropExisting,
	}); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Println("Restore completed successfully")
	return nil
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	sourceContainer := fs.String("source", "", "Source container name (required)")
	fs.StringVar(sourceContainer, "s", "", "Source container name (shorthand)")
	targetContainer := fs.String("target", "", "Target container name (required)")
	fs.StringVar(targetContainer, "t", "", "Target container name (shorthand)")
	dbName := fs.String("database", "postgres", "Database name")
	fs.StringVar(dbName, "d", "postgres", "Database name (shorthand)")
	dbUser := fs.String("user", "postgres", "Database user")
	fs.StringVar(dbUser, "u", "postgres", "Database user (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *sourceContainer == "" || *targetContainer == "" {
		fmt.Fprintln(os.Stderr, "Error: --source and --target flags are required")
		fs.Usage()
		return fmt.Errorf("missing required flags")
	}

	// Initialize services
	dockerSvc := docker.NewService()
	backupSvc := backup.NewService(dockerSvc)

	// Perform verification
	fmt.Printf("Verifying databases match between '%s' and '%s'...\n", *sourceContainer, *targetContainer)
	match, err := backupSvc.Verify(backup.VerifyConfig{
		SourceContainer: *sourceContainer,
		TargetContainer: *targetContainer,
		DatabaseName:    *dbName,
		DatabaseUser:    *dbUser,
	})
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if match {
		fmt.Println("✓ Databases match - verification successful")
	} else {
		fmt.Println("✗ Databases do not match")
		return fmt.Errorf("database verification failed")
	}

	return nil
}

func runTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	sourceContainer := fs.String("source", "", "Source container name (required)")
	fs.StringVar(sourceContainer, "s", "", "Source container name (shorthand)")
	targetContainer := fs.String("target", "", "Target container name (required)")
	fs.StringVar(targetContainer, "t", "", "Target container name (shorthand)")
	dbName := fs.String("database", "postgres", "Database name")
	fs.StringVar(dbName, "d", "postgres", "Database name (shorthand)")
	dbUser := fs.String("user", "postgres", "Database user")
	fs.StringVar(dbUser, "u", "postgres", "Database user (shorthand)")
	outputDir := fs.String("output", "./backups", "Output directory for backup file")
	fs.StringVar(outputDir, "o", "./backups", "Output directory for backup file (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *sourceContainer == "" || *targetContainer == "" {
		fmt.Fprintln(os.Stderr, "Error: --source and --target flags are required")
		fs.Usage()
		return fmt.Errorf("missing required flags")
	}

	// Initialize services
	dockerSvc := docker.NewService()
	backupSvc := backup.NewService(dockerSvc)

	// Step 1: Backup from source
	fmt.Println("Step 1: Creating backup from source container...")
	backupPath, err := backupSvc.Backup(backup.Config{
		ContainerName: *sourceContainer,
		DatabaseName:  *dbName,
		DatabaseUser:  *dbUser,
		OutputDir:     *outputDir,
		Timestamp:     time.Now(),
	})
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	fmt.Printf("✓ Backup created: %s\n\n", backupPath)

	// Step 2: Restore to target
	fmt.Println("Step 2: Restoring backup to target container...")
	if err := backupSvc.Restore(backup.RestoreConfig{
		ContainerName: *targetContainer,
		DatabaseName:  *dbName,
		DatabaseUser:  *dbUser,
		BackupPath:    backupPath,
		DropExisting:  true,
	}); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	fmt.Println("✓ Restore completed")

	// Step 3: Verify databases match
	fmt.Println("\nStep 3: Verifying databases match...")
	match, err := backupSvc.Verify(backup.VerifyConfig{
		SourceContainer: *sourceContainer,
		TargetContainer: *targetContainer,
		DatabaseName:    *dbName,
		DatabaseUser:    *dbUser,
	})
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if match {
		fmt.Println("✓ Test passed - databases match!")
		fmt.Printf("\nBackup file: %s\n", backupPath)
	} else {
		return fmt.Errorf("test failed - databases do not match")
	}

	return nil
}

func printUsage() {
	fmt.Println(`back-it-up - PostgreSQL database backup CLI tool

Usage:
  back-it-up <command> [flags]

Commands:
  backup      Backup a PostgreSQL database from a Docker container
  restore     Restore a PostgreSQL database to a Docker container
  verify      Verify two databases contain the same data
  test        Backup, restore, and verify in one command
  help        Show this help message

Backup Flags:
  -c, --container string   Docker container name (required)
  -d, --database string    Database name (default "postgres")
  -u, --user string        Database user (default "postgres")
  -o, --output string      Output directory for backup file (default "./backups")

Restore Flags:
  -c, --container string   Docker container name (required)
  -f, --file string        Backup file path (required)
  -d, --database string    Database name (default "postgres")
  -u, --user string        Database user (default "postgres")
  --drop                   Drop existing database before restore

Verify Flags:
  -s, --source string      Source container name (required)
  -t, --target string      Target container name (required)
  -d, --database string    Database name (default "postgres")
  -u, --user string        Database user (default "postgres")

Test Flags:
  -s, --source string      Source container name (required)
  -t, --target string      Target container name (required)
  -d, --database string    Database name (default "postgres")
  -u, --user string        Database user (default "postgres")
  -o, --output string      Output directory for backup file (default "./backups")

Examples:
  # Backup
  back-it-up backup -c my-postgres-container -d mydb

  # Restore
  back-it-up restore -c test-postgres -f ./backups/mydb_2025_12_21_14_30_45.sql.gz --drop

  # Verify
  back-it-up verify -s prod-postgres -t test-postgres -d mydb

  # Full test (backup, restore, verify)
  back-it-up test -s prod-postgres -t test-postgres -d mydb`)
}
