package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/protibimbok/mgit/internal/config"
)

var httpsPattern = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)

// resolveRemoteURL turns a GitHub HTTPS URL or a <key>:<user>/<repo> shorthand
// into a git@hub.<key>:<user>/<repo> SSH URL. The returned profile is nil when
// the URL was left unchanged.
func resolveRemoteURL(raw string, cfg *config.Config, promptLabel string) (string, *config.Profile, error) {
	switch {
	case strings.HasPrefix(raw, "https://github.com/") || strings.HasPrefix(raw, "http://github.com/"):
		path := strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://")
		path = strings.TrimPrefix(path, "github.com/")
		path = strings.TrimSuffix(path, ".git")

		p, err := chooseProfile(cfg, promptLabel)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("git@hub.%s:%s", p.Key, path), p, nil

	case strings.Contains(raw, ":") && !strings.Contains(raw, "://") && !strings.HasPrefix(raw, "git@"):
		parts := strings.SplitN(raw, ":", 2)
		key, path := parts[0], parts[1]
		p := cfg.FindByKey(key)
		if p == nil {
			return "", nil, fmt.Errorf("unknown profile key %q — run 'mgit list' to see available profiles", key)
		}
		path = strings.TrimSuffix(path, ".git")
		return fmt.Sprintf("git@hub.%s:%s", key, path), p, nil

	default:
		return raw, nil, nil
	}
}

func profileLabels(profiles []config.Profile) []string {
	labels := make([]string, len(profiles))
	for i, p := range profiles {
		labels[i] = p.Label
	}
	return labels
}
