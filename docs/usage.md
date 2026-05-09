# Usage Guide

Complete reference for all tmpo commands and features.

## Basic Commands

### `tmpo start [description]`

Start tracking time for the current project. Automatically detects the project name from:

1. `--project` flag (global project)
2. `.tmporc` configuration file (if present)
3. Git repository name
4. Current directory name

**Options:**

- `--project NAME` / `-p NAME` - Track time for a specific global project

**Examples:**

```bash
tmpo start                             # Start tracking (auto-detect)
tmpo start "Fix authentication bug"    # Start with description
tmpo start --project "Client Work"     # Track a global project from anywhere
tmpo start -p "Consulting" "Code review"  # Short flag with description
```

### `tmpo stop`

Stop the currently running time entry.

```bash
tmpo stop
```

### `tmpo pause`

Pause the currently running time entry. This is useful for taking quick breaks without losing context. The paused session can be resumed with `tmpo resume`.

```bash
tmpo pause
# Output:
# [tmpo] Paused tracking my-project
#     Session Duration: 45m 23s
#     Use 'tmpo resume' to continue tracking
```

**How it works:**

- Stops the current time entry (records end time)
- Use `tmpo resume` to start a new entry with the same project and description
- Each pause creates a separate time entry, giving you a detailed audit trail

### `tmpo resume`

Resume time tracking by starting a new session with the same project and description as the last paused (or stopped) session.

**Options:**

- `--project NAME` / `-p NAME` - Resume tracking for a specific global project

**Examples:**

```bash
tmpo resume                            # Resume current project
tmpo resume --project "Client Work"    # Resume a global project from anywhere
# Output:
# [tmpo] Resumed tracking time for Client Work
#     Description: Implementing feature
```

**Use cases:**

- Continue work after a break
- Resume after accidentally stopping the timer
- Quickly restart the same task
- Resume a global project from any directory

