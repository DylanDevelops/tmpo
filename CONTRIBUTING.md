# Contributing to tmpo

Thank you for your interest in contributing to tmpo! This document provides guidelines and instructions for contributing to the project.

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Git

### Setting Up Your Development Environment

1. [Fork](https://github.com/DylanDevelops/tmpo/fork) the repository
2. Clone your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/tmpo.git
   cd tmpo
   ```

3. Add the upstream repository:

   ```bash
   git remote add upstream https://github.com/DylanDevelops/tmpo.git
   ```

## Development Workflow

### Building

```bash
# Build for local development
go build -o tmpo .

# Run the binary
./tmpo --help
```

### Development Mode

To prevent corrupting your real tmpo data during development, use the `TMPO_DEV` environment variable:

```bash
# Enable development mode (uses ~/.tmpo-dev/ instead of ~/.tmpo/)
export TMPO_DEV=1

# Now all commands use the development database
./tmpo start "Testing new feature"
./tmpo status
./tmpo stop
```

**Data Locations:**

- **Production mode** (default):
  - Database: `~/.tmpo/tmpo.db`
  - Global config: `~/.tmpo/config.yaml`
- **Development mode** (`TMPO_DEV=1`):
  - Database: `~/.tmpo-dev/tmpo.db`
  - Global config: `~/.tmpo-dev/config.yaml`

> [!NOTE]
> The `export TMPO_DEV=1` command only applies to your **current terminal session**. When you close the terminal, it resets to production mode. This is intentional for safety - you must explicitly enable dev mode each time.

**Making it persistent (optional):**

If you prefer to always use dev mode, add it to your shell profile:

```bash
# For zsh (macOS default)
echo 'export TMPO_DEV=1' >> ~/.zshrc

# For bash
echo 'export TMPO_DEV=1' >> ~/.bashrc
```

Then restart your terminal or run `source ~/.zshrc` (or `source ~/.bashrc`).

**Benefits of development mode:**

- Your real time tracking data and settings stay safe
- You can test database and config changes without risk
- You can easily clean up test data and config (`rm -rf ~/.tmpo-dev/`)

### Building with Version Information

To build with version information injected (useful for testing version display):

```bash
go build -ldflags "-X github.com/DylanDevelops/tmpo/cmd/utilities.Version=0.1.0 \
  -X github.com/DylanDevelops/tmpo/cmd/utilities.Commit=$(git rev-parse --short HEAD) \
  -X github.com/DylanDevelops/tmpo/cmd/utilities.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o tmpo .
```

> [!NOTE]
> This is an example - you can modify the version number (e.g., `0.1.0`) or any other injected values to suit your testing needs.

This is useful when you want to:

- Test version display locally (`./tmpo --version`)
- Build a binary with specific version info
- Verify version injection is working correctly

For production releases, goreleaser handles version injection automatically.

### Testing

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...
```

### Building Releases

```bash
# Build with goreleaser (for testing release builds)
goreleaser build --snapshot --clean
```

## Project Structure

```text
tmpo/
├── cmd/                # CLI commands (Using Cobra)
│   ├── root.go         # Root command with RootCmd() constructor
│   ├── tracking/       # Time tracking commands (start, stop, pause, resume, status)
│   ├── entries/        # Entry management (edit, delete, manual)
│   ├── history/        # History commands (log, stats, export)
│   ├── milestones/     # Milestone management (start, finish, status, list)
│   ├── setup/          # Setup commands (init)
│   ├── config/         # Global configuration (config/settings/preferences)
│   └── utilities/      # Utility commands (version)
├── internal/
│   ├── settings/       # Configuration management (.tmporc and global config)
│   ├── storage/        # SQLite database layer
│   ├── project/        # Project detection logic
│   ├── export/         # Export functionality (CSV, JSON)
│   ├── currency/       # Currency formatting and symbol handling
│   ├── update/         # Update checking from GitHub releases
│   └── ui/             # UI helpers (formatting, colors, printing)
├── docs/               # User documentation
│   ├── usage.md
│   └── configuration.md
├── main.go
└── README.md
```

### Key Directories

- **`cmd/`**: Contains all CLI command implementations using Cobra
  - **`cmd/tracking/`**: Time tracking commands (start, stop, pause, resume, status)
  - **`cmd/entries/`**: Entry management commands (edit, delete, manual)
  - **`cmd/history/`**: History and reporting commands (log, stats, export)
  - **`cmd/setup/`**: Setup and initialization commands (init)
  - **`cmd/utilities/`**: Utility commands and version information (version)
  - **`cmd/config/`**: Global configuration command (config/settings/preferences)
  - **`cmd/milestones/`**: Milestone management commands (start, finish, status, list)
- **`internal/settings/`**: Configuration management (`.tmporc` files and global `config.yaml`)
- **`internal/storage/`**: SQLite database operations, models, and migrations
- **`internal/project/`**: Project name detection logic (git/directory/config)
- **`internal/export/`**: Export functionality (CSV, JSON)
- **`internal/currency/`**: Currency formatting and symbol handling for billing
- **`internal/update/`**: Update checking from GitHub releases
- **`internal/ui/`**: UI helpers for formatting, colors, and terminal output
- **`docs/`**: User-facing documentation and guides

### Data Storage

All user data is stored locally in:

```text
~/.tmpo/              # Production (default)
  ├── tmpo.db         # SQLite database
  ├── config.yaml     # Global configuration (optional)
  └── projects.yaml   # Global projects registry (optional)

~/.tmpo-dev/          # Development (when TMPO_DEV=1)
  ├── tmpo.db         # SQLite database
  ├── config.yaml     # Global configuration (optional)
  └── projects.yaml   # Global projects registry (optional)
```

The database schema includes:

- **time_entries**: Time tracking entries (start/end times, project, description, hourly rate, milestone)
- **milestones**: Project milestones for organizing work
- **settings**: Migration tracking and other metadata (e.g., `001_utc_timestamps`)
- Automatic indexing for fast queries

> [!IMPORTANT]
> All timestamps are stored in UTC and converted to the user's configured timezone for display.

> [!NOTE]
> See [Development Mode](#development-mode) for information on using the development database during local development.

### How Project Detection Works

When a user runs `tmpo start`, the project name is detected in this priority order:

1. **`--project` flag** - Explicitly specified global project (e.g., `tmpo start --project "My Project"`)
2. **`.tmporc` file** - Searches current directory and all parent directories
3. **Git repository** - Uses `git rev-parse --show-toplevel` to find repo root
4. **Directory name** - Falls back to current directory name

**Global Projects:**

Users can create global projects with `tmpo init --global`, which stores project configurations in `~/.tmpo/projects.yaml`. These projects can be tracked from any directory using the `--project` flag.

This logic lives in `internal/project/detect.go`.

### Database Migrations

When you need to modify the database schema, use the migration system in `internal/storage/migrations.go`:

1. **Add a migration constant:**

   ```go
   const (
       Migration001_UTCTimestamps = "001_utc_timestamps"
       Migration002_YourFeature   = "002_your_feature"  // New migration
   )
   ```

2. **Create your migration function:**

   ```go
   func (d *Database) migrateYourFeature() error {
       completed, err := d.hasMigrationRun(Migration002_YourFeature)
       if err != nil || completed {
           return err
       }

       tx, err := d.db.Begin()
       if err != nil {
           return fmt.Errorf("failed to begin transaction: %w", err)
       }
       defer func() {
           if err != nil {
               tx.Rollback()
           }
       }()

       // Your migration logic here

       // Mark complete
       _, err = tx.Exec(
           "INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
           Migration002_YourFeature, "completed", time.Now().UTC(),
       )
       if err != nil {
           return err
       }

       return tx.Commit()
   }
   ```

3. **Register in `runMigrations()`:**

   ```go
   func (d *Database) runMigrations() error {
       if err := d.migrateTimestampsToUTC(); err != nil {
           return fmt.Errorf("timestamp UTC migration failed: %w", err)
       }
       if err := d.migrateYourFeature(); err != nil {
           return fmt.Errorf("your feature migration failed: %w", err)
       }
       return nil
   }
   ```

**Important:** Migrations run automatically on `Initialize()` and are wrapped in transactions for safety. If a migration fails, all changes are rolled back.

## Making Changes

### Branching

Create a feature branch from `main`:

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names such as:

- `feature/add-pause-command`
- `fix/status-display-bug`
- `docs/update-readme`

### Code Style

- Follow standard Go conventions and use `gofmt`
- Write clear, descriptive commit messages
- Add comments for complex logic
- Keep functions focused and modular

### Commit Messages

Use clear, imperative commit messages:

```text
Add pause/resume functionality

- Implement pause command to temporarily stop tracking
- Add resume command to continue paused sessions
- Update status command to show paused state
```

## Submitting Changes

1. Ensure all tests pass: `go test -v ./...`
2. Commit your changes with descriptive messages
3. Push to your fork:

   ```bash
   git push origin feature/your-feature-name
   ```

4. Open a Pull Request against the `main` branch
5. Describe your changes and link any related issues

### Pull Request Guidelines

- **Use a conventional-commit-style PR title** (e.g. `feat: add pause command`, `fix: correct status display`). The prefix determines the PR's category in the auto-generated release notes — supported: `feat`, `fix`, `perf`, `refactor`, `docs`, `build`, `ci`, `chore`.
- Provide a clear description of the changes
- Reference any related issues (e.g., "Fixes #123")
- Ensure tests pass and code builds successfully
- Be responsive to feedback and questions

Reviews can take a few iterations, especially for large contributions. Don't be disheartened if you feel it takes time - we just want to ensure each contribution is high-quality and that any outstanding questions are resolved, captured or documented for posterity.

## Releasing (Maintainers)

Pushing a `vX.Y.Z` tag triggers the [Release workflow](.github/workflows/release.yml): GoReleaser builds the cross-platform binaries, publishes the GitHub release, and updates the Homebrew tap.

Release notes are generated automatically from the merged pull requests — there is no manual changelog. Each PR is one entry, attributed to its author and grouped by its `type:` label, which the [PR labeler](.github/workflows/label-prs.yml) applies from the PR title. Add the `ignore-for-release` label to omit a PR.

To preview the next release's notes before tagging:

```bash
gh api --method POST repos/DylanDevelops/tmpo/releases/generate-notes \
  -f tag_name=vX.Y.Z -f previous_tag_name=<last-tag> --jq .body
```

## Distribution Packaging

We appreciate packaging efforts for various package managers and distributions! However, to keep maintenance focused on the core application and avoid setting precedent for supporting every distribution method, we have the following policy:

### What's Maintained In-Tree

- **GoReleaser configuration** (`.goreleaser.yml`) - handles official releases for multiple platforms
- **Core build files** - Go modules, source code, and build scripts

### What Should Live Externally

Distribution-specific packaging should be maintained outside this repository:

- **Nix flakes and modules** - Contribute to [nixpkgs](https://github.com/NixOS/nixpkgs) or maintain in a separate repo
- **Homebrew formulas** - Maintained in our custom tap repository at [`DylanDevelops/homebrew-tap`](https://github.com/DylanDevelops/homebrew-tap). Once `tmpo` meets the required repository popularity metrics, we will submit it to `homebrew-core`.
- **Linux packages** - AUR (Arch), APT/RPM repos, Snap, Flatpak, etc.
- **System configuration** - Systemd units, init scripts, etc.
- **Other package managers** - Scoop (Windows), Chocolatey, etc.

### Why This Policy?

Maintaining distribution-specific packaging in-tree creates ongoing maintenance burden:

- Each config format change requires updating multiple package definitions
- Testing across different package managers becomes complex
- Sets precedent for accepting every packaging request
- Most package ecosystems prefer maintaining packages in their own repositories anyway

### How to Contribute Packaging

If you'd like to package tmpo for your preferred distribution:

1. **Create the package** in the appropriate repository (nixpkgs, AUR, etc.)
2. **Open an issue** with the link to your package
3. **We'll add a link** in our installation documentation to help users find it

This way, tmpo remains accessible across platforms while keeping maintenance focused on what we do best—building a great time tracking tool.

We may consider dedicated support for specific platforms in the future if we see a large user base, but for now, community-maintained packages with documentation links work best for everyone.

## Reporting Issues

When reporting bugs or requesting features, please:

1. Check existing issues first to avoid duplicates
2. Use the issue templates provided
3. Include relevant details:
   - tmpo version (`tmpo --version`)
   - Operating system
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior

## Questions?

Feel free to [open an issue](https://github.com/DylanDevelops/tmpo/issues/new/choose) for questions or discussions about:

- Feature ideas
- Implementation approaches
- Project architecture

## Code of Conduct

Be respectful and constructive in all interactions. We're all here to make tmpo better!

---

Thank you for contributing to tmpo!
