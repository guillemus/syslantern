package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"syslantern/db"
	"syslantern/views"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleAgentsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetAuthenticatedUser(r)
	agentID := chi.URLParam(r, "agentID")

	agent, ok := s.getVisibleAgent(w, r, user.TeamID, agentID)
	if !ok {
		return
	}

	metrics, err := s.agentMetricsData(ctx)
	if err != nil {
		s.Logger.Warn("agent page: load metrics", "team_id", user.TeamID, "agent_id", agentID, "err", err)
		http.Error(w, "Could not load this agent.", http.StatusInternalServerError)
		return
	}

	data := s.agentPageData(r, agent)
	data.Metrics = metrics
	s.Renderer.RenderAgentPage(w, data)
}

func (s *Server) getVisibleAgent(w http.ResponseWriter, r *http.Request, teamID int64, agentID string) (db.Agent, bool) {
	agent, err := s.DB.GetAgent(r.Context(), db.GetAgentParams{
		ID:     agentID,
		TeamID: teamID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return db.Agent{}, false
	} else if err != nil {
		s.Logger.Warn("agent page: get agent", "team_id", teamID, "agent_id", agentID, "err", err)
		http.Error(w, "Could not load this agent.", http.StatusInternalServerError)
		return db.Agent{}, false
	}

	if agent.Status == db.AgentStatusDeleted {
		http.NotFound(w, r)
		return db.Agent{}, false
	}
	return agent, true
}

func (s *Server) agentPageData(r *http.Request, agent db.Agent) views.AgentPageData {
	hostID := ""
	if agent.HostID.Valid {
		hostID = agent.HostID.String
	}

	return views.AgentPageData{
		ID:             agent.ID,
		Status:         string(agent.Status),
		Name:           agent.Name,
		Version:        agent.Version,
		HostID:         hostID,
		CreatedAt:      agent.CreatedAt,
		UpdatedAt:      agent.UpdatedAt,
		InstallCommand: s.agentInstallCommand(r, agent.ApiKey),
	}
}

func (s *Server) HandleAgentsEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetAuthenticatedUser(r)
	agentID := chi.URLParam(r, "agentID")
	snapshotReceivedC := s.BusSnapshotProcessed.Subscribe(ctx)

	sse := datastar.NewSSE(w, r)
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-snapshotReceivedC:
			if evt.TeamID != user.TeamID || evt.AgentID != agentID {
				continue
			}
			if evt.Type != SnapshotProcessedTypeMetrics {
				continue
			}

			data, err := s.agentMetricsData(ctx)
			if err != nil {
				s.Logger.Warn("agent events: load metrics", "team_id", user.TeamID, "agent_id", agentID, "err", err)
				return
			}

			s.Renderer.PatchAgentMetrics(sse, data)
		}
	}
}

func (s *Server) agentMetricsData(ctx context.Context) (views.DashboardMetricsData, error) {
	latestCPU, cpuExists, err := s.latestCPUSample(ctx)
	if err != nil {
		return views.DashboardMetricsData{}, err
	}
	latestMemory, memoryExists, err := s.latestMemorySample(ctx)
	if err != nil {
		return views.DashboardMetricsData{}, err
	}
	latestDisks, err := s.DB.ListLatestDiskSamples(ctx)
	if err != nil {
		return views.DashboardMetricsData{}, err
	}

	latestTimes := []string{latestCPU.ObservedAt, latestMemory.ObservedAt}
	latestTimes = append(latestTimes, latestDiskSampleTimes(latestDisks)...)
	latest := latestSampleTime(latestTimes...)
	since := latest.Add(-6 * time.Hour)
	if latest.IsZero() {
		since = time.Now().Add(-6 * time.Hour)
	}
	sinceValue := since.Format(time.RFC3339Nano)

	cpuSamples, err := s.DB.ListCPUSamplesSince(ctx, sinceValue)
	if err != nil {
		return views.DashboardMetricsData{}, err
	}
	memorySamples, err := s.DB.ListMemorySamplesSince(ctx, sinceValue)
	if err != nil {
		return views.DashboardMetricsData{}, err
	}
	diskSamples, err := s.DB.ListDiskSamplesSince(ctx, sinceValue)
	if err != nil {
		return views.DashboardMetricsData{}, err
	}

	return views.DashboardMetricsData{
		Stats: views.DashboardStatsData{
			HasMetrics:           cpuExists || memoryExists || len(latestDisks) > 0,
			CPUUsedPercent:       latestCPU.UsedPercent,
			CPUCoresLogical:      int(latestCPU.CoresLogical),
			MemoryUsedBytes:      uint64(latestMemory.VirtualUsedBytes),
			MemoryAvailableBytes: uint64(latestMemory.VirtualAvailableBytes),
			MemoryTotalBytes:     uint64(latestMemory.VirtualTotalBytes),
			Disks:                dashboardDisks(latestDisks),
		},
		Analytics: views.DashboardAnalyticsData{
			HasAnalytics: len(cpuSamples) > 0 || len(memorySamples) > 0 || len(diskSamples) > 0,
			Since:        since,
			CPU:          dashboardCPUHistory(cpuSamples),
			Memory:       dashboardMemoryHistory(memorySamples),
			Disks:        dashboardDiskHistory(diskSamples),
		},
	}, nil
}

