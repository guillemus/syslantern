-- name: GetUserByEmailQuery :one
SELECT sqlc.embed(users)
FROM users
WHERE email = @email;

-- name: GetUserByIDQuery :one
SELECT sqlc.embed(users)
FROM users
WHERE id = @id;

-- name: CreateUserQuery :one
INSERT INTO users (email, password_hash)
VALUES (@email, @password_hash)
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
