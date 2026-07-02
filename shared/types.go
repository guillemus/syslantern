package shared

import "time"

type IngestEvent struct {
	LiveSnapshot *LiveSnapshot `json:"live_snapshot,omitempty"`
	Logs         []LogEvent    `json:"logs,omitempty"`
}

type AgentStatus string

const (
	AgentStatusCreated  AgentStatus = "created"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusDeleted  AgentStatus = "deleted"
	AgentStatusPaused   AgentStatus = "paused"
	AgentStatusResuming AgentStatus = "resuming"
)

func (s AgentStatus) ShouldAgentPoll() bool {
	switch s {
	case AgentStatusPaused:
		return true
	case AgentStatusCreated, AgentStatusRunning, AgentStatusDeleted, AgentStatusResuming:
		return false
	}
	return false
}

func (s AgentStatus) ShouldAgentSendMetrics() bool {
	switch s {
	case AgentStatusCreated, AgentStatusResuming, AgentStatusRunning:
		return true
	case AgentStatusDeleted, AgentStatusPaused:
		return false
	}
	return false
}

type IngestResult struct {
	AgentStatus AgentStatus `json:"agent_status"`
}

type AgentConfig struct {
	AgentStatus AgentStatus `json:"agent_status"`
}

type LiveSnapshot struct {
	ID      string          `json:"id"`
	Agent   Agent           `json:"agent"`
	Host    Host            `json:"host"`
	SentAt  time.Time       `json:"sent_at"`
	Metrics MetricsSnapshot `json:"metrics"`
}

type Agent struct {
	Version string `json:"version"`
}

type Host struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type MetricsSnapshot struct {
	ObservedAt    time.Time   `json:"observed_at"`
	CPU           CPUUsage    `json:"cpu"`
	VirtualMemory MemoryUsage `json:"virtual_memory"`
	SwapMemory    MemoryUsage `json:"swap_memory"`
	Disk          DiskMetrics `json:"disk"`
}

type LogEvent struct {
	ID         string            `json:"id"`
	Host       Host              `json:"host"`
	SentAt     time.Time         `json:"sent_at"`
	ObservedAt time.Time         `json:"observed_at"`
	Source     string            `json:"source"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Message    string            `json:"message"`
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

type DiskMetrics struct {
	Total      DiskUsage   `json:"total"`
	Partitions []DiskUsage `json:"partitions"`
}

type DiskUsage struct {
	Device      string  `json:"device"`
	Mount       string  `json:"mount"`
	Filesystem  string  `json:"filesystem"`
	UsedPercent float64 `json:"used_percent"`
	UsedBytes   uint64  `json:"used_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	TotalBytes  uint64  `json:"total_bytes"`
}

type CPUAnalyticsSample struct {
	ObservedAt time.Time `json:"observed_at"`
	CPU        CPUUsage  `json:"cpu"`
}

type MemoryAnalyticsSample struct {
	ObservedAt    time.Time   `json:"observed_at"`
	VirtualMemory MemoryUsage `json:"virtual_memory"`
	SwapMemory    MemoryUsage `json:"swap_memory"`
}

type DiskAnalyticsSample struct {
	ObservedAt time.Time `json:"observed_at"`
	IsTotal    bool      `json:"is_total"`
	Disk       DiskUsage `json:"disk"`
}

type AgentCommand struct {
	AgentID string  `json:"agent_id"`
	Command Command `json:"command"`
}

type Command struct {
	AnalyticsSnapshot *AnalyticsSnapshotCommand `json:"analytics_snapshot,omitempty"`
}

type AnalyticsSnapshotCommand struct {
	Since time.Time `json:"since"`
}
