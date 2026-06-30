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

		rr := sendIngest(t, s, "", ingestBody(t))

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		assertNoSnapshotSaved(t, s)
	})

	t.Run("rejects invalid api key", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, db.AgentStatusCreated)

		rr := sendIngest(t, s, "wrong-api-key", ingestBody(t))

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		assertNoSnapshotSaved(t, s)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		s := newTestServer(t)
		createIngestAgentFixture(t, s, db.AgentStatusCreated)

		rr := sendIngest(t, s, "api-key-a", `{`)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assertNoSnapshotSaved(t, s)
	})

	for _, tc := range []struct {
		name           string
		initialStatus  db.AgentStatus
		responseStatus shared.AgentStatus
		finalStatus    db.AgentStatus
		savesSnapshot  bool
	}{
		{"created agent saves snapshot and becomes running", db.AgentStatusCreated, shared.AgentStatusRunning, db.AgentStatusRunning, true},
		{"running agent saves snapshot and stays running", db.AgentStatusRunning, shared.AgentStatusRunning, db.AgentStatusRunning, true},
		{"resuming agent saves snapshot and becomes running", db.AgentStatusResuming, shared.AgentStatusRunning, db.AgentStatusRunning, true},
		{"paused agent does not save snapshot", db.AgentStatusPaused, shared.AgentStatusPaused, db.AgentStatusPaused, false},
		{"deleted agent does not save snapshot", db.AgentStatusDeleted, shared.AgentStatusDeleted, db.AgentStatusDeleted, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(t)
			createIngestAgentFixture(t, s, tc.initialStatus)
			events := s.BusSnapshotProcessed.Subscribe(t.Context())

			rr := sendIngest(t, s, "api-key-a", ingestBody(t))

			require.Equal(t, http.StatusOK, rr.Code)
			assertIngestResult(t, rr, tc.responseStatus)
			assertAgentStatus(t, s, tc.finalStatus)
			if tc.savesSnapshot {
				assertSnapshotSaved(t, s)
				assertSnapshotProcessedEvent(t, events, SnapshotProcessedTypeMetrics)
				return
			}
			assertNoSnapshotSaved(t, s)
		})
	}

	for _, tc := range []struct {
		name           string
		initialStatus  db.AgentStatus
		responseStatus shared.AgentStatus
		finalStatus    db.AgentStatus
		savesLogs      bool
	}{
		{"created agent saves logs and becomes running", db.AgentStatusCreated, shared.AgentStatusRunning, db.AgentStatusRunning, true},
		{"running agent saves logs and stays running", db.AgentStatusRunning, shared.AgentStatusRunning, db.AgentStatusRunning, true},
		{"resuming agent saves logs and becomes running", db.AgentStatusResuming, shared.AgentStatusRunning, db.AgentStatusRunning, true},
		{"paused agent does not save logs", db.AgentStatusPaused, shared.AgentStatusPaused, db.AgentStatusPaused, false},
		{"deleted agent does not save logs", db.AgentStatusDeleted, shared.AgentStatusDeleted, db.AgentStatusDeleted, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer(t)
			createIngestAgentFixture(t, s, tc.initialStatus)
			events := s.BusSnapshotProcessed.Subscribe(t.Context())

			rr := sendIngest(t, s, "api-key-a", logIngestBody(t))

			require.Equal(t, http.StatusOK, rr.Code)
			assertIngestResult(t, rr, tc.responseStatus)
			assertAgentStatus(t, s, tc.finalStatus)
			assertNoSnapshotSaved(t, s)
			if tc.savesLogs {
				assertSnapshotProcessedEvent(t, events, SnapshotProcessedTypeLogs)
				require.Equal(t, 1, countRows(t, s, "log_entries"))
				return
			}
			assertNoSnapshotProcessedEvent(t, events)
			require.Equal(t, 0, countRows(t, s, "log_entries"))
		})
	}
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
		require.Equal(t, invalidAPIKey+"\n", rr.Body.String())
	})

	t.Run("stores first host id and allows install", func(t *testing.T) {
		s := newTestServer(t)
		createAgentAlreadyRegisteredFixture(t, s, "")

		rr := sendAgentAlreadyRegistered(
			t, s,
			`{"api_key":"api-key-a","host_id":"host-a"}`)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, allowInstall, rr.Body.String())

		agent, notFound, err := s.DB.GetAgentFromAPIKey(t.Context(), "api-key-a")
		require.NoError(t, err)
		require.False(t, notFound)
		require.True(t, agent.HostID.Valid)
		require.Equal(t, "host-a", agent.HostID.String)
	})

	t.Run("blocks different host id", func(t *testing.T) {
		s := newTestServer(t)
		createAgentAlreadyRegisteredFixture(t, s, "host-a")

		rr := sendAgentAlreadyRegistered(
			t, s,
			`{"api_key":"api-key-a","host_id":"host-b"}`)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, duplicatedHost, rr.Body.String())
	})

	t.Run("allows same host id", func(t *testing.T) {
		s := newTestServer(t)
		createAgentAlreadyRegisteredFixture(t, s, "host-a")

		rr := sendAgentAlreadyRegistered(
			t, s,
			`{"api_key":"api-key-a","host_id":"host-a"}`)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, allowInstall, rr.Body.String())
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

func createIngestAgentFixture(t *testing.T, s *Server, status db.AgentStatus) {
	t.Helper()

	createAgentAlreadyRegisteredFixture(t, s, "host-a")
	_, err := s.DB.GetDB().ExecContext(
		t.Context(),
		`UPDATE agents SET status = ? WHERE id = 'agent-a'`,
		status,
	)
	require.NoError(t, err)
}

func ingestBody(t *testing.T) string {
	t.Helper()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	body, err := json.Marshal(shared.IngestEvent{LiveSnapshot: &shared.LiveSnapshot{
		SentAt: now,
		Agent:  shared.Agent{Version: "agent-version-a"},
		Metrics: shared.MetricsSnapshot{
			ObservedAt: now,
			CPU: shared.CPUUsage{
				PerCorePercent: []float64{},
			},
		},
	}})
	require.NoError(t, err)
	return string(body)
}

func logIngestBody(t *testing.T) string {
	t.Helper()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	body, err := json.Marshal(shared.IngestEvent{Logs: []shared.LogEvent{{
		ID:         "log-a",
		SentAt:     now,
		ObservedAt: now,
		Source:     "journal",
		Unit:       "ssh.service",
		Priority:   "6",
		Message:    "accepted login",
	}}})
	require.NoError(t, err)
	return string(body)
}

func assertIngestResult(t *testing.T, rr *httptest.ResponseRecorder, status shared.AgentStatus) {
	t.Helper()

	var result shared.IngestResult
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	require.Equal(t, status, result.AgentStatus)
}

func assertAgentStatus(t *testing.T, s *Server, status db.AgentStatus) {
	t.Helper()

	agent, notFound, err := s.DB.GetAgentFromAPIKey(t.Context(), "api-key-a")
	require.NoError(t, err)
	require.False(t, notFound)
	require.Equal(t, status, agent.Status)
}

func assertSnapshotSaved(t *testing.T, s *Server) {
	t.Helper()

	require.Equal(t, 1, countRows(t, s, "cpu_samples"))
	require.Equal(t, 1, countRows(t, s, "memory_samples"))
	require.Equal(t, 1, countRows(t, s, "disk_samples"))

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

func assertSnapshotProcessedEvent(t *testing.T, events <-chan EventSnapshotProcessed, eventType SnapshotProcessedType) {
	t.Helper()

	select {
	case event := <-events:
		require.Equal(t, eventType, event.Type)
		require.Equal(t, "agent-a", event.AgentID)
	case <-time.After(time.Second):
		t.Fatalf("expected snapshot processed event: type=%s", eventType)
	}
}

func assertNoSnapshotProcessedEvent(t *testing.T, events <-chan EventSnapshotProcessed) {
	t.Helper()

	select {
	case event := <-events:
		t.Fatalf("unexpected snapshot processed event: type=%s agent_id=%s", event.Type, event.AgentID)
	case <-time.After(100 * time.Millisecond):
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

func createAgentAlreadyRegisteredFixture(t *testing.T, s *Server, hostID string) {
	t.Helper()

	user, err := s.DB.CreateUserAndTeam(t.Context(), "agent-a@example.com", "hash")
	require.NoError(t, err)

	err = s.DB.InsertAgent(t.Context(), db.InsertAgentParams{
		ID:     "agent-a",
		TeamID: user.TeamID,
		Name:   "agent-a",
		HostID: sql.NullString{String: hostID, Valid: hostID != ""},
		ApiKey: "api-key-a",
	})
	require.NoError(t, err)
}
