# Back It Up

A PostgreSQL database backup CLI tool for Docker containers. Backup, restore, and verify PostgreSQL databases running in Docker with ease.

## Features

- ✅ **Backup** - Create compressed `.sql.gz` backups from PostgreSQL containers
- ✅ **Restore** - Restore backups to any PostgreSQL container
- ✅ **Verify** - Compare two databases to ensure data integrity
- ✅ **Test** - Full backup → restore → verify workflow in one command
- ✅ **Detailed Error Messages** - Clear PostgreSQL error output for debugging

## Installation

### Build from Source

```bash
git clone https://github.com/iostate/back-it-up.git
cd back-it-up
go build -o biu ./cmd
```

### Add to PATH (Optional)

```bash
# macOS/Linux
sudo mv biu /usr/local/bin/

# Or add to your PATH
export PATH=$PATH:$(pwd)
```

## Prerequisites

- Docker installed and running
- PostgreSQL container(s) running
- Go 1.21+ (for building from source)

## Usage

### Basic Commands

```bash
biu <command> [flags]
```

### Commands

- `backup` - Backup a PostgreSQL database from a Docker container
- `restore` - Restore a PostgreSQL database to a Docker container
- `verify` - Verify two databases contain the same data
- `test` - Backup, restore, and verify in one command
- `help` - Show help message

## Examples

### Backup a Database

Create a compressed backup of your PostgreSQL database:

```bash
biu backup -c postgres-db -d myapp -u myuser
```

**Flags:**
- `-c, --container` - Docker container name (required)
- `-d, --database` - Database name (default: "postgres")
- `-u, --user` - Database user (default: "postgres")
- `-o, --output` - Output directory (default: "./backups")

**Output:**
```
Verifying container 'postgres-db' exists...
Starting backup...
Backup completed successfully: backups/myapp_2025_12_21_14_30_45.sql.gz
```

### Restore a Database

Restore a backup to a PostgreSQL container:

```bash
biu restore -c postgres-test -d myapp -u myuser -f backups/myapp_2025_12_21_14_30_45.sql.gz --drop
```

**Flags:**
- `-c, --container` - Docker container name (required)
- `-f, --file` - Backup file path (required)
- `-d, --database` - Database name (default: "postgres")
- `-u, --user` - Database user (default: "postgres")
- `--drop` - Drop existing database before restore

**Output:**
```
Restoring backup to container 'postgres-test'...
Restore completed successfully
```

### Verify Two Databases Match

Compare two databases to ensure they contain identical data:

```bash
biu verify -s postgres-prod -t postgres-test -d myapp -u myuser
```

**Flags:**
- `-s, --source` - Source container name (required)
- `-t, --target` - Target container name (required)
- `-d, --database` - Database name (default: "postgres")
- `-u, --user` - Database user (default: "postgres")

**Output:**
```
Verifying databases match between 'postgres-prod' and 'postgres-test'...
✓ Databases match - verification successful
```

### Full Test Workflow

Backup from source, restore to target, and verify they match:

```bash
biu test -s postgres-prod -t postgres-test -d myapp -u myuser
```

**Flags:**
- `-s, --source` - Source container name (required)
- `-t, --target` - Target container name (required)
- `-d, --database` - Database name (default: "postgres")
- `-u, --user` - Database user (default: "postgres")
- `-o, --output` - Output directory (default: "./backups")

**Output:**
```
Step 1: Creating backup from source container...
✓ Backup created: backups/myapp_2025_12_21_14_30_45.sql.gz

Step 2: Restoring backup to target container...
✓ Restore completed

Step 3: Verifying databases match...
✓ Test passed - databases match!

Backup file: backups/myapp_2025_12_21_14_30_45.sql.gz
```

## Common Use Cases

### 1. Production Backup

```bash
# Daily backup of production database
biu backup -c prod-postgres -d myapp -u dbuser -o /backups/prod
```

### 2. Clone Database for Testing

```bash
# Backup production and restore to test environment
biu backup -c prod-postgres -d myapp -u dbuser
biu restore -c test-postgres -d myapp -u dbuser -f backups/myapp_2025_12_21_14_30_45.sql.gz --drop
```

### 3. Verify Migration Success

```bash
# After migrating data, verify it matches the original
biu verify -s old-postgres -t new-postgres -d myapp -u dbuser
```

### 4. Automated Testing

```bash
# Test entire backup/restore pipeline
biu test -s prod-postgres -t test-postgres -d myapp -u dbuser
```

## Finding Your Database User

If you don't know your PostgreSQL user, check your container environment:

```bash
docker exec <container-name> env | grep POSTGRES_USER
```

Or inspect the container:

```bash
docker inspect <container-name> | grep POSTGRES_USER
```

## Backup File Format

Backups are saved as gzip-compressed SQL dumps:

**Filename format:** `{database}_{YYYY_MM_DD_HH_MM_SS}.sql.gz`

**Example:** `myapp_2025_12_21_14_30_45.sql.gz`

## Troubleshooting

### Error: role "postgres" does not exist

You're using the wrong database user. Find your user with:

```bash
docker exec <container-name> env | grep POSTGRES_USER
```

Then use the correct user with `-u`:

```bash
biu backup -c mycontainer -d mydb -u correctuser
```

### Error: container not found

Ensure your Docker container is running:

```bash
docker ps | grep <container-name>
```

### Error: database does not exist

Check available databases:

```bash
docker exec <container-name> psql -U <user> -c "\l"
```

## Architecture

```
back-it-up/
├── cmd/
│   ├── main.go          # CLI entry point
│   └── commands.go      # Command implementations
├── internal/
│   ├── backup/
│   │   ├── service.go   # Backup/restore/verify logic
│   │   └── config.go    # Configuration types
│   └── docker/
│       └── service.go   # Docker operations
└── backups/             # Default output directory
```

### Design Principles

- **Clean Architecture** - Separation of concerns with internal packages
- **Interface-based** - Easy to test and extend
- **Error Transparency** - Detailed PostgreSQL error messages
- **Zero Dependencies** - Uses only Go standard library (except build tools)

## Development

### Build

```bash
go build -o biu ./cmd
```

### Run Tests

```bash
go test ./...
```

### Project Structure

- `cmd/` - CLI application code
- `internal/backup/` - Backup service and configuration
- `internal/docker/` - Docker container operations
- `backups/` - Default backup output directory

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Author

Created by [iostate](https://github.com/iostate)

## Roadmap

Future enhancements:
- [ ] Support for custom pg_dump options
- [ ] Scheduled backups with cron integration
- [ ] S3/cloud storage support
- [ ] Backup rotation and retention policies
- [ ] Multiple database backup in one command
- [ ] Progress bars for large backups
- [ ] Email notifications on backup completion
