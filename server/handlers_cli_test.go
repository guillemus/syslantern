package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"syslantern/db"
	"syslantern/shared"
)

func TestHandleIngest(t *testing.T) {
	t.Run("rejects missing api key", func(t *testing.T) {
		s := newTestServer(t)

		rr := sendIngest(t, s, "", mustJSON(t, testLiveSnapshot()))

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		require.Equal(t, 0, countRows(t, s, "cpu_samples"))
	})

	t.Run("rejects invalid api key", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, "agent-a", "api-key-a", db.AgentStatusCreated)

		rr := sendIngest(t, s, "wrong-api-key", mustJSON(t, testLiveSnapshot()))

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		require.Equal(t, 0, countRows(t, s, "cpu_samples"))
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, "agent-a", "api-key-a", db.AgentStatusCreated)

		rr := sendIngest(t, s, "api-key-a", `{`)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		require.Equal(t, 0, countRows(t, s, "cpu_samples"))
	})

	t.Run("created agent saves snapshot and becomes running", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, "agent-a", "api-key-a", db.AgentStatusCreated)
		events := s.BusSnapshotProcessed.Subscribe(t.Context())

		rr := sendIngest(t, s, "api-key-a", mustJSON(t, testLiveSnapshot()))

		require.Equal(t, http.StatusOK, rr.Code)
		assertIngestResult(t, rr, shared.AgentStatusRunning)
		assertAgentStatus(t, s, "api-key-a", db.AgentStatusRunning)
		assertSnapshotSaved(t, s)
		assertSnapshotProcessedEvent(t, events, "agent-a")
	})

	t.Run("running agent saves snapshot and stays running", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, "agent-a", "api-key-a", db.AgentStatusRunning)

		rr := sendIngest(t, s, "api-key-a", mustJSON(t, testLiveSnapshot()))

		require.Equal(t, http.StatusOK, rr.Code)
		assertIngestResult(t, rr, shared.AgentStatusRunning)
		assertAgentStatus(t, s, "api-key-a", db.AgentStatusRunning)
		assertSnapshotSaved(t, s)
	})

	t.Run("resuming agent saves snapshot and becomes running", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, "agent-a", "api-key-a", db.AgentStatusResuming)
		events := s.BusSnapshotProcessed.Subscribe(t.Context())

		rr := sendIngest(t, s, "api-key-a", mustJSON(t, testLiveSnapshot()))

		require.Equal(t, http.StatusOK, rr.Code)
		assertIngestResult(t, rr, shared.AgentStatusRunning)
		assertAgentStatus(t, s, "api-key-a", db.AgentStatusRunning)
		assertSnapshotSaved(t, s)
		assertSnapshotProcessedEvent(t, events, "agent-a")
	})

	t.Run("paused agent does not save snapshot", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, "agent-a", "api-key-a", db.AgentStatusPaused)

		rr := sendIngest(t, s, "api-key-a", mustJSON(t, testLiveSnapshot()))

		require.Equal(t, http.StatusOK, rr.Code)
		assertIngestResult(t, rr, shared.AgentStatusPaused)
		assertAgentStatus(t, s, "api-key-a", db.AgentStatusPaused)
		assertNoSnapshotSaved(t, s)
	})

	t.Run("deleted agent does not save snapshot", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, "agent-a", "api-key-a", db.AgentStatusDeleted)

		rr := sendIngest(t, s, "api-key-a", mustJSON(t, testLiveSnapshot()))

		require.Equal(t, http.StatusOK, rr.Code)
		assertIngestResult(t, rr, shared.AgentStatusDeleted)
		assertAgentStatus(t, s, "api-key-a", db.AgentStatusDeleted)
		assertNoSnapshotSaved(t, s)
	})
}

func TestHandleAgentAlreadyRegistered(t *testing.T) {
	t.Run("rejects invalid payload", func(t *testing.T) {
		s := newTestServer(t)

		rr := sendAgentAlreadyRegistered(t, s, `{}`)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "PARSE_ERROR\n", rr.Body.String())
	})

	t.Run("rejects unknown api key", func(t *testing.T) {
		s := newTestServer(t)

		rr := sendAgentAlreadyRegistered(
			t, s,
			`{"api_key":"missing","host_id":"host-a"}`)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, INVALID_API_KEY+"\n", rr.Body.String())
	})

	t.Run("stores first host id and allows install", func(t *testing.T) {
		s := newTestServer(t)
		createAgentAlreadyRegisteredFixture(t, s, "agent-a", "api-key-a", "")

		rr := sendAgentAlreadyRegistered(
			t, s,
			`{"api_key":"api-key-a","host_id":"host-a"}`)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, ALLOW_INSTALL, rr.Body.String())

		agent, notFound, err := s.DB.GetAgentFromAPIKey(t.Context(), "api-key-a")
		require.NoError(t, err)
		require.False(t, notFound)
		require.True(t, agent.HostID.Valid)
		require.Equal(t, "host-a", agent.HostID.String)
	})

	t.Run("blocks different host id", func(t *testing.T) {
		s := newTestServer(t)
		createAgentAlreadyRegisteredFixture(
			t, s,
			"agent-a", "api-key-a", "host-a",
		)

		rr := sendAgentAlreadyRegistered(
			t, s,
			`{"api_key":"api-key-a","host_id":"host-b"}`)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, DUPLICATED_HOST, rr.Body.String())
	})

	t.Run("allows same host id", func(t *testing.T) {
		s := newTestServer(t)
		createAgentAlreadyRegisteredFixture(
			t, s,
			"agent-a", "api-key-a", "host-a")

		rr := sendAgentAlreadyRegistered(
			t, s,
			`{"api_key":"api-key-a","host_id":"host-a"}`)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, ALLOW_INSTALL, rr.Body.String())
	})
}

