CREATE TABLE IF NOT EXISTS cpu_samples (
    id INTEGER PRIMARY KEY,
    observed_at TEXT NOT NULL,
    used_percent REAL NOT NULL,
    cores_logical INTEGER NOT NULL,
    cores_physical INTEGER NOT NULL,
    per_core_percent TEXT NOT NULL,
    load_1m REAL NOT NULL,
    load_5m REAL NOT NULL,
    load_15m REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS cpu_samples_observed_idx ON cpu_samples (observed_at);

CREATE TABLE IF NOT EXISTS memory_samples (
    id INTEGER PRIMARY KEY,
    observed_at TEXT NOT NULL,
    virtual_used_percent REAL NOT NULL,
    virtual_used_bytes INTEGER NOT NULL,
    virtual_available_bytes INTEGER NOT NULL,
    virtual_total_bytes INTEGER NOT NULL,
    swap_used_percent REAL NOT NULL,
    swap_used_bytes INTEGER NOT NULL,
    swap_available_bytes INTEGER NOT NULL,
    swap_total_bytes INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS memory_samples_observed_idx ON memory_samples (observed_at);

CREATE TABLE IF NOT EXISTS disk_samples (
    id INTEGER PRIMARY KEY,
    observed_at TEXT NOT NULL,
    is_total INTEGER NOT NULL,
    device TEXT NOT NULL,
    mount TEXT NOT NULL,
    filesystem TEXT NOT NULL,
    used_percent REAL NOT NULL,
    used_bytes INTEGER NOT NULL,
    free_bytes INTEGER NOT NULL,
    total_bytes INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS disk_samples_observed_idx ON disk_samples (observed_at);
CREATE INDEX IF NOT EXISTS disk_samples_total_observed_idx ON disk_samples (is_total, observed_at);
CREATE INDEX IF NOT EXISTS disk_samples_mount_observed_idx ON disk_samples (mount, observed_at);
