package sshutil

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

func generateEd25519Key(keyPath, comment string) error {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ed25519 key: %w", err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(priv, comment)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		return fmt.Errorf("write private key: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(priv.Public())
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}
	pubLine := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))
	if comment != "" {
		pubLine += " " + comment
	}
	if err := os.WriteFile(keyPath+".pub", []byte(pubLine+"\n"), 0644); err != nil {
		return fmt.Errorf("write public key: %w", err)
	}
	return nil
}
