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

-- name: deleteOldMemorySamples :exec
DELETE FROM memory_samples
WHERE observed_at < @cutoff;
