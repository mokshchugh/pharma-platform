-- The simulator (internal/simulation/simulator.go) produces roughly 1 part
-- every 1.0-1.3s per machine (speed 75-100, 100ms tick). The oee_targets
-- default of 60s/part (deploy/postgres/init/002_oee_schema.sql) assumes a
-- much slower process, which inflated performance/OEE past 100%. Seed a
-- realistic ideal cycle time so simulated performance stays in range.
INSERT INTO oee_targets (machine_id, availability_target, performance_target, quality_target,
                          ideal_cycle_time_seconds, planned_production_time_seconds)
SELECT id, 0.90, 0.95, 0.98, 1.0, 28800
FROM machines
ON CONFLICT (machine_id) DO NOTHING;
