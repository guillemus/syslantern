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
		RunE: func(cmd *cobra.Command, args []string) error {
			StartAgent(context.Background())
			return nil
		},
	}

	cmd.SetOut(out)
	cmd.SetErr(out)

	return cmd
}
