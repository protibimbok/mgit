# mgit

Manage multiple GitHub SSH profiles on a single machine. `mgit` wraps `git` — all unknown subcommands are forwarded transparently to git.

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap protibimbok/pkg-dist
brew install mgit
```

### apt (Debian / Ubuntu)

```bash
# One-time: install the signing key
curl -fsSL \
  https://github.com/protibimbok/pkg-dist/raw/master/public.gpg \
  | sudo gpg --dearmor \
  -o /usr/share/keyrings/protibimbok.gpg

# One-time: add the repository
echo "deb [signed-by=/usr/share/keyrings/protibimbok.gpg] \
  https://protibimbok.github.io/pkg-dist/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/protibimbok.list

sudo apt update
sudo apt install mgit
```

### pacman / AUR (Arch Linux)

```bash
yay -S mgit-bin
# or
paru -S mgit-bin
```

### rpm (Fedora / RHEL / openSUSE)

Download the `.rpm` from the [latest release](https://github.com/protibimbok/mgit/releases/latest):

```bash
sudo rpm -i mgit_linux_amd64.rpm
```

### Alpine Linux

```bash
# Download the .apk from the latest release, then:
sudo apk add --allow-untrusted mgit_linux_amd64.apk
```

### Shell installer (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/protibimbok/mgit/main/scripts/install.sh | bash
```

Installs to `/usr/local/bin` by default. Override with `INSTALL_DIR=/your/path`.

### go install

```bash
go install github.com/protibimbok/mgit@latest
```

---

## Quick start

```bash
# 1. Create your first profile (generates SSH key + updates ~/.ssh/config)
mgit gen

# 2. Add the printed public key to GitHub → Settings → SSH and GPG keys

# 3. Clone a repo using the profile key
mgit clone work:myorg/myrepo

# 4. In an existing repo, init and attach a profile
mgit init
```

---

## Commands

### `mgit gen`

Interactively creates a new SSH profile.

```
Prompts:
  Full name        → stored in profile, used for git config
  Email            → SSH key comment + git config
  Key              → short alias used in URLs (e.g. "work", "personal")
  Label            → display name (defaults to key)

Creates:
  ~/.ssh/mgit_<key>          private key (ed25519)
  ~/.ssh/mgit_<key>.pub      public key
  ~/.ssh/config entry:
      Host hub.<key>
        HostName github.com
        User git
        IdentityFile ~/.ssh/mgit_<key>
        IdentitiesOnly yes
  ~/.config/mgit/profiles.json entry
```

After running, copy the printed public key to GitHub.

---

### `mgit init`

Initialises a git repo (if needed) and sets `user.name` / `user.email` from a profile.

```bash
mgit init
```

---

### `mgit clone`

Clones a repository using a profile key or an HTTPS URL.

```bash
# key:user/repo  →  git clone git@hub.<key>:<user>/<repo>
mgit clone work:myorg/myrepo
mgit clone personal:myname/dotfiles

# HTTPS URL → prompts you to pick a profile
mgit clone https://github.com/myorg/myrepo

# Extra git flags are forwarded
mgit clone work:myorg/myrepo --depth 1 --branch dev
```

---

### `mgit list`

Lists all registered profiles.

```
KEY           LABEL             EMAIL                          SSH KEY
---           -----             -----                          -------
work          Work              me@company.com                 /home/me/.ssh/mgit_work
personal      Personal          me@gmail.com                   /home/me/.ssh/mgit_personal
```

---

### `mgit fix [remote]`

Rewrites a GitHub HTTPS remote to SSH using a chosen profile. Default remote is `origin`.

```bash
mgit fix          # fixes origin
mgit fix upstream # fixes a different remote
```

---

### `mgit del [key]`

Removes a profile, its SSH key pair, and its `~/.ssh/config` block.

```bash
mgit del work    # remove by key
mgit del         # prompts if key not given
```

---

### Pass-through to git

Any unrecognised subcommand is forwarded to git with all arguments intact:

```bash
mgit status
mgit log --oneline -10
mgit push origin main
mgit stash pop
```

---

## How it works

Each profile gets a Host alias in `~/.ssh/config` that maps `hub.<key>` to `github.com` with the right identity file. URLs like `git@hub.work:myorg/myrepo` are resolved by SSH before git ever sees them.

```
mgit clone work:myorg/myrepo
        │
        └─► git clone git@hub.work:myorg/myrepo
                               │
                        SSH resolves to:
                        HostName github.com
                        IdentityFile ~/.ssh/mgit_work
```

Profiles are stored in `~/.config/mgit/profiles.json`.

---

## Release setup (for maintainers)

Homebrew and apt distribution are managed centrally in [protibimbok/pkg-dist](https://github.com/protibimbok/pkg-dist). This repo only builds binaries and publishes GitHub Releases.

### Required GitHub secrets (this repo)

| Secret | Purpose | Required |
|--------|---------|----------|
| `GITHUB_TOKEN` | Create GitHub Releases | Auto-provided |
| `PKG_DIST_TOKEN` | Trigger pkg-dist update after release | For Homebrew + apt |
| `AUR_KEY` | SSH private key for AUR updates | For AUR |

### Creating a release

```bash
git tag v1.0.0
git push origin v1.0.0
```

The release workflow builds binaries, publishes to GitHub Releases, updates AUR, and notifies pkg-dist to update Homebrew casks and the apt repository.

See [pkg-dist](https://github.com/protibimbok/pkg-dist) for signing keys, apt repo setup, and GitHub Pages configuration.
