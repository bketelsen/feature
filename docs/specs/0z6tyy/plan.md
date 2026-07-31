# Implementation plan: Upgrade Go to 1.26.5

## Approach

Pin the module, lint target, development container, GitHub Actions workflows, and Woodpecker image to Go 1.26.5 without changing application behavior or unrelated tool versions. Preserve every existing workflow trigger and command. Regenerate module metadata with Go 1.26.5, update the two documents that currently state older toolchain versions, and validate using the repository's Task targets plus the commands embedded in its CI workflows.

## Ordered implementation steps

### 1. Establish the toolchain and change boundary

1. Work from `/tmp/spec-forge.0Z6TYy` and inspect `git status --short` before editing so existing work, including this specification and plan, is not overwritten.
2. Confirm the validation shell is actually running the requested toolchain:

   ```sh
   go version
   GOTOOLCHAIN=local go version
   ```

   Both validations used for the upgrade must report `go1.26.5`; `GOTOOLCHAIN=local` prevents an implicit download or switch from hiding a mismatched local installation.
3. Keep source under `cmd/` and `internal/`, dependency versions, GitHub Action versions, GoReleaser settings, Task definitions, and product-facing `feature go` behavior unchanged unless a concrete Go 1.26.5 incompatibility is found during validation.

### 2. Align the module and lint target

Files:

- `go.mod`
- `.golangci.yaml`
- `go.sum` only if Go 1.26.5 module maintenance produces a deterministic change

Changes:

1. Replace `go 1.24.2` in `go.mod` with `go 1.26.5`. Do not add a separate `toolchain` directive; the exact module directive is the repository's canonical minimum.
2. Replace `.golangci.yaml`'s `run.go: "1.24"` value with `"1.26.5"`. Leave the v2 schema, enabled linters, settings, and exclusions untouched.
3. Run `GOTOOLCHAIN=local go mod tidy` after the directive change. Review the resulting `go.mod`/`go.sum` diff and retain only changes produced by tidy under Go 1.26.5; do not run a general dependency update. The current dependency graph is expected to remain unchanged apart from the `go` directive, but the generated output is authoritative.

### 3. Pin local development tooling

Files:

- `.devcontainer/Dockerfile`
- `.devcontainer/devcontainer.json`

Changes:

1. Change the Dockerfile default from `ARG VARIANT=1.24-bookworm` to `ARG VARIANT=1.26.5-bookworm`. Preserve `FROM golang:${VARIANT}` and the Bookworm distribution.
2. Change the existing `ghcr.io/devcontainers/features/go:1` options object from `{}` to:

   ```json
   {
       "version": "1.26.5"
   }
   ```

   Do not change the feature revision or the other configured development-container features.
3. Review the resulting JSON-with-comments structure and Dockerfile diff to ensure no unrelated formatting or container customization changes were introduced.

### 4. Pin every CI and release path

Files:

- `.github/workflows/go.yml`
- `.github/workflows/release.yml`
- `.woodpecker/workflow.yaml`

Changes:

1. In `.github/workflows/go.yml`, explicitly replace the outlier `go-version: '1.23'` value in the existing `actions/setup-go@v4` step with `go-version: '1.26.5'`; this workflow is not on the 1.24 value used by the other older selectors.
2. In `.github/workflows/release.yml`, replace the moving `stable` input to the existing `actions/setup-go@v5` step with the quoted exact value `'1.26.5'`.
3. In `.woodpecker/workflow.yaml`, change the build image to `golang:1.26.5`.
4. Diff each workflow and verify that only its Go selector changed:
   - `go.yml` still triggers on pushes and pull requests to `main` and retains `go build -v ./cmd/feature` followed by `go test -v ./...`.
   - `release.yml` still triggers on tag pushes, keeps its permissions and secrets, and retains GoReleaser Pro with `args: release --clean`.
   - Woodpecker still triggers on pushes to `main` and retains `go build ./cmd/feature` followed by `./feature --version`.

### 5. Synchronize maintained documentation

