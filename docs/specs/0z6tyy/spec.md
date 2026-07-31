# Upgrade the Go toolchain to 1.26.5

## Problem statement

The repository does not currently use one consistent Go toolchain. The module declares Go 1.24.2, the lint configuration targets Go 1.24, the development container starts from Go 1.24 while its Go feature is left on that feature's default, the main GitHub Actions workflow installs Go 1.23, the release workflow follows the moving `stable` alias, and Woodpecker uses a Go 1.24 image. As a result, local builds, linting, pull-request validation, and releases can compile the same source with different language and standard-library versions. The moving release alias also makes release builds non-reproducible with respect to the Go patch version.

Upgrade every repository-controlled Go toolchain selector to the exact requested release, Go 1.26.5, and update the documentation that describes those selectors. Builds, tests, lint configuration, development environments, and release automation must agree on that version.

## Grounded references

The following checked-in locations currently control or document the Go version and are affected:

- `go.mod:3` declares `go 1.24.2`. This is the module's minimum Go version and the version inherited by Go commands in `Taskfile.yml`, the scripts, and GoReleaser hooks.
- `.golangci.yaml:1-4` is a golangci-lint v2 configuration whose `run.go` target is `1.24`.
- `.devcontainer/Dockerfile:1-2` defaults `VARIANT` to `1.24-bookworm` and uses it in `FROM golang:${VARIANT}`.
- `.devcontainer/devcontainer.json:2-9` builds that Dockerfile and also enables `ghcr.io/devcontainers/features/go:1` with an empty options object. The Go feature therefore is not currently pinned by this repository and can replace or augment the image's Go installation independently of the Dockerfile default.
- `.github/workflows/go.yml:17-28` is the push and pull-request build/test workflow. Its `Set up Go` step uses `actions/setup-go@v4` with `go-version: '1.23'` before running `go build -v ./cmd/feature` and `go test -v ./...`.
- `.github/workflows/release.yml:17-38` is the tag-triggered GoReleaser workflow. Its `Set up Go` step uses `actions/setup-go@v5` with the moving `go-version: stable` value before invoking GoReleaser Pro.
- `.woodpecker/workflow.yaml:1-10` builds pushes to `main` in `golang:1.24` and then executes the resulting binary.
- `CONTRIBUTING.md:6-14` documents `Go 1.24+` as a development prerequisite.
- `yeti/build-release.md:22-24`, `yeti/build-release.md:51-70` describe the lint target and each CI toolchain as Go 1.24, Go 1.23, `stable`, and `golang:1.24`, respectively.

The following Go consumers do not contain their own toolchain selector, but their behavior must be validated under the upgraded module/tooling configuration:

- `Taskfile.yml:15-73` runs module tidy, build, generated-doc, test, coverage, formatting, and lint commands from the active Go environment.
- `.goreleaser.yaml:6-29` runs Go commands in release pre-hooks and builds `./cmd/feature`; in GitHub releases it inherits the toolchain installed by `.github/workflows/release.yml`.
- `scripts/clidocs.sh`, `scripts/completions.sh`, and `scripts/manpages.sh` invoke `go run`, while `scripts/pre-commit.sh` invokes golangci-lint.

## Scope

### In scope

1. **Module and lint language target**
   - Change the `go` directive in `go.mod` to the exact patch version `1.26.5`.
   - Change `.golangci.yaml`'s `run.go` value to `1.26.5`, so lint analysis uses the same Go language/standard-library target as the module.
   - Run module maintenance with Go 1.26.5 and commit any deterministic `go.mod` or `go.sum` changes it produces. The toolchain upgrade alone is not a reason to upgrade unrelated dependencies.

2. **Local and containerized development tooling**
   - Change `.devcontainer/Dockerfile`'s default image variant to `1.26.5-bookworm`, preserving the existing Bookworm base choice while pinning the Go patch release.
   - Configure the existing `ghcr.io/devcontainers/features/go:1` entry in `.devcontainer/devcontainer.json` with its supported `version` option set to `1.26.5`. The active Go executable in a container built from the checked-in configuration must therefore not depend on that feature's moving default.

3. **All checked-in CI and release workflows**
   - Change the Go setup input in `.github/workflows/go.yml` from `1.23` to the exact string `1.26.5`.
   - Change the Go setup input in `.github/workflows/release.yml` from `stable` to the exact string `1.26.5`.
   - Change the Woodpecker build image in `.woodpecker/workflow.yaml` from `golang:1.24` to `golang:1.26.5`.
   - Preserve the workflows' existing triggers, commands, permissions, action/distribution versions, and secrets contract unless a change is strictly required for Go 1.26.5 compatibility.

