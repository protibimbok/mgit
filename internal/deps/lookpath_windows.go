//go:build windows

package deps

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func lookPath(name string) (string, error) {
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	return findWindowsExecutable(name)
}

func findWindowsExecutable(name string) (string, error) {
	if !strings.Contains(name, ".") {
		for _, ext := range windowsExts() {
			if p, err := searchWindowsPaths(name + ext); err == nil {
				return p, nil
			}
		}
	}
	return searchWindowsPaths(name)
}

func searchWindowsPaths(file string) (string, error) {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if p, err := statExecutable(filepath.Join(dir, file)); err == nil {
			return p, nil
		}
		if alt := rewriteSystem32Dir(dir); alt != dir {
			if p, err := statExecutable(filepath.Join(alt, file)); err == nil {
				return p, nil
			}
		}
	}
	for _, dir := range wellKnownDirs(file) {
		if p, err := statExecutable(filepath.Join(dir, file)); err == nil {
			return p, nil
		}
	}
	return "", exec.ErrNotFound
}

func rewriteSystem32Dir(dir string) string {
	if runtime.GOARCH != "386" {
		return dir
	}
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = `C:\Windows`
	}
	sys32 := filepath.Join(systemRoot, "System32")
	clean := filepath.Clean(dir)
	rel, err := filepath.Rel(sys32, clean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return dir
	}
	return filepath.Join(systemRoot, "Sysnative", rel)
}

func wellKnownDirs(file string) []string {
	var dirs []string
	if base := strings.ToLower(filepath.Base(file)); base == "ssh-keygen" || base == "ssh-keygen.exe" {
		if programFiles := os.Getenv("ProgramFiles"); programFiles != "" {
			dirs = append(dirs, filepath.Join(programFiles, "Git", "usr", "bin"))
		}
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		if runtime.GOARCH == "386" {
			dirs = append(dirs, filepath.Join(systemRoot, "Sysnative", "OpenSSH"))
		} else {
			dirs = append(dirs, filepath.Join(systemRoot, "System32", "OpenSSH"))
		}
	}
	return dirs
}

func windowsExts() []string {
	pathext := os.Getenv("PATHEXT")
	if pathext == "" {
		return []string{".exe", ".com", ".bat", ".cmd"}
	}
	var exts []string
	for _, ext := range strings.Split(pathext, ";") {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		exts = append(exts, strings.ToLower(ext))
	}
	return exts
}

func statExecutable(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return "", os.ErrNotExist
	}
	return path, nil
}