Files:

- `CONTRIBUTING.md`
- `yeti/build-release.md`

Review without an expected edit:

- `README.md`
- `CLAUDE.md`

Changes:

1. Update the prerequisite in `CONTRIBUTING.md` from Go 1.24+ to Go 1.26.5+.
2. Update all four stale descriptions in `yeti/build-release.md`:
   - golangci-lint targets Go 1.26.5;
   - the GitHub build/test workflow sets up Go 1.26.5;
   - the GitHub release workflow sets up Go 1.26.5 rather than `stable`;
   - Woodpecker uses `golang:1.26.5`.
3. Review `README.md` and `CLAUDE.md` as directed by `CLAUDE.md`. Do not alter them merely to create a diff: neither currently declares the repository development toolchain, and `README.md`'s “Install latest version of Go” text describes the shipped `feature go` command rather than this repository's build environment.

### 6. Perform static consistency checks

1. Inspect all active selectors and their context:

   ```sh
   git grep -nE '(^go 1\.|run:|go-version:|ARG VARIANT=|image: golang:|Go 1\.)' -- \
     go.mod .golangci.yaml .devcontainer .github/workflows \
     .woodpecker CONTRIBUTING.md yeti/build-release.md
   ```

2. Confirm the operational selectors are exactly:
   - `go.mod`: `go 1.26.5`;
   - `.golangci.yaml`: `run.go` is `1.26.5`;
   - devcontainer Docker variant: `1.26.5-bookworm`;
   - devcontainer Go feature option: `1.26.5`;
   - both `setup-go` inputs: `1.26.5`;
   - Woodpecker image: `golang:1.26.5`.
3. Search the affected configuration and documentation for stale selectors:

   ```sh
   git grep -nE '1\.23|1\.24(\.2)?|go-version:[[:space:]]*stable' -- \
     go.mod .golangci.yaml .devcontainer .github/workflows \
     .woodpecker CONTRIBUTING.md yeti/build-release.md
   ```

   This command must return no matches. Do not treat `.goreleaser.yaml`'s existing `templ@latest` as a Go toolchain selector.
4. Run whitespace/error checking:

   ```sh
   git diff --check
   ```

## Test strategy

### Module, build, and unit tests

Run the same core commands used by `.github/workflows/go.yml`, forcing the already-installed Go 1.26.5 toolchain:

```sh
GOTOOLCHAIN=local go build -v ./cmd/feature
GOTOOLCHAIN=local go test -v ./...
```

The test command must include and pass `internal/oci` and `internal/options`; `cmd/feature` currently has no test files.

Then exercise the repository's fuller local CI path from `Taskfile.yml`:

```sh
GOTOOLCHAIN=local task ci
```

`task ci` runs `task setup` (`go mod tidy`), `task build` (`go build ./cmd/feature` and `go run ./cmd/feature gendocs`), and `task test` (race detection, atomic coverage, and a 15-minute timeout). The direct `gendocs` call is known to rewrite the six tracked `docs/feature*.md` files even on the pristine checkout: it emits Cobra's raw heading levels and does not perform the heading/version post-processing in `scripts/clidocs.sh`. That post-processing is invoked by `scripts/commit.sh` through `task commit`, not by `task build` or `task ci`. Treat those six generated-document changes as a pre-existing Task behavior rather than an upgrade result, inspect them only to confirm that they are confined to `docs/feature*.md`, and discard them before reviewing the Go-upgrade diff:

```sh
git checkout -- docs/feature.md docs/feature_completion*.md
```

After that cleanup, review module files and the rest of the working tree for unexplained changes.

To prove module tidiness in a clean post-change checkout, run:

```sh
GOTOOLCHAIN=local go mod tidy
git diff --exit-code -- go.mod go.sum
```

During implementation, compare tidy output with the intentional module edit before using `git diff --exit-code`; in CI or final verification this command is run after the expected upgrade diff is committed or otherwise established as the baseline.

### Lint configuration

