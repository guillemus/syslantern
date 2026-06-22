-- name: GetUserByEmailQuery :one
SELECT sqlc.embed(users)
FROM users
WHERE email = @email;

-- name: GetUserByIDQuery :one
SELECT sqlc.embed(users)
FROM users
WHERE id = @id;

-- name: GetFirstUserByTeamIDQuery :one
SELECT sqlc.embed(users)
FROM users
WHERE team_id = @team_id
ORDER BY id
LIMIT 1;

-- name: GetTeamByIDQuery :one
SELECT sqlc.embed(teams)
FROM teams
WHERE id = @id;

-- name: CreateTeamQuery :one
INSERT INTO teams (name, agent_api_key)
VALUES (@name, @agent_api_key)
RETURNING *;

-- name: CreateUserQuery :one
INSERT INTO users (team_id, email, password_hash)
VALUES (@team_id, @email, @password_hash)
RETURNING *;

-- name: DeleteSessionQuery :exec
DELETE FROM sessions
WHERE token = @token;

-- name: FindSessionQuery :one
SELECT data
FROM sessions
WHERE token = @token
AND expiry > @now;

-- name: CommitSessionQuery :exec
INSERT INTO sessions (token, data, expiry)
VALUES (@token, @data, @expiry)
ON CONFLICT(token) DO UPDATE SET data = excluded.data, expiry = excluded.expiry;

-- name: GetTeamByAgentAPIKeyQuery :one
SELECT sqlc.embed(teams)
FROM teams
WHERE agent_api_key = @agent_api_key;

-- name: UpsertAgentForUserQuery :one
INSERT INTO agents (id, user_id, name, version)
VALUES (@id, @user_id, @name, @version)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    version = excluded.version,
    updated_at = CURRENT_TIMESTAMP
WHERE agents.user_id = excluded.user_id
RETURNING *;

-- name: ListAgentsForUserQuery :many
SELECT sqlc.embed(agents)
FROM agents
WHERE user_id = @user_id
ORDER BY updated_at DESC;

-- name: GetAgentForUserQuery :one
SELECT sqlc.embed(agents)
FROM agents
WHERE id = @id
AND user_id = @user_id;

-- name: TouchAgentQuery :exec
UPDATE agents
SET version = @version,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id;

-- name: CreateCPUSampleQuery :exec
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

-- name: CreateMemorySampleQuery :exec
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

-- name: CreateDiskSampleQuery :exec
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

-- name: GetLatestCPUSampleQuery :one
SELECT sqlc.embed(cpu_samples)
FROM cpu_samples
ORDER BY observed_at DESC
LIMIT 1;

-- name: ListCPUSamplesSinceQuery :many
SELECT sqlc.embed(cpu_samples)
FROM cpu_samples
WHERE observed_at >= @since
ORDER BY observed_at;

-- name: GetLatestMemorySampleQuery :one
SELECT sqlc.embed(memory_samples)
FROM memory_samples
ORDER BY observed_at DESC
LIMIT 1;

-- name: ListMemorySamplesSinceQuery :many
SELECT sqlc.embed(memory_samples)
FROM memory_samples
WHERE observed_at >= @since
ORDER BY observed_at;

-- name: ListLatestDiskSamplesQuery :many
SELECT sqlc.embed(disk_samples)
FROM disk_samples
WHERE observed_at = (
    SELECT MAX(observed_at)
    FROM disk_samples
)
ORDER BY is_total DESC, free_bytes, mount;

-- name: ListDiskSamplesSinceQuery :many
SELECT sqlc.embed(disk_samples)
FROM disk_samples
WHERE observed_at >= @since
ORDER BY observed_at, is_total DESC, mount;

-- name: ListDiskSamplesForMountSinceQuery :many
SELECT sqlc.embed(disk_samples)
FROM disk_samples
WHERE mount = @mount
  AND observed_at >= @since
ORDER BY observed_at;

-- name: DeleteOldCPUSamplesQuery :exec
DELETE FROM cpu_samples
WHERE observed_at < @cutoff;

-- name: DeleteOldMemorySamplesQuery :exec
DELETE FROM memory_samples
WHERE observed_at < @cutoff;

-- name: DeleteOldDiskSamplesQuery :exec
DELETE FROM disk_samples
WHERE observed_at < @cutoff;
