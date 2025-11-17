package langenv

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Language string

const (
	LanguageGo Language = "go"
	// later: LanguageNode, LanguageTerraform, ...
)

func IsInstalled(lang Language) bool {
	switch lang {
	case LanguageGo:
		_, err := exec.LookPath("go")
		return err == nil
	default:
		return false
	}
}

// InstallCommand returns an *exec.Cmd that attempts to install the language.
//
// Linux-only:
//   - apt-get (Debian/Ubuntu)
//   - dnf (Fedora/RHEL/...)
//   - pacman (Arch/Manjaro/...)
//
// Other OS: returns an error (no automatic install).
func InstallCommand(lang Language) (*exec.Cmd, error) {
	switch lang {
	case LanguageGo:
		return installGoCommand()
	default:
		return nil, fmt.Errorf("no installer defined for language: %s", lang)
	}
}

// RunWithOutput launches the command, reads stdout and stderr line-by-line,
// and pushes each line into the provided channel.
//
// Caller is responsible for closing the channel AFTER this returns
// (or in a wrapper).
func RunWithOutput(cmd *exec.Cmd, ch chan<- string) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read stderr in a goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
	}()

	// Read stdout in current goroutine
	stdoutScanner := bufio.NewScanner(stdout)
	for stdoutScanner.Scan() {
		ch <- stdoutScanner.Text()
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// ExpandPathEnv expands $HOME, $USER, etc. in path-like values,
// and also handles "~/something".
func ExpandPathEnv(s string) string {
	s = os.ExpandEnv(s)

	if strings.HasPrefix(s, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, strings.TrimPrefix(s, "~/"))
		}
	}

	return s
}

// ------------ Linux-only installer helpers ------------

func installGoCommand() (*exec.Cmd, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("automatic Go installation is only supported on Linux; please install Go manually")
	}

	pm := detectLinuxPackageManager()
	switch pm {
	case "apt":
		// Debian / Ubuntu
		return exec.Command("sh", "-c", "sudo apt-get update && sudo apt-get install -y golang-go"), nil
	case "dnf":
		// Fedora / RHEL / CentOS
		return exec.Command("sh", "-c", "sudo dnf install -y golang"), nil
	case "pacman":
		// Arch / Manjaro
		return exec.Command("sh", "-c", "sudo pacman -Sy --noconfirm go"), nil
	default:
		return nil, fmt.Errorf("unsupported Linux distro for automatic Go install; please install Go manually")
	}
}

func detectLinuxPackageManager() string {
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		id := parseOsReleaseID(string(data))
		switch id {
		case "ubuntu", "debian", "linuxmint", "pop":
			return "apt"
		case "fedora", "rhel", "centos", "rocky", "almalinux":
			return "dnf"
		case "arch", "manjaro", "endeavouros":
			return "pacman"
		}
	}

	// Fallback: check common binaries
	if existsInPath("apt-get") {
		return "apt"
	}
	if existsInPath("dnf") {
		return "dnf"
	}
	if existsInPath("pacman") {
		return "pacman"
	}

	return ""
}

func parseOsReleaseID(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			val := strings.TrimPrefix(line, "ID=")
			val = strings.Trim(val, `"`)
			val = strings.TrimSpace(val)
			return val
		}
	}
	return ""
}

func existsInPath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}
