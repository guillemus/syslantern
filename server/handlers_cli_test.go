package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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

		agent, err := s.DB.GetAgentByAPIKey(t.Context(), "api-key-a")
		require.NoError(t, err)
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

	_, err = s.DB.ExecContext(
		t.Context(),
		`
		INSERT INTO agents (id, team_id, name, version, status, host_id, api_key)
		VALUES (?, ?, ?, '', 'created', NULLIF(?, ''), ?)
	`,
		agentID,
		user.TeamID,
		agentID,
		hostID,
		apiKey,
	)
	require.NoError(t, err)
}
