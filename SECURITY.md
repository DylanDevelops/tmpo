# Security Policy

## Supported Versions

We release security updates for the following versions of tmpo:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1.0 | :x:                |

We recommend always using the latest stable release to ensure you have the most recent security patches.

## Security Considerations

tmpo is a local-first CLI tool that stores time tracking data on your machine. Here are some security aspects to be aware of:

### **Local Data Storage**

- All time entries are stored in a SQLite database at `$HOME/.tmpo/tmpo.db`
- The database is only accessible to your user account (standard file permissions apply)
- No data is transmitted over the network

### **Configuration Files**

- `.tmporc` files may contain project-specific configuration including hourly rates
- These files are stored in plain text and inherit directory permissions
- Be cautious when committing `.tmporc` files to version control if they contain sensitive rate information

### **Git Integration**

- tmpo uses Git commands for automatic project detection
- Only basic Git metadata (repository name) is accessed
- No Git credentials or remote repository data is used

## Reporting a Vulnerability

If you discover a security vulnerability in tmpo, please report it responsibly:

### **Preferred Method:**

[Report via GitHub Security Advisories](https://github.com/DylanDevelops/tmpo/security/advisories/new)

### **What to Include:**

- Description of the vulnerability
- Steps to reproduce the issue
- Affected versions (if known)
- Potential impact
- Any suggested fixes (optional)

### **Response Timeline:**

- We will acknowledge your report within 48 hours
- We will provide an initial assessment within 5 business days
- We will work to release a fix as quickly as possible depending on severity

### **Responsible Disclosure:**

Please do not publicly disclose the vulnerability until we have had a chance to address it and release a fix. We will credit security researchers who report valid vulnerabilities (unless you prefer to remain anonymous).

## Security Best Practices

When using tmpo, we recommend:

- Keep your tmpo binary updated to the latest version
- Be mindful of what information you include in time entry descriptions
- Review `.tmporc` files before committing them to public repositories
- Use appropriate file permissions for your `~/.tmpo/` directory
- Regularly backup your time tracking data if it's business-critical
