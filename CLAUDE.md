# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

tmpo is a minimal CLI time tracker for developers built with Go. It uses Cobra for CLI commands, SQLite for local storage, and supports automatic project detection via Git or `.tmporc` configuration files.

## Build and Development Commands

**Build the project:**
```bash
go build -o tmpo .
```

**Run tests:**
```bash
go test -v ./...
```

**Test the binary:**
```bash
./tmpo --version
./tmpo --help
```

**Run locally without building:**
```bash
go run main.go [command]
```

**Test release build locally:**
```bash
goreleaser build --snapshot --clean
```

## Architecture

### Core Components

**CLI Layer** (`cmd/`):
- Uses Cobra for command structure
- Commands organized in subdirectories: tracking/, entries/, history/, setup/, utilities/
- Each command is a constructor function that returns `*cobra.Command`
- All commands explicitly registered in `cmd/root.go` RootCmd() function
- Version information is injected via ldflags during build

**Storage Layer** (`internal/storage/`):
- `db.go`: Database wrapper around `*sql.DB` with all query methods
- `models.go`: TimeEntry struct with Duration() and IsRunning() helper methods
- Uses modernc.org/sqlite (pure Go implementation)
- Database location: `$HOME/.tmpo/tmpo.db`
- Schema: time_entries table with id, project_name, start_time, end_time, description, hourly_rate

**Configuration** (`internal/settings/`):
- **Per-Project Configuration**: YAML-based config using `.tmporc` files
  - Config fields: project_name, hourly_rate, description
  - FindAndLoad() searches upward through parent directories for `.tmporc`
  - Supports per-project configuration by placing `.tmporc` in project root
- **Global Configuration** (`~/.tmpo/config.yaml`):
  - Managed via `tmpo config` command
  - Settings: currency, date_format, time_format, timezone
  - Currency is stored globally for consistent billing display
  - Date format is selectable (MM/DD/YYYY, DD/MM/YYYY, YYYY-MM-DD)
  - Time format is selectable (24-hour, 12-hour (AM/PM))
  - Timezone uses IANA format with validation (e.g., America/New_York, UTC)
  - Located at `$HOME/.tmpo/config.yaml`
  - If missing, defaults are used (USD, empty formats, local timezone)
  - LoadGlobalConfig() returns defaults if file doesn't exist (no error)

**Project Detection** (`internal/project/`):
- Three-tier detection strategy:
  1. `.tmporc` file (highest priority)
  2. Git repository name via `git rev-parse --show-toplevel`
  3. Current directory name (fallback)
- Helper functions: FindTmporc(), GetGitRepoName(), IsInGitRepo(), GetGitRoot()

**Export** (`internal/export/`):
- `csv.go`: Export to CSV format
- `json.go`: Export to JSON format
- Used by `export` command with filtering options

### Key Patterns

**Database Initialization:**
Every command that accesses the database follows this pattern:
```go
db, err := storage.Initialize()
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
defer db.Close()
```

**Project Name Detection:**
The `internal/project.DetectConfiguredProject()` function implements the priority: .tmporc config → Git repository → directory name

**Time Entry States:**
- Running: EndTime is nil
- Stopped: EndTime is non-nil
- Methods: TimeEntry.IsRunning(), TimeEntry.Duration()

**Hourly Rate Handling:**
HourlyRate is optional (*float64). Stored as sql.NullFloat64 in database queries and converted to/from pointer for the TimeEntry struct.

**Config Template Pattern:**
The `.tmporc` file generation uses a template-based approach to ensure all fields are visible to users:
- Template is defined as `configTemplate` constant in `internal/settings/config.go`
- Located directly below the `Config` struct for easy maintenance
- When adding new fields to `Config`, update both the struct AND the template (marked with IMPORTANT comments)

## Important Notes

**SQLite Driver:**
Uses `modernc.org/sqlite` (pure Go, no CGO) instead of mattn/go-sqlite3. This is important for cross-compilation. macOS builds keep CGO enabled, but Linux/Windows disable it (CGO_ENABLED=0).

**Version Injection:**
Version, Commit, and Date are injected at build time via ldflags:
```
-X github.com/DylanDevelops/tmpo/cmd/utilities.Version={{.Version}}
-X github.com/DylanDevelops/tmpo/cmd/utilities.Commit={{.Commit}}
-X github.com/DylanDevelops/tmpo/cmd/utilities.Date={{.Date}}
```

**Command Registration:**
New commands should be created as constructor functions (e.g., `StartCmd()`, `StopCmd()`) that return `*cobra.Command`. Each command registers its own flags before returning. Commands are then registered in `cmd/root.go` RootCmd() by calling `cmd.AddCommand(tracking.StartCmd())`.

**Interactive Prompts:**
The `manual` command uses `github.com/manifoldco/promptui` for interactive prompts (date/time input).
