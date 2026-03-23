# Data Model

## FeatureOptions

Defined in `internal/options/options.go`. Parsed from `devcontainer-feature.json` files via `GetOptionsForFeature(root, feature)`.

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

## JSON Source

The `devcontainer-feature.json` file lives at `{featureRoot}/src/{feature}/devcontainer-feature.json`. Example structure:

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
