// Package oci provides support for pulling devcontainer features from OCI registries.
package oci

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"oras.land/oras-go/v2/registry/remote"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// FeatureRef represents a parsed OCI feature reference.
type FeatureRef struct {
	Registry  string // e.g., "ghcr.io"
	Namespace string // e.g., "devcontainers-extra/features"
	Name      string // e.g., "go-task"
	Tag       string // e.g., "1" (default: "latest")
}

// Repository returns the full repository path (namespace/name).
func (r FeatureRef) Repository() string {
	return r.Namespace + "/" + r.Name
}

// IsOCIRef returns true if the argument looks like an OCI reference.
// It checks whether the first path segment contains a dot, indicating a registry hostname.
func IsOCIRef(arg string) bool {
	firstSlash := strings.Index(arg, "/")
	if firstSlash < 0 {
		return false
	}
	return strings.Contains(arg[:firstSlash], ".")
}

// ParseFeatureRef parses an OCI feature reference string into its components.
// Expected format: registry/namespace[/...]/name[:tag]
// Examples:
//
//	ghcr.io/devcontainers/features/go:1
//	ghcr.io/devcontainers-extra/features/go-task:latest
func ParseFeatureRef(ref string) (FeatureRef, error) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return FeatureRef{}, fmt.Errorf("invalid OCI feature reference %q: expected registry/namespace/name[:tag]", ref)
	}

	registry := parts[0]
	rest := parts[1] // e.g., "devcontainers-extra/features/go-task:1"

	// Split the remaining path into segments
	segments := strings.Split(rest, "/")
	if len(segments) < 2 {
		return FeatureRef{}, fmt.Errorf("invalid OCI feature reference %q: need at least registry/namespace/name", ref)
	}

	// Last segment is name[:tag]
	last := segments[len(segments)-1]
	namespace := strings.Join(segments[:len(segments)-1], "/")

	name := last
	tag := "latest"
	if colonIdx := strings.LastIndex(last, ":"); colonIdx >= 0 {
		name = last[:colonIdx]
		tag = last[colonIdx+1:]
		if tag == "" {
			tag = "latest"
		}
	}

	if name == "" {
		return FeatureRef{}, fmt.Errorf("invalid OCI feature reference %q: empty feature name", ref)
	}

	return FeatureRef{
		Registry:  registry,
		Namespace: namespace,
		Name:      name,
		Tag:       tag,
	}, nil
}

// PullFeature pulls a devcontainer feature from an OCI registry and extracts it to cacheDir.
// If the feature is already cached (install.sh exists) and forceUpdate is false, it returns
// the cached path without re-downloading.
func PullFeature(ctx context.Context, ref FeatureRef, cacheDir string, forceUpdate bool) (string, error) {
	// Check cache
	if !forceUpdate {
		if _, err := os.Stat(filepath.Join(cacheDir, "install.sh")); err == nil {
			return cacheDir, nil
		}
	}

	// Clean cache dir for fresh pull
	if err := os.RemoveAll(cacheDir); err != nil {
		return "", fmt.Errorf("clearing cache directory: %w", err)
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("creating cache directory: %w", err)
	}

	// Connect to registry
	repo, err := remote.NewRepository(ref.Registry + "/" + ref.Repository())
	if err != nil {
		return "", fmt.Errorf("connecting to registry: %w", err)
	}
	// Use anonymous auth (sufficient for public registries)
	repo.PlainHTTP = false

	// Fetch the manifest
	descriptor, rc, err := repo.FetchReference(ctx, ref.Tag)
	if err != nil {
		return "", fmt.Errorf("fetching manifest for %s:%s: %w (if this is a private registry, try 'docker login %s' first)", ref.Repository(), ref.Tag, err, ref.Registry)
	}
	defer rc.Close()

	manifestBytes, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("reading manifest: %w", err)
	}

	// Parse manifest to find the feature layer
	var layerDesc ocispec.Descriptor
	switch descriptor.MediaType {
	case ocispec.MediaTypeImageManifest, "application/vnd.docker.distribution.manifest.v2+json":
		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			return "", fmt.Errorf("parsing manifest: %w", err)
		}
		if len(manifest.Layers) == 0 {
			return "", fmt.Errorf("OCI manifest for %s has no layers", ref.Repository())
		}
		// Prefer the devcontainer-specific media type, fall back to first layer
		layerDesc = manifest.Layers[0]
		for _, l := range manifest.Layers {
			if l.MediaType == "application/vnd.devcontainers.layer.v1+tar" {
				layerDesc = l
				break
			}
		}
	default:
		return "", fmt.Errorf("unsupported manifest media type: %s", descriptor.MediaType)
	}

	// Fetch the layer blob
	layerRC, err := repo.Fetch(ctx, layerDesc)
	if err != nil {
		return "", fmt.Errorf("fetching feature layer: %w", err)
	}
	defer layerRC.Close()

	// Extract the tgz to cacheDir
	if err := extractTGZ(layerRC, cacheDir); err != nil {
		return "", fmt.Errorf("extracting feature archive: %w", err)
	}

	// Validate required files exist
	for _, required := range []string{"install.sh", "devcontainer-feature.json"} {
		if _, err := os.Stat(filepath.Join(cacheDir, required)); os.IsNotExist(err) {
			return "", fmt.Errorf("extracted feature is missing required file %q", required)
		}
	}

	// Ensure install.sh is executable
	if err := os.Chmod(filepath.Join(cacheDir, "install.sh"), 0o755); err != nil {
		return "", fmt.Errorf("making install.sh executable: %w", err)
	}

	return cacheDir, nil
}

// maxExtractFileSize is the maximum allowed size for a single file extracted from a tar archive.
// This prevents decompression bombs from exhausting disk space.
const maxExtractFileSize = 500 * 1024 * 1024 // 500 MB

// extractTGZ extracts a .tgz archive from r into targetDir with zip-slip prevention.
func extractTGZ(r io.Reader, targetDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar entry: %w", err)
		}

		// Zip-slip prevention: ensure the path stays within targetDir
		target := filepath.Join(targetDir, filepath.Clean(header.Name))
		if !strings.HasPrefix(target, filepath.Clean(targetDir)+string(os.PathSeparator)) && target != filepath.Clean(targetDir) {
			return fmt.Errorf("tar entry %q attempts path traversal outside target directory", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if header.Size > maxExtractFileSize {
				return fmt.Errorf("tar entry %q size %d exceeds maximum allowed size %d", header.Name, header.Size, maxExtractFileSize)
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, io.LimitReader(tr, maxExtractFileSize+1)); err != nil {
				f.Close()
				return err
			}
			fi, err := f.Stat()
			f.Close()
			if err != nil {
				return err
			}
			if fi.Size() > maxExtractFileSize {
				return fmt.Errorf("tar entry %q decompressed to %d bytes, exceeding maximum allowed size %d", header.Name, fi.Size(), maxExtractFileSize)
			}
		}
	}
	return nil
}
