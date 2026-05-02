# Feature - Overview

## Purpose

`feature` is a Go CLI tool that installs [DevContainer features](https://github.com/devcontainers/features) on host systems outside of devcontainers — such as VPSes, VMs, and bare-metal Linux machines. It supports both built-in features (from a local clone of the upstream `devcontainers/features` repo) and community features pulled from OCI registries (e.g., `ghcr.io`). It runs feature install scripts as root and persists feature environment variables via `/etc/profile.d/` scripts.

## Architecture

```
cmd/feature/           CLI entry point and command logic
  main.go              Root cobra command, dual-path install orchestration
  root.go              Helper functions (root check, repo clone, install, env update)
  config.go            Viper-based configuration setup
  gendocs_command.go   Hidden command to generate CLI markdown docs

internal/oci/          OCI registry client for community features
  oci.go               Reference parsing, artifact pulling, tgz extraction

internal/options/      Feature metadata parsing
  options.go           Parses devcontainer-feature.json from built-in or arbitrary paths

scripts/               Build, commit, and release automation
docs/                  User-facing documentation (Docsify site)
```

### Execution Flow

The CLI accepts a feature argument that is either a short name (built-in) or an OCI reference (community). The path is chosen by `oci.IsOCIRef()`, which checks whether the first path segment contains a `.` (indicating a registry hostname).

#### Built-in Feature Path (`feature node`)

1. **Root check** — `checkRootUser()` verifies UID 0
2. **Repository setup** — `ensureRepo()` clones `devcontainers/features` to `--featureRoot` (default `~/.features`), optionally pulls updates
3. **Metadata** — `options.GetOptionsForFeature(root, feature)` reads `{featureRoot}/src/{feature}/devcontainer-feature.json`
4. **Dependency install** — `ensureCommonUtils()` installs `common-utils` if marker `/usr/local/etc/vscode-dev-containers/common` is absent
5. **Feature install** — `installFeature()` runs `{featureDir}/install.sh` via bash
6. **Environment update** — `updateEnvironment()` reads `ContainerEnv` from metadata and writes `/etc/profile.d/devcontainer-{id}.sh`

#### OCI Community Feature Path (`feature ghcr.io/devcontainers-extra/features/go-task:1`)

1. **Root check** — same as above
2. **Parse reference** — `oci.ParseFeatureRef()` splits into registry, namespace, name, and tag
3. **Pull & cache** — `oci.PullFeature()` downloads the OCI artifact to `{featureRoot}-oci/{registry}/{namespace}/{name}/{tag}/`, skipping download if cached (unless `--updateRepo` forces refresh)
4. **Metadata** — `options.GetOptionsForPath(featureDir)` reads `devcontainer-feature.json` from the extracted directory
5. **Feature install** — same `installFeature()` as built-in path
6. **Environment update** — same `updateEnvironment()` as built-in path

## Key Patterns

### Root-Only Execution
The tool requires root because feature install scripts modify system packages and directories. Enforced at command start via `checkRootUser()`.

### Dual Feature Sources
Features can come from two sources, unified into a single install flow:
- **Built-in**: Local clone of `github.com/devcontainers/features` at `{featureRoot}/src/{name}/`
- **Community/OCI**: Pulled from OCI registries (e.g., `ghcr.io`) to `{featureRoot}-oci/` cache

### OCI Reference Detection
`oci.IsOCIRef()` classifies arguments: if the first path segment (before `/`) contains a `.`, it's an OCI reference (e.g., `ghcr.io/...`). Otherwise it's a built-in short name. This avoids ambiguity since bare feature names like `node` have no dots.

### OCI Pulling & Caching
`oci.PullFeature()` uses `oras.land/oras-go/v2` for registry operations with anonymous auth (sufficient for public registries). It handles both OCI image manifests and Docker v2 manifests. The layer is extracted from a `.tgz` with zip-slip prevention. Caching is tag-based: if `install.sh` exists in the cache dir, re-download is skipped unless `--updateRepo` is set.

### Feature Metadata
Each feature has a `devcontainer-feature.json` parsed into `FeatureOptions` (see [data-model.md](data-model.md)). Two entry points:
- `GetOptionsForFeature(root, feature)` — for built-in features, constructs path from root + `src/` + name
- `GetOptionsForPath(featureDir)` — for any feature directory (used by OCI path)

### Environment Persistence
Container environment variables from feature metadata are written to `/etc/profile.d/devcontainer-{id}.sh` as `export KEY="VALUE"` lines (values are safely quoted via Go's `%q` verb), making them available in new shell sessions. Environment variable keys are validated to contain only alphanumeric characters and underscores to prevent shell injection.

### Configuration Layering
Viper binds flags and environment variables with prefix `FEATURE_`. Dots and dashes in flag names are stripped for env var matching:
- `--featureRoot` → `FEATURE_FEATUREROOT`
- `--updateRepo` → `FEATURE_UPDATEREPO`

### CLI Framework
Cobra for commands, Charmbracelet (lipgloss, fang) for terminal styling, `automaxprocs` for container-aware GOMAXPROCS. The `NOCOLOR` env var disables colored output.

## Configuration

| Flag / Env Var | Default | Description |
|---|---|---|
| `-r, --featureRoot` / `FEATURE_FEATUREROOT` | `~/.features` | Local path for cloned features repo (OCI cache goes to `{value}-oci/`) |
| `-u, --updateRepo` / `FEATURE_UPDATEREPO` | `false` | Pull latest features before install; also forces OCI cache refresh |
| `NOCOLOR` | (unset) | Disable colored terminal output |

## Build & Release

- **Build**: `task build` — compiles binary and generates docs
- **Lint**: golangci-lint with strict ruleset (see `.golangci.yaml`)
- **Release**: GoReleaser Pro builds static Linux binaries (amd64, arm64), generates manpages and shell completions
- **CI (GitHub Actions)**: `go.yml` on push/PR to main — builds and tests; `release.yml` on tags — runs GoReleaser Pro
- **CI (Woodpecker)**: Pipeline on push to main — builds and runs version check
- **Versioning**: Semantic versioning via `svu`, conventional commits enforced

See also:
- [Data Model](data-model.md) — `FeatureOptions` struct, `FeatureRef` struct, and JSON schema
- [Build & Release](build-release.md) — Detailed build, lint, CI, and release configuration
