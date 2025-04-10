package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bketelsen/feature/internal/options"
	"github.com/bketelsen/toolbox"
	"github.com/bketelsen/toolbox/cobra"
	"github.com/bketelsen/toolbox/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"go.uber.org/automaxprocs/maxprocs"

	goversion "github.com/bketelsen/toolbox/go-version"
	"github.com/spf13/viper"
)

var (
	// The global config object
	appname = "feature"

	version   = ""
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
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
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// ldflags
// Default: '-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=goreleaser'.
var bversion = buildVersion(version, commit, date, builtBy, treeState)

// NewRootCommand creates a new root command for the application
func NewRootCommand() (*cobra.Command, *viper.Viper) {
	config := setupConfig()

	// Define our command
	rootCmd := &cobra.Command{
		Use:   "feature [featurename]",
		Short: "Install devcontainer features",
		Args:  cobra.MinimumNArgs(1),

		Long: ui.Long("Install devcontainer features on the host system.",
			ui.Example{
				Description: "Install nodejs",
				Command:     "feature node",
			},
			ui.Example{
				Description: "Install Go",
				Command:     "feature go",
			},
		),
		Version: bversion.String(),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default slog logger to the cobra command
			slog.SetDefault(cmd.Logger)

			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			err := checkRootUser(cmd)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			featureRoot, _ := cmd.Flags().GetString("featureRoot")
			featureRoot = toolbox.ExpandPath(featureRoot)
			err = ensureRepo(cmd, featureRoot)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}

			cmd.Logger.Info("Feature root", "path", featureRoot)
			// check if the featureRoot exists
			if _, err := os.Stat(featureRoot); os.IsNotExist(err) {
				// create the featureRoot
				errmd := os.MkdirAll(featureRoot, 0o755)
				if errmd != nil {
					cmd.PrintErrln(errmd)
					os.Exit(1)
				}
			}

			opts, err := options.GetOptionsForFeature(featureRoot, args[0])
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			err = ensureCommonUtils(cmd, featureRoot)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			err = installFeature(cmd, featureRoot, args[0])
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			err = updateEnvironment(cmd, opts)
			if err != nil {
				cmd.PrintErrln(err)
				os.Exit(1)
			}
			ui.Success("Feature installed successfully.", "Restart your shell to apply the changes.")
		},
	}
	rootCmd.Flags().StringP("featureRoot", "r", "~/.features", "Location to checkout feature repository")
	_ = config.BindPFlag("featureRoot", rootCmd.PersistentFlags().Lookup("featureRoot"))
	rootCmd.Flags().BoolP("updateRepo", "u", false, "Update the feature repository")
	_ = config.BindPFlag("updateRepo", rootCmd.PersistentFlags().Lookup("updateRepo"))

	return rootCmd, config
}

// https://www.asciiart.eu/text-to-ascii-art to make your own
// just make sure the font doesn't have backticks in the letters or
// it will break the string quoting
var asciiName = `
███████╗███████╗ █████╗ ████████╗██╗   ██╗██████╗ ███████╗
██╔════╝██╔════╝██╔══██╗╚══██╔══╝██║   ██║██╔══██╗██╔════╝
█████╗  █████╗  ███████║   ██║   ██║   ██║██████╔╝█████╗
██╔══╝  ██╔══╝  ██╔══██║   ██║   ██║   ██║██╔══██╗██╔══╝
██║     ███████╗██║  ██║   ██║   ╚██████╔╝██║  ██║███████╗
╚═╝     ╚══════╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝
`

// buildVersion builds the version info for the application
func buildVersion(version, commit, date, builtBy, treeState string) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails(appname, "Install devcontainer features anywhere.", "https://bketelsen.github.io/feature"),
		goversion.WithASCIIName(asciiName),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if treeState != "" {
				i.GitTreeState = treeState
			}
			if date != "" {
				i.BuildDate = date
			}
			if version != "" {
				i.GitVersion = version
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)
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
