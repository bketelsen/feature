/*
Copyright © 2025 Brian Ketelsen <bketelsen@gmail.com>
*/
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

func GetOptionsForFeature(root, feature string) (FeatureOptions, error) {
	// make sure the featureRoot exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return FeatureOptions{}, err
	}

	// make sure the feature exists
	if _, err := os.Stat(filepath.Join(root, "src", feature)); os.IsNotExist(err) {
		return FeatureOptions{}, fmt.Errorf("feature %s not found", feature)
	}

	// Construct the path to the devcontainer-feature.json file
	jsonPath := filepath.Join(root, "src", feature, "devcontainer-feature.json")
	// Read the file
	bb, err := os.ReadFile(jsonPath)
	if err != nil {
		return FeatureOptions{}, err
	}

	// Unmarshal the file into a FeatureOptions struct
	var fo FeatureOptions
	err = json.Unmarshal(bb, &fo)
	if err != nil {
		return FeatureOptions{}, err
	}

	// Return the struct
	return fo, nil

}
