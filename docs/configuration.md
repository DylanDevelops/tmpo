# Configuration Guide

Learn how to configure tmpo for your projects and workflow.

## Storage Location

All time tracking data is stored locally on your machine:

```
~/.tmpo/
  └── tmpo.db          # SQLite database
```

Your data never leaves your machine. The database file can be backed up, copied, or version controlled if desired.

## Project Configuration

### The `.tmporc` File

Place a `.tmporc` file in your project root to customize tracking settings for that project. When you run tmpo commands from within the project directory (or any subdirectory), it will automatically use these settings.

### Creating a Configuration File

Use `tmpo init` to create a `.tmporc` file using an interactive form:

```bash
cd ~/projects/my-project
tmpo init
# You'll be prompted for:
# - Project name (defaults to auto-detected name)
# - Hourly rate (optional, press Enter to skip)
# - Description (optional, press Enter to skip)
# - Currency code (optional, defaults to USD)
```

For quick setup without prompts, use the `--accept-defaults` flag:

```bash
tmpo init --accept-defaults
# Creates .tmporc with auto-detected project name and default values
```

This creates a `.tmporc` file in the current directory.

### File Format

The `.tmporc` file uses YAML format:

```yaml
# tmpo project configuration
# This file configures time tracking settings for this project

# Project name (used to identify time entries)
project_name: My Awesome Project

# [OPTIONAL] Hourly rate for billing calculations (set to 0 to disable)
hourly_rate: 125.50

# [OPTIONAL] Description for this project
description: Client project for Acme Corp

# [OPTIONAL] Currency code for billing display (USD, EUR, GBP, JPY, etc.)
currency: USD
```

### Configuration Fields

#### `project_name` (required)

The name used to identify time entries for this project. This overrides automatic detection from git or directory names.

**Example:**

```yaml
project_name: Client Website Redesign
```

#### `hourly_rate` (optional)

Your billing rate per hour. When set, tmpo will calculate estimated earnings based on tracked time. The currency symbol displayed is determined by the `currency` field (defaults to USD if not specified).

**Example:**

```yaml
hourly_rate: 150.00
currency: USD  # Will display as $150.00
```

Set to `0` or omit to disable rate tracking:

```yaml
hourly_rate: 0
```

#### `description` (optional)

A longer description or notes about the project. This is for your reference and doesn't affect time tracking.

**Example:**

```yaml
description: Q1 2024 website redesign for Acme Corp. Main contact: john@acme.com
```

#### `currency` (optional)

The ISO 4217 currency code for displaying billing rates and earnings. This determines the currency symbol shown in stats, reports, and summaries.

**Supported Currencies:**

tmpo supports 30+ currencies including:

- **Americas:** USD ($), CAD (CA$), BRL (R$), MXN (MX$)
- **Europe:** EUR (€), GBP (£), CHF (Fr), SEK (kr), NOK (kr)
- **Asia:** JPY (¥), CNY (¥), INR (₹), KRW (₩), SGD (S$)
- **Oceania:** AUD (A$), NZD (NZ$)

And many more. See the [full currency code list](https://en.wikipedia.org/wiki/ISO_4217#Active_codes).

**Examples:**

```yaml
# US-based project
hourly_rate: 150
currency: USD
# Displays as: $150.00

# European project
hourly_rate: 120
currency: EUR
# Displays as: €120.00

# UK project
hourly_rate: 100
currency: GBP
# Displays as: £100.00

# Japanese project
hourly_rate: 15000
currency: JPY
# Displays as: ¥15000.00
```

If not specified or if an unknown currency code is provided, `currency` defaults to USD. Currency codes are case-insensitive (USD, usd, or Usd all work).

## Project Detection Priority

When you run `tmpo start`, the project name is determined in this order:

1. **`.tmporc` file** - If present in current directory or any parent directory
2. **Git repository name** - The name of the git repository root folder
3. **Current directory name** - The name of your current working directory

This means you can override automatic detection by adding a `.tmporc` file.

### Example Scenarios

#### **Scenario 1:** Git repo with custom name

```bash
# Directory: ~/code/website-2024/
# Git repo name: website-2024
# No .tmporc file
tmpo start
# → Tracks to project "website-2024"
```

#### **Scenario 2:** With .tmporc override

```bash
# Directory: ~/code/website-2024/
# .tmporc contains: project_name: "Acme Website"
tmpo start
# → Tracks to project "Acme Website"
```

#### **Scenario 3:** Subdirectory detection

```bash
# Directory: ~/code/my-project/src/components/
# .tmporc exists at: ~/code/my-project/.tmporc
tmpo start
# → Uses .tmporc from project root
```

## Multi-Project Setup

### Separate Projects with Different Rates

Create a `.tmporc` in each project directory using `tmpo init`:

```bash
# Client A - $150/hour
cd ~/projects/client-a
tmpo init
# Project name: Client A - Web Development
# Hourly rate: 150
# Description: [press Enter to skip]
# Currency code: USD

# Client B - £175/hour
cd ~/projects/client-b
tmpo init
# Project name: Client B - Game Development
# Hourly rate: 175
# Description: [press Enter to skip]
# Currency code: GBP

# Personal project - no billing
cd ~/projects/my-app
tmpo init --accept-defaults  # Quick setup with defaults
```

Alternatively, you can manually create `.tmporc` files:

```bash
# Client configuration
cat > ~/projects/client-project/.tmporc << EOF
project_name: Client Project - Web Development
hourly_rate: 150.00
currency: USD
EOF
```

### Monorepo with Sub-Projects

In a monorepo, you can track different sub-projects separately:

```bash
# Main repo tracks to "My Company Platform"
cd ~/monorepo
echo "project_name: My Company Platform" > .tmporc

# But frontend team tracks separately
cd ~/monorepo/frontend
echo "project_name: Platform - Frontend" > .tmporc

# And backend team tracks separately
cd ~/monorepo/backend
echo "project_name: Platform - Backend" > .tmporc
```

## Version Control

### Should I commit `.tmporc`?

**Yes, commit it** *if*:

- Your team wants shared project naming
- It's an open source project and contributors might want to track time
- The configuration has no sensitive information

**Don't commit it** *if*:

- The hourly rate is personal/confidential
- Each team member prefers their own project naming

### Using `.gitignore`

To keep `.tmporc` files local:

```bash
echo ".tmporc" >> .gitignore
```

Or create a global gitignore:

```bash
echo ".tmporc" >> ~/.gitignore_global
git config --global core.excludesfile ~/.gitignore_global
```

## Migrating Data

### Backing Up Your Data

```bash
# Create a backup of your time tracking database
cp ~/.tmpo/tmpo.db ~/backups/tmpo-backup-$(date +%Y%m%d).db
```

### Moving to a New Machine

```bash
# On old machine
cp ~/.tmpo/tmpo.db ~/tmpo-export.db

# Transfer file to new machine, then:
mkdir -p ~/.tmpo
cp ~/tmpo-export.db ~/.tmpo/tmpo.db
```

### Exporting for External Tools

Use `tmpo export` to get your data in portable formats:

```bash
# Export everything to CSV
tmpo export --output all-time-data.csv

# Export to JSON for programmatic access
tmpo export --format json --output all-time-data.json
```

See the [Usage Guide](usage.md#tmpo-export) for more export options.
