# macOS Installation Guide

This guide will walk you through installing tmpo on macOS.

## Prerequisites

- macOS 11 (Big Sur) or later
- For building from source: Go 1.25 or later

## Method 1: Homebrew (Recommended)

The easiest way to install tmpo on macOS is via [Homebrew](https://brew.sh/).

```bash
brew install DylanDevelops/tap/tmpo
```

> [!NOTE]
> Homebrew 6 and later require non-official taps to be trusted before
> they can be installed (this is enabled by default). If you see a trust error,
> run the following once, then re-run the install command above:
>
> ```bash
> brew trust --tap DylanDevelops/tap
> ```

### Verify Installation

```bash
tmpo --version
```

### Updating

```bash
brew upgrade tmpo
```

### Uninstalling via Homebrew

```bash
brew uninstall tmpo

# Optionally, delete your tmpo data
rm -rf ~/.tmpo
```

## Method 2: Download Pre-built Binary

### Step 1: Download the Binary

1. Visit the [tmpo releases page](https://github.com/DylanDevelops/tmpo/releases)
2. Download the appropriate file for your Mac:
   - **Apple Silicon (M1/M2/M3/M4)**: `tmpo_X.X.X_Darwin_arm64.tar.gz`
   - **Intel Mac**: `tmpo_X.X.X_Darwin_x86_64.tar.gz`

> [!NOTE]
> Replace `X.X.X` with the latest version number, e.g., `0.1.0`

### Step 2: Extract and Install

Open Terminal and run:

```bash
# Navigate to your Downloads folder
cd ~/Downloads

# Extract the archive (adjust filename for your architecture and version)
tar -xzf tmpo_0.1.0_Darwin_arm64.tar.gz

# Move to /usr/local/bin (may require sudo)
sudo mv tmpo /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/tmpo
```

Note: Replace `0.1.0` with the actual version you downloaded.

### Step 3: Handle macOS Gatekeeper

macOS may block the binary because it's not signed with an Apple Developer ID. To allow it:

```bash
# Remove the quarantine attribute
sudo xattr -d com.apple.quarantine /usr/local/bin/tmpo
```

Alternatively, if you see a security warning:

1. Open "System Settings" → "Privacy & Security"
2. Scroll down to the Security section
3. Click "Allow Anyway" next to the tmpo warning

### Step 4: Verify Installation

```bash
tmpo --version
```

You should see the tmpo version information.

## Method 3: Build from Source

### Step 1: Install Go

If you don't have Go installed:

```bash
# Using Homebrew
brew install go

# Verify installation
go version
```

Or download from [golang.org/dl](https://golang.org/dl/).

### Step 2: Clone and Build

```bash
# Clone the repository
git clone https://github.com/DylanDevelops/tmpo.git
cd tmpo

# Build the binary
go build -o tmpo .

# Move to PATH
sudo mv tmpo /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/tmpo
```

### Step 3: Verify Installation

```bash
tmpo --version
```

## Alternative Installation Locations

If you prefer not to use `/usr/local/bin/`, you can install to your home directory:

```bash
# Create a bin directory in your home folder
mkdir -p ~/bin

# Move tmpo there
mv tmpo ~/bin/

# Add to PATH (add this line to ~/.zshrc or ~/.bash_profile)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc

# Reload your shell configuration
source ~/.zshrc
```

## Troubleshooting

### "tmpo: command not found"

This means tmpo is not in your PATH. Check:

```bash
# Verify the file exists
ls -l /usr/local/bin/tmpo

# Check if /usr/local/bin is in PATH
echo $PATH | grep "/usr/local/bin"
```

If it's not in your PATH:

```bash
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### "tmpo cannot be opened because the developer cannot be verified"

Run the xattr command to remove the quarantine flag:

```bash
sudo xattr -d com.apple.quarantine /usr/local/bin/tmpo
```

### Permission Denied

If you get permission errors:

```bash
# Make sure the file is executable
sudo chmod +x /usr/local/bin/tmpo

# Or if installed in ~/bin
chmod +x ~/bin/tmpo
```

### Architecture Mismatch

If you see an error about architecture, make sure you downloaded the correct binary:

```bash
# Check your Mac's architecture
uname -m

# arm64 = Apple Silicon (M1/M2/M3)
# x86_64 = Intel
```

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
# If installed via Homebrew
brew uninstall tmpo

# If installed manually
sudo rm /usr/local/bin/tmpo

# Optionally, delete your tmpo data
rm -rf ~/.tmpo
```

If you installed to `~/bin`:

```bash
rm ~/bin/tmpo
rm -rf ~/.tmpo
```