func sendIngest(
	t *testing.T, s *Server, apiKey string, json string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(json))
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func createIngestAgentFixture(
	t *testing.T, s *Server, agentID string, apiKey string, status db.AgentStatus,
) {
	t.Helper()

	createAgentAlreadyRegisteredFixture(t, s, agentID, apiKey, "host-a")
	_, err := s.DB.GetDB().ExecContext(
		t.Context(),
		`UPDATE agents SET status = ? WHERE id = ?`,
		status,
		agentID,
	)
	require.NoError(t, err)
}

func mustJSON(t *testing.T, snapshot shared.LiveSnapshot) string {
	t.Helper()

	body, err := json.Marshal(shared.IngestEvent{LiveSnapshot: &snapshot})
	require.NoError(t, err)
	return string(body)
}

func testLiveSnapshot() shared.LiveSnapshot {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	return shared.LiveSnapshot{
		ID:     "snapshot-a",
		SentAt: now,
		Agent: shared.Agent{
			Version: "agent-version-a",
		},
		Host: shared.Host{
			ID:   "host-a",
			Name: "host-a-name",
			OS:   "linux",
			Arch: "amd64",
		},
		Metrics: shared.MetricsSnapshot{
			ObservedAt: now,
			CPU: shared.CPUUsage{
				UsedPercent:    12.5,
				CoresLogical:   4,
				CoresPhysical:  2,
				PerCorePercent: []float64{10, 20, 5, 15},
				Load1M:         0.1,
				Load5M:         0.2,
				Load15M:        0.3,
			},
			VirtualMemory: shared.MemoryUsage{
				UsedPercent:    40,
				UsedBytes:      400,
				AvailableBytes: 600,
				TotalBytes:     1000,
			},
			SwapMemory: shared.MemoryUsage{
				UsedPercent:    0,
				UsedBytes:      0,
				AvailableBytes: 1000,
				TotalBytes:     1000,
			},
			Disk: shared.DiskMetrics{
				Total: shared.DiskUsage{
					Device:      "/dev/sda1",
					Mount:       "/",
					Filesystem:  "ext4",
					UsedPercent: 50,
					UsedBytes:   500,
					FreeBytes:   500,
					TotalBytes:  1000,
				},
				Partitions: []shared.DiskUsage{
					{
						Device:      "/dev/sda1",
						Mount:       "/",
						Filesystem:  "ext4",
						UsedPercent: 50,
						UsedBytes:   500,
						FreeBytes:   500,
						TotalBytes:  1000,
					},
				},
			},
		},
	}
}

func assertIngestResult(t *testing.T, rr *httptest.ResponseRecorder, status shared.AgentStatus) {
	t.Helper()

	var result shared.IngestResult
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	require.Equal(t, status, result.AgentStatus)
}

func assertAgentStatus(t *testing.T, s *Server, apiKey string, status db.AgentStatus) {
	t.Helper()

	agent, notFound, err := s.DB.GetAgentFromAPIKey(t.Context(), apiKey)
	require.NoError(t, err)
	require.False(t, notFound)
	require.Equal(t, status, agent.Status)
}

func assertSnapshotSaved(t *testing.T, s *Server) {
	t.Helper()

	require.Equal(t, 1, countRows(t, s, "cpu_samples"))
	require.Equal(t, 1, countRows(t, s, "memory_samples"))
	require.Equal(t, 2, countRows(t, s, "disk_samples"))

	var version string
	err := s.DB.GetDB().QueryRowContext(
		t.Context(), `SELECT version FROM agents WHERE id = 'agent-a'`,
	).Scan(&version)
	require.NoError(t, err)
	require.Equal(t, "agent-version-a", version)
}

func assertNoSnapshotSaved(t *testing.T, s *Server) {
	t.Helper()

	require.Equal(t, 0, countRows(t, s, "cpu_samples"))
	require.Equal(t, 0, countRows(t, s, "memory_samples"))
	require.Equal(t, 0, countRows(t, s, "disk_samples"))
}

func assertSnapshotProcessedEvent(
	t *testing.T, events <-chan EventSnapshotProcessed, agentID string,
) {
	t.Helper()

	select {
	case event := <-events:
		require.Equal(t, agentID, event.AgentID)
	case <-time.After(time.Second):
		t.Fatal("expected snapshot processed event")
	}
}

func countRows(t *testing.T, s *Server, table string) int {
	t.Helper()

	var count int
	err := s.DB.GetDB().QueryRowContext(
		t.Context(), "SELECT COUNT(*) FROM "+table,
	).Scan(&count)
	require.NoError(t, err)
	return count
}

func sendAgentAlreadyRegistered(
	t *testing.T, s *Server, json string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost, "/agents/already-registered", strings.NewReader(json))
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func createAgentAlreadyRegisteredFixture(
	t *testing.T, s *Server, agentID string, apiKey string, hostID string,
) {
	t.Helper()

	user, err := s.DB.CreateUserAndTeam(t.Context(), agentID+"@example.com", "hash")
	require.NoError(t, err)

	err = s.DB.InsertAgent(t.Context(), db.InsertAgentParams{
		ID:     agentID,
		TeamID: user.TeamID,
		Name:   agentID,
		HostID: sql.NullString{String: hostID, Valid: hostID != ""},
		ApiKey: apiKey,
	})
	require.NoError(t, err)
}
