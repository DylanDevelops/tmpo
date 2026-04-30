# tmpo CLI

> Set the `tmpo` — A minimal CLI time tracker for developers.

![screenshot of tmpo start and tmpo stats](https://github.com/user-attachments/assets/ce6c684e-04a6-48d0-a6b9-77349d1b3ec8)

**tmpo** allows you to track time effortlessly with simple commands that live in your terminal. Track time with automatic project detection, organize work into milestones, view stats, and export data; all without leaving your workspace.

## About

**tmpo** is a lightweight, developer-friendly time tracking tool designed to integrate seamlessly with your terminal workflow. It automatically detects your project context from Git repositories or configuration files, making time tracking as simple as `tmpo start` and `tmpo stop`.

### Why tmpo?

- **🚀 Fast & Lightweight** - Built in Go, tmpo starts instantly and uses minimal resources
- **🎯 Automatic Project Detection** - Detects project names from Git repos or `.tmporc` configuration files
- **🌐 Global Projects** - Track time for any project from any directory without configuration files
- **🎯 Milestone Tracking** - Organize time entries into sprints, releases, or project phases
- **💾 Local & Private Storage** - All data stored locally in SQLite - your time tracking stays private
- **🔒 Built-in Backups** - Create, restore, and manage database backups with a single command
- **📊 Rich Reporting** - View stats, export to CSV/JSON, and track hourly rates
- **⚡ Zero Configuration Needed** - Works out of the box, configure only when you need to

## Installation

### Installation Guides

See the installation guide for your platform:

- [macOS Installation Guide](docs/installation/macos_installation.md)
- [Linux Installation Guide](docs/installation/linux_installation.md)
- [Windows Installation Guide](docs/installation/windows_installation.md)

### Build from Source

```bash
git clone https://github.com/DylanDevelops/tmpo.git
cd tmpo
go build -o tmpo .
```

## Quick Start

```bash
# Start tracking (auto-detects project)
tmpo start

# Check status
tmpo status

# Stop tracking
tmpo stop

# View statistics
tmpo stats

# Organize work into milestones
tmpo milestone start "Sprint 1"
```

### Track Projects From Anywhere

Create global projects to track time from any directory:

```bash
# Create a global project
tmpo init --global

# Track time from anywhere on your system
cd /tmp
tmpo start --project "Client Work" "Implementing new feature"
tmpo stop

# View entries from anywhere
tmpo log --project "Client Work"
```

For detailed usage and all commands, see the [Usage Guide](docs/usage.md).

## Configuration

### Global Settings

Set your preferences for currency, date/time formats, and timezone:

```bash
tmpo config
```

This opens an interactive wizard to configure:

- Currency (USD, EUR, GBP, JPY, and 30+ more)
- Date format (MM/DD/YYYY, DD/MM/YYYY, or YYYY-MM-DD)
- Time format (24-hour or 12-hour)
- Timezone (IANA format like America/New_York)
- Export path (default directory for exported files)

### Per-Project Settings

**Local Projects (directory-specific):**

```bash
# Create a .tmporc file in your project directory
tmpo init

# Or skip prompts and use defaults
tmpo init --accept-defaults
```

**Global Projects (track from anywhere):**

```bash
# Create a global project you can use from any directory
tmpo init --global

# Now track time from anywhere
tmpo start --project "Project Name"
```

See the [Configuration Guide](docs/configuration.md) for details.

## Feedback

Found a bug or have an idea for a feature you'd like to see in tmpo? [Open an issue](https://github.com/DylanDevelops/tmpo/issues/new/choose) and our team will be able to help.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Made with ❤️ by [Dylan Ravel](https://github.com/DylanDevelops) and you!
