package oci

import (
	"testing"
)

func TestIsOCIRef(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"ghcr.io/devcontainers/features/go:1", true},
		{"ghcr.io/devcontainers-extra/features/go-task:latest", true},
		{"docker.io/library/feature:1", true},
		{"go", false},
		{"node", false},
		{"common-utils", false},
		{"devcontainers/features/go", false}, // no dot in first segment
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsOCIRef(tt.input)
			if got != tt.want {
				t.Errorf("IsOCIRef(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFeatureRef(t *testing.T) {
	tests := []struct {
		input   string
		want    FeatureRef
		wantErr bool
	}{
		{
			input: "ghcr.io/devcontainers/features/go:1",
			want: FeatureRef{
				Registry:  "ghcr.io",
				Namespace: "devcontainers/features",
				Name:      "go",
				Tag:       "1",
			},
		},
		{
			input: "ghcr.io/devcontainers-extra/features/go-task:latest",
			want: FeatureRef{
				Registry:  "ghcr.io",
				Namespace: "devcontainers-extra/features",
				Name:      "go-task",
				Tag:       "latest",
			},
		},
		{
			input: "ghcr.io/devcontainers/features/node",
			want: FeatureRef{
				Registry:  "ghcr.io",
				Namespace: "devcontainers/features",
				Name:      "node",
				Tag:       "latest",
			},
		},
		{
			input: "docker.io/myorg/features/mytool:2.0",
			want: FeatureRef{
				Registry:  "docker.io",
				Namespace: "myorg/features",
				Name:      "mytool",
				Tag:       "2.0",
			},
		},
		{
			input:   "ghcr.io/onlyone",
			wantErr: true,
		},
		{
			input:   "ghcr.io",
			wantErr: true,
		},
		{
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFeatureRef(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFeatureRef(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFeatureRef(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseFeatureRef(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
