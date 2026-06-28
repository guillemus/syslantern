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
	cmd := &cobra.Command{
		Use:   "syslantern",
		Short: "Syslantern agent",
		RunE: func(_ *cobra.Command, _ []string) error {
			StartAgent(context.Background())
			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the agent version",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(out, version)
			return err
		},
	})

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set agent configuration values",
	}
	setCmd.AddCommand(&cobra.Command{
		Use:   "apikey <key>",
		Short: "Set the agent API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := SetAPIKey(args[0]); err != nil {
				return err
			}
			_, err := fmt.Fprintln(out, "Config saved.")
			return err
		},
	})
	cmd.AddCommand(setCmd)

	cmd.SetOut(out)
	cmd.SetErr(out)

	return cmd
}
