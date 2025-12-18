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

```
tmpo/
├── cmd/                 # CLI commands (Using Cobra)
│   ├── command1.go
│   ├── command2.go
│   ├── command3.go
│   ├── ...
├── internal/
│   ├── config/         # Configuration management (.tmporc files)
│   ├── storage/        # SQLite database layer
│   ├── project/        # Project detection logic
│   └── export/         # Export functionality
├── docs/               # User documentation
│   ├── usage.md
│   └── configuration.md
├── main.go
└── README.md
```

### Key Directories

- **`cmd/`**: Contains all CLI command implementations using Cobra
- **`internal/config/`**: Handles `.tmporc` file parsing and configuration
- **`internal/storage/`**: SQLite database operations and models
- **`internal/project/`**: Project name detection logic (git/directory/config)
- **`internal/export/`**: Export functionality used by commands
- **`docs/`**: User-facing documentation and guides

### Data Storage

All user data is stored locally in:

```
~/.tmpo/
  └── tmpo.db          # SQLite database
```

The database schema includes:

- Time entries (start/end times, project, description)
- Project metadata (derived from entries)
- Automatic indexing for fast queries

### How Project Detection Works

When a user runs `tmpo start`, the project name is detected in this priority order:

1. **`.tmporc` file** - Searches current directory and all parent directories
2. **Git repository** - Uses `git rev-parse --show-toplevel` to find repo root
3. **Directory name** - Falls back to current directory name

This logic lives in `internal/project/detector.go`.

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

```
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

- Provide a clear description of the changes
- Reference any related issues (e.g., "Fixes #123")
- Ensure tests pass and code builds successfully
- Be responsive to feedback and questions

Reviews can take a few iterations, especially for large contributions. Don't be disheartened if you feel it takes time - we just want to ensure each contribution is high-quality and that any outstanding questions are resolved, captured or documented for posterity.

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
