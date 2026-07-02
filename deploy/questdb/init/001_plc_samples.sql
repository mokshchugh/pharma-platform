CREATE TABLE IF NOT EXISTS plc_samples (
    timestamp TIMESTAMP,
    plc_id SYMBOL,
    tag_id SYMBOL,
    value DOUBLE,
    quality BYTE
)
TIMESTAMP(timestamp)
PARTITION BY DAY;