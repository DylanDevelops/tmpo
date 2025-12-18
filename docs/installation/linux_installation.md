# Linux Installation Guide

This guide will walk you through installing tmpo on Linux.

## Prerequisites

- Linux kernel 3.10 or later (most modern distributions)
- For building from source: Go 1.25 or later

## Method 1: Download Pre-built Binary (Recommended)

### Step 1: Download the Binary

1. Visit the [tmpo releases page](https://github.com/DylanDevelops/tmpo/releases)
2. Download the appropriate file for your architecture:
   - **x86_64 (64-bit)**: `tmpo_X.X.X_Linux_x86_64.tar.gz`
   - **ARM64**: `tmpo_X.X.X_Linux_arm64.tar.gz`

> [!NOTE]  
> Replace `X.X.X` with the latest version number, e.g., `0.1.0`

You can also download using curl or wget:

```bash
# For x86_64 (replace 0.1.0 with the latest version)
wget https://github.com/DylanDevelops/tmpo/releases/download/v0.1.0/tmpo_0.1.0_Linux_x86_64.tar.gz

# Or using curl
curl -LO https://github.com/DylanDevelops/tmpo/releases/download/v0.1.0/tmpo_0.1.0_Linux_x86_64.tar.gz
```

### Step 2: Extract and Install

**System-wide installation (requires sudo):**

```bash
# Extract the archive (replace version and architecture as needed)
tar -xzf tmpo_0.1.0_Linux_x86_64.tar.gz

# Move to /usr/local/bin
sudo mv tmpo /usr/local/bin/

# Make executable (usually already set)
sudo chmod +x /usr/local/bin/tmpo
```

**User installation (no sudo required):**

```bash
# Create a bin directory in your home folder if it doesn't exist
mkdir -p ~/bin

# Extract and move (replace version and architecture as needed)
tar -xzf tmpo_0.1.0_Linux_x86_64.tar.gz
mv tmpo ~/bin/

# Make executable
chmod +x ~/bin/tmpo

# Add to PATH if not already (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Step 3: Verify Installation

```bash
tmpo --version
```

You should see the tmpo version information.

## Method 2: Build from Source

### Step 1: Install Go

**Debian/Ubuntu:**

```bash
sudo apt update
sudo apt install golang-go
```

**Fedora:**

```bash
sudo dnf install golang
```

**Arch Linux:**

```bash
sudo pacman -S go
```

**Or download from [golang.org/dl](https://golang.org/dl/):**

```bash
# Download and install Go 1.25 or later
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### Step 2: Clone and Build

```bash
# Install git if needed
sudo apt install git  # Debian/Ubuntu
sudo dnf install git  # Fedora
sudo pacman -S git    # Arch

# Clone the repository
git clone https://github.com/DylanDevelops/tmpo.git
cd tmpo

# Build the binary
go build -o tmpo .

# Move to PATH
sudo mv tmpo /usr/local/bin/
# Or for user installation
mv tmpo ~/bin/

# Make executable
sudo chmod +x /usr/local/bin/tmpo
# Or for user installation
chmod +x ~/bin/tmpo
```

### Step 3: Verify Installation

```bash
tmpo --version
```

## Determining Your Architecture

If you're not sure which binary to download:

```bash
uname -m
```

Output mapping:

- `x86_64` → Download `tmpo_X.X.X_Linux_x86_64.tar.gz`
- `aarch64` or `arm64` → Download `tmpo_X.X.X_Linux_arm64.tar.gz`

Replace `X.X.X` with the actual version number.

## Troubleshooting

### "tmpo: command not found"

The binary is not in your PATH. Check:

```bash
# Verify the file exists
ls -l /usr/local/bin/tmpo
# Or for user installation
ls -l ~/bin/tmpo

# Check if the directory is in PATH
echo $PATH
```

Add to PATH if needed:

```bash
# For /usr/local/bin (should already be there)
echo $PATH | grep "/usr/local/bin"

# For ~/bin, add to your shell config
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Permission Denied

If you get permission errors:

```bash
# Make sure the file is executable
sudo chmod +x /usr/local/bin/tmpo
# Or for user installation
chmod +x ~/bin/tmpo

# If you're trying to write to /usr/local/bin without sudo
sudo mv tmpo /usr/local/bin/
```

### "cannot execute binary file: Exec format error"

You downloaded the wrong architecture. Check your system:

```bash
uname -m
```

And download the correct binary for your architecture.

### Database Permission Issues

If tmpo can't create or access the database:

```bash
# Check if the directory exists and is writable
ls -la ~/.tmpo

# If it doesn't exist, create it
mkdir -p ~/.tmpo

# Fix permissions if needed
chmod 755 ~/.tmpo
```

## Distribution-Specific Notes

### Ubuntu/Debian

If you prefer `/usr/bin` over `/usr/local/bin`:

```bash
sudo mv tmpo /usr/bin/
sudo chmod +x /usr/bin/tmpo
```

### Fedora/RHEL

SELinux may require additional context:

```bash
sudo chcon -t bin_t /usr/local/bin/tmpo
```

### Arch Linux

Consider creating a PKGBUILD for easier installation and updates.

## Next Steps

Once installed, check out the [Usage Guide](../usage.md) to learn how to use tmpo, or get started quickly with:

```bash
# Navigate to your project directory
cd ~/path/to/your/project

# Start tracking time
tmpo start

# Check status
tmpo status

# Stop tracking
tmpo stop

# View statistics
tmpo stats
```

## Uninstalling

To uninstall tmpo:

```bash
# Remove the binary
sudo rm /usr/local/bin/tmpo
# Or for user installation
rm ~/bin/tmpo

# Optionally, delete your tmpo data
rm -rf ~/.tmpo
```