4. **Documentation consistency**
   - Update `CONTRIBUTING.md` to state that Go 1.26.5 or newer is required for source development, consistent with the module's exact minimum patch release.
   - Update the existing version descriptions in `yeti/build-release.md` so the lint target, GitHub build workflow, GitHub release workflow, and Woodpecker image accurately describe their new pinned values.
   - Review `README.md` and `CLAUDE.md` as required by `CLAUDE.md:5-7`; do not add unrelated text where those files do not describe the development toolchain. The product-facing `feature go` behavior described in `README.md:41-45` is distinct from the repository build toolchain.

5. **Compatibility validation**
   - Validate module tidiness, compilation, and the complete Go test suite with Go 1.26.5.
   - Validate that the golangci-lint v2 configuration accepts the Go 1.26.5 target. Existing source findings from enabled linters are not part of this toolchain-only change.
   - Validate the development-container configuration sufficiently to show that its active `go` executable is Go 1.26.5.

### Out of scope

- Changes to CLI behavior, commands, flags, OCI handling, or other application source under `cmd/` and `internal/`.
- General dependency refreshes in `go.mod`/`go.sum`, including updates to the `svu` tool dependency, except for changes emitted by `go mod tidy` solely as a consequence of the Go directive upgrade.
- Upgrading GitHub Action major versions, GoReleaser, golangci-lint, Task, devcontainer feature revisions, or other non-Go-toolchain tools unless the existing version cannot operate with Go 1.26.5.
- Changing the devcontainer's Debian Bookworm base, CI triggers, release targets, archive formats, credentials, permissions, or secret names.
- Pinning the version of Go that the shipped `feature go` command installs on a target machine; this request concerns the toolchain used to develop, validate, and release this repository.
- Fixing pre-existing lint findings or unrelated formatting/documentation issues.

## Acceptance criteria

1. **One exact module and lint target**
   - `go.mod` contains exactly `go 1.26.5` as its Go directive.
   - `.golangci.yaml` contains `run.go: "1.26.5"` (equivalent YAML formatting is acceptable), and `golangci-lint config verify --config .golangci.yaml` succeeds with a golangci-lint v2 release that supports Go 1.26.

2. **Development environment is pinned**
   - `.devcontainer/Dockerfile` defaults to `ARG VARIANT=1.26.5-bookworm` and continues to build from `golang:${VARIANT}`.
   - The existing Go feature in `.devcontainer/devcontainer.json` has `"version": "1.26.5"` rather than an empty options object.
   - Building and starting the checked-in devcontainer without overriding `VARIANT` or the feature option makes `go version` report `go1.26.5`.

3. **Every CI path selects Go 1.26.5**
   - The `Set up Go` steps in both `.github/workflows/go.yml` and `.github/workflows/release.yml` use the quoted exact value `1.26.5`; neither uses `1.23`, `1.24`, `stable`, or another moving Go-version alias.
   - `.woodpecker/workflow.yaml` uses `image: golang:1.26.5` for its build step.
   - The workflows retain their existing triggers and their workflow-specific commands: `.github/workflows/go.yml` still runs for pushes and pull requests to `main` and keeps `go build -v ./cmd/feature` and `go test -v ./...`; `.github/workflows/release.yml` still runs for tag pushes and keeps its GoReleaser Pro `release --clean` step; and `.woodpecker/workflow.yaml` still runs for pushes to `main` and keeps `go build ./cmd/feature` and `./feature --version`.
   - A run of each available CI path exposes Go 1.26.5 to its Go commands. In particular, GoReleaser and its Go pre-hooks execute after the release workflow's pinned setup step.

4. **The upgraded toolchain builds a tidy, working module**
   - In an environment where `go version` reports `go1.26.5`, `GOTOOLCHAIN=local go mod tidy` completes, and a subsequent `git diff --exit-code -- go.mod go.sum` is clean.
   - In that same environment, `GOTOOLCHAIN=local go build -v ./cmd/feature` succeeds.
   - In that same environment, `GOTOOLCHAIN=local go test -v ./...` succeeds, including the tests in `internal/oci` and `internal/options`.

5. **Version documentation matches configuration**
   - `CONTRIBUTING.md` identifies Go 1.26.5+ as the source-development prerequisite.
   - `yeti/build-release.md` identifies the lint target and both GitHub workflows as Go 1.26.5 and the Woodpecker image as `golang:1.26.5`.
   - No active, checked-in Go toolchain selector remains on Go 1.23, Go 1.24/1.24.2, or the `stable`/`latest` alias. This applies to `go.mod`, `.golangci.yaml`, `.devcontainer/`, `.github/workflows/`, and `.woodpecker/`; uses of `latest` for a non-toolchain dependency such as the existing `templ@latest` GoReleaser hook are not Go version selectors.
