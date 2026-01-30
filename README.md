# Hours Signer

A tool to add signature blocks to hours/timesheet PDFs.

## Features

- **Interactive TUI** - Run without arguments for a guided interface
- **Setup Wizard** - Automatic configuration on first run
- Adds employee and manager signature blocks to the last page of a PDF
- Pre-fills employee date with current date (Dutch format: dd-mm-yyyy)
- Embeds employee signature image
- Configurable via config file or command-line flags
- Output filename defaults to `Urenstaat-<year>-<month>-Joel-Grimberg.pdf`

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap joelgrimberg/hours-signer https://github.com/joelgrimberg/hours-signer
brew install hours-signer
```

### Build from Source

```bash
go build -o hours-signer .
mv hours-signer /usr/local/bin/
```

## Quick Start

### Interactive Mode (TUI)

Simply run without any arguments:

```bash
hours-signer
```

On first run, a setup wizard will guide you through configuration:
1. **Signature path** - Path to your signature image
2. **Employee name** - Your name
3. **Manager name** - Your manager's name

After setup, use the main menu to:
- **[s]** Sign a PDF - Opens a file picker to select your timesheet
- **[c]** Configure - Re-run the setup wizard
- **[q]** Quit

### CLI Mode

For scripting or quick use, pass flags directly:

```bash
hours-signer -input timesheet.pdf
```

## Configuration

### Config File Location

```
~/.config/hours-signer/config.json
```

### Initialize Config (CLI)

```bash
hours-signer -init
```

Creates a config file with default values:

```json
{
  "signature_path": "~/.config/hours-signer/signature.png",
  "employee_name": "Joël Grimberg",
  "manager_name": "Rob van der Pouw Kraan"
}
```

### Config Options

| Option | Description | Default |
|--------|-------------|---------|
| `signature_path` | Path to signature image (PNG/JPG). **Required.** | `""` |
| `employee_name` | Default employee name | `"Joël Grimberg"` |
| `manager_name` | Default manager name | `"Rob van der Pouw Kraan"` |

### Setting Up Your Signature

1. Copy your signature image to the config directory:
   ```bash
   mkdir -p ~/.config/hours-signer
   cp /path/to/your/signature.png ~/.config/hours-signer/signature.png
   ```

2. Run the app and configure via the TUI, or edit the config file directly.

### View Current Config

```bash
hours-signer -show-config
```

## Usage

```bash
# Basic usage
hours-signer -input timesheet.pdf

# Specify output filename
hours-signer -input timesheet.pdf -output signed.pdf

# Override employee/manager names
hours-signer -input timesheet.pdf -employee "John Doe" -manager "Jane Smith"

# Use a specific signature file (overrides config)
hours-signer -input timesheet.pdf -signature /path/to/signature.png
```

## Command-Line Flags

| Flag | Description |
|------|-------------|
| `-input` | Input PDF file (required in CLI mode) |
| `-output` | Output PDF file (default: `Urenstaat-<year>-<month>-Joel-Grimberg.pdf`) |
| `-employee` | Employee name (default: from config) |
| `-manager` | Manager name (default: from config) |
| `-signature` | Path to signature image (default: from config) |
| `-init` | Initialize config file with defaults |
| `-show-config` | Show current configuration |

## Output

The tool adds to the last page of the PDF:

**Left side (Employee):**
- Werknemer: [name]
- Datum: [current date]
- Handtekening: [signature image]

**Right side (Manager):**
- Manager: [name]
- Datum: [empty - to be filled manually]
- Handtekening: [empty - to be signed manually]
