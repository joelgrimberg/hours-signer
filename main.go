package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ============================================================================
// Configuration
// ============================================================================

type Config struct {
	SignaturePath string `json:"signature_path"`
	EmployeeName  string `json:"employee_name"`
	ManagerName   string `json:"manager_name"`
}

func DefaultConfig() Config {
	return Config{
		SignaturePath: "",
		EmployeeName:  "Joël Grimberg",
		ManagerName:   "Rob van der Pouw Kraan",
	}
}

func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "hours-signer", "config.json")
}

func ConfigExists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

func LoadConfig() Config {
	cfg := DefaultConfig()
	configPath := ConfigPath()
	if configPath == "" {
		return cfg
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	return cfg
}

func SaveConfig(cfg Config) error {
	configPath := ConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine config path")
	}
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// ============================================================================
// Styles
// ============================================================================

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

// ============================================================================
// TUI Model
// ============================================================================

type screen int

const (
	screenSetupWelcome screen = iota
	screenSetupSignature
	screenSetupEmployee
	screenSetupManager
	screenSetupConfirm
	screenMain
	screenFilePicker
	screenSigning
	screenResult
)

type model struct {
	screen       screen
	config       Config
	configExists bool

	// Text inputs for setup
	inputs      []textinput.Model
	focusIndex  int

	// File picker
	filepicker   filepicker.Model
	selectedFile string

	// Result
	resultMsg string
	resultErr error

	// Window size
	width  int
	height int
}

func initialModel() model {
	configExists := ConfigExists()
	var cfg Config
	if configExists {
		cfg = LoadConfig()
	} else {
		cfg = DefaultConfig()
	}

	// Create text inputs
	inputs := make([]textinput.Model, 3)

	// Signature path input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "~/.config/hours-signer/signature.png"
	inputs[0].CharLimit = 256
	inputs[0].Width = 50

	// Employee name input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Joël Grimberg"
	inputs[1].SetValue(cfg.EmployeeName)
	inputs[1].CharLimit = 100
	inputs[1].Width = 50

	// Manager name input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Rob van der Pouw Kraan"
	inputs[2].SetValue(cfg.ManagerName)
	inputs[2].CharLimit = 100
	inputs[2].Width = 50

	// File picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".pdf"}
	fp.CurrentDirectory, _ = os.Getwd()

	startScreen := screenMain
	if !configExists {
		startScreen = screenSetupWelcome
	}

	return model{
		screen:       startScreen,
		config:       cfg,
		configExists: configExists,
		inputs:       inputs,
		filepicker:   fp,
	}
}

func (m model) Init() tea.Cmd {
	if m.screen == screenMain {
		return nil
	}
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.screen == screenMain || m.screen == screenSetupWelcome || m.screen == screenResult {
				return m, tea.Quit
			}
		case "esc":
			if m.screen == screenFilePicker {
				m.screen = screenMain
				return m, nil
			}
		}
	}

	switch m.screen {
	case screenSetupWelcome:
		return m.updateSetupWelcome(msg)
	case screenSetupSignature:
		return m.updateSetupSignature(msg)
	case screenSetupEmployee:
		return m.updateSetupEmployee(msg)
	case screenSetupManager:
		return m.updateSetupManager(msg)
	case screenSetupConfirm:
		return m.updateSetupConfirm(msg)
	case screenMain:
		return m.updateMain(msg)
	case screenFilePicker:
		return m.updateFilePicker(msg)
	case screenResult:
		return m.updateResult(msg)
	}

	return m, nil
}

func (m model) updateSetupWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter", " ":
			m.screen = screenSetupSignature
			m.inputs[0].Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updateSetupSignature(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			val := m.inputs[0].Value()
			if val == "" {
				// Signature path is required - don't proceed
				return m, nil
			}
			m.config.SignaturePath = val
			m.screen = screenSetupEmployee
			m.inputs[0].Blur()
			m.inputs[1].Focus()
			return m, textinput.Blink
		}
	}
	var cmd tea.Cmd
	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m model) updateSetupEmployee(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			val := m.inputs[1].Value()
			if val == "" {
				val = m.inputs[1].Placeholder
			}
			m.config.EmployeeName = val
			m.screen = screenSetupManager
			m.inputs[1].Blur()
			m.inputs[2].Focus()
			return m, textinput.Blink
		}
	}
	var cmd tea.Cmd
	m.inputs[1], cmd = m.inputs[1].Update(msg)
	return m, cmd
}

