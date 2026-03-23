/*
Copyright © 2025 Brian Ketelsen <bketelsen@gmail.com>
*/

// Package options provides a way to read and parse the devcontainer-feature.json file
package options

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FeatureOptions struct {
	ID               string            `json:"id"`
	Version          string            `json:"version"`
	Name             string            `json:"name"`
	DocumentationURL string            `json:"documentationURL"`
	Description      string            `json:"description"`
	Options          map[string]Option `json:"options"`
	Init             bool              `json:"init"`
	Customizations   Customizations    `json:"customizations"`
	ContainerEnv     map[string]string `json:"containerEnv"`
	CapAdd           []string          `json:"capAdd"`
	SecurityOpt      []string          `json:"securityOpt"`
	InstallsAfter    []string          `json:"installsAfter"`
}
type Option struct {
	Type        string   `json:"type"`
	Proposals   []string `json:"proposals"`
	Default     any      `json:"default"`
	Description string   `json:"description"`
}

type Vscode struct {
	Extensions []string `json:"extensions"`
}
type Customizations struct {
	Vscode Vscode `json:"vscode"`
}

// GetOptionsForPath reads and parses devcontainer-feature.json from an arbitrary directory.
func GetOptionsForPath(featureDir string) (FeatureOptions, error) {
	jsonPath := filepath.Join(featureDir, "devcontainer-feature.json")
	bb, err := os.ReadFile(jsonPath)
	if err != nil {
		return FeatureOptions{}, fmt.Errorf("feature metadata not found at %s: %w", jsonPath, err)
	}

	var fo FeatureOptions
	if err := json.Unmarshal(bb, &fo); err != nil {
		return FeatureOptions{}, err
	}
	return fo, nil
}

// GetOptionsForFeature reads and parses devcontainer-feature.json for a built-in feature.
func GetOptionsForFeature(root, feature string) (FeatureOptions, error) {
	featureDir := filepath.Join(root, "src", feature)
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		return FeatureOptions{}, fmt.Errorf("feature %s not found", feature)
	}
	return GetOptionsForPath(featureDir)
}
