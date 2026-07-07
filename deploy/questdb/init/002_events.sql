CREATE TABLE IF NOT EXISTS alarms (
    timestamp TIMESTAMP,
    machine_id SYMBOL,
    tag_name SYMBOL,
    severity SYMBOL,
    message STRING,
    acknowledged BOOLEAN
)
TIMESTAMP(timestamp)
PARTITION BY DAY;

CREATE TABLE IF NOT EXISTS events (
    timestamp TIMESTAMP,
    machine_id SYMBOL,
    event_type SYMBOL,
    batch_id SYMBOL,
    payload STRING
)
TIMESTAMP(timestamp)
PARTITION BY DAY;

CREATE TABLE IF NOT EXISTS logs (
    timestamp TIMESTAMP,
    level SYMBOL,
    component SYMBOL,
    message STRING
)
TIMESTAMP(timestamp)
PARTITION BY DAY;
