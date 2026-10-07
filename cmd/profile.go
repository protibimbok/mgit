package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/prompt"
)

const profileConfigKey = "mgit.profile"

var hubRemotePattern = regexp.MustCompile(`^git@hub\.([^:]+):`)

func chooseProfile(cfg *config.Config, label string) (*config.Profile, error) {
	if len(cfg.Profiles) == 0 {
		return nil, fmt.Errorf("no profiles found — run 'mgit new' to create one")
	}

	if p, how := detectRepoProfile(cfg); p != nil {
		fmt.Printf("Using profile %q (%s)\n", p.Key, how)
		return p, nil
	}

	if len(cfg.Profiles) == 1 {
		p := &cfg.Profiles[0]
		fmt.Printf("Using profile %q (only profile configured)\n", p.Key)
		return p, nil
	}

	idx, err := prompt.Select(label, profileLabels(cfg.Profiles))
	if err != nil {
		return nil, err
	}
	p := &cfg.Profiles[idx]
	if insideGitRepo("") {
		rememberProfile("", p.Key)
	}
	return p, nil
}

func detectRepoProfile(cfg *config.Config) (*config.Profile, string) {
	if !insideGitRepo("") {
		return nil, ""
	}

	if key := gitConfigLocal("", profileConfigKey); key != "" {
		if p := cfg.FindByKey(key); p != nil {
			return p, "configured for this repo"
		}
	}

	for _, url := range gitRemoteURLs() {
		if key := profileKeyFromRemoteURL(url); key != "" {
			if p := cfg.FindByKey(key); p != nil {
				return p, "matches existing remote"
			}
		}
	}

	if email := gitConfigLocal("", "user.email"); email != "" {
		var match *config.Profile
		for i := range cfg.Profiles {
			if strings.EqualFold(cfg.Profiles[i].Email, email) {
				if match != nil {
					match = nil // ambiguous: several profiles share this email
					break
				}
				match = &cfg.Profiles[i]
			}
		}
		if match != nil {
			return match, "matches repo user.email"
		}
	}

	return nil, ""
}

// profileKeyFromRemoteURL extracts <key> from a git@hub.<key>:user/repo URL.
func profileKeyFromRemoteURL(url string) string {
	m := hubRemotePattern.FindStringSubmatch(strings.TrimSpace(url))
	if m == nil {
		return ""
	}
	return m[1]
}

func rememberProfile(dir, key string) {
	_ = gitIn(dir, "config", "--local", profileConfigKey, key).Run()
}

// applyProfileToRepo writes user.name, user.email and mgit.profile into the
// repository at dir ("" for the current directory).
func applyProfileToRepo(dir string, p *config.Profile) error {
	for _, kv := range [][2]string{
		{"user.name", p.Name},
		{"user.email", p.Email},
		{profileConfigKey, p.Key},
	} {
		if err := gitIn(dir, "config", "--local", kv[0], kv[1]).Run(); err != nil {
			return err
		}
	}
	return nil
}

func insideGitRepo(dir string) bool {
	out, err := gitIn(dir, "rev-parse", "--is-inside-work-tree").Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func gitConfigLocal(dir, key string) string {
	out, err := gitIn(dir, "config", "--local", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitRemoteURLs() []string {
	out, err := gitIn("", "config", "--local", "--get-regexp", `^remote\..*\.url$`).Output()
	if err != nil {
		return nil
	}
	var urls []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 {
			urls = append(urls, fields[1])
		}
	}
	return urls
}

func gitIn(dir string, args ...string) *exec.Cmd {
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	return exec.Command("git", args...)
}
