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

-- name: deleteOldDiskSamples :exec
DELETE FROM disk_samples
WHERE observed_at < @cutoff;
