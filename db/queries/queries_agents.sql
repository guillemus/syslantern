-- name: upsertAgentForTeam :one
INSERT INTO agents (id, team_id, name, version, status, api_key)
VALUES (@id, @team_id, @name, @version, @status, @api_key)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    version = excluded.version,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP
WHERE agents.team_id = excluded.team_id
RETURNING *;

-- name: InsertAgent :exec
INSERT INTO agents (id, team_id, name, version, status, host_id, api_key)
VALUES (?, ?, ?, '', 'created', NULLIF(?, ''), ?);

-- name: ListAgentsForTeam :many
SELECT agents.*
FROM agents
WHERE team_id = @team_id
AND STATUS != 'deleted'
ORDER BY updated_at DESC;

-- name: GetAgentForTeam :one
SELECT agents.*
FROM agents
WHERE id = @id
AND team_id = @team_id;

-- name: GetAgentByAPIKey :one
SELECT agents.*
FROM agents
WHERE api_key = @api_key;

-- name: updateAgentHostID :exec
UPDATE agents
SET host_id = @host_id,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: touchAgentForTeam :execrows
UPDATE agents
SET version = @version,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id
AND team_id = @team_id;

-- name: setAgentStatusForTeam :exec
UPDATE agents
SET status = @status,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id
AND team_id = @team_id;
