//go:build windows

package sshutil

func generateKeyAt(keyPath, email string) error {
	return generateEd25519Key(keyPath, email)
}
