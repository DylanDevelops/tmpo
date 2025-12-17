# tmpo CLI

> Set the `tmpo` — A minimal CLI time tracker for developers.

![screenshot of tmpo start and tmpo stats](https://github.com/user-attachments/assets/2142f136-6f0b-44b8-9043-7810b8b25142)

**tmpo** allows you to track time effortlessly with simple commands that live in your terminal. Track time with automatic project detection, view stats, and export data; all without leaving your workspace.

## About

**tmpo** is a lightweight, developer-friendly time tracking tool designed to integrate seamlessly with your terminal workflow. It automatically detects your project context from Git repositories or configuration files, making time tracking as simple as `tmpo start` and `tmpo stop`.

### Why tmpo?

- **🚀 Fast & Lightweight** - Built in Go, tmpo starts instantly and uses minimal resources
- **🎯 Automatic Project Detection** - Detects project names from Git repos or `.tmporc` configuration files
- **💾 Local & Private Storage** - All data stored locally in SQLite - your time tracking stays private
- **📊 Rich Reporting** - View stats, export to CSV/JSON, and track hourly rates
- **⚡ Zero Configuration Needed** - Works out of the box, configure only when you need to

## Installation

### Homebrew (macOS/Linux)

```bash
brew install tmpo
```

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/DylanDevelops/tmpo/releases).

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
```

For detailed usage and all commands, see the [Usage Guide](docs/usage.md).

## Configuration

Optionally create a `.tmporc` file in your project to customize settings:

```bash
# Interactive form (prompts for name, rate, description)
tmpo init

# Or skip prompts and use defaults
tmpo init --accept-defaults
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
