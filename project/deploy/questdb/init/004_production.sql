CREATE TABLE IF NOT EXISTS production_counts (
    timestamp TIMESTAMP,
    machine_id SYMBOL,
    batch_id SYMBOL,
    good_count DOUBLE,
    bad_count DOUBLE,
    total_count DOUBLE,
    cycle_time DOUBLE,
    speed DOUBLE
)
TIMESTAMP(timestamp)
PARTITION BY DAY;

CREATE TABLE IF NOT EXISTS machine_state (
    timestamp TIMESTAMP,
    machine_id SYMBOL,
    state SYMBOL,
    speed DOUBLE,
    setpoint DOUBLE,
    load_percent DOUBLE
)
TIMESTAMP(timestamp)
PARTITION BY DAY;
