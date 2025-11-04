# semtag

Semantic version tagging tool - automatically calculate version numbers based on Git commits and branches.

## Features

- 🚀 Automatically determine next version following Conventional Commits
- 📌 Get the current highest version tag from the branch
- 🌿 Automatically append pre-release suffixes based on branch (e.g. beta, alpha, rc)
- 🔄 Support both stable and pre-release versioning strategies

## Installation

### Using `go install`

```bash
go install github.com/google-internal/semtag/cmd/semtag@latest
```

### Download from Release

Download the pre-built binary for your platform:

```bash
# macOS (arm64)
curl -L https://github.com/google-internal/semtag/releases/latest/download/semtag_darwin_arm64 -o semtag
chmod +x semtag
sudo mv semtag /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/google-internal/semtag/releases/latest/download/semtag_linux_amd64 -o semtag
chmod +x semtag
sudo mv semtag /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/google-internal/semtag.git
cd semtag
go build -o semtag ./cmd/semtag
```

## Usage

### Commands

`semtag` defaults to `next` command (i.e., `semtag` equals `semtag next`).

| Command | Description |
|---------|-------------|
| `next` | Calculate the next version based on commits and branch |
| `current` | Get the current highest version tag (returns `0.0.0` if no tags exist) |

### Basic Usage

```bash
# Calculate next semantic version
semtag next

# Get current version
semtag current
```

### Create and Push Tags

```bash
# Calculate next version with prefix
VERSION=$(semtag next -p v)

# Create annotated tag
git tag -a "$VERSION" -m "Release ${VERSION}"

# Push tag
git push origin "refs/tags/${VERSION}"
```

### Options

- `--prefix`, `-p`: Version prefix (default: empty string)
- `--branch-suffix`: Custom branch to pre-release suffix mapping (format: `branch:suffix`)

### Help

```bash
semtag --help
semtag next --help
```

## Version Calculation Rules

### Conventional Commits

`semtag` follows [Conventional Commits](https://www.conventionalcommits.org/) to determine version bumps:

#### Major (x.0.0)

Breaking changes trigger major version bump:

```bash
# Commit message with '!' marker
git commit -m "feat!: breaking API change"

# Commit body with BREAKING CHANGE
git commit -m "feat: new feature" -m "BREAKING CHANGE: API redesign"
```

#### Minor (0.x.0)

New features trigger minor version bump:

```bash
git commit -m "feat: add export functionality"
git commit -m "feat(auth): implement user authentication"
```

#### Patch (0.0.x)

Bug fixes trigger patch version bump:

```bash
git commit -m "fix: resolve login issue"
git commit -m "fix(api): correct API response"
```

### Branch Pre-release Suffixes

`semtag` automatically appends pre-release suffixes based on the current branch:

| Branch | Suffix | Example |
|--------|--------|---------|
| `beta`, `develop`, `dev`, `staging` | `-beta.N` | `1.2.3-beta.1` |
| `alpha`, `experimental`, `wip` | `-alpha.N` | `1.2.3-alpha.1` |
| `rc`, `release`, `release/*` | `-rc.N` | `1.2.3-rc.1` |
| `hotfix`, `hotfix/*` | `-beta.N` | `1.2.3-beta.1` |
| `feature/*` | `-alpha.N` | `1.2.3-alpha.1` |
| `main`, `master` | No suffix | `1.2.3` |

## License

MIT License

