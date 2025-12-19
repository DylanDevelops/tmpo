# Usage Guide

Complete reference for all tmpo commands and features.

## Basic Commands

### `tmpo start [description]`

Start tracking time for the current project. Automatically detects the project name from:

1. `.tmporc` configuration file (if present)
2. Git repository name
3. Current directory name

**Examples:**

```bash
tmpo start                             # Start tracking
tmpo start "Fix authentication bug"    # Start with description
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

```bash
tmpo resume
# Output:
# [tmpo] Resumed tracking time for my-project
#     Description: Implementing feature
```

**Use cases:**

- Continue work after a break
- Resume after accidentally stopping the timer
- Quickly restart the same task

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

**Examples:**

```bash
tmpo log             # Show recent entries
tmpo log --limit 50  # Show more entries
```

### `tmpo stats`

Display statistics about your tracked time.

**Options:**

- `--today` - Show only today's statistics
- `--week` - Show this week's statistics
- `--month` - Show this month's statistics

**Examples:**

```bash
tmpo stats          # All-time stats
tmpo stats --today  # Today's stats
tmpo stats --week   # This week's stats
```

## Project Configuration

### `tmpo init`

Create a `.tmporc` configuration file for the current project using an interactive form. You'll be prompted to enter:

- **Project name** - Defaults to auto-detected name from Git repo or directory
- **Hourly rate** - Optional billing rate (press Enter to skip)
- **Description** - Optional project description (press Enter to skip)

**Interactive Mode (default):**

```bash
tmpo init
# [tmpo] Initialize Project Configuration
# Project name (my-project): [Enter custom name or press Enter for default]
# Hourly rate (press Enter to skip): 150
# Description (press Enter to skip): Client website redesign
```

**Quick Mode:**

Use the `--accept-defaults` flag to skip all prompts and use auto-detected defaults:

```bash
tmpo init --accept-defaults   # Creates .tmporc with defaults, no prompts
```

This creates a `.tmporc` file with:

- Project name from Git repo or directory name
- Hourly rate of 0 (disabled)
- Empty description

See [Configuration Guide](configuration.md) for details on the `.tmporc` file format and manual editing.

## Advanced Features

### `tmpo manual`

Create manual time entries for past work using an interactive prompt.

```bash
tmpo manual
# Prompts for:
# - Project name
# - Start date and time
# - End date and time
# - Description
```

This is useful for:

- Recording time before you started using tmpo
- Adding entries when you forgot to start the timer
- Correcting tracking mistakes

### `tmpo edit`

Edit an existing time entry using an interactive menu. Select an entry and modify its start time, end time, or description.

**Options:**

- `--show-all-projects` - Show project selection before entry selection

**Examples:**

```bash
tmpo edit                        # Edit entries from current project
tmpo edit --show-all-projects    # Select project first, then entry
```

**Interactive Flow:**

1. Select an entry from the list (shows completed entries only)
2. Edit start date and time (press Enter to keep current value)
3. Edit end date and time (press Enter to keep current value)
4. Edit description (press Enter to keep current value)
5. Review your changes with a diff view
6. Confirm to save or discard changes

**When to use:**

- Correct accidentally recorded times
- Fix typos in descriptions
- Adjust times when you forgot to stop the timer
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

### `tmpo export`

Export your time tracking data to CSV or JSON.

**Options:**

- `--format [csv|json]` - Output format (default: csv)
- `--project "Name"` - Filter by specific project
- `--today` - Export only today's entries
- `--week` - Export this week's entries
- `--month` - Export this month's entries
- `--output filename` - Specify output file path

**Examples:**

```bash
tmpo export                              # Export all as CSV
tmpo export --format json                # Export as JSON
tmpo export --project "My Project"       # Filter by project
tmpo export --today                      # Export today's entries
tmpo export --week                       # Export this week
tmpo export --output timesheet.csv       # Specify output file
```

**CSV Format:**

```csv
Project,Description,Start,End,Duration (hours)
my-project,Implementing feature,2024-01-15 14:30:00,2024-01-15 16:45:00,2.25
```

**JSON Format:**

```json
[
  {
    "project": "my-project",
    "description": "Implementing feature",
    "start": "2024-01-15T14:30:00Z",
    "end": "2024-01-15T16:45:00Z",
    "duration_hours": 2.25
  }
]
```

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

Create a `.tmporc` file in each project directory with different hourly rates:

```bash
cd ~/projects/client-a
tmpo init --name "Client A" --rate 150

cd ~/projects/client-b
tmpo init --name "Client B" --rate 175
```

Now `tmpo start` will automatically track to the correct project when you're in each directory.

### Tracking Without Descriptions

You can always start tracking immediately and add context later by checking your git commits:

```bash
tmpo start
# ... do work ...
tmpo stop

# Later, correlate with git log to recall what you did
git log --since="2 hours ago" --oneline
```
