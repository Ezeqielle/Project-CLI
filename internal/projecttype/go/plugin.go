package goproject

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ezeqielle/pcli/internal/langenv"
	"github.com/ezeqielle/pcli/internal/postplugin"
)

type GoPlugin struct{}

func New() *GoPlugin {
	return &GoPlugin{}
}

func (p *GoPlugin) ID() string {
	return "go"
}

func (p *GoPlugin) DisplayName() string {
	return "Go"
}

func (p *GoPlugin) Description() string {
	return "Create a Go project using module path workflow"
}

func (p *GoPlugin) NewWizard() tea.Model {
	return NewGoWizardModel()
}

// -------------------------------------------
// GO WIZARD MODEL
// -------------------------------------------

type goWizardStep int

const (
	goStepModulePath goWizardStep = iota
	goStepSummary
	goStepInstallPrompt
	goStepInstalling
	goStepDone
)

type installProgressMsg struct{}
type installLogMsg struct {
	Line string
}
type installFinishedMsg struct {
	Err error
}

type GoWizardModel struct {
	step goWizardStep

	modulePath string
	projectDir string
	errMsg     string

	modulePathInput textinput.Model

	progress      progress.Model
	progressValue float64

	installEvents chan tea.Msg

	logLines []string
}

func NewGoWizardModel() GoWizardModel {
	defaultModule := loadDefaultModulePath()

	ti := textinput.New()
	ti.Placeholder = defaultModule
	ti.SetValue(defaultModule)
	ti.Focus()

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	return GoWizardModel{
		step:            goStepModulePath,
		modulePathInput: ti,
		progress:        prog,
		logLines:        make([]string, 0, 64),
	}
}

func (m GoWizardModel) Init() tea.Cmd {
	return nil
}

