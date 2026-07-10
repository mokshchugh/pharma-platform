CREATE TABLE IF NOT EXISTS plc_samples (
    timestamp TIMESTAMP,
    machine_id SYMBOL,
    machine_name SYMBOL,
    tag_name SYMBOL,
    value DOUBLE,
    quality BYTE
)
TIMESTAMP(timestamp)
PARTITION BY DAY;
