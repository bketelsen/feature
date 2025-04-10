/*
Copyright © 2025 Brian Ketelsen <bketelsen@gmail.com>
*/
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bketelsen/feature/internal/options"
	"github.com/bketelsen/toolbox/cobra"
)

func checkRootUser(_ *cobra.Command) error {
	if os.Geteuid() != 0 {
		return errors.New("this command must be run as root")
	}
	return nil
}

func ensureRepo(cmd *cobra.Command, featureRoot string) error {
	// check if the featureRoot exists
	if _, err := os.Stat(featureRoot); os.IsNotExist(err) {
		// create the featureRoot
		errmd := os.MkdirAll(featureRoot, 0o755)
		if errmd != nil {
			return errmd
		}
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
			_, _ = fmt.Fprintf(f, "export %s=%s\n", k, v)
		}
	}
	return nil
}