func (m model) updateSetupManager(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			val := m.inputs[2].Value()
			if val == "" {
				val = m.inputs[2].Placeholder
			}
			m.config.ManagerName = val
			m.screen = screenSetupConfirm
			m.inputs[2].Blur()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.inputs[2], cmd = m.inputs[2].Update(msg)
	return m, cmd
}

func (m model) updateSetupConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter", "y":
			if err := SaveConfig(m.config); err != nil {
				m.resultErr = err
				m.screen = screenResult
				return m, nil
			}
			m.configExists = true
			m.screen = screenMain
			return m, nil
		case "n":
			m.screen = screenSetupSignature
			m.inputs[0].Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "s", "1":
			m.screen = screenFilePicker
			return m, m.filepicker.Init()
		case "c", "2":
			// Pre-fill inputs with current config values
			m.inputs[0].SetValue(m.config.SignaturePath)
			m.inputs[1].SetValue(m.config.EmployeeName)
			m.inputs[2].SetValue(m.config.ManagerName)
			m.inputs[0].Focus()
			m.screen = screenSetupSignature
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updateFilePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
		m.screen = screenSigning

		// Sign the PDF
		now := time.Now()
		output := fmt.Sprintf("Urenstaat-%d-%02d-Joel-Grimberg.pdf", now.Year(), now.Month())

		err := signPDF(path, output, m.config.EmployeeName, m.config.ManagerName, m.config.SignaturePath)
		if err != nil {
			m.resultErr = err
			m.resultMsg = ""
		} else {
			m.resultErr = nil
			m.resultMsg = output
		}
		m.screen = screenResult
		return m, nil
	}

	return m, cmd
}

func (m model) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter", " ":
			m.screen = screenMain
			m.resultErr = nil
			m.resultMsg = ""
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenSetupWelcome:
		return m.viewSetupWelcome()
	case screenSetupSignature:
		return m.viewSetupSignature()
	case screenSetupEmployee:
		return m.viewSetupEmployee()
	case screenSetupManager:
		return m.viewSetupManager()
	case screenSetupConfirm:
		return m.viewSetupConfirm()
	case screenMain:
		return m.viewMain()
	case screenFilePicker:
		return m.viewFilePicker()
	case screenSigning:
		return m.viewSigning()
	case screenResult:
		return m.viewResult()
	}
	return ""
}

func (m model) viewSetupWelcome() string {
	s := titleStyle.Render("📝 Hours Signer - Setup") + "\n\n"
	s += "Welcome! No configuration file found.\n"
	s += "Let's set up your signature settings.\n\n"
	s += subtitleStyle.Render(fmt.Sprintf("Config will be saved to: %s", ConfigPath())) + "\n\n"
	s += helpStyle.Render("Press Enter to continue • q to quit")
	return s
}

func (m model) viewSetupSignature() string {
	s := titleStyle.Render("📝 Hours Signer - Setup (1/3)") + "\n\n"
	s += "Enter the path to your signature image (PNG/JPG).\n"
	s += subtitleStyle.Render("Example: ~/.config/hours-signer/signature.png") + "\n\n"
	s += "Signature path:\n"
	s += m.inputs[0].View() + "\n\n"
	s += helpStyle.Render("Press Enter to continue")
	return s
}

func (m model) viewSetupEmployee() string {
	s := titleStyle.Render("📝 Hours Signer - Setup (2/3)") + "\n\n"
	s += "Enter the employee name.\n\n"
	s += "Employee name:\n"
	s += m.inputs[1].View() + "\n\n"
	s += helpStyle.Render("Press Enter to continue")
	return s
}

func (m model) viewSetupManager() string {
	s := titleStyle.Render("📝 Hours Signer - Setup (3/3)") + "\n\n"
	s += "Enter the manager name.\n\n"
	s += "Manager name:\n"
	s += m.inputs[2].View() + "\n\n"
	s += helpStyle.Render("Press Enter to continue")
	return s
}

func (m model) viewSetupConfirm() string {
	s := titleStyle.Render("📝 Hours Signer - Confirm Setup") + "\n\n"
	s += "Please confirm your settings:\n\n"

	s += fmt.Sprintf("  Signature:  %s\n", m.config.SignaturePath)
	s += fmt.Sprintf("  Employee:   %s\n", m.config.EmployeeName)
	s += fmt.Sprintf("  Manager:    %s\n\n", m.config.ManagerName)

	s += helpStyle.Render("Press y/Enter to save • n to edit again")
	return s
}