func (s *Server) latestCPUSample(ctx context.Context) (db.CpuSample, bool, error) {
	sample, err := s.DB.GetLatestCPUSample(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return db.CpuSample{}, false, nil
	}
	return sample, true, err
}

func (s *Server) latestMemorySample(ctx context.Context) (db.MemorySample, bool, error) {
	sample, err := s.DB.GetLatestMemorySample(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return db.MemorySample{}, false, nil
	}
	return sample, true, err
}

func dashboardCPUHistory(samples []db.CpuSample) []views.DashboardCPUHistoryData {
	points := make([]views.DashboardCPUHistoryData, 0, len(samples))
	for _, sample := range samples {
		var perCore []float64
		if err := json.Unmarshal([]byte(sample.PerCorePercent), &perCore); err != nil {
			perCore = nil
		}
		points = append(points, views.DashboardCPUHistoryData{
			ObservedAt:     parseSampleTime(sample.ObservedAt),
			UsedPercent:    sample.UsedPercent,
			CoresLogical:   int(sample.CoresLogical),
			CoresPhysical:  int(sample.CoresPhysical),
			PerCorePercent: perCore,
			Load1M:         sample.Load1m,
			Load5M:         sample.Load5m,
			Load15M:        sample.Load15m,
		})
	}
	return points
}

func dashboardMemoryHistory(samples []db.MemorySample) []views.DashboardMemoryHistoryData {
	points := make([]views.DashboardMemoryHistoryData, 0, len(samples))
	for _, sample := range samples {
		points = append(points, views.DashboardMemoryHistoryData{
			ObservedAt:            parseSampleTime(sample.ObservedAt),
			VirtualUsedPercent:    sample.VirtualUsedPercent,
			VirtualUsedBytes:      uint64(sample.VirtualUsedBytes),
			VirtualAvailableBytes: uint64(sample.VirtualAvailableBytes),
			VirtualTotalBytes:     uint64(sample.VirtualTotalBytes),
			SwapUsedPercent:       sample.SwapUsedPercent,
			SwapUsedBytes:         uint64(sample.SwapUsedBytes),
			SwapAvailableBytes:    uint64(sample.SwapAvailableBytes),
			SwapTotalBytes:        uint64(sample.SwapTotalBytes),
		})
	}
	return points
}

func dashboardDiskHistory(samples []db.DiskSample) []views.DashboardDiskHistoryData {
	points := make([]views.DashboardDiskHistoryData, 0, len(samples))
	for _, sample := range samples {
		points = append(points, dashboardDiskHistoryPoint(sample))
	}
	return points
}

func dashboardDisks(samples []db.DiskSample) []views.DashboardDiskData {
	disks := make([]views.DashboardDiskData, 0, len(samples))
	for _, sample := range samples {
		if sample.IsTotal == 1 {
			continue
		}
		disks = append(disks, views.DashboardDiskData{
			Mount:       sample.Mount,
			FreeBytes:   uint64(sample.FreeBytes),
			UsedPercent: sample.UsedPercent,
			TotalBytes:  uint64(sample.TotalBytes),
		})
	}
	return disks
}

func dashboardDiskHistoryPoint(sample db.DiskSample) views.DashboardDiskHistoryData {
	return views.DashboardDiskHistoryData{
		ObservedAt:  parseSampleTime(sample.ObservedAt),
		IsTotal:     sample.IsTotal == 1,
		Mount:       sample.Mount,
		Device:      sample.Device,
		Filesystem:  sample.Filesystem,
		UsedPercent: sample.UsedPercent,
		UsedBytes:   uint64(sample.UsedBytes),
		FreeBytes:   uint64(sample.FreeBytes),
		TotalBytes:  uint64(sample.TotalBytes),
	}
}

func latestDiskSampleTimes(samples []db.DiskSample) []string {
	values := make([]string, 0, len(samples))
	for _, sample := range samples {
		values = append(values, sample.ObservedAt)
	}
	return values
}

func latestSampleTime(values ...string) time.Time {
	var latest time.Time
	for _, value := range values {
		parsed := parseSampleTime(value)
		if parsed.After(latest) {
			latest = parsed
		}
	}
	return latest
}

func parseSampleTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
