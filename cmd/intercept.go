package cmd

import (
	"strings"

	"github.com/protibimbok/mgit/internal/config"
)

// TryInterceptGitArgs inspects git passthrough arguments and may rewrite them.
func TryInterceptGitArgs(args []string) ([]string, bool, error) {
	if len(args) < 4 || args[0] != "remote" || args[1] != "add" {
		return args, false, nil
	}

	urlIdx := len(args) - 1
	rawURL := args[urlIdx]
	if !isTransformableRemoteURL(rawURL) {
		return args, false, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, false, err
	}

	sshURL, transformed, err := resolveRemoteURL(rawURL, cfg, "Select profile")
	if err != nil {
		return nil, false, err
	}
	if !transformed {
		return args, false, nil
	}

	newArgs := make([]string, len(args))
	copy(newArgs, args)
	newArgs[urlIdx] = sshURL
	return newArgs, true, nil
}

func isTransformableRemoteURL(url string) bool {
	return strings.HasPrefix(url, "https://github.com/") ||
		strings.HasPrefix(url, "http://github.com/") ||
		(strings.Contains(url, ":") && !strings.Contains(url, "://") && !strings.HasPrefix(url, "git@"))
}
