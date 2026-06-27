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

-- name: deleteOldCPUSamples :exec
DELETE FROM cpu_samples
WHERE observed_at < @cutoff;
