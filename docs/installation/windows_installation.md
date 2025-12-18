# Windows Installation Guide

This guide will walk you through installing tmpo on Windows.

## Prerequisites

- Windows 10 or later
- For building from source: Go 1.25 or later

## Method 1: Download Pre-built Binary (Recommended)

### Step 1: Download the Binary

1. Visit the [tmpo releases page](https://github.com/DylanDevelops/tmpo/releases)
2. Download the appropriate file for your system:
   - **x86_64 (64-bit Intel/AMD)**: `tmpo_X.X.X_Windows_x86_64.zip`
   - **ARM64 (ARM-based Windows)**: `tmpo_X.X.X_Windows_arm64.zip`

> [!NOTE]
> Replace `X.X.X` with the latest version number, e.g., `0.1.0`

3. Extract the ZIP file to a location of your choice (e.g., `C:\Program Files\tmpo\`)

### Step 2: Add tmpo to PATH

To use tmpo from any directory, add it to your system PATH:

**Using PowerShell (Recommended):**

```powershell
# Add to user PATH (doesn't require admin)
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$tmpoPath = "C:\Program Files\tmpo"  # Adjust this to your installation path
[Environment]::SetEnvironmentVariable("Path", "$userPath;$tmpoPath", "User")
```

**Using System Settings (GUI):**

1. Press `Win + X` and select "System"
2. Click "Advanced system settings"
3. Click "Environment Variables"
4. Under "User variables", select "Path" and click "Edit"
5. Click "New" and add the path to your tmpo directory (e.g., `C:\Program Files\tmpo`)
6. Click "OK" on all dialogs

### Step 3: Verify Installation

Open a new Command Prompt or PowerShell window and run:

```powershell
tmpo --version
```

You should see the tmpo version information.

### Step 4: Start Using tmpo

```powershell
tmpo --help
```

## Method 2: Build from Source

### Step 1: Install Go

1. Download and install Go from [golang.org/dl](https://golang.org/dl/)
2. Verify installation:

```powershell
go version
```

### Step 2: Clone and Build

```powershell
# Clone the repository
git clone https://github.com/DylanDevelops/tmpo.git
cd tmpo

# Build the binary
go build -o tmpo.exe .
```

### Step 3: Move Binary to PATH

Move the built `tmpo.exe` to a directory in your PATH, or add the current directory to PATH as described in Method 1.

### Step 4: Verify Installation

```powershell
tmpo --version
```

## Determining Your System Architecture

If you're not sure which binary to download, open PowerShell and run:

```powershell
$env:PROCESSOR_ARCHITECTURE
```

Output mapping:

- `AMD64` → Download `tmpo_X.X.X_Windows_x86_64.zip`
- `ARM64` → Download `tmpo_X.X.X_Windows_arm64.zip`

Replace `X.X.X` with the actual version number.

## Troubleshooting

### "tmpo is not recognized as an internal or external command"

This means tmpo is not in your PATH. Make sure you:

1. Added the correct directory to your PATH
2. Opened a new terminal window after modifying PATH
3. The `tmpo.exe` file exists in the directory you added

### Permission Denied Errors

If you get permission errors when extracting or running tmpo:

1. Try running your terminal as Administrator
2. Extract the binary to your user directory instead (e.g., `C:\Users\YourName\bin\`)

### Windows SmartScreen Warning

Windows may show a SmartScreen warning for the binary. This is normal for newly released software. You can click "More info" and then "Run anyway" to proceed.

## Next Steps

Once installed, check out the [Usage Guide](../usage.md) to learn how to use tmpo, or get started quickly with:

```powershell
# Navigate to your project directory
cd C:\path\to\your\project

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

1. Delete the tmpo binary from your installation directory
2. Remove the directory from your PATH (reverse the steps in "Add tmpo to PATH")
3. Optionally, delete your tmpo data:

   ```powershell
   Remove-Item -Recurse -Force "$env:USERPROFILE\.tmpo"
   ```
