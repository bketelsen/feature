# Feature - Overview

## Purpose

`feature` is a Go CLI tool that installs [DevContainer features](https://github.com/devcontainers/features) on host systems outside of devcontainers — such as VPSes, VMs, and bare-metal Linux machines. It clones the upstream devcontainers/features repository, runs feature install scripts as root, and persists the feature's environment variables via `/etc/profile.d/` scripts.

## Architecture

```
cmd/feature/           CLI entry point and command logic
  main.go              Root cobra command, flag definitions, run orchestration
  root.go              Core installation logic (clone, install, env update)
  config.go            Viper-based configuration setup
  gendocs_command.go   Hidden command to generate CLI markdown docs

internal/options/      Feature metadata parsing
  options.go           Parses devcontainer-feature.json into FeatureOptions struct

scripts/               Build, commit, and release automation
docs/                  User-facing documentation (Docsify site)
```

### Execution Flow

1. **Root check** — `checkRootUser()` verifies UID 0
2. **Repository setup** — `ensureRepo()` clones `devcontainers/features` to `--featureRoot` (default `~/.features`), optionally pulls updates
3. **Dependency install** — `ensureCommonUtils()` installs `common-utils` if marker `/usr/local/etc/vscode-dev-containers/common` is absent
4. **Feature install** — `installFeature()` runs `{featureRoot}/src/{feature}/install.sh` via bash
5. **Environment update** — `updateEnvironment()` reads `ContainerEnv` from `devcontainer-feature.json` and writes `/etc/profile.d/devcontainer-{id}.sh`

## Key Patterns

### Root-Only Execution
The tool requires root because feature install scripts modify system packages and directories. Enforced at command start via `checkRootUser()`.

### Git-Based Feature Source
Features come from a local clone of `github.com/devcontainers/features`. The `--updateRepo` / `-u` flag triggers a `git pull` before installation. Feature scripts live at `{featureRoot}/src/{name}/install.sh`.

### Feature Metadata
Each feature has a `devcontainer-feature.json` parsed into `FeatureOptions` (see [data-model.md](data-model.md)). The `ContainerEnv` map drives environment variable persistence.

### Environment Persistence
Container environment variables from feature metadata are written to `/etc/profile.d/devcontainer-{id}.sh` as `export KEY=VALUE` lines, making them available in new shell sessions.

### Configuration Layering
Viper binds flags and environment variables with prefix `FEATURE_`. Dots and dashes in flag names are stripped for env var matching:
- `--featureRoot` → `FEATURE_FEATUREROOT`
- `--updateRepo` → `FEATURE_UPDATEREPO`

### CLI Framework
Cobra for commands, Charmbracelet (lipgloss, fang) for terminal styling, `automaxprocs` for container-aware GOMAXPROCS. The `NOCOLOR` env var disables colored output.

## Configuration

| Flag / Env Var | Default | Description |
|---|---|---|
| `-r, --featureRoot` / `FEATURE_FEATUREROOT` | `~/.features` | Local path for cloned features repo |
| `-u, --updateRepo` / `FEATURE_UPDATEREPO` | `false` | Pull latest features before install |
| `NOCOLOR` | (unset) | Disable colored terminal output |

## Build & Release

- **Build**: `task build` — compiles binary and generates docs
- **Lint**: golangci-lint with strict ruleset (see `.golangci.yaml`)
- **Release**: GoReleaser builds static Linux binaries (amd64, arm64), generates manpages and shell completions
- **CI**: Woodpecker pipeline on push to main — builds and runs version check
- **Versioning**: Semantic versioning via `svu`, conventional commits enforced

See also:
- [Data Model](data-model.md) — `FeatureOptions` struct and JSON schema
- [Build & Release](build-release.md) — Detailed build, lint, CI, and release configuration
