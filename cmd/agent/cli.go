package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCommand(os.Stdout).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand(out io.Writer) *cobra.Command {
	var cmd cobra.Command
	cmd.Use = "syslantern"
	cmd.Short = "Syslantern agent"
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		StartAgent(context.Background())
		return nil
	}

	var versionCmd cobra.Command
	versionCmd.Use = "version"
	versionCmd.Short = "Print the agent version"
	versionCmd.Args = cobra.NoArgs
	versionCmd.RunE = func(_ *cobra.Command, _ []string) error {
		_, err := fmt.Fprintln(out, version)
		return err
	}
	cmd.AddCommand(&versionCmd)

	var setCmd cobra.Command
	setCmd.Use = "set"
	setCmd.Short = "Set agent configuration values"

	var setAPIKeyCmd cobra.Command
	setAPIKeyCmd.Use = "apikey <key>"
	setAPIKeyCmd.Short = "Set the agent API key"
	setAPIKeyCmd.Args = cobra.ExactArgs(1)
	setAPIKeyCmd.RunE = func(_ *cobra.Command, args []string) error {
		if err := SetAPIKey(args[0]); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "Config saved.")
		return err
	}
	setCmd.AddCommand(&setAPIKeyCmd)
	cmd.AddCommand(&setCmd)

	cmd.SetOut(out)
	cmd.SetErr(out)

	return &cmd
}
