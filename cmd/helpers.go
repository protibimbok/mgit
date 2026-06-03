package cmd

import (
	"fmt"

	"github.com/protibimbok/mgit/internal/config"
)

func profileLabels(profiles []config.Profile) []string {
	labels := make([]string, len(profiles))
	for i, p := range profiles {
		labels[i] = fmt.Sprintf("%s [%s] <%s>", p.Label, p.Key, p.Email)
	}
	return labels
}
