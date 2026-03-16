package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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
		Example: `  # Install nodejs
  feature node

  # Install Go
  feature go`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRootUser(cmd); err != nil {
				return err
			}
			featureRoot, _ := cmd.Flags().GetString("featureRoot")
			featureRoot = expandPath(featureRoot)
			if err := ensureRepo(cmd, featureRoot); err != nil {
				return err
			}

			slog.Info("Feature root", "path", featureRoot)
			// check if the featureRoot exists
			if _, err := os.Stat(featureRoot); os.IsNotExist(err) {
				// create the featureRoot
				if err := os.MkdirAll(featureRoot, 0o755); err != nil {
					return err
				}
			}

			opts, err := options.GetOptionsForFeature(featureRoot, args[0])
			if err != nil {
				return err
			}
			if err := ensureCommonUtils(cmd, featureRoot); err != nil {
				return err
			}
			if err := installFeature(cmd, featureRoot, args[0]); err != nil {
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
	_ = config.BindPFlag("featureRoot", rootCmd.PersistentFlags().Lookup("featureRoot"))
	rootCmd.Flags().BoolP("updateRepo", "u", false, "Update the feature repository")
	_ = config.BindPFlag("updateRepo", rootCmd.PersistentFlags().Lookup("updateRepo"))

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
