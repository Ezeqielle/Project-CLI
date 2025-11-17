package global

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// GlobalPlugin implements a post-create plugin that
// adds global + type-specific files/folders.
type GlobalPlugin struct{}

func New() *GlobalPlugin {
	return &GlobalPlugin{}
}

func (p *GlobalPlugin) ID() string {
	return "global"
}

func (p *GlobalPlugin) DisplayName() string {
	return "Global post-create populater"
}

func (p *GlobalPlugin) NewWizard(projectPath, projectType string) tea.Model {
	return NewModel(projectPath, projectType)
}

// ---------- Wizard model ----------

type step int

const (
	stepGlobal step = iota
	stepTypeSpecific
	stepDone
)

type item struct {
	ID       string
	Label    string
	Selected bool
}

type Model struct {
	step step

	projectPath string
	projectType string

	cursor       int
	globalItems  []item
	typeItems    []item
	errMsg       string
	applySummary []string
}

func NewModel(projectPath, projectType string) Model {
	globalItems := []item{
		{ID: "global_env", Label: "Create .env file", Selected: false},
		{ID: "global_notes", Label: "Create notes/ folder", Selected: false},
		{ID: "global_readme", Label: "Create README.md file", Selected: false},
		{ID: "global_gitignore", Label: "Create .gitignore file", Selected: false},
		{ID: "global_makefile", Label: "Create Makefile", Selected: false},
	}

	var typeItems []item

	switch projectType {
	case "go":
		typeItems = []item{
			{ID: "go_cmd", Label: "Create cmd/ folder", Selected: true},
			{ID: "go_internal", Label: "Create internal/ folder", Selected: true},
			{ID: "go_pkg", Label: "Create pkg/ folder", Selected: true},
			{ID: "go_tests", Label: "Create tests/ folder", Selected: false},
			{ID: "go_gen", Label: "Create gen/ folder", Selected: false},
			{ID: "go_api", Label: "Create api/ folder", Selected: false},
		}
	default:
		typeItems = nil
	}

	return Model{
		step:        stepGlobal,
		projectPath: projectPath,
		projectType: projectType,
		cursor:      0,
		globalItems: globalItems,
		typeItems:   typeItems,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {

		case stepGlobal:
			switch msg.String() {
			case "up", "k":
				if len(m.globalItems) == 0 {
					return m, nil
				}
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.globalItems) - 1
				}
				return m, nil

			case "down", "j":
				if len(m.globalItems) == 0 {
					return m, nil
				}
				m.cursor++
				if m.cursor >= len(m.globalItems) {
					m.cursor = 0
				}
				return m, nil

			case " ":
				if len(m.globalItems) == 0 {
					return m, nil
				}
				m.globalItems[m.cursor].Selected = !m.globalItems[m.cursor].Selected
				return m, nil

			case "enter":
				m.step = stepTypeSpecific
				m.cursor = 0
				return m, nil

			case "esc":
				m.step = stepDone
				m.applySummary = []string{"Skipped population."}
				return m, nil

			case "ctrl+c":
				return m, tea.Quit
			}

		case stepTypeSpecific:
			switch msg.String() {
			case "up", "k":
				if len(m.typeItems) == 0 {
					return m, nil
				}
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.typeItems) - 1
				}
				return m, nil

			case "down", "j":
				if len(m.typeItems) == 0 {
					return m, nil
				}
				m.cursor++
				if m.cursor >= len(m.typeItems) {
					m.cursor = 0
				}
				return m, nil

			case " ":
				if len(m.typeItems) == 0 {
					return m, nil
				}
				m.typeItems[m.cursor].Selected = !m.typeItems[m.cursor].Selected
				return m, nil

			case "enter":
				summary, err := m.applySelections()
				m.applySummary = summary
				if err != nil {
					m.errMsg = err.Error()
				}
				m.step = stepDone
				return m, nil

			case "esc":
				m.step = stepGlobal
				m.cursor = 0
				return m, nil

			case "ctrl+c":
				return m, tea.Quit
			}

		case stepDone:
			// any key exits
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	switch m.step {
	case stepGlobal:
		return m.viewGlobal()
	case stepTypeSpecific:
		return m.viewTypeSpecific()
	case stepDone:
		return m.viewDone()
	}
	return ""
}

func (m Model) viewGlobal() string {
	var b strings.Builder

	b.WriteString("Post-create – Global options\n\n")
	b.WriteString("Project: " + m.projectPath + "\n\n")
	b.WriteString("Select global items to add (space to toggle, enter to continue):\n\n")

	if len(m.globalItems) == 0 {
		b.WriteString("  (no global options yet)\n")
	} else {
		for i, it := range m.globalItems {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			check := " "
			if it.Selected {
				check = "x"
			}
			b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, check, it.Label))
		}
	}

	if m.errMsg != "" {
		b.WriteString("\nError: " + m.errMsg + "\n")
	}

	b.WriteString("\n[↑/↓] Move  [space] Toggle  [enter] Next  [esc] Skip  [ctrl+c] Quit\n")

	return b.String()
}

