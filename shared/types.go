package shared

import "time"

type EventBatch struct {
	ID     string    `json:"id"`
	Agent  Agent     `json:"agent"`
	Host   Host      `json:"host"`
	SentAt time.Time `json:"sent_at"`
	Events []Event   `json:"events"`
}

type Agent struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type Host struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type Event struct {
	ID         string       `json:"id"`
	ObservedAt time.Time    `json:"observed_at"`
	CPU        *CPUUsage    `json:"cpu,omitempty"`
	Memory     *MemoryUsage `json:"memory,omitempty"`
	Disk       *DiskUsage   `json:"disk,omitempty"`
}

type CPUUsage struct {
	UsedPercent    float64   `json:"used_percent"`
	CoresLogical   int       `json:"cores_logical"`
	CoresPhysical  int       `json:"cores_physical"`
	PerCorePercent []float64 `json:"per_core_percent"`
	Load1M         float64   `json:"load_1m"`
	Load5M         float64   `json:"load_5m"`
	Load15M        float64   `json:"load_15m"`
}

type MemoryUsage struct {
	UsedPercent    float64 `json:"used_percent"`
	UsedBytes      uint64  `json:"used_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	TotalBytes     uint64  `json:"total_bytes"`
}

type DiskUsage struct {
	Mount       string  `json:"mount"`
	Filesystem  string  `json:"filesystem"`
	UsedPercent float64 `json:"used_percent"`
	UsedBytes   uint64  `json:"used_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	TotalBytes  uint64  `json:"total_bytes"`
}

type Command struct{}