func (m model) viewMain() string {
	s := titleStyle.Render("📝 Hours Signer") + "\n\n"

	sigPath := m.config.SignaturePath
	if sigPath == "" {
		sigPath = errorStyle.Render("(not configured)")
	}
	s += subtitleStyle.Render("Current configuration:") + "\n"
	s += fmt.Sprintf("  Employee:   %s\n", m.config.EmployeeName)
	s += fmt.Sprintf("  Manager:    %s\n", m.config.ManagerName)
	s += fmt.Sprintf("  Signature:  %s\n\n", sigPath)

	if m.config.SignaturePath == "" {
		s += errorStyle.Render("⚠ Signature not configured - press c to configure") + "\n\n"
	}

	s += "What would you like to do?\n\n"
	s += "  [s] Sign a PDF\n"
	s += "  [c] Configure settings\n\n"

	s += helpStyle.Render("Press s to sign • c to configure • q to quit")
	return s
}

func (m model) viewFilePicker() string {
	s := titleStyle.Render("📂 Select PDF File") + "\n\n"
	s += m.filepicker.View() + "\n\n"
	s += helpStyle.Render("Enter to select • Esc to cancel")
	return s
}

func (m model) viewSigning() string {
	s := titleStyle.Render("⏳ Signing PDF...") + "\n\n"
	s += fmt.Sprintf("Processing: %s\n", m.selectedFile)
	return s
}

func (m model) viewResult() string {
	if m.resultErr != nil {
		s := errorStyle.Render("❌ Error") + "\n\n"
		s += fmt.Sprintf("%v\n\n", m.resultErr)
		s += helpStyle.Render("Press Enter to continue")
		return s
	}

	s := successStyle.Render("✓ PDF Signed Successfully!") + "\n\n"
	s += fmt.Sprintf("Output: %s\n\n", m.resultMsg)
	s += helpStyle.Render("Press Enter to continue")
	return s
}

// ============================================================================
// PDF Signing Logic
// ============================================================================

func getSignatureData(signaturePath string) ([]byte, error) {
	if signaturePath == "" {
		return nil, fmt.Errorf("signature path is required - please configure it first")
	}

	// Expand ~ to home directory
	if signaturePath[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			signaturePath = filepath.Join(home, signaturePath[1:])
		}
	}

	data, err := os.ReadFile(signaturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature file: %w", err)
	}
	return data, nil
}

