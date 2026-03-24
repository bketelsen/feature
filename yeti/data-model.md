# Data Model

## FeatureOptions

Defined in `internal/options/options.go`. Parsed from `devcontainer-feature.json` files.

Two entry points:
- `GetOptionsForFeature(root, feature)` — reads from `{root}/src/{feature}/devcontainer-feature.json` (built-in features)
- `GetOptionsForPath(featureDir)` — reads from `{featureDir}/devcontainer-feature.json` (any path, used by OCI flow)

```go
type FeatureOptions struct {
    ID               string            // Feature identifier (e.g., "go", "node")
    Version          string            // Semantic version
    Name             string            // Display name
    DocumentationURL string            // URL to feature documentation
    Description      string            // Human-readable description
    Options          map[string]Option // User-configurable options
    Init             bool              // Whether feature needs init
    Customizations   Customizations    // IDE-specific customizations
    ContainerEnv     map[string]string // Environment variables to persist
    CapAdd           []string          // Linux capabilities required
    SecurityOpt      []string          // Security options
    InstallsAfter    []string          // Feature dependencies
}
```

### Option

Individual configuration option for a feature:

```go
type Option struct {
    Type        string   // Data type: "string", "boolean", etc.
    Proposals   []string // Suggested values
    Default     any      // Default value
    Description string   // Human-readable description
}
```

### Customizations

IDE-specific configuration, currently only VS Code:

```go
type Customizations struct {
    Vscode Vscode
}

type Vscode struct {
    Extensions []string // VS Code extension IDs to install
}
```

## FeatureRef

Defined in `internal/oci/oci.go`. Represents a parsed OCI feature reference for community features.

```go
type FeatureRef struct {
    Registry  string // e.g., "ghcr.io"
    Namespace string // e.g., "devcontainers-extra/features"
    Name      string // e.g., "go-task"
    Tag       string // e.g., "1" (default: "latest")
}
```

### Parsing

`ParseFeatureRef(ref)` splits an OCI reference string:
- Registry = first path segment (before first `/`)
- Namespace = middle segments joined by `/`
- Name and Tag = last segment split on `:` (tag defaults to `"latest"`)
- Requires at least 3 path segments (registry + namespace + name)

Examples:
- `ghcr.io/devcontainers/features/go:1` → `{ghcr.io, devcontainers/features, go, 1}`
- `ghcr.io/devcontainers-extra/features/go-task` → `{ghcr.io, devcontainers-extra/features, go-task, latest}`

### Cache Layout

OCI features are cached at `{featureRoot}-oci/{registry}/{namespace}/{name}/{tag}/`. The cache directory contains the extracted feature artifact including `install.sh` and `devcontainer-feature.json`.

## JSON Source

The `devcontainer-feature.json` file lives at `{featureRoot}/src/{feature}/devcontainer-feature.json` for built-in features, or in the OCI cache directory for community features. Example structure:

```json
{
  "id": "go",
  "version": "1.2.3",
  "name": "Go",
  "description": "Installs Go and common tools",
  "options": {
    "version": {
      "type": "string",
      "proposals": ["latest", "1.21"],
      "default": "latest",
      "description": "Go version to install"
    }
  },
  "containerEnv": {
    "GOPATH": "/go"
  },
  "installsAfter": ["ghcr.io/devcontainers/features/common-utils"]
}
```

## Key Usage

- `ContainerEnv` is read by `updateEnvironment()` in `cmd/feature/root.go` to write `/etc/profile.d/devcontainer-{id}.sh`
- `Options` are not currently passed to install scripts — the tool uses feature defaults
- `InstallsAfter` is not currently used for dependency resolution; only `common-utils` is handled as a hardcoded prerequisite
