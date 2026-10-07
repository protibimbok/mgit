//go:build !windows

package sshutil

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/protibimbok/mgit/internal/deps"
)

func generateKeyAt(keyPath, email string) error {
	if err := deps.RequireSSHKeygen(); err != nil {
		return err
	}

	c := exec.Command("ssh-keygen", "-t", "ed25519", "-C", email, "-f", keyPath, "-N", "")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("ssh-keygen failed: %w", err)
	}
	return nil
}
