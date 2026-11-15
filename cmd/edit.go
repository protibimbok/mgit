package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/protibimbok/mgit/internal/config"
	"github.com/protibimbok/mgit/internal/prompt"
	"github.com/protibimbok/mgit/internal/sshutil"
)

var editCmd = &cobra.Command{
	Use:   "edit [key|email]",
	Short: "Edit a profile's name, email, label or key",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runEdit,
}

func runEdit(_ *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles found — run 'mgit new' to create one")
	}

	var profile *config.Profile
	if len(args) == 1 {
		profile, err = findProfile(cfg, args[0])
		if err != nil {
			return err
		}
	} else {
		idx, err := prompt.Select("Select profile to edit", profileLabels(cfg.Profiles))
		if err != nil {
			return err
		}
		profile = &cfg.Profiles[idx]
	}

	fmt.Printf("Editing profile %q (press Enter to keep the current value)\n", profile.Key)
	name, err := prompt.Ask("Full name", profile.Name)
	if err != nil {
		return err
	}
	email, err := prompt.Ask("Email", profile.Email)
	if err != nil {
		return err
	}
	label, err := prompt.Ask("Label", profile.Label)
	if err != nil {
		return err
	}
	newKey, err := prompt.Ask("Key", profile.Key)
	if err != nil {
		return err
	}

	oldKey := profile.Key
	keyChanged := newKey != oldKey
	if name == profile.Name && email == profile.Email && label == profile.Label && !keyChanged {
		fmt.Println("No changes.")
		return nil
	}

	// Decide before touching anything whether this repo uses the profile.
	repoUsesProfile := insideGitRepo("") && repoUsesKey(oldKey)

	if keyChanged {
		if err := validateNewKey(cfg, newKey); err != nil {
			return err
		}
		fmt.Printf("\nWarning: renaming %q to %q removes the SSH alias hub.%s.\n", oldKey, newKey, oldKey)
		fmt.Printf("Existing repos with git@hub.%s: remotes will fail to push/pull/fetch until you\n", oldKey)
		fmt.Printf("run 'mgit init %s' in each of them (it rewrites the remotes).\n", newKey)
		if repoUsesProfile {
			fmt.Println("The current repo will be updated automatically.")
		}
		ok, err := prompt.Confirm("Continue?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
		if err := renameProfileKey(profile, newKey); err != nil {
			return err
		}
	}

	profile.Name, profile.Email, profile.Label = name, email, label
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("Profile %q updated.\n", profile.Key)

	// Keep the current repo in sync if it uses this profile.
	if repoUsesProfile {
		if keyChanged {
			rewriteRepoRemotes(oldKey, newKey)
		}
		if err := applyProfileToRepo("", profile); err != nil {
			fmt.Printf("warning: could not update this repo's git config: %v\n", err)
		} else {
			fmt.Println("Updated user.name / user.email in this repo.")
		}
	}
	return nil
}

// findProfile looks a profile up by key, falling back to a unique email match.
func findProfile(cfg *config.Config, keyOrEmail string) (*config.Profile, error) {
	if p := cfg.FindByKey(keyOrEmail); p != nil {
		return p, nil
	}
	var match *config.Profile
	for i := range cfg.Profiles {
		if strings.EqualFold(cfg.Profiles[i].Email, keyOrEmail) {
			if match != nil {
				return nil, fmt.Errorf("several profiles use email %q — specify the profile key instead", keyOrEmail)
			}
			match = &cfg.Profiles[i]
		}
	}
	if match == nil {
		return nil, fmt.Errorf("profile %q not found — run 'mgit list' to see available profiles", keyOrEmail)
	}
	return match, nil
}

func validateNewKey(cfg *config.Config, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if strings.ContainsAny(key, " \t:/\\@") {
		return fmt.Errorf("key %q must not contain spaces, ':', '/', '\\' or '@'", key)
	}
	if cfg.FindByKey(key) != nil {
		return fmt.Errorf("profile %q already exists", key)
	}
	return nil
}

// renameProfileKey moves the SSH key pair to mgit_<newKey> (when it lives at
// the default location) and replaces the hub.<oldKey> SSH config block.
func renameProfileKey(p *config.Profile, newKey string) error {
	oldKey := p.Key
	keyPath := p.SSHKey

	oldDefault, err := sshutil.KeyPath(oldKey)
	if err != nil {
		return err
	}
	if p.SSHKey == oldDefault {
		newPath, err := sshutil.KeyPath(newKey)
		if err != nil {
			return err
		}
		for _, path := range []string{newPath, newPath + ".pub"} {
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("SSH key %s already exists", path)
			}
		}
		if err := os.Rename(oldDefault, newPath); err != nil {
			return fmt.Errorf("failed to rename SSH key: %w", err)
		}
		if err := os.Rename(oldDefault+".pub", newPath+".pub"); err != nil && !os.IsNotExist(err) {
			_ = os.Rename(newPath, oldDefault) // roll back the private key
			return fmt.Errorf("failed to rename SSH public key: %w", err)
		}
		keyPath = newPath
	}

	if err := sshutil.RemoveFromSSHConfig(oldKey); err != nil {
		fmt.Printf("warning: could not remove hub.%s from SSH config: %v\n", oldKey, err)
	}
	if err := sshutil.AddToSSHConfig(newKey, keyPath); err != nil {
		return fmt.Errorf("failed to update SSH config: %w", err)
	}

	p.Key = newKey
	p.SSHKey = keyPath
	return nil
}

// repoUsesKey reports whether the current repo is tied to the profile key,
// either via mgit.profile or a git@hub.<key>: remote.
func repoUsesKey(key string) bool {
	if gitConfigLocal("", profileConfigKey) == key {
		return true
	}
	for _, url := range gitRemoteURLs() {
		if profileKeyFromRemoteURL(url) == key {
			return true
		}
	}
	return false
}

// rewriteRepoRemotes points git@hub.<oldKey>: remotes in the current repo at
// hub.<newKey>.
func rewriteRepoRemotes(oldKey, newKey string) {
	out, err := gitIn("", "config", "--local", "--get-regexp", `^remote\..*\.url$`).Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || profileKeyFromRemoteURL(fields[1]) != oldKey {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(fields[0], "remote."), ".url")
		newURL := "git@hub." + newKey + ":" + strings.SplitN(fields[1], ":", 2)[1]
		if err := gitIn("", "remote", "set-url", name, newURL).Run(); err != nil {
			fmt.Printf("warning: could not update remote %q: %v\n", name, err)
			continue
		}
		fmt.Printf("Remote %q → %s\n", name, newURL)
	}
}
