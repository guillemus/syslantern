package main

import (
	"app/shared"
	"fmt"
	"os"
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
		Events: nil,
	}

	events, err := collectEvents(now)
	if err != nil {
		return shared.EventBatch{}, err
	}
	batch.Events = events

	return batch, nil
}

func collectEvents(now time.Time) ([]shared.Event, error) {
	events := make([]shared.Event, 0, 10)
	nextID := 1

	cpuEvent, nextID, err := createCPUEvent(now, nextID)
	if err != nil {
		return nil, err
	}
	events = append(events, cpuEvent)

	memoryEvent, nextID, err := createMemoryEvent(now, nextID)
	if err != nil {
		return nil, err
	}
	events = append(events, memoryEvent)

	diskEvents, nextID, err := createDiskEvents(now, nextID)
	if err != nil {
		return nil, err
	}
	events = append(events, diskEvents...)

	return events, nil
}

func createCPUEvent(now time.Time, sequence int) (shared.Event, int, error) {
	perCoreUsage, err := cpu.Percent(0, true)
	if err != nil {
		return shared.Event{}, sequence, err
	}
	logicalCores, err := cpu.Counts(true)
	if err != nil {
		return shared.Event{}, sequence, err
	}
	physicalCores, err := cpu.Counts(false)
	if err != nil {
		return shared.Event{}, sequence, err
	}
	loadAvg, err := load.Avg()
	if err != nil {
		return shared.Event{}, sequence, err
	}

	var totalUsage float64
	for _, usage := range perCoreUsage {
		totalUsage += usage
	}
	overallUsage := totalUsage / float64(len(perCoreUsage))

	return shared.Event{
		ID:         "evt_" + eventID(now, sequence),
		ObservedAt: now,
		CPU: &shared.CPUUsage{
			UsedPercent:    overallUsage,
			CoresLogical:   logicalCores,
			CoresPhysical:  physicalCores,
			PerCorePercent: perCoreUsage,
			Load1M:         loadAvg.Load1,
			Load5M:         loadAvg.Load5,
			Load15M:        loadAvg.Load15,
		},
	}, sequence + 1, nil
}

func createMemoryEvent(now time.Time, sequence int) (shared.Event, int, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return shared.Event{}, sequence, err
	}

	return shared.Event{
		ID:         "evt_" + eventID(now, sequence),
		ObservedAt: now,
		Memory: &shared.MemoryUsage{
			UsedPercent:    memory.UsedPercent,
			UsedBytes:      memory.Used,
			AvailableBytes: memory.Available,
			TotalBytes:     memory.Total,
		},
	}, sequence + 1, nil
}

func createDiskEvents(now time.Time, sequence int) ([]shared.Event, int, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, sequence, err
	}

	events := make([]shared.Event, 0, len(partitions))
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return nil, sequence, err
		}
		events = append(events, shared.Event{
			ID:         "evt_" + eventID(now, sequence),
			ObservedAt: now,
			Disk: &shared.DiskUsage{
				Mount:       usage.Path,
				Filesystem:  partition.Fstype,
				UsedPercent: usage.UsedPercent,
				UsedBytes:   usage.Used,
				FreeBytes:   usage.Free,
				TotalBytes:  usage.Total,
			},
		})
		sequence++
	}

	return events, sequence, nil
}

func eventID(t time.Time, sequence int) string {
	return fmt.Sprintf("%d_%d", t.UnixNano(), sequence)
}