Validate the repository's golangci-lint v2 configuration and its new language target:

```sh
golangci-lint config verify --config .golangci.yaml
```

The repository's normal lint command is:

```sh
task lint
# equivalent to:
golangci-lint run --config ./.golangci.yaml ./...
```

Run it to detect a Go 1.26.5 compatibility regression, but do not expand this change to fix pre-existing source findings. In the current checkout, the lint task reports existing `noctx` findings for `exec.Command` calls in `cmd/feature/root.go`; acceptance for this upgrade requires successful config verification and no new unsupported-version/configuration failure, not unrelated source cleanup.

### Woodpecker command parity

Exercise the commands retained from `.woodpecker/workflow.yaml` under Go 1.26.5:

```sh
GOTOOLCHAIN=local go build ./cmd/feature
./feature --version
rm -f ./feature
```

### Development-container validation

Where the Dev Container CLI and Docker are available, build/start the checked-in configuration without version overrides and inspect the active executable:

```sh
devcontainer up --workspace-folder .
devcontainer exec --workspace-folder . go version
```

The second command must report `go version go1.26.5 ...`. This end-to-end check is necessary because both the Dockerfile base and the Go feature can affect the final `PATH`. As a lower-level prerequisite, the Dockerfile itself can be checked with:

```sh
docker build -f .devcontainer/Dockerfile -t feature-devcontainer-go-1.26.5 .
docker run --rm feature-devcontainer-go-1.26.5 go version
```

The lower-level image check must also report Go 1.26.5, but it does not exercise features and therefore does not validate the `ghcr.io/devcontainers/features/go:1` option. If the Dev Container CLI is unavailable, record the full feature-aware check as not run, verify the `"version": "1.26.5"` feature option by configuration review/static parity, and use the Docker build only to validate the `golang:1.26.5-bookworm` base. Do not present the Docker-only result as end-to-end devcontainer validation.

### Release configuration

The tag workflow requires repository secrets and must not be simulated by creating or pushing a release tag. Verify its unchanged GoReleaser configuration locally when GoReleaser Pro-compatible tooling is available using the repository's existing Task target:

```sh
task release-check
```

Final release-path confidence comes from the static workflow review plus a real tag-triggered GitHub Actions run: its setup step must install Go 1.26.5 before the unchanged GoReleaser Pro `release --clean` action executes.

## Expected file set

Expected modifications:

- `go.mod`
- `.golangci.yaml`
- `.devcontainer/Dockerfile`
- `.devcontainer/devcontainer.json`
- `.github/workflows/go.yml`
- `.github/workflows/release.yml`
- `.woodpecker/workflow.yaml`
- `CONTRIBUTING.md`
- `yeti/build-release.md`

Conditionally modified:

- `go.sum`, only if `GOTOOLCHAIN=local go mod tidy` under Go 1.26.5 changes checksums or the dependency graph.

Reviewed but not expected to change:

- `README.md`
- `CLAUDE.md`
- `Taskfile.yml`
- `.goreleaser.yaml`
- Go source under `cmd/` and `internal/`

## Completion checklist

- Every repository-controlled Go selector is pinned to 1.26.5, with Bookworm retained only as the devcontainer image suffix.
- Workflow triggers, actions, commands, permissions, and secret names are unchanged apart from Go version inputs/images.
- Documentation matches the final configuration.
- Module tidy, direct GitHub build/test commands, `task ci`, golangci-lint configuration verification, and Woodpecker command parity complete as described; the known raw `docs/feature*.md` regeneration from `task ci` is discarded rather than committed.
- When the Dev Container CLI is available, the feature-aware `devcontainer up`/`devcontainer exec ... go version` check reports Go 1.26.5. When it is unavailable, that check is explicitly recorded as not run, the Go feature's `"version": "1.26.5"` option is verified by configuration review, and the Docker-only check validates the base image without being represented as feature validation.
- `git diff --check` passes, and the final diff contains no regenerated `docs/feature*.md`, application-source changes, or unrelated dependency/tool upgrades.