func signPDF(inputPath, outputPath, employeeName, managerName, signaturePath string) error {
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF context: %w", err)
	}

	pageCount := ctx.PageCount
	conf := pdfmodel.NewDefaultConfiguration()

	sigData, err := getSignatureData(signaturePath)
	if err != nil {
		return err
	}

	sigFile, err := os.CreateTemp("", "signature-*.png")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(sigFile.Name())

	if _, err := sigFile.Write(sigData); err != nil {
		return fmt.Errorf("failed to write signature: %w", err)
	}
	sigFile.Close()

	pageSelection := []string{fmt.Sprintf("%d", pageCount)}
	currentDate := time.Now().Format("02-01-2006")

	employeeLabel := fmt.Sprintf("Werknemer: %s", employeeName)
	wmEmployeeLabel, err := api.TextWatermark(employeeLabel, "font:Helvetica, points:10, pos:bl, off:40 210, scale:1 abs, rot:0", true, false, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to create employee label watermark: %w", err)
	}

	employeeDateLabel := fmt.Sprintf("Datum: %s", currentDate)
	wmEmployeeDate, err := api.TextWatermark(employeeDateLabel, "font:Helvetica, points:10, pos:bl, off:40 195, scale:1 abs, rot:0", true, false, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to create employee date watermark: %w", err)
	}

	wmHandtekeningLabel, err := api.TextWatermark("Handtekening:", "font:Helvetica, points:10, pos:bl, off:40 180, scale:1 abs, rot:0", true, false, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to create handtekening label watermark: %w", err)
	}

	wmSignature, err := api.ImageWatermark(sigFile.Name(), "pos:bl, off:120 90, scale:.35 abs, rot:0", true, false, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to create signature watermark: %w", err)
	}

	managerLabel := fmt.Sprintf("Manager: %s", managerName)
	wmManagerLabel, err := api.TextWatermark(managerLabel, "font:Helvetica, points:10, pos:bl, off:350 210, scale:1 abs, rot:0", true, false, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to create manager label watermark: %w", err)
	}

	wmManagerDate, err := api.TextWatermark("Datum:", "font:Helvetica, points:10, pos:bl, off:350 195, scale:1 abs, rot:0", true, false, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to create manager date watermark: %w", err)
	}

	wmManagerHandtekening, err := api.TextWatermark("Handtekening:", "font:Helvetica, points:10, pos:bl, off:350 180, scale:1 abs, rot:0", true, false, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to create manager handtekening label watermark: %w", err)
	}

	reader := bytes.NewReader(inputData)
	var buf bytes.Buffer

	if err := api.AddWatermarks(reader, &buf, pageSelection, wmEmployeeLabel, conf); err != nil {
		return fmt.Errorf("failed to add employee label: %w", err)
	}

	reader = bytes.NewReader(buf.Bytes())
	buf.Reset()
	if err := api.AddWatermarks(reader, &buf, pageSelection, wmEmployeeDate, conf); err != nil {
		return fmt.Errorf("failed to add employee date: %w", err)
	}

	reader = bytes.NewReader(buf.Bytes())
	buf.Reset()
	if err := api.AddWatermarks(reader, &buf, pageSelection, wmHandtekeningLabel, conf); err != nil {
		return fmt.Errorf("failed to add handtekening label: %w", err)
	}

	reader = bytes.NewReader(buf.Bytes())
	buf.Reset()
	if err := api.AddWatermarks(reader, &buf, pageSelection, wmSignature, conf); err != nil {
		return fmt.Errorf("failed to add signature: %w", err)
	}

	reader = bytes.NewReader(buf.Bytes())
	buf.Reset()
	if err := api.AddWatermarks(reader, &buf, pageSelection, wmManagerLabel, conf); err != nil {
		return fmt.Errorf("failed to add manager label: %w", err)
	}

	reader = bytes.NewReader(buf.Bytes())
	buf.Reset()
	if err := api.AddWatermarks(reader, &buf, pageSelection, wmManagerDate, conf); err != nil {
		return fmt.Errorf("failed to add manager date: %w", err)
	}

	reader = bytes.NewReader(buf.Bytes())
	buf.Reset()
	if err := api.AddWatermarks(reader, &buf, pageSelection, wmManagerHandtekening, conf); err != nil {
		return fmt.Errorf("failed to add manager handtekening label: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// ============================================================================
// Main
// ============================================================================

func main() {
	// Check if any flags are provided (CLI mode)
	if len(os.Args) > 1 && (strings.HasPrefix(os.Args[1], "-") || strings.HasPrefix(os.Args[1], "--")) {
		runCLI()
		return
	}

	// No flags - run TUI
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runCLI() {
	cfg := LoadConfig()

	inputFile := flag.String("input", "", "Input PDF file (required)")
	outputFile := flag.String("output", "", "Output PDF file (default: Urenstaat-<year>-<month>-Joel-Grimberg.pdf)")
	employeeName := flag.String("employee", cfg.EmployeeName, "Employee name")
	managerName := flag.String("manager", cfg.ManagerName, "Manager name")
	signaturePath := flag.String("signature", cfg.SignaturePath, "Path to signature image (PNG/JPG)")
	initConfig := flag.Bool("init", false, "Initialize config file with defaults")
	showConfig := flag.Bool("show-config", false, "Show current configuration")
	flag.Parse()

	if *initConfig {
		if err := SaveConfig(DefaultConfig()); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config file created at: %s\n", ConfigPath())
		os.Exit(0)
	}

	if *showConfig {
		fmt.Printf("Config file: %s\n", ConfigPath())
		fmt.Printf("Employee name: %s\n", cfg.EmployeeName)
		fmt.Printf("Manager name: %s\n", cfg.ManagerName)
		if cfg.SignaturePath != "" {
			fmt.Printf("Signature path: %s\n", cfg.SignaturePath)
		} else {
			fmt.Println("Signature: (not configured)")
		}
		os.Exit(0)
	}

	if *inputFile == "" {
		fmt.Println("Error: -input is required")
		flag.Usage()
		os.Exit(1)
	}

	output := *outputFile
	if output == "" {
		now := time.Now()
		output = fmt.Sprintf("Urenstaat-%d-%02d-Joel-Grimberg.pdf", now.Year(), now.Month())
	}

	if err := signPDF(*inputFile, output, *employeeName, *managerName, *signaturePath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created signed PDF: %s\n", output)
	fmt.Printf("  Employee: %s\n", *employeeName)
	fmt.Printf("  Manager: %s\n", *managerName)
}
