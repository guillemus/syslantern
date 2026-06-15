package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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
	cmd.AddCommand(newUsageCommand(out))

	return cmd
}

func newUsageCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Print RAM and CPU usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			memory, err := mem.VirtualMemory()
			if err != nil {
				return err
			}

			cpuInfo, err := cpu.Info()
			if err != nil {
				return err
			}

			cpuUsage, err := cpu.Percent(500*time.Millisecond, true)
			if err != nil {
				return err
			}
			if len(cpuUsage) == 0 {
				return fmt.Errorf("cpu usage unavailable")
			}

			fmt.Fprintf(out, "RAM: %.1f%% (%s / %s)\n", memory.UsedPercent, formatBytes(memory.Used), formatBytes(memory.Total))
			for i, usage := range cpuUsage {
				fmt.Fprintf(out, "CPU %d: %.1f%%", i, usage)
				if i < len(cpuInfo) && cpuInfo[i].ModelName != "" {
					fmt.Fprintf(out, " %s", cpuInfo[i].ModelName)
				}
				fmt.Fprintln(out)
			}
			return nil
		},
	}
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes)
	for _, suffix := range []string{"KB", "MB", "GB", "TB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}

	return fmt.Sprintf("%.1f PB", value/unit)
}
