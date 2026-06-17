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
    used_percent,
    used_bytes,
    available_bytes,
    total_bytes
) VALUES (
    @observed_at,
    @used_percent,
    @used_bytes,
    @available_bytes,
    @total_bytes
);

-- name: CreateDiskSampleQuery :exec
INSERT INTO disk_samples (
    observed_at,
    mount,
    filesystem,
    used_percent,
    used_bytes,
    free_bytes,
    total_bytes
) VALUES (
    @observed_at,
    @mount,
    @filesystem,
    @used_percent,
    @used_bytes,
    @free_bytes,
    @total_bytes
);

-- name: DeleteOldCPUSamplesQuery :exec
DELETE FROM cpu_samples
WHERE observed_at < @cutoff;

-- name: DeleteOldMemorySamplesQuery :exec
DELETE FROM memory_samples
WHERE observed_at < @cutoff;

-- name: DeleteOldDiskSamplesQuery :exec
DELETE FROM disk_samples
WHERE observed_at < @cutoff;
