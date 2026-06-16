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
		Use:   "openlogs",
		Short: "OpenLogs CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.AddCommand(newStartCommand())

	return cmd
}

func newStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts emitting events",
		RunE: func(cmd *cobra.Command, args []string) error {
			StartAgent(context.Background())
			return nil
		},
	}
}
