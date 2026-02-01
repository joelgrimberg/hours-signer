# Hours Signer

PDF signing tool that adds employee/manager signature blocks to timesheet PDFs.

## Architecture

```
main.go
├── Config management (~/.config/hours-signer/config.json)
├── TUI (Bubble Tea)
│   ├── Setup wizard (first run)
│   ├── Main menu (sign/configure)
│   └── File picker
├── CLI mode (flags: -input, -output, -version, etc.)
├── Version checker (GitHub API)
└── PDF signing (pdfcpu watermarks)
```

## Features

- **Interactive TUI**: File picker, setup wizard, main menu
- **CLI mode**: Pass flags for scripting
- **Config file**: Stores signature path, employee/manager names
- **Version check**: Shows current version + update availability on startup
- **PDF signing**: Adds signature blocks to last page using pdfcpu watermarks

## Tech Stack

- **Go 1.23+**
- **Bubble Tea** - TUI framework
- **Lipgloss** - TUI styling
- **pdfcpu** - PDF manipulation
- **GoReleaser** - Build/release automation

## Release

```bash
git tag v1.x.x
git push origin v1.x.x
# GitHub Actions builds + updates Homebrew formula
brew update && brew upgrade hours-signer
```

## Install

```bash
brew tap joelgrimberg/hours-signer https://github.com/joelgrimberg/hours-signer
brew install hours-signer
```
