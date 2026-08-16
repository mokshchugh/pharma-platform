CREATE TABLE IF NOT EXISTS alarm_acks (
    id SERIAL PRIMARY KEY,
    machine_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    alarm_ts TIMESTAMPTZ NOT NULL,
    acked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_alarm_acks_lookup ON alarm_acks(machine_id, tag_name, alarm_ts);
