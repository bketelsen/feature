/*
Copyright © 2025 Brian Ketelsen <bketelsen@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bketelsen/feature/internal/options"
	"github.com/bketelsen/toolbox"
	"github.com/bketelsen/toolbox/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "feature [featurename]",
	Short: "Install devcontainer features",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Install devcontainer features`,

	Run: func(cmd *cobra.Command, args []string) {
		err := checkRootUser(cmd)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		err = ensureRepo(cmd)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		featureRoot, _ := cmd.Flags().GetString("featureRoot")
		featureRoot = toolbox.ExpandPath(featureRoot)
		// check if the featureRoot exists
		if _, err := os.Stat(featureRoot); os.IsNotExist(err) {
			// create the featureRoot
			os.MkdirAll(featureRoot, 0755)
		}
		if strings.HasPrefix(featureRoot, "~/") {
			home, _ := os.UserHomeDir()
			featureRoot = filepath.Join(home, featureRoot[2:])
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
		cmd.Println("Feature installed successfully.\n\n Please restart your shell to apply the changes.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.feature.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("featureRoot", "r", "~/.features", "Location to checkout feature repository")
	rootCmd.Flags().BoolP("updateRepo", "u", false, "Update the feature repository")

}

func checkRootUser(_ *cobra.Command) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command must be run as root")
	}
	return nil
}
func ensureRepo(cmd *cobra.Command) error {
	featureRoot, _ := cmd.Flags().GetString("featureRoot")
	// check if the featureRoot exists
	if _, err := os.Stat(featureRoot); os.IsNotExist(err) {
		// create the featureRoot
		os.MkdirAll(featureRoot, 0755)
	}
	if strings.HasPrefix(featureRoot, "~/") {
		home, _ := os.UserHomeDir()
		featureRoot = filepath.Join(home, featureRoot[2:])
	}

	// checkout the repo to the featureRoot
	command := "git"
	args := []string{"clone", "https://github.com/devcontainers/features", featureRoot}
	if _, err := os.Stat(featureRoot + "/.git"); os.IsNotExist(err) {
		cerr := exec.Command(command, args...).Run()
		if cerr != nil {
			return cerr
		}
	}

	updateRepo, _ := cmd.Flags().GetBool("updateRepo")
	if updateRepo {
		args = []string{"pull"}
		cerr := exec.Command(command, args...).Run()
		if cerr != nil {
			return cerr
		}
	}
	return nil

}

func ensureCommonUtils(_ *cobra.Command, featureRoot string) error {
	marker := "/usr/local/etc/vscode-dev-containers/common"
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		fmt.Println("Installing common utils feature, please wait...")

		featureInstallPath := filepath.Join(featureRoot, "src", "common-utils", "install.sh")
		if _, err := os.Stat(featureInstallPath); os.IsNotExist(err) {
			return fmt.Errorf("feature %s does not have an install script", "common-utils")
		}
		command := "bash"
		args := []string{featureInstallPath}
		cerr := exec.Command(command, args...).Run()
		if cerr != nil {
			return cerr
		}
	}
	return nil

}
func installFeature(_ *cobra.Command, featureRoot, feature string) error {
	fmt.Println("Installing requested feature, please wait...")

	featureInstallPath := filepath.Join(featureRoot, "src", feature, "install.sh")
	if _, err := os.Stat(featureInstallPath); os.IsNotExist(err) {
		return fmt.Errorf("feature %s does not have an install script", feature)
	}
	command := "bash"
	args := []string{featureInstallPath}
	cerr := exec.Command(command, args...).Run()
	if cerr != nil {
		return cerr
	}
	return nil
}

func updateEnvironment(_ *cobra.Command, opts options.FeatureOptions) error {
	fmt.Println("Setting environment variables...")
	if opts.ContainerEnv == nil {
		return nil
	}

	if len(opts.ContainerEnv) > 0 {
		// create a file in /etc/profile.d to set the environment variables
		profile := fmt.Sprintf("/etc/profile.d/devcontainer-%s.sh", opts.ID)
		fmt.Println("Creating profile file", profile)
		f, err := os.Create(profile)
		if err != nil {
			return err
		}
		defer f.Close()
		for k, v := range opts.ContainerEnv {
			f.WriteString(fmt.Sprintf("export %s=%s\n", k, v))
		}
	}
	return nil
}