func (m GoWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Always update progress bar
	var pCmd tea.Cmd
	updatedModel, pCmd := m.progress.Update(msg)
	m.progress = updatedModel.(progress.Model)
	if pCmd != nil {
		cmds = append(cmds, pCmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {

		case goStepModulePath:
			switch msg.String() {
			case "enter":
				m.modulePath = strings.TrimSpace(m.modulePathInput.Value())
				if m.modulePath == "" {
					m.errMsg = "module path cannot be empty"
					return m, tea.Batch(cmds...)
				}
				m.projectDir = previewProjectDir(m.modulePath)
				m.errMsg = ""
				m.step = goStepSummary
				return m, tea.Batch(cmds...)

			case "ctrl+c":
				return m, tea.Quit
			}

		case goStepSummary:
			switch msg.String() {
			case "enter":
				if !langenv.IsInstalled(langenv.LanguageGo) {
					m.step = goStepInstallPrompt
					m.errMsg = ""
					return m, tea.Batch(cmds...)
				}

				dir, err := createGoProject(m.modulePath)
				if err != nil {
					m.errMsg = err.Error()
					m.projectDir = dir
					return m, tea.Batch(cmds...)
				}
				m.projectDir = dir
				m.errMsg = ""

				// After project creation, hand off to the first post-create plugin
				postPlugins := postplugin.All()
				if len(postPlugins) > 0 {
					return postPlugins[0].NewWizard(m.projectDir, "go"), nil
				}

				// If no post-create plugin is registered, fall back to the simple done screen
				m.step = goStepDone
				return m, tea.Batch(cmds...)

			case "esc":
				m.step = goStepModulePath
				m.errMsg = ""
				return m, tea.Batch(cmds...)

			case "ctrl+c":
				return m, tea.Quit
			}

		case goStepInstallPrompt:
			switch msg.String() {
			case "y", "Y", "enter":
				cmd, err := langenv.InstallCommand(langenv.LanguageGo)
				if err != nil {
					m.errMsg = err.Error()
					m.step = goStepSummary
					return m, tea.Batch(cmds...)
				}

				m.step = goStepInstalling
				m.errMsg = ""
				m.progressValue = 0.0
				m.progress.SetPercent(0.0)
				m.logLines = nil

				m.installEvents = make(chan tea.Msg)
				go runInstallWithOutput(cmd, m.installEvents)

				cmds = append(cmds, waitInstallEvent(m.installEvents))
				return m, tea.Batch(cmds...)

			case "n", "N", "esc":
				m.errMsg = "Go is required to create a Go project. Please install it and retry."
				m.step = goStepSummary
				return m, tea.Batch(cmds...)

			case "ctrl+c":
				return m, tea.Quit
			}

		case goStepInstalling:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			}

		case goStepDone:
			// any key exits
			return m, tea.Quit
		}

	case installLogMsg:
		if m.step == goStepInstalling {
			m.appendLogLine(msg.Line)

			// bump progress a bit per line
			m.progressValue += 0.02
			if m.progressValue > 0.95 {
				m.progressValue = 0.95
			}
			m.progress.SetPercent(m.progressValue)

			if m.installEvents != nil {
				cmds = append(cmds, waitInstallEvent(m.installEvents))
			}
		}

	case installFinishedMsg:
		if m.step == goStepInstalling {
			if msg.Err != nil {
				m.errMsg = fmt.Sprintf("Go installation failed: %v", msg.Err)
			} else {
				m.errMsg = "Go installation succeeded."
				m.progressValue = 1.0
				m.progress.SetPercent(1.0)
			}
			m.step = goStepSummary
			m.installEvents = nil
		}
	}

	// Update text input in module path step
	if m.step == goStepModulePath {
		var tiCmd tea.Cmd
		m.modulePathInput, tiCmd = m.modulePathInput.Update(msg)
		if tiCmd != nil {
			cmds = append(cmds, tiCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *GoWizardModel) appendLogLine(line string) {
	if line == "" {
		return
	}
	m.logLines = append(m.logLines, line)
	const maxLines = 50
	if len(m.logLines) > maxLines {
		m.logLines = m.logLines[len(m.logLines)-maxLines:]
	}
}

func (m GoWizardModel) View() string {
	switch m.step {

	case goStepModulePath:
		var errLine string
		if m.errMsg != "" {
			errLine = "\n\nError: " + m.errMsg
		}

		return "Go project – module path\n\n" +
			m.modulePathInput.View() + "\n\n" +
			"[enter] Continue   [ctrl+c] Quit" +
			errLine + "\n"

	case goStepSummary:
		var b strings.Builder

		b.WriteString("Summary – Go project\n\n")
		b.WriteString(fmt.Sprintf("Module path:  %s\n", m.modulePath))
		b.WriteString(fmt.Sprintf("Project path: %s\n\n", previewProjectDir(m.modulePath)))

		if m.errMsg != "" {
			b.WriteString("Info: " + m.errMsg + "\n\n")
		}

		b.WriteString("[enter] Create   [esc] Back   [ctrl+c] Quit\n")

		return b.String()

	case goStepInstallPrompt:
		return "Go is not installed on this system.\n\n" +
			"Do you want to install Go now?\n\n" +
			"[y] Yes   [n] No   [ctrl+c] Quit\n"

	case goStepInstalling:
		var b strings.Builder

		b.WriteString("Installing Go...\n\n")
		b.WriteString(m.progress.View())
		b.WriteString("\n\nLogs (latest):\n")

		for _, line := range m.logLines {
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString("\n[ctrl+c] Cancel\n")

		return b.String()

	case goStepDone:
		return fmt.Sprintf(
			"Go project created.\n\nModule path:  %s\nProject path: %s\n\n[any key] Exit\n",
			m.modulePath,
			m.projectDir,
		)
	}

	return ""
}

// -------------------------------------------
// Install streaming helpers
// -------------------------------------------

func runInstallWithOutput(cmd *exec.Cmd, ch chan<- tea.Msg) {
	defer close(ch)

	lines := make(chan string)

	// Reader goroutine: transform lines into tea.Msg
	go func() {
		for line := range lines {
			ch <- installLogMsg{Line: line}
			ch <- installProgressMsg{}
		}
	}()

	// Blocking call – runs command and streams output to `lines`
	err := langenv.RunWithOutput(cmd, lines)
	// Done with output, close lines so reader goroutine stops
	close(lines)

	ch <- installFinishedMsg{Err: err}
}

func waitInstallEvent(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return installFinishedMsg{Err: nil}
		}
		return msg
	}
}

// -------------------------------------------
// Env, paths, project creation
// -------------------------------------------

func loadDefaultModulePath() string {
	data, err := os.ReadFile(".env")
	if err != nil {
		return "github.com/you/your-service"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "DEFAULT_GO_PROJECT_MODULE_PATH=") {
			val := strings.TrimPrefix(line, "DEFAULT_GO_PROJECT_MODULE_PATH=")
			val = strings.TrimSpace(val)
			if val != "" {
				return langenv.ExpandPathEnv(val)
			}
		}
	}

	return "github.com/you/your-service"
}

func loadDefaultProjectBasePath() string {
	data, err := os.ReadFile(".env")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "DEFAULT_GO_PROJECT_PATH=") {
				val := strings.TrimPrefix(line, "DEFAULT_GO_PROJECT_PATH=")
				val = strings.TrimSpace(val)
				if val != "" {
					return langenv.ExpandPathEnv(val)
				}
			}
		}
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, "Documents")
	}

	return "."
}

func deriveProjectNameFromModule(modulePath string) string {
	modulePath = strings.TrimSpace(modulePath)
	if modulePath == "" {
		return "go-project"
	}
	parts := strings.Split(modulePath, "/")
	return parts[len(parts)-1]
}

func previewProjectDir(modulePath string) string {
	base := loadDefaultProjectBasePath()
	name := deriveProjectNameFromModule(modulePath)
	return filepath.Join(base, name)
}

func createGoProject(modulePath string) (string, error) {
	modulePath = strings.TrimSpace(modulePath)
	if modulePath == "" {
		return "", fmt.Errorf("module path cannot be empty")
	}

	base := loadDefaultProjectBasePath()
	name := deriveProjectNameFromModule(modulePath)
	projectDir := filepath.Join(base, name)

	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return projectDir, fmt.Errorf("failed to create project directory: %w", err)
	}

	modInit := exec.Command("go", "mod", "init", modulePath)
	modInit.Dir = projectDir

	if out, err := modInit.CombinedOutput(); err != nil {
		return projectDir, fmt.Errorf("go mod init failed: %v\n%s", err, string(out))
	}

	modTidy := exec.Command("go", "mod", "tidy")
	modTidy.Dir = projectDir

	if out, err := modTidy.CombinedOutput(); err != nil {
		return projectDir, fmt.Errorf("go mod tidy failed: %v\n%s", err, string(out))
	}

	return projectDir, nil
}
