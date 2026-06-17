package deps

import (
	"fmt"
	"runtime"
	"strings"
)

func RequireGit() error {
	if _, err := lookPath("git"); err == nil {
		return nil
	}
	return fmt.Errorf("%s", gitMissingMessage())
}

func RequireSSHKeygen() error {
	_, err := SSHKeygenPath()
	return err
}

// SSHKeygenPath resolves ssh-keygen on PATH, including Windows fallbacks when
// a 32-bit process cannot see System32 via WOW64 filesystem redirection.
func SSHKeygenPath() (string, error) {
	p, err := lookPath("ssh-keygen")
	if err != nil {
		return "", fmt.Errorf("%s", sshKeygenMissingMessage())
	}
	return p, nil
}

func GitMissingMessage() string {
	return gitMissingMessage()
}

func SSHKeygenMissingMessage() string {
	return sshKeygenMissingMessage()
}

func gitMissingMessage() string {
	var b strings.Builder
	b.WriteString("mgit: git is not installed or not on your PATH.\n\n")
	b.WriteString("  mgit wraps git — install Git first, then restart your terminal.\n\n")
	switch runtime.GOOS {
	case "windows":
		b.WriteString("  Windows:\n")
		b.WriteString("    winget install Git.Git\n")
		b.WriteString("    — or — https://git-scm.com/download/win\n")
		b.WriteString("    Choose \"Git from the command line and also from 3rd-party software\" during setup.\n")
	case "darwin":
		b.WriteString("  macOS:\n")
		b.WriteString("    brew install git\n")
		b.WriteString("    — or — install Xcode Command Line Tools: xcode-select --install\n")
	default:
		b.WriteString("  Linux:\n")
		b.WriteString("    Use your package manager, e.g. apt install git / dnf install git\n")
	}
	return b.String()
}

func sshKeygenMissingMessage() string {
	var b strings.Builder
	b.WriteString("mgit: ssh-keygen is not installed or not on your PATH.\n\n")
	b.WriteString("  mgit gen creates SSH keys using ssh-keygen.\n\n")
	switch runtime.GOOS {
	case "windows":
		b.WriteString("  Windows:\n")
		b.WriteString("    Install OpenSSH Client and restart your terminal.\n")
	default:
		b.WriteString("  macOS / Linux:\n")
		b.WriteString("    OpenSSH client is usually pre-installed. If missing: brew install openssh / apt install openssh-client\n")
	}
	return b.String()
}
