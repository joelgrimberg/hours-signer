# Hours Signer

A tool to add signature blocks to hours/timesheet PDFs. Available in both Python and Go versions.

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
# Build the binary
go build -o hours-signer .

# Optionally, move to a directory in your PATH
mv hours-signer /usr/local/bin/
```

### Python Version

```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install pypdf reportlab
```

## Quick Start

### Interactive Mode (TUI)

Simply run without any arguments:

```bash
./hours-signer
```

On first run, a setup wizard will guide you through configuration:
1. **Signature path** - Path to your signature image (or leave empty for embedded)
2. **Employee name** - Your name
3. **Manager name** - Your manager's name

After setup, use the main menu to:
- **[s]** Sign a PDF - Opens a file picker to select your timesheet
- **[c]** Configure - Re-run the setup wizard
- **[q]** Quit

### CLI Mode

For scripting or quick use, pass flags directly:

```bash
./hours-signer -input timesheet.pdf
```

## Configuration

### Config File Location

The Go version uses a JSON config file stored at:

```
~/.config/hours-signer/config.json
```

### Initialize Config (CLI)

Create a default config file via command line:

```bash
./hours-signer -init
```

This creates a config file with default values:

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
./hours-signer -show-config
```

## Usage

### Go Version

```bash
# Basic usage (uses defaults from config or embedded values)
./hours-signer -input timesheet.pdf

# Specify output filename
./hours-signer -input timesheet.pdf -output signed.pdf

# Override employee/manager names
./hours-signer -input timesheet.pdf -employee "John Doe" -manager "Jane Smith"

# Use a specific signature file (overrides config)
./hours-signer -input timesheet.pdf -signature /path/to/signature.png
```

### Python Version

```bash
source venv/bin/activate
python hours_signer.py -input timesheet.pdf
python hours_signer.py -input timesheet.pdf -output signed.pdf
python hours_signer.py -input timesheet.pdf -employee "John Doe" -manager "Jane Smith"
```

## Command-Line Flags

### Go Version

| Flag | Description |
|------|-------------|
| `-input` | Input PDF file (required) |
| `-output` | Output PDF file (default: `Urenstaat-<year>-<month>-Joel-Grimberg.pdf`) |
| `-employee` | Employee name (default: from config) |
| `-manager` | Manager name (default: from config) |
| `-signature` | Path to signature image (default: from config or embedded) |
| `-init` | Initialize config file with defaults |
| `-show-config` | Show current configuration |

### Python Version

| Flag | Description |
|------|-------------|
| `-input` | Input PDF file (required) |
| `-output` | Output PDF file (default: `Urenstaat-<year>-<month>-Joel-Grimberg.pdf`) |
| `-employee` | Employee name (default: `Joël Grimberg`) |
| `-manager` | Manager name (default: `Rob van der Pouw Kraan`) |

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

