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
