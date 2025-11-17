# pcli – Project Creation CLI

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?style=flat&logo=linux&logoColor=black)
![Bash Script](https://img.shields.io/badge/bash_script-%23121011.svg?style=flat&logo=gnu-bash&logoColor=white)

A fast, extensible, interactive CLI for creating and bootstrapping development projects.

---

## ⭐ What is pcli?

`pcli` is an interactive **TUI-based project generator** built with **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**, designed to:

- Create projects for multiple languages (Go supported now, more coming)
- Detect or install required language runtimes (Linux: apt, dnf, pacman)
- Scaffold post-create resources (global files & language‑specific folders)
- Use a **fully plugin‑driven architecture** for maximum extensibility

It is meant to be your central command for starting new projects *the same way every time*, but without the constraints of template systems.

---

## 🚀 Features

### ✔️ Interactive Bubble Tea Wizard

- Arrow‑key navigation  
- Space to toggle  
- Real‑time logs during installation  
- Output‑driven progress bar  
- Clean UX and clear steps  

### ✔️ Project Creation Plugins

Each project type is implemented as a plugin under:

```bash
internal/projecttype/<type>/
```

Example: **Go plugin**

- Detects if Go is installed
- If missing: prompts the user and installs Go
- Creates the project folder using actual Go commands:

```bash
go mod init <module>
go mod tidy
```

### ✔️ Post‑Create Plugins

Run immediately after the project is created.

Two kinds:

1. **Global additions**
   - `.env`
   - `notes/`
2. **Type‑specific scaffolding**
   - Go: `cmd/`, `internal/`, `pkg/`

Plugins live under:

```bash
internal/postplugin/
```

### ✔️ Automatic Linux Language Installation

Supported package managers:

- `apt` (Debian / Ubuntu)
- `dnf` (Fedora / RHEL)
- `pacman` (Arch / Manjaro)

Logs are streamed in real‑time into the UI.

---

## 📁 Project Structure

```bash
pcli/
├── cmd/pcli/
│   └── main.go                # Entrypoint
│
├── internal/
│   ├── plugins/               # Project creation plugin loader
│   ├── projecttype/
│   │   └── go/                # Go project creator plugin
│   │       └── plugin.go
│   │
│   ├── postplugin/            # Post-create plugin system
│   │   ├── plugin.go
│   │   ├── register_all.go
│   │   └── global/
│   │       └── global.go      # Global + type-specific folder creator
│   │
│   ├── langenv/               # Language installation checker
│   │   └── langenv.go
│   │
│   └── ui/                    # Root UI screens (type chooser)
│
└── README.md
```

---

## ⚙️ Configuration (`.env`)

Defaults for project paths and module names:

```env
DEFAULT_GO_PROJECT_MODULE_PATH=github.com/$USER/myservice
DEFAULT_GO_PROJECT_PATH=$HOME/Documents/projects
```

Environment variables like `$HOME`, `$USER` and `~/…` are automatically expanded.

---

## 🧪 Usage

### Run the CLI

```bash
go run ./cmd/pcli
```

This is destined for development use. For production, see the install section below.

### Install the CLI

```bash
sudo make install
```

This will build and copy the `pcli` binary to the path present in the [Makefile](Makefile) $GOBIN variable (default: `/usr/local/go/bin`).

### Steps

1. Choose the project type  
2. Enter module path (for Go)  
3. Confirm summary  
4. If Go missing → decide whether to install  
5. After project creation → post-create plugin runs  
6. Choose global and language-specific files/folders  
7. pcli applies everything and exits  

---

## 🧱 Architecture Overview

```bash
    ┌───────────────────────┐
    │  Type Chooser (TUI)   │
    └───────────┬───────────┘
                │
                ▼
     ┌──────────────────────┐
     │ ProjectType Plugin   │  (Go, TS, Terraform…)
     └──────────┬───────────┘
                │ creates project
                ▼
     ┌──────────────────────┐
     │  PostCreate Plugins  │  (global + type-specific)
     └──────────┬───────────┘
                │ scaffolds files
                ▼
       Project ready to use!
```

---

## 🛣️ Roadmap

- [ ] Node/TypeScript plugin  
- [ ] Terraform plugin  
- [ ] Git initializer plugin  
- [ ] CI/CD plugin (GitHub Actions)  
- [ ] “Language Manager” tool (install runtimes anytime)  
- [ ] Plugin metadata system  
- [ ] Automatic project templates for frameworks  

---

## 📜 License

[MIT License](LICENSE)

---

## ❤️ Contributing

Contributions welcome!  
The plugin architecture is intentionally simple so you can easily:

- Add a new language
- Add new scaffolding actions
- Add new post-create plugins
- Add new installation routines

Just open an issue or PR.
