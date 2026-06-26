-- name: GetUserByEmail :one
SELECT users.*
FROM users
WHERE email = @email;

-- name: GetUserByID :one
SELECT users.*
FROM users
WHERE id = @id;

-- name: GetTeamByID :one
SELECT teams.*  
FROM teams
WHERE id = @id;

-- name: createTeam :one
INSERT INTO teams (name)
VALUES (@name)
RETURNING *;

-- name: createUser :one
INSERT INTO users (team_id, email, password_hash)
VALUES (@team_id, @email, @password_hash)
RETURNING *;

-- name: deleteSession :exec
DELETE FROM sessions
WHERE token = @token;

-- name: findSession :one
SELECT data
FROM sessions
WHERE token = @token
AND expiry > @now;

-- name: commitSession :exec
INSERT INTO sessions (token, data, expiry)
VALUES (@token, @data, @expiry)
ON CONFLICT(token) DO UPDATE SET data = excluded.data, expiry = excluded.expiry;

-- name: GetTeamByAgentAPIKey :one
SELECT teams.* 
FROM teams
JOIN agents ON agents.team_id = teams.id
WHERE agents.api_key = @agent_api_key;

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

-- name: createCPUSample :exec
INSERT INTO cpu_samples (
    observed_at,
    used_percent,
    cores_logical,
    cores_physical,
    per_core_percent,
    load_1m,
    load_5m,
    load_15m
) VALUES (
    @observed_at,
    @used_percent,
    @cores_logical,
    @cores_physical,
    @per_core_percent,
    @load_1m,
    @load_5m,
    @load_15m
);

-- name: createMemorySample :exec
INSERT INTO memory_samples (
    observed_at,
    virtual_used_percent,
    virtual_used_bytes,
    virtual_available_bytes,
    virtual_total_bytes,
    swap_used_percent,
    swap_used_bytes,
    swap_available_bytes,
    swap_total_bytes
) VALUES (
    @observed_at,
    @virtual_used_percent,
    @virtual_used_bytes,
    @virtual_available_bytes,
    @virtual_total_bytes,
    @swap_used_percent,
    @swap_used_bytes,
    @swap_available_bytes,
    @swap_total_bytes
);

-- name: createDiskSample :exec
INSERT INTO disk_samples (
    observed_at,
    is_total,
    device,
    mount,
    filesystem,
    used_percent,
    used_bytes,
    free_bytes,
    total_bytes
) VALUES (
    @observed_at,
    @is_total,
    @device,
    @mount,
    @filesystem,
    @used_percent,
    @used_bytes,
    @free_bytes,
    @total_bytes
);

-- name: GetLatestCPUSample :one
SELECT cpu_samples.*
FROM cpu_samples
ORDER BY observed_at DESC
LIMIT 1;

-- name: ListCPUSamplesSince :many
SELECT cpu_samples.*
FROM cpu_samples
WHERE observed_at >= @since
ORDER BY observed_at;

-- name: GetLatestMemorySample :one
SELECT memory_samples.*
FROM memory_samples
ORDER BY observed_at DESC
LIMIT 1;

-- name: ListMemorySamplesSince :many
SELECT memory_samples.*
FROM memory_samples
WHERE observed_at >= @since
ORDER BY observed_at;

-- name: ListLatestDiskSamples :many
SELECT disk_samples.*
FROM disk_samples
WHERE observed_at = (
    SELECT MAX(observed_at)
    FROM disk_samples
)
ORDER BY is_total DESC, free_bytes, mount;

-- name: ListDiskSamplesSince :many
SELECT disk_samples.*
FROM disk_samples
WHERE observed_at >= @since
ORDER BY observed_at, is_total DESC, mount;

-- name: ListDiskSamplesForMountSince :many
SELECT disk_samples.*
FROM disk_samples
WHERE mount = @mount
  AND observed_at >= @since
ORDER BY observed_at;

-- name: deleteOldCPUSamples :exec
DELETE FROM cpu_samples
WHERE observed_at < @cutoff;

-- name: deleteOldMemorySamples :exec
DELETE FROM memory_samples
WHERE observed_at < @cutoff;

-- name: deleteOldDiskSamples :exec
DELETE FROM disk_samples
WHERE observed_at < @cutoff;
