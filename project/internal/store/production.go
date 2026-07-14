package store

import (
	"fmt"
	"time"

	"pharma-platform/internal/models"
	"pharma-platform/internal/postgres"
)

type ProductionStore struct {
	client *postgres.Client
}

func NewProductionStore(client *postgres.Client) *ProductionStore {
	return &ProductionStore{client: client}
}

func (s *ProductionStore) GetActiveRuns() []models.ProductionRun {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	rows, err := db.Query(
		`SELECT pr.id, pr.machine_id, COALESCE(m.machine_name, ''), pr.batch_id, pr.product_name,
		        pr.target_qty, pr.good_qty, pr.bad_qty, pr.start_time, pr.end_time, pr.status, pr.created_at
		 FROM production_runs pr
		 JOIN machines m ON m.id = pr.machine_id
		 WHERE pr.status = 'running'
		 ORDER BY pr.start_time DESC`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var runs []models.ProductionRun
	for rows.Next() {
		var r models.ProductionRun
		if err := rows.Scan(&r.ID, &r.MachineID, &r.MachineName, &r.BatchID, &r.ProductName,
			&r.TargetQty, &r.GoodQty, &r.BadQty, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt); err != nil {
			continue
		}
		runs = append(runs, r)
	}

	return runs
}

func (s *ProductionStore) GetActiveRun(machineID int) *models.ProductionRun {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	var r models.ProductionRun
	err := db.QueryRow(
		`SELECT pr.id, pr.machine_id, COALESCE(m.machine_name, ''), pr.batch_id, pr.product_name,
		        pr.target_qty, pr.good_qty, pr.bad_qty, pr.start_time, pr.end_time, pr.status, pr.created_at
		 FROM production_runs pr
		 JOIN machines m ON m.id = pr.machine_id
		 WHERE pr.machine_id = $1 AND pr.status = 'running'`,
		machineID,
	).Scan(&r.ID, &r.MachineID, &r.MachineName, &r.BatchID, &r.ProductName,
		&r.TargetQty, &r.GoodQty, &r.BadQty, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil
	}

	return &r
}

func (s *ProductionStore) ListRuns(machineID int, limit int) []models.ProductionRun {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	rows, err := db.Query(
		`SELECT pr.id, pr.machine_id, COALESCE(m.machine_name, ''), pr.batch_id, pr.product_name,
		        pr.target_qty, pr.good_qty, pr.bad_qty, pr.start_time, pr.end_time, pr.status, pr.created_at
		 FROM production_runs pr
		 JOIN machines m ON m.id = pr.machine_id
		 WHERE pr.machine_id = $1
		 ORDER BY pr.start_time DESC
		 LIMIT $2`,
		machineID, limit,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var runs []models.ProductionRun
	for rows.Next() {
		var r models.ProductionRun
		if err := rows.Scan(&r.ID, &r.MachineID, &r.MachineName, &r.BatchID, &r.ProductName,
			&r.TargetQty, &r.GoodQty, &r.BadQty, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt); err != nil {
			continue
		}
		runs = append(runs, r)
	}

	return runs
}

func (s *ProductionStore) CreateRun(machineID int, batchID string, productName string, targetQty int) (*models.ProductionRun, error) {
	db := s.client.DB()
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var r models.ProductionRun
	err := db.QueryRow(
		`INSERT INTO production_runs (machine_id, batch_id, product_name, target_qty)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, machine_id, batch_id, product_name, target_qty, good_qty, bad_qty, start_time, end_time, status, created_at`,
		machineID, batchID, productName, targetQty,
	).Scan(&r.ID, &r.MachineID, &r.BatchID, &r.ProductName, &r.TargetQty,
		&r.GoodQty, &r.BadQty, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}

	db.QueryRow(`SELECT COALESCE(machine_name, '') FROM machines WHERE id = $1`, machineID).Scan(&r.MachineName)

	return &r, nil
}

func (s *ProductionStore) UpdateRunCounts(id int, good int, bad int) error {
	db := s.client.DB()
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(
		`UPDATE production_runs SET good_qty = $1, bad_qty = $2 WHERE id = $3`,
		good, bad, id,
	)
	if err != nil {
		return fmt.Errorf("update run counts: %w", err)
	}

	return nil
}

func (s *ProductionStore) CompleteRun(id int) error {
	db := s.client.DB()
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(
		`UPDATE production_runs SET end_time = now(), status = 'completed' WHERE id = $1 AND status = 'running'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("complete run: %w", err)
	}

	return nil
}

func (s *ProductionStore) GetActiveDowntime(machineID int) *models.DowntimeEvent {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	var d models.DowntimeEvent
	err := db.QueryRow(
		`SELECT de.id, de.machine_id, COALESCE(m.machine_name, ''), de.start_time, de.end_time,
		        de.reason, de.category, de.duration_seconds, de.created_at
		 FROM downtime_events de
		 JOIN machines m ON m.id = de.machine_id
		 WHERE de.machine_id = $1 AND de.end_time IS NULL`,
		machineID,
	).Scan(&d.ID, &d.MachineID, &d.MachineName, &d.StartTime, &d.EndTime,
		&d.Reason, &d.Category, &d.DurationSeconds, &d.CreatedAt)
	if err != nil {
		return nil
	}

	return &d
}

func (s *ProductionStore) ListDowntime(machineID int, limit int) []models.DowntimeEvent {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	rows, err := db.Query(
		`SELECT de.id, de.machine_id, COALESCE(m.machine_name, ''), de.start_time, de.end_time,
		        de.reason, de.category, de.duration_seconds, de.created_at
		 FROM downtime_events de
		 JOIN machines m ON m.id = de.machine_id
		 WHERE de.machine_id = $1
		 ORDER BY de.start_time DESC
		 LIMIT $2`,
		machineID, limit,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var events []models.DowntimeEvent
	for rows.Next() {
		var d models.DowntimeEvent
		if err := rows.Scan(&d.ID, &d.MachineID, &d.MachineName, &d.StartTime, &d.EndTime,
			&d.Reason, &d.Category, &d.DurationSeconds, &d.CreatedAt); err != nil {
			continue
		}
		events = append(events, d)
	}

	return events
}

func (s *ProductionStore) StartDowntime(machineID int, reason string, category string) (*models.DowntimeEvent, error) {
	db := s.client.DB()
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var d models.DowntimeEvent
	err := db.QueryRow(
		`INSERT INTO downtime_events (machine_id, reason, category)
		 VALUES ($1, $2, $3)
		 RETURNING id, machine_id, start_time, end_time, reason, category, duration_seconds, created_at`,
		machineID, reason, category,
	).Scan(&d.ID, &d.MachineID, &d.StartTime, &d.EndTime, &d.Reason, &d.Category, &d.DurationSeconds, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("start downtime: %w", err)
	}

	db.QueryRow(`SELECT COALESCE(machine_name, '') FROM machines WHERE id = $1`, machineID).Scan(&d.MachineName)

	return &d, nil
}

func (s *ProductionStore) EndDowntime(id int) error {
	db := s.client.DB()
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(
		`UPDATE downtime_events
		 SET end_time = now(),
		     duration_seconds = COALESCE(EXTRACT(EPOCH FROM now() - start_time)::INTEGER, 0)
		 WHERE id = $1 AND end_time IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("end downtime: %w", err)
	}

	return nil
}

func (s *ProductionStore) GetOEETargets(machineID int) *models.OEETargets {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	var t models.OEETargets
	err := db.QueryRow(
		`SELECT id, machine_id, availability_target, performance_target, quality_target,
		        ideal_cycle_time_seconds, planned_production_time_seconds
		 FROM oee_targets
		 WHERE machine_id = $1`,
		machineID,
	).Scan(&t.ID, &t.MachineID, &t.AvailabilityTarget, &t.PerformanceTarget,
		&t.QualityTarget, &t.IdealCycleTimeSeconds, &t.PlannedProductionTimeSeconds)
	if err != nil {
		return nil
	}

	return &t
}

func (s *ProductionStore) UpsertOEETargets(t *models.OEETargets) error {
	db := s.client.DB()
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.Exec(
		`INSERT INTO oee_targets (machine_id, availability_target, performance_target, quality_target,
		                          ideal_cycle_time_seconds, planned_production_time_seconds)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (machine_id) DO UPDATE SET
		     availability_target = EXCLUDED.availability_target,
		     performance_target = EXCLUDED.performance_target,
		     quality_target = EXCLUDED.quality_target,
		     ideal_cycle_time_seconds = EXCLUDED.ideal_cycle_time_seconds,
		     planned_production_time_seconds = EXCLUDED.planned_production_time_seconds`,
		t.MachineID, t.AvailabilityTarget, t.PerformanceTarget, t.QualityTarget,
		t.IdealCycleTimeSeconds, t.PlannedProductionTimeSeconds,
	)
	if err != nil {
		return fmt.Errorf("upsert oee targets: %w", err)
	}

	return nil
}

func (s *ProductionStore) CalculateOEE(machineID int, window time.Duration) *models.OEEResponse {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	startWindow := time.Now().Add(-window)

	targets := s.GetOEETargets(machineID)

	availTarget := 0.95
	perfTarget := 0.95
	qualTarget := 0.99
	idealCycle := 60.0
	plannedSec := 28800

	if targets != nil {
		availTarget = targets.AvailabilityTarget
		perfTarget = targets.PerformanceTarget
		qualTarget = targets.QualityTarget
		idealCycle = targets.IdealCycleTimeSeconds
		plannedSec = targets.PlannedProductionTimeSeconds
	}

	var goodQty, badQty int
	err := db.QueryRow(
		`SELECT COALESCE(SUM(good_qty), 0), COALESCE(SUM(bad_qty), 0)
		 FROM production_runs
		 WHERE machine_id = $1 AND start_time >= $2`,
		machineID, startWindow,
	).Scan(&goodQty, &badQty)
	if err != nil {
		goodQty = 0
		badQty = 0
	}

	var downtimeSec int
	err = db.QueryRow(
		`SELECT COALESCE(SUM(duration_seconds), 0)
		 FROM downtime_events
		 WHERE machine_id = $1 AND start_time >= $2`,
		machineID, startWindow,
	).Scan(&downtimeSec)
	if err != nil {
		downtimeSec = 0
	}

	var machineName string
	db.QueryRow(`SELECT COALESCE(machine_name, '') FROM machines WHERE id = $1`, machineID).Scan(&machineName)

	totalParts := goodQty + badQty
	runTimeSec := plannedSec - downtimeSec
	if runTimeSec < 0 {
		runTimeSec = 0
	}

	var availability, performance, quality, overall float64

	if plannedSec > 0 {
		availability = float64(runTimeSec) / float64(plannedSec)
	} else {
		availability = 0
	}

	if runTimeSec > 0 && totalParts > 0 {
		performance = (idealCycle * float64(totalParts)) / float64(runTimeSec)
	} else {
		performance = 0
	}

	if totalParts > 0 {
		quality = float64(goodQty) / float64(totalParts)
	} else {
		quality = 0
	}

	overall = availability * performance * quality

	return &models.OEEResponse{
		MachineID:   machineID,
		MachineName: machineName,
		OEE: models.OEEValues{
			Availability: availability,
			Performance:  performance,
			Quality:      quality,
			Overall:      overall,
			TargetOEE:    availTarget * perfTarget * qualTarget,
		},
		DowntimeSec: downtimeSec,
		RunTimeSec:  runTimeSec,
		GoodParts:   goodQty,
		BadParts:    badQty,
		TotalParts:  totalParts,
	}
}

func (s *ProductionStore) GetDashboardSummary() *models.DashboardSummary {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	var totalMachines int
	db.QueryRow(`SELECT COUNT(*) FROM machines`).Scan(&totalMachines)

	var runningMachines int
	db.QueryRow(
		`SELECT COUNT(DISTINCT machine_id) FROM production_runs WHERE status = 'running'`,
	).Scan(&runningMachines)

	stoppedMachines := totalMachines - runningMachines
	if stoppedMachines < 0 {
		stoppedMachines = 0
	}

	var todayGood, todayBad int
	db.QueryRow(
		`SELECT COALESCE(SUM(good_qty), 0), COALESCE(SUM(bad_qty), 0)
		 FROM production_runs
		 WHERE start_time >= CURRENT_DATE`,
	).Scan(&todayGood, &todayBad)

	rows, err := db.Query(`SELECT id, COALESCE(machine_name, '') FROM machines ORDER BY id`)
	if err != nil {
		return &models.DashboardSummary{
			TotalMachines:   totalMachines,
			RunningMachines: runningMachines,
			StoppedMachines: stoppedMachines,
			TodayGoodParts:  todayGood,
			TodayBadParts:   todayBad,
		}
	}
	defer rows.Close()

	var totalOEE float64
	var machineCount int
	var briefs []models.MachineBrief

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}

		var hasActiveRun bool
		var status string
		db.QueryRow(
			`SELECT COUNT(*) > 0 FROM production_runs WHERE machine_id = $1 AND status = 'running'`,
			id,
		).Scan(&hasActiveRun)
		if hasActiveRun {
			status = "running"
		} else {
			status = "stopped"
		}

		oeeResp := s.CalculateOEE(id, 24*time.Hour)
		var oeeScore float64
		if oeeResp != nil {
			oeeScore = oeeResp.OEE.Overall
		}

		briefs = append(briefs, models.MachineBrief{
			ID:       id,
			Name:     name,
			Status:   status,
			OEEScore: oeeScore,
		})
		totalOEE += oeeScore
		machineCount++
	}

	var overallOEE float64
	if machineCount > 0 {
		overallOEE = totalOEE / float64(machineCount)
	}

	return &models.DashboardSummary{
		TotalMachines:   totalMachines,
		RunningMachines: runningMachines,
		StoppedMachines: stoppedMachines,
		ActiveAlarms:    0,
		CriticalAlarms:  0,
		OverallOEE:      overallOEE,
		TodayGoodParts:  todayGood,
		TodayBadParts:   todayBad,
		MachineStates:   briefs,
	}
}
