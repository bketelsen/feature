package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bketelsen/feature/internal/oci"
	"github.com/bketelsen/feature/internal/options"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	// The global config object
	appname = "feature"

	version = ""
	commit  = ""
)

// LogLevelIDs Maps 3rd party enumeration values to their textual representations
var LogLevelIDs = map[slog.Level][]string{
	slog.LevelDebug: {"debug"},
	slog.LevelInfo:  {"info"},
	slog.LevelWarn:  {"warn"},
	slog.LevelError: {"error"},
}

func main() {
	cmd, config := NewRootCommand()

	cmd.AddCommand(NewGendocsCommand(config))
	err := fang.Execute(
		context.Background(),
		cmd,
		fang.WithVersion(version),
		fang.WithCommit(commit),
	)
	if err != nil {
		os.Exit(1)
	}
}

// NewRootCommand creates a new root command for the application
func NewRootCommand() (*cobra.Command, *viper.Viper) {
	config := setupConfig()

	// Define our command
	rootCmd := &cobra.Command{
		Use:   "feature [featurename]",
		Short: "Install devcontainer features",
		Args:  cobra.MinimumNArgs(1),
		Long:  "Install devcontainer features on the host system.",
		Example: `  # Install nodejs (built-in)
  feature node

  # Install Go (built-in)
  feature go

  # Install a community feature from an OCI registry
  feature ghcr.io/devcontainers-extra/features/go-task:1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRootUser(cmd); err != nil {
				return err
			}
			featureRoot, _ := cmd.Flags().GetString("featureRoot")
			featureRoot = expandPath(featureRoot)

			slog.Info("Feature root", "path", featureRoot)

			var featureDir string
			var opts options.FeatureOptions

			if oci.IsOCIRef(args[0]) {
				// OCI community feature
				ref, err := oci.ParseFeatureRef(args[0])
				if err != nil {
					return err
				}
				updateRepo, _ := cmd.Flags().GetBool("updateRepo")
				cacheDir := filepath.Join(featureRoot+"-oci", ref.Registry, ref.Namespace, ref.Name, ref.Tag)
				featureDir, err = oci.PullFeature(cmd.Context(), ref, cacheDir, updateRepo)
				if err != nil {
					return err
				}
				opts, err = options.GetOptionsForPath(featureDir)
				if err != nil {
					return err
				}
			} else {
				// Built-in feature
				if err := ensureRepo(cmd, featureRoot); err != nil {
					return err
				}
				featureDir = filepath.Join(featureRoot, "src", args[0])
				var err error
				opts, err = options.GetOptionsForFeature(featureRoot, args[0])
				if err != nil {
					return err
				}
				if err := ensureCommonUtils(cmd, featureRoot); err != nil {
					return err
				}
			}

			if err := installFeature(cmd, featureDir); err != nil {
				return err
			}
			if err := updateEnvironment(cmd, opts); err != nil {
				return err
			}
			fmt.Println("Feature installed successfully. Restart your shell to apply the changes.")
			return nil
		},
	}
	rootCmd.Flags().StringP("featureRoot", "r", "~/.features", "Location to checkout feature repository")
	_ = config.BindPFlag("featureRoot", rootCmd.Flags().Lookup("featureRoot"))
	rootCmd.Flags().BoolP("updateRepo", "u", false, "Update the feature repository")
	_ = config.BindPFlag("updateRepo", rootCmd.Flags().Lookup("updateRepo"))

	return rootCmd, config
}

// expandPath expands a leading ~ to the user's home directory
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

func init() {
	// enable colored output on github actions et al
	if os.Getenv("NOCOLOR") != "" {
		lipgloss.DefaultRenderer().SetColorProfile(termenv.Ascii)
	}
	// automatically set GOMAXPROCS to match available CPUs.
	// GOMAXPROCS will be used as the default value for the --parallelism flag.
	if _, err := maxprocs.Set(); err != nil {
		fmt.Println("failed to set GOMAXPROCS")
	}
}
