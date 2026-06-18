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

-- name: DeleteOldCPUSamplesQuery :exec
DELETE FROM cpu_samples
WHERE observed_at < @cutoff;

-- name: DeleteOldMemorySamplesQuery :exec
DELETE FROM memory_samples
WHERE observed_at < @cutoff;

-- name: DeleteOldDiskSamplesQuery :exec
DELETE FROM disk_samples
WHERE observed_at < @cutoff;