> [!TIP]
> Made a mistake? Run `tmpo undo` to instantly revert your last action — start, stop, pause, resume, manual entry creation, or deletion. See [`tmpo undo`](#tmpo-undo) for details.

### `tmpo status`

View the current tracking session with elapsed time.

```bash
tmpo status
# Output:
# [tmpo] Currently tracking: my-project
#     Started: 2:30 PM
#     Duration: 1h 23m
#     Description: Implementing feature
```

### `tmpo log`

View your time tracking history.

**Options:**

- `--limit N` - Show N most recent entries (default: 20)
- `--milestone "name"` - Filter entries by milestone name
- `--project "name"` - Filter entries by project name
- `--today` - Show only today's entries
- `--week` - Show this week's entries
- `--date` - Show a specific date's entries

**Examples:**

```bash
tmpo log                            # Show recent entries
tmpo log --limit 50                 # Show more entries
tmpo log --project "Client Work"    # Filter by global project
tmpo log --milestone "Sprint 1"     # Filter by milestone
tmpo log --today                    # Show today's entries
tmpo log --week                     # Show this week's entries
tmpo log --date "2026-01-15"        # Show January 15th, 2026 entries
```

### `tmpo stats`

Display statistics about your tracked time.

**Options:**

- `--today` - Show only today's statistics
- `--week` - Show this week's statistics
- `--month` - Show this month's statistics
- `--date` - Show a specific date's statistics

**Examples:**

```bash
tmpo stats          # All-time stats
tmpo stats --today  # Today's stats
tmpo stats --week   # This week's stats
tmpo stats --date "2026-01-15" # January 15th, 2026 entries
```

## Configuration

### `tmpo config`

Configure global user preferences that apply across all projects. This includes:

- **Currency** - Your preferred currency for displaying earnings (USD, EUR, GBP, etc.)
- **Date Format** - Choose between MM/DD/YYYY, DD/MM/YYYY, or YYYY-MM-DD
- **Time Format** - Choose between 24-hour (15:30) or 12-hour (3:30 PM)
- **Timezone** - IANA timezone for your location (e.g., America/New_York)
- **Export Path** - Default directory for exported files (type "clear" to remove)

**Usage:**

```bash
tmpo config
# [tmpo] Global tmpo Configuration
# Current settings:
#   Currency:    USD
#   Date format: MM/DD/YYYY
#   Time format: 12-hour (AM/PM)
#   Timezone:    (local)
#   Export path: (current directory)
#
# Currency code (press Enter for USD): EUR
# Select date format: [use arrow keys]
# Select time format: [use arrow keys]
# Timezone (press Enter for local): Europe/London
# Export path (press Enter to keep current): ~/Documents/timesheets
#
# [tmpo] Configuration saved to ~/.tmpo/config.yaml
```

Settings are stored in `~/.tmpo/config.yaml` and affect how times and currency are displayed throughout tmpo.

See [Configuration Guide](configuration.md#global-configuration) for more details.

### `tmpo init`

Create a project configuration using an interactive form.

**Options:**

- `--global` / `-g` - Create a global project (track from any directory)
- `--accept-defaults` / `-a` - Skip prompts and use defaults

**Types of Projects:**

1. **Local Projects** (default) - Creates a `.tmporc` file in current directory
2. **Global Projects** (with `--global`) - Can be tracked from any directory

You'll be prompted to enter:

- **Project name** - Defaults to auto-detected name from Git repo or directory
- **Hourly rate** - Optional billing rate (press Enter to skip)
- **Description** - Optional project description (press Enter to skip)
- **Export path** - Optional default export directory (press Enter to skip)

**Examples:**

```bash
# Create a local project (.tmporc in current directory)
tmpo init
# [tmpo] Initialize Project Configuration
# Project name (my-project): [Enter custom name or press Enter for default]
# Hourly rate (press Enter to skip): 150
# Description (press Enter to skip): Client website redesign
# Export path (press Enter to skip): ~/Documents/client-exports

# Create a global project (track from anywhere)
tmpo init --global
# [tmpo] Initialize Global Project
# Project name: Client Work
# Hourly rate: 150
# Description: Consulting work
# Export path: ~/exports/client

# Quick setup with defaults
tmpo init --accept-defaults           # Local with defaults
tmpo init --global --accept-defaults  # Global with defaults
```

**Using Global Projects:**

Once created, track global projects from any directory:

```bash
cd /tmp
tmpo start --project "Client Work" "Working on feature"
tmpo log --project "Client Work"
tmpo resume --project "Client Work"
```

See [Configuration Guide](configuration.md) for details on project configuration.

## Milestone Management

Milestones help you organize time entries into time-boxed periods like sprints, releases, or project phases. When a milestone is active, all new time entries are automatically tagged with it.

### `tmpo milestone start [name]`

Start a new milestone for the current project, or the one specified. All new time entries will be automatically tagged with this milestone until you finish it.

**Options:**

- `--project "name"` / `-p "name"` - Start a milestone for a specific global project

**Examples:**

```bash
tmpo milestone start "Sprint 1"
tmpo milestone start "Release 2.0"
tmpo milestone start "Q1 Planning" --project "GlobalProject"
```

**Notes:**

- Only one milestone can be active per project at a time
- Starting a milestone when one is already active will show an error
- New time entries created with `tmpo start` are automatically tagged

### `tmpo milestone finish`

Finish the currently active milestone for the current project, or the one specified. This stops auto-tagging new entries and marks the milestone as completed.

**Options:**

- `--project "name"` / `-p "name"` - Finish a milestone for a specific global project

```bash
tmpo milestone finish
tmpo milestone finish --project "GlobalProject"
```

### `tmpo milestone status`

Show detailed information about the currently active milestone.

**Options:**

- `--project "name"` / `-p "name"` - Show milestone status for a specific global project

```bash
tmpo milestone status
tmpo milestone status --project "GlobalProject"
```

### `tmpo milestone list`

List all milestones for the current project, grouped by active and finished.

**Options:**

- `--project "name"` - Show milestones for a specific project
- `--all` - Show milestones from all projects

**Examples:**

```bash
tmpo milestone list                     # List milestones for current project
tmpo milestone list --project "webapp"  # List for specific project
tmpo milestone list --all               # List all milestones
```

**Output:**

```text
[tmpo] Milestones for my-project

─── Active ───
  Sprint 2
    Started: 9:00 AM  Duration: 2d 5h  Entries: 12

─── Finished ───
  Sprint 1
    Dec 1 9:00 AM - Dec 14 5:00 PM  Duration: 1w 6d 8h  Entries: 47
```

## Advanced Features

### `tmpo manual`

Create manual time entries for past work using an interactive prompt.

**Options:**

- `--project NAME` / `-p NAME` - Create entry for a specific global project

**Examples:**

```bash
tmpo manual                          # Create entry for current project
tmpo manual --project "Client Work"  # Create entry for global project
# Prompts for:
# - Project name (pre-filled if --project is used)
# - Start date and time (date format follows your config setting)
# - End date and time (date format follows your config setting)
# - Description
# - Milestone (optional, if milestones exist for the project)
```

> [!NOTE]
> Date input format adapts to your configured date format (`tmpo config`). For example, if you've set DD/MM/YYYY format, enter dates as "25-12-2024" rather than "12-25-2024".

This is useful for:

- Recording time before you started using tmpo
- Adding entries when you forgot to start the timer
- Correcting tracking mistakes
- Manually assigning entries to specific milestones (even finished ones)

### `tmpo edit`

Edit an existing time entry using an interactive menu. Select an entry and modify its start time, end time, description, or milestone assignment.

**Options:**

- `--project NAME` / `-p NAME` - Edit entries for a specific global project
- `--show-all-projects` - Show project selection before entry selection

**Examples:**

```bash
tmpo edit                           # Edit entries from current project
tmpo edit --project "Client Work"   # Edit entries for global project
tmpo edit --show-all-projects       # Select project first, then entry
```

**Interactive Flow:**

1. Select an entry from the list (shows completed entries only)
2. Edit start date and time (dates use your configured format - press Enter to keep current value)
3. Edit end date and time (dates use your configured format - press Enter to keep current value)
4. Edit description (press Enter to keep current value)
5. Assign to milestone (optional - select from available milestones or "(None)" to remove)
6. Review your changes with a diff view
7. Confirm to save or discard changes

**Milestone Assignment with Date Warnings:**

When assigning an entry to a milestone, tmpo checks if the entry's date falls within the milestone's timeframe. If the entry is outside the milestone's date range, you'll see an informative warning:

```text
⚠️  Entry not within milestone timeframe
Entry starts (Jan 5, 2024) before milestone began (Jan 10, 2024)
This is allowed - milestones are organizational tags and work with any date range.

Assign this entry to the milestone? [Yes/No]
```

You can freely assign entries to any milestone regardless of dates - milestones are organizational tags, not strict time boundaries. This is useful for reorganizing historical entries or handling edge cases.

**When to use:**

- Correct accidentally recorded times
- Fix typos in descriptions
- Adjust times when you forgot to stop the timer
- Reassign entries to different milestones
- Add milestone tags to entries created before the milestone existed
- Update entries after reviewing your work log

### `tmpo delete`

Delete a time entry using an interactive menu. Select an entry and confirm deletion.

**Options:**

- `--show-all-projects` - Show project selection before entry selection

**Examples:**

```bash
tmpo delete                        # Delete entries from current project
tmpo delete --show-all-projects    # Select project first, then entry
```

**Interactive Flow:**

1. Select an entry from the list (shows all entries, including running ones)
2. Review the entry details
3. Confirm deletion (defaults to "No" for safety)

**When to use:**

- Remove duplicate entries
- Delete test/accidental entries
- Clean up your time tracking history

### `tmpo undo`

Revert the most recent time-tracking action in case of a mistake. A single confirmation prompt is shown before any change is made.

**Supported actions:**

| Action        | What undo does                                    |
|---------------|---------------------------------------------------|
| `tmpo start`  | Deletes the entry that was just created           |
| `tmpo stop`   | Re-opens the stopped entry as a running timer     |
| `tmpo pause`  | Re-opens the paused entry as a running timer      |
| `tmpo resume` | Deletes the new entry that was created to resume  |
| `tmpo manual` | Deletes the manually created entry                |
| `tmpo delete` | Restores the deleted entry with its original data |

**Examples:**

```bash
tmpo undo
# ↩️  Last action: Stopped tracking my-project
#
# Undo this action? [y/N]: y
# ↩️  Undo successful.

tmpo undo
# ⚠️  Nothing to undo.
```

**Notes:**

- Only the single most recent action can be undone — tmpo does not maintain a full history stack
- Running `tmpo undo` a second time will show "Nothing to undo"
- Undoing a `stop` or `pause` will fail if another timer is already running — stop it first with `tmpo stop`

### `tmpo export`

Export your time tracking data to CSV or JSON.

**Options:**

- `--format [csv|json]` - Output format (default: csv)
- `--project "Name"` - Filter by specific project
- `--milestone "Name"` - Filter by milestone name
- `--today` - Export only today's entries
- `--week` - Export this week's entries
- `--output filename` - Specify output file path

**Examples:**

```bash
tmpo export                              # Export all as CSV
tmpo export --format json                # Export as JSON
tmpo export --project "Client Work"      # Filter by global project
tmpo export --milestone "Sprint 1"       # Filter by milestone
tmpo export --today                      # Export today's entries
tmpo export --week                       # Export this week
tmpo export --output timesheet.csv       # Specify output file
tmpo export --project "Consulting" --format json  # Global project to JSON
```

**CSV Format:**

```csv
Project,Start Time,End Time,Duration (hours),Description,Milestone
my-project,2024-01-15 14:30:00,2024-01-15 16:45:00,2.25,Implementing feature,Sprint 1
```

**JSON Format:**

```json
[
  {
    "project": "my-project",
    "start_time": "2024-01-15T14:30:00-05:00",
    "end_time": "2024-01-15T16:45:00-05:00",
    "duration_hours": 2.25,
    "description": "Implementing feature",
    "milestone": "Sprint 1"
  }
]
```

## Backup Management

tmpo can back up and restore your entire database. Backups are plain SQLite files stored in `~/.tmpo/backups/` and can be created, listed, restored, or deleted at any time.

### `tmpo backup create`

Create a snapshot of your database. The backup file is named `tmpo-YYYYMMDD-HHMMSS.db`.

**Examples:**

```bash
tmpo backup create
# 💾  Backup created successfully
#
#   File:   tmpo-20260429-202433.db
#   Path:   /Users/you/.tmpo/backups/tmpo-20260429-202433.db
#   Size:   44.0 KB
```

> [!NOTE]
> A timer must not be running when you create a backup. Stop any active session with `tmpo stop` first.

### `tmpo backup list`

List all existing backups with their ID, creation date, size, and schema version status.

**Examples:**

```bash
tmpo backup list
#
#   ID    Created                       Size      Schema
#
#   1     04/29/2026 8:24 PM            44.0 KB   ✅ v1 (current)
#   2     04/28/2026 9:00 AM            43.5 KB   ✅ v1 (current)
```

**Notes:**

- IDs are assigned newest-first and can be used with `--id` in `restore` and `delete`
- Schema shows whether the backup's database schema matches the current binary
- Backups created before a tmpo upgrade may show ⚠️ `v0 (outdated)` — they can still be restored, and migrations will run automatically after

### `tmpo backup restore`

Restore your database from a backup. **This overwrites your current database.**

**Options:**

- `--id VALUE` / `-i VALUE` - Backup ID or filename to restore, skips interactive selection

**Examples:**

```bash
tmpo backup restore                              # Interactive selection prompt
tmpo backup restore --id 2                       # Restore by ID
tmpo backup restore -i tmpo-20260428-090000.db   # Restore by filename
```

**Interactive Flow:**

1. Select a backup from the list using arrow keys
2. If the backup schema is outdated, a warning is shown (migrations will auto-run after restore)
3. Confirm restoration before overwriting your current database

### `tmpo backup delete`

Permanently delete a backup. **This cannot be undone.**

**Options:**

- `--id VALUE` / `-i VALUE` - Backup ID or filename to delete, skips interactive selection

**Examples:**

```bash
tmpo backup delete                              # Interactive selection prompt
tmpo backup delete --id 2                       # Delete by ID
tmpo backup delete -i tmpo-20260428-090000.db   # Delete by filename
```

**Interactive Flow:**

1. Select a backup from the list using arrow keys
2. Confirm permanent deletion before proceeding

## Tips and Workflows

### Taking Breaks with Pause/Resume

Use pause and resume for quick breaks without losing context:

```bash
tmpo start "Implementing authentication"
# ... work for a while ...
tmpo pause    # Take a lunch break
# ... break time ...
tmpo resume   # Continue same task
# ... more work ...
tmpo stop     # Done for the day
```

This creates separate entries for each work session, making it easy to see your actual working time versus break time when reviewing your log.

### Quick Daily Review

```bash
# See what you worked on today
tmpo stats --today
tmpo log --limit 10
```

### Weekly Timesheet Export

```bash
# Export this week's entries for invoicing
tmpo export --week --output timesheet-$(date +%Y-%m-%d).csv
```

### Multi-Project Workflow

**Option 1: Local .tmporc Files (directory-based)**

Create a `.tmporc` file in each project directory with different hourly rates:

```bash
# Client A - $150/hour
cd ~/projects/client-a
tmpo init
# Interactive prompts:
# Project name (client-a): Client A
# Hourly rate: 150
# Description: [press Enter to skip]

# Client B - $175/hour
cd ~/projects/client-b
tmpo init
# Interactive prompts:
# Project name (client-b): Client B
# Hourly rate: 175
# Description: [press Enter to skip]
```

Now `tmpo start` will automatically track to the correct project when you're in each directory.

**Option 2: Global Projects (track from anywhere)**

Create global projects once, then track from any directory:

```bash
# Create global projects
tmpo init --global
# Project name: Client A
# Hourly rate: 150

tmpo init --global
# Project name: Client B
# Hourly rate: 175

tmpo init --global
# Project name: Personal Projects
# Hourly rate: 0

# Now track from anywhere
cd /tmp
tmpo start --project "Client A" "Working on feature X"
# ... work ...
tmpo stop

tmpo start --project "Client B" "Code review"
# ... work ...
tmpo stop

# Switch projects without changing directories
tmpo start --project "Personal Projects" "Weekend coding"
```

**Option 3: Mix Both**

Use global projects for non-directory work (consulting, meetings) and local `.tmporc` files for code projects:

```bash
# Global projects for general work
tmpo init --global
# Project name: Consulting
# Hourly rate: 200

# Local .tmporc for specific codebases
cd ~/projects/client-website
tmpo init
# Project name: Client Website
# Hourly rate: 150
```

### Tracking Without Descriptions

You can always start tracking immediately and add context later by checking your git commits:

```bash
tmpo start
# ... do work ...
tmpo stop

# Later, correlate with git log to recall what you did
git log --since="2 hours ago" --oneline
```

### Sprint/Milestone-Based Workflow

Organize your work by sprints or project phases using milestones:

```bash
# Start a new sprint
tmpo milestone start "Sprint 5"

# All your work during this sprint is automatically tagged
tmpo start "Implement user authentication"
# ... work ...
tmpo stop

tmpo start "Fix bug #123"
# ... work ...
tmpo stop

# Review progress at any time
tmpo milestone status

# View all entries for this sprint
tmpo log --milestone "Sprint 5"

# When the sprint ends
tmpo milestone finish

# Review completed sprint
tmpo milestone list
```

### Retrospective Analysis

Use milestones to analyze your work across different phases:

```bash
# Compare time spent across different milestones
tmpo milestone list --all

# Export specific milestone data for reporting
tmpo export --milestone "Sprint 1"

# Get detailed breakdown by milestone
tmpo stats  # Shows breakdown by project and milestone
```
