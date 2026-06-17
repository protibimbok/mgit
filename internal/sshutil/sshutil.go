package sshutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/protibimbok/mgit/internal/deps"
)

func KeyPath(key string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "mgit_"+key), nil
}

func GenerateKey(key, email string) (string, error) {
	sshKeygen, err := deps.SSHKeygenPath()
	if err != nil {
		return "", err
	}

	keyPath, err := KeyPath(key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return "", err
	}
	if _, err := os.Stat(keyPath); err == nil {
		return "", fmt.Errorf("SSH key %s already exists", keyPath)
	}

	c := exec.Command(sshKeygen, "-t", "ed25519", "-C", email, "-f", keyPath, "-N", "")
	c.Env = os.Environ()
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("ssh-keygen failed: %w", err)
	}
	return keyPath, nil
}

func AddToSSHConfig(key, keyPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".ssh", "config")

	// Don't add a duplicate block
	if data, err := os.ReadFile(configPath); err == nil {
		if strings.Contains(string(data), "Host hub."+key+"\n") ||
			strings.Contains(string(data), "Host hub."+key+"\r\n") {
			return nil
		}
	}

	identity := sshIdentityFile(keyPath)
	entry := fmt.Sprintf("\nHost hub.%s\n  HostName github.com\n  User git\n  IdentityFile %s\n  IdentitiesOnly yes\n", key, identity)
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(entry)
	return err
}

func RemoveFromSSHConfig(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".ssh", "config")

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	target := "Host hub." + key
	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	var out []string
	skip := false

	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == target {
			skip = true
			for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
				out = out[:len(out)-1]
			}
			continue
		}
		if skip {
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				skip = false
			} else {
				continue
			}
		}
		out = append(out, line)
	}

	return os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0600)
}

// FormatSSHConfigEntry builds an SSH config host block for testing.
func FormatSSHConfigEntry(key, keyPath string) string {
	identity := sshIdentityFile(keyPath)
	return fmt.Sprintf("\nHost hub.%s\n  HostName github.com\n  User git\n  IdentityFile %s\n  IdentitiesOnly yes\n", key, identity)
}

func sshIdentityFile(path string) string {
	return strings.ReplaceAll(filepath.ToSlash(path), "\\", "/")
}
