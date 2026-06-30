-- name: createLogEntry :exec
INSERT INTO log_entries (
    id,
    team_id,
    agent_id,
    observed_at,
    received_at,
    source,
    unit,
    priority,
    message
) VALUES (
    @id,
    @team_id,
    @agent_id,
    @observed_at,
    @received_at,
    @source,
    @unit,
    @priority,
    @message
);

-- name: ListAgentLogEntries :many
SELECT * FROM log_entries
WHERE agent_id = @agent_id AND team_id = @team_id
ORDER BY observed_at DESC
LIMIT @limit;
