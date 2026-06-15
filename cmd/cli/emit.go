package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

type EventBatch struct {
	BatchID string       `json:"batch_id"`
	Agent   BatchAgent   `json:"agent"`
	Host    BatchHost    `json:"host"`
	SentAt  time.Time    `json:"sent_at"`
	Events  []BatchEvent `json:"events"`
}

type BatchAgent struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type BatchHost struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type BatchEvent struct {
	ID         string        `json:"id"`
	ObservedAt time.Time     `json:"observed_at"`
	Type       string        `json:"type"`
	Source     string        `json:"source"`
	Payload    MetricPayload `json:"payload"`
}

type MetricPayload struct {
	Name   string         `json:"name"`
	Value  float64        `json:"value"`
	Unit   string         `json:"unit"`
	Fields map[string]any `json:"fields"`
}

func StartEmitter() {
	for {
		time.Sleep(1 * time.Second)
		batch, err := collectBatch()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
}

func collectBatch() (EventBatch, error) {
	now := time.Now().UTC()
	hostInfo, err := host.Info()
	if err != nil {
		return EventBatch{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return EventBatch{}, err
	}

	batch := EventBatch{
		BatchID: "batch_" + eventID(now, 0),
		Agent: BatchAgent{
			ID:      hostInfo.HostID,
			Version: "0.1.0",
		},
		Host: BatchHost{
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
		return EventBatch{}, err
	}
	batch.Events = events

	return batch, nil
}

func collectEvents(now time.Time) ([]BatchEvent, error) {
	events := make([]BatchEvent, 0, 10)
	nextID := 1

	cpuUsage, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}
	cpuCores, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}
	loadAvg, err := load.Avg()
	if err != nil {
		return nil, err
	}

	for core, usage := range cpuUsage {
		events = append(events, newCPUEvent(now, nextID, core, usage, cpuCores, loadAvg))
		nextID++
	}

	memory, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	events = append(events, newMemoryEvent(now, nextID, memory))
	nextID++

	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return nil, err
		}
		events = append(events, newDiskEvent(now, nextID, partition, usage))
		nextID++
	}

	return events, nil
}

func newCPUEvent(now time.Time, sequence int, core int, usage float64, cores int, loadAvg *load.AvgStat) BatchEvent {
	return BatchEvent{
		ID:         "evt_" + eventID(now, sequence),
		ObservedAt: now,
		Type:       "metric",
		Source:     "system.cpu",
		Payload: MetricPayload{
			Name:  "cpu.usage",
			Value: usage,
			Unit:  "percent",
			Fields: map[string]any{
				"core":     core,
				"cores":    cores,
				"load_1m":  loadAvg.Load1,
				"load_5m":  loadAvg.Load5,
				"load_15m": loadAvg.Load15,
			},
		},
	}
}

func newMemoryEvent(now time.Time, sequence int, memory *mem.VirtualMemoryStat) BatchEvent {
	return BatchEvent{
		ID:         "evt_" + eventID(now, sequence),
		ObservedAt: now,
		Type:       "metric",
		Source:     "system.memory",
		Payload: MetricPayload{
			Name:  "memory.usage",
			Value: memory.UsedPercent,
			Unit:  "percent",
			Fields: map[string]any{
				"used_bytes":      memory.Used,
				"available_bytes": memory.Available,
				"total_bytes":     memory.Total,
			},
		},
	}
}

func newDiskEvent(now time.Time, sequence int, partition disk.PartitionStat, usage *disk.UsageStat) BatchEvent {
	return BatchEvent{
		ID:         "evt_" + eventID(now, sequence),
		ObservedAt: now,
		Type:       "metric",
		Source:     "system.disk",
		Payload: MetricPayload{
			Name:  "disk.usage",
			Value: usage.UsedPercent,
			Unit:  "percent",
			Fields: map[string]any{
				"mount":           usage.Path,
				"device":          partition.Device,
				"filesystem":      partition.Fstype,
				"used_bytes":      usage.Used,
				"available_bytes": usage.Free,
				"total_bytes":     usage.Total,
			},
		},
	}
}

func eventID(t time.Time, sequence int) string {
	return fmt.Sprintf("%d_%d", t.UnixNano(), sequence)
}
