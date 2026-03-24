# Build & Release

## Task Runner

The project uses [Task](https://taskfile.dev/) (`Taskfile.yml`) with `GO111MODULE=on` and `GOPROXY=https://proxy.golang.org,direct`.

| Task | Purpose |
|---|---|
| `task dev` | Install git pre-commit hook |
| `task setup` | `go mod tidy` |
| `task build` | Compile binary + generate CLI docs |
| `task install` | `go install ./cmd/feature` |
| `task test` | Run tests with `-race -cover` |
| `task cover` | Open HTML coverage report |
| `task fmt` | Format with `gofumpt` |
| `task lint` | Run golangci-lint |
| `task ci` | Full CI: setup → build → test |
| `task commit` | Interactive conventional commit via `gum` |
| `task release` | Tag next semver via `svu` and push |
| `task goreleaser` | Build release artifacts |

## Linting (`.golangci.yaml`)

Go 1.24 target. Notable rules:
- **Forbidden imports**: `github.com/pkg/errors` (use stdlib), `math/rand$` (use v2), `ioutil.*`
- **Tag conventions**: JSON → goCamel, YAML → snake_case
- Linters include: bodyclose, gocritic, misspell, revive, unconvert, unparam, and others

## Key Dependencies

| Module | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI command framework |
| `github.com/spf13/viper` | Configuration (flags, env vars) |
| `github.com/charmbracelet/lipgloss` + `fang` | Terminal styling and command wiring |
| `go.uber.org/automaxprocs` | Container-aware GOMAXPROCS |
| `oras.land/oras-go/v2` | OCI registry client for community features |
| `github.com/opencontainers/image-spec` | OCI image/manifest types |

## GoReleaser (`.goreleaser.yaml`)

- **Targets**: Linux amd64, arm64 only
- **Static linking**: `CGO_ENABLED=0`
- **ldflags**: Injects `version`, `commit`, `date`, `builtBy` into `main` package
- **Archives**: tar.gz format, includes README, LICENSE, completions, manpages
- **Changelog**: GitHub API, grouped by type (Features, Security, Bug fixes, Docs), filters test/chore commits

## CI — Woodpecker (`.woodpecker/workflow.yaml`)

Triggers on push to `main`:
1. `go build ./cmd/feature`
2. `./feature --version`

## Versioning (`.svu.yaml`)

Semantic versioning via [svu](https://github.com/caarlos0/svu):
- Tag prefix: `v`
- Tag mode: `all`
- v0 mode enabled (pre-1.0 breaking changes in minor bumps)

## Pre-commit Hook (`scripts/pre-commit.sh`)

1. Runs `gofmt` on staged `.go` files
2. Runs `golangci-lint --new --fix`
3. Re-stages modified files

## Documentation Generation

The hidden `gendocs` command generates markdown docs from the cobra command tree into `docs/`. The `scripts/clidocs.sh` script post-processes headings and updates the version in `_coverpage.md`.

## Shell Completions & Manpages

Generated via `scripts/completions.sh` (bash, zsh, fish → `completions/`) and `scripts/manpages.sh` (gzipped manpage → `manpages/`). Both are included in release archives.
