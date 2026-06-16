package server

import (
	"app/shared"
	"app/validate"
	"app/views"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleBatch(w http.ResponseWriter, r *http.Request) {
	var payload shared.EventBatch

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Warn("receive stats: parse request", "err", err)
		http.Error(w, "Send an event batch as JSON.", http.StatusBadRequest)
		return
	}

	if err := s.BatchBus.Emit(r.Context(), payload); err != nil {
		s.Logger.Warn("receive stats: emit batch", "err", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderDashboard(w)
}

func (s *Server) HandleDashboardEvents(w http.ResponseWriter, r *http.Request) {
	events := make(chan shared.EventBatch, 16)
	cancel := s.BatchBus.Subscribe(r.Context(), func(evt shared.EventBatch) error {
		select {
		case events <- evt:
		default:
		}
		return nil
	})
	defer cancel()

	sse := datastar.NewSSE(w, r)
	for {
		select {
		case <-r.Context().Done():
			return
		case batch := <-events:
			html, err := s.Renderer.RenderDashboardStatsHTML(dashboardStats(batch))
			if err != nil {
				s.Logger.Warn("dashboard events: render stats", "err", err)
				continue
			}
			if err := sse.PatchElements(html); err != nil {
				s.Logger.Warn("dashboard events: patch stats", "err", err)
				return
			}
		}
	}
}

func dashboardStats(batch shared.EventBatch) views.DashboardStatsData {
	stats := views.DashboardStatsData{
		HostName:   batch.Host.Name,
		HostDetail: fmt.Sprintf("%s %s", batch.Host.OS, batch.Host.Arch),
		UpdatedAt:  fmt.Sprintf("updated %s", batch.SentAt.Format(time.RFC3339)),
		Disks:      []views.DashboardDiskData{},
	}

	var cpuTotal float64
	var cpuCount int
	var cpuCores uint64

	for _, event := range batch.Events {
		switch event.Payload.Name {
		case "cpu.usage":
			cpuTotal += event.Payload.Value
			cpuCount++
			if cpuCores == 0 {
				cpuCores = uintField(event.Payload.Fields, "cores")
			}
		case "memory.usage":
			stats.MemoryAvailable = formatBytes(uintField(event.Payload.Fields, "available_bytes"))
			stats.MemoryDescription = fmt.Sprintf("%s used of %s", formatBytes(uintField(event.Payload.Fields, "used_bytes")), formatBytes(uintField(event.Payload.Fields, "total_bytes")))
		case "disk.usage":
			used := event.Payload.Value
			freePercent := 100 - used
			freeBytes := uintField(event.Payload.Fields, "available_bytes")
			totalBytes := uintField(event.Payload.Fields, "total_bytes")
			stats.Disks = append(stats.Disks, views.DashboardDiskData{
				Mount:            stringField(event.Payload.Fields, "mount"),
				Free:             formatBytes(freeBytes),
				FreeBytes:        freeBytes,
				FreePercent:      percent(freePercent),
				FreePercentValue: freePercent,
				Used:             percent(used),
				Total:            formatBytes(totalBytes),
				Meaning:          diskMeaning(used),
				StatusClass:      diskStatusClass(used),
			})
		}
	}

	if cpuCount > 0 {
		cpuUsage := cpuTotal / float64(cpuCount)
		stats.CPUHeadroom = percent(100 - cpuUsage)
		stats.CPUDescription = fmt.Sprintf("%s currently used across %d cores", percent(cpuUsage), cpuCores)
	}

	sort.Slice(stats.Disks, func(i, j int) bool {
		return stats.Disks[i].FreePercentValue < stats.Disks[j].FreePercentValue
	})

	if len(stats.Disks) > 0 {
		lowest := stats.Disks[0]
		stats.LowestDiskFree = lowest.FreePercent
		stats.LowestDiskDesc = fmt.Sprintf("%s has %s free", lowest.Mount, lowest.Free)
	}

	return stats
}

func percent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	value := float64(bytes)
	for _, suffix := range []string{"KB", "MB", "GB", "TB", "PB"} {
		value = value / unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f EB", value/unit)
}

func diskMeaning(used float64) string {
	if used >= 90 {
		return "Near full"
	}
	if used >= 80 {
		return "High usage, not full yet"
	}
	return "Healthy"
}

func diskStatusClass(used float64) string {
	if used >= 90 {
		return "text-red-300"
	}
	if used >= 80 {
		return "text-amber-300"
	}
	return "text-emerald-300"
}

func stringField(fields map[string]any, key string) string {
	value, ok := fields[key]
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}

func uintField(fields map[string]any, key string) uint64 {
	value, ok := fields[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case uint64:
		return typed
	case uint:
		return uint64(typed)
	case int:
		return uint64(typed)
	case int64:
		return uint64(typed)
	case float64:
		return uint64(typed)
	default:
		return 0
	}
}