func (m Model) viewTypeSpecific() string {
	var b strings.Builder

	b.WriteString("Post-create – Type-specific options\n\n")
	b.WriteString("Project: " + m.projectPath + "\n")
	b.WriteString("Type: " + m.projectType + "\n\n")
	b.WriteString("Select type-specific items to add (space to toggle, enter to apply):\n\n")

	if len(m.typeItems) == 0 {
		b.WriteString("  (no type-specific options for this project type yet)\n")
	} else {
		for i, it := range m.typeItems {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			check := " "
			if it.Selected {
				check = "x"
			}
			b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, check, it.Label))
		}
	}

	if m.errMsg != "" {
		b.WriteString("\nError: " + m.errMsg + "\n")
	}

	b.WriteString("\n[↑/↓] Move  [space] Toggle  [enter] Apply  [esc] Back  [ctrl+c] Quit\n")

	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder

	b.WriteString("Post-create – Result\n\n")
	b.WriteString("Project: " + m.projectPath + "\n\n")

	if len(m.applySummary) == 0 {
		b.WriteString("No changes were applied.\n")
	} else {
		for _, line := range m.applySummary {
			b.WriteString("- " + line + "\n")
		}
	}

	if m.errMsg != "" {
		b.WriteString("\nError: " + m.errMsg + "\n")
	}

	b.WriteString("\n[any key] Exit\n")

	return b.String()
}

func (m *Model) applySelections() ([]string, error) {
	var summary []string

	// Global items
	for _, it := range m.globalItems {
		if !it.Selected {
			continue
		}

		switch it.ID {
		case "global_env":
			envPath := filepath.Join(m.projectPath, ".env")
			if _, err := os.Stat(envPath); err == nil {
				summary = append(summary, ".env already exists (skipped)")
				continue
			}
			if err := os.WriteFile(envPath, []byte(""), 0o644); err != nil {
				return summary, fmt.Errorf("failed to create .env: %w", err)
			}
			summary = append(summary, "Created .env file")

		case "global_notes":
			notesPath := filepath.Join(m.projectPath, "notes")
			if err := os.MkdirAll(notesPath, 0o755); err != nil {
				return summary, fmt.Errorf("failed to create notes/ folder: %w", err)
			}
			summary = append(summary, "Created notes/ folder")

		case "global_readme":
			readmePath := filepath.Join(m.projectPath, "README.md")
			if _, err := os.Stat(readmePath); err == nil {
				summary = append(summary, "README.md already exists (skipped)")
				continue
			}
			if err := os.WriteFile(readmePath, []byte("# Project Title\n\nProject description.\n"), 0o644); err != nil {
				return summary, fmt.Errorf("failed to create README.md: %w", err)
			}
			summary = append(summary, "Created README.md file")

		case "global_gitignore":
			gitignorePath := filepath.Join(m.projectPath, ".gitignore")
			if _, err := os.Stat(gitignorePath); err == nil {
				summary = append(summary, ".gitignore already exists (skipped)")
				continue
			}
			if err := os.WriteFile(gitignorePath, []byte("# Ignore files\n.env\nnotes/\n"), 0o644); err != nil {
				return summary, fmt.Errorf("failed to create .gitignore: %w", err)
			}
			summary = append(summary, "Created .gitignore file")

		case "global_makefile":
			makefilePath := filepath.Join(m.projectPath, "Makefile")
			if _, err := os.Stat(makefilePath); err == nil {
				summary = append(summary, "Makefile already exists (skipped)")
				continue
			}
			if err := os.WriteFile(makefilePath, []byte("all:\n\t@echo \"Build commands go here\"\n"), 0o644); err != nil {
				return summary, fmt.Errorf("failed to create Makefile: %w", err)
			}
			summary = append(summary, "Created Makefile")
		}
	}

	// Type-specific items
	for _, it := range m.typeItems {
		if !it.Selected {
			continue
		}

		switch m.projectType {
		case "go":
			switch it.ID {
			case "go_cmd":
				if err := os.MkdirAll(filepath.Join(m.projectPath, "cmd"), 0o755); err != nil {
					return summary, fmt.Errorf("failed to create cmd/: %w", err)
				}
				summary = append(summary, "Created cmd/ folder")

			case "go_internal":
				if err := os.MkdirAll(filepath.Join(m.projectPath, "internal"), 0o755); err != nil {
					return summary, fmt.Errorf("failed to create internal/: %w", err)
				}
				summary = append(summary, "Created internal/ folder")

			case "go_pkg":
				if err := os.MkdirAll(filepath.Join(m.projectPath, "pkg"), 0o755); err != nil {
					return summary, fmt.Errorf("failed to create pkg/: %w", err)
				}
				summary = append(summary, "Created pkg/ folder")

			case "go_tests":
				if err := os.MkdirAll(filepath.Join(m.projectPath, "tests"), 0o755); err != nil {
					return summary, fmt.Errorf("failed to create tests/: %w", err)
				}
				summary = append(summary, "Created tests/ folder")
			case "go_gen":
				if err := os.MkdirAll(filepath.Join(m.projectPath, "gen"), 0o755); err != nil {
					return summary, fmt.Errorf("failed to create gen/: %w", err)
				}
				summary = append(summary, "Created gen/ folder")
			case "go_api":
				if err := os.MkdirAll(filepath.Join(m.projectPath, "api"), 0o755); err != nil {
					return summary, fmt.Errorf("failed to create api/: %w", err)
				}
				summary = append(summary, "Created api/ folder")
			}
		}
	}

	if len(summary) == 0 {
		summary = []string{"No items were selected."}
	}

	return summary, nil
}
