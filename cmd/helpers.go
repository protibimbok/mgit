package cmd

import "github.com/protibimbok/mgit/internal/config"

func profileLabels(profiles []config.Profile) []string {
	labels := make([]string, len(profiles))
	for i, p := range profiles {
		labels[i] = p.Label
	}
	return labels
}
