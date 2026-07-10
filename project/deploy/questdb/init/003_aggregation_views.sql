CREATE MATERIALIZED VIEW IF NOT EXISTS plc_samples_1m AS (
  SELECT machine_id, machine_name, tag_name, timestamp,
    avg(value) as avg_value,
    min(value) as min_value,
    max(value) as max_value,
    count() as sample_count
  FROM plc_samples
  SAMPLE BY 1m
)
TIMESTAMP(timestamp)
PARTITION BY DAY;

CREATE MATERIALIZED VIEW IF NOT EXISTS plc_samples_1h AS (
  SELECT machine_id, machine_name, tag_name, timestamp,
    avg(value) as avg_value,
    min(value) as min_value,
    max(value) as max_value,
    count() as sample_count
  FROM plc_samples
  SAMPLE BY 1h
)
TIMESTAMP(timestamp)
PARTITION BY MONTH;

CREATE MATERIALIZED VIEW IF NOT EXISTS plc_samples_1d AS (
  SELECT machine_id, machine_name, tag_name, timestamp,
    avg(value) as avg_value,
    min(value) as min_value,
    max(value) as max_value,
    count() as sample_count
  FROM plc_samples
  SAMPLE BY 1d
)
TIMESTAMP(timestamp)
PARTITION BY YEAR;

CREATE MATERIALIZED VIEW IF NOT EXISTS plc_samples_1w AS (
  SELECT machine_id, machine_name, tag_name, timestamp,
    avg(value) as avg_value,
    min(value) as min_value,
    max(value) as max_value,
    count() as sample_count
  FROM plc_samples
  SAMPLE BY 7d
)
TIMESTAMP(timestamp)
PARTITION BY YEAR;
