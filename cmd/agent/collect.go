package main

import (
	"app/shared"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

func collectBatch() (shared.EventBatch, error) {
	now := time.Now().UTC()
	hostInfo, err := host.Info()
	if err != nil {
		return shared.EventBatch{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return shared.EventBatch{}, err
	}

	batch := shared.EventBatch{
		ID: "batch_" + eventID(now, 0),
		Agent: shared.Agent{
			ID:      hostInfo.HostID,
			Version: "0.1.0",
		},
		Host: shared.Host{
			ID:   hostInfo.HostID,
			Name: hostname,
			OS:   hostInfo.OS,
			Arch: hostInfo.KernelArch,
		},
		SentAt: now,
	}

	metrics, err := collectMetrics(now)
	if err != nil {
		return shared.EventBatch{}, err
	}
	batch.Metrics = metrics

	return batch, nil
}

func collectMetrics(now time.Time) (shared.MetricsSnapshot, error) {
	cpuUsage, err := createCPUUsage()
	if err != nil {
		return shared.MetricsSnapshot{}, err
	}

	virtualMemory, err := createVirtualMemoryUsage()
	if err != nil {
		return shared.MetricsSnapshot{}, err
	}

	swapMemory, err := createSwapMemoryUsage()
	if err != nil {
		return shared.MetricsSnapshot{}, err
	}

	diskMetrics, err := createDiskMetrics()
	if err != nil {
		return shared.MetricsSnapshot{}, err
	}

	return shared.MetricsSnapshot{
		ObservedAt:    now,
		CPU:           cpuUsage,
		VirtualMemory: virtualMemory,
		SwapMemory:    swapMemory,
		Disk:          diskMetrics,
	}, nil
}

func createCPUUsage() (shared.CPUUsage, error) {
	perCoreUsage, err := cpu.Percent(0, true)
	if err != nil {
		return shared.CPUUsage{}, err
	}
	logicalCores, err := cpu.Counts(true)
	if err != nil {
		return shared.CPUUsage{}, err
	}
	physicalCores, err := cpu.Counts(false)
	if err != nil {
		return shared.CPUUsage{}, err
	}
	loadAvg, err := load.Avg()
	if err != nil {
		return shared.CPUUsage{}, err
	}

	var totalUsage float64
	for _, usage := range perCoreUsage {
		totalUsage += usage
	}
	overallUsage := totalUsage / float64(len(perCoreUsage))

	return shared.CPUUsage{
		UsedPercent:    overallUsage,
		CoresLogical:   logicalCores,
		CoresPhysical:  physicalCores,
		PerCorePercent: perCoreUsage,
		Load1M:         loadAvg.Load1,
		Load5M:         loadAvg.Load5,
		Load15M:        loadAvg.Load15,
	}, nil
}

func createVirtualMemoryUsage() (shared.MemoryUsage, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return shared.MemoryUsage{}, err
	}

	return shared.MemoryUsage{
		UsedPercent:    memory.UsedPercent,
		UsedBytes:      memory.Used,
		AvailableBytes: memory.Available,
		TotalBytes:     memory.Total,
	}, nil
}

func createSwapMemoryUsage() (shared.MemoryUsage, error) {
	swap, err := mem.SwapMemory()
	if err != nil {
		return shared.MemoryUsage{}, err
	}

	return shared.MemoryUsage{
		UsedPercent:    swap.UsedPercent,
		UsedBytes:      swap.Used,
		AvailableBytes: swap.Free,
		TotalBytes:     swap.Total,
	}, nil
}

func createDiskMetrics() (shared.DiskMetrics, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return shared.DiskMetrics{}, err
	}

	diskUsages := make([]shared.DiskUsage, 0, len(partitions))
	for _, partition := range partitions {
		if skipDiskPartition(partition) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return shared.DiskMetrics{}, err
		}

		diskUsages = append(diskUsages, shared.DiskUsage{
			Device:      partition.Device,
			Mount:       usage.Path,
			Filesystem:  partition.Fstype,
			UsedPercent: usage.UsedPercent,
			UsedBytes:   usage.Used,
			FreeBytes:   usage.Free,
			TotalBytes:  usage.Total,
		})
	}

	return shared.DiskMetrics{
		Total:      createTotalDiskUsage(diskUsages),
		Partitions: diskUsages,
	}, nil
}

var (
	skippedFilesystems = []string{
		"autofs",
		"cgroup",
		"cgroup2",
		"debugfs",
		"devfs",
		"devpts",
		"devtmpfs",
		"fusectl",
		"mqueue",
		"overlay",
		"proc",
		"pstore",
		"ramfs",
		"securityfs",
		"squashfs",
		"sysfs",
		"tmpfs",
		"tracefs",
	}
	skippedMountPrefixes = []string{
		"/dev",
		"/proc",
		"/run",
		"/snap",
		"/sys",
		"/var/lib/docker/overlay2",
	}
)

func skipDiskPartition(partition disk.PartitionStat) bool {

	filesystem := strings.ToLower(partition.Fstype)
	if slices.Contains(skippedFilesystems, filesystem) {
		return true
	}
	for _, prefix := range skippedMountPrefixes {
		if partition.Mountpoint == prefix || strings.HasPrefix(partition.Mountpoint, prefix+"/") {
			return true
		}
	}
	return false
}

func createTotalDiskUsage(partitions []shared.DiskUsage) shared.DiskUsage {
	var total shared.DiskUsage
	total.Mount = "__total__"
	total.Filesystem = "mounted"
	seenDevices := make(map[string]bool, len(partitions))
	for _, partition := range partitions {
		device := partition.Device
		if device == "" {
			device = partition.Mount
		}
		// Bind mounts and filesystem submounts can report the same backing device
		// multiple times; count each device once so the aggregate stays usable-disk total.
		if seenDevices[device] {
			continue
		}
		seenDevices[device] = true

		total.UsedBytes += partition.UsedBytes
		total.FreeBytes += partition.FreeBytes
		total.TotalBytes += partition.TotalBytes
	}
	if total.TotalBytes > 0 {
		total.UsedPercent = float64(total.UsedBytes) / float64(total.TotalBytes) * 100
	}
	return total
}

func eventID(t time.Time, sequence int) string {
	return fmt.Sprintf("%d_%d", t.UnixNano(), sequence)
}
