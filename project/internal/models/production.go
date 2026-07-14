package models

import "time"

type ProductionRun struct {
	ID           int        `json:"id"`
	MachineID    int        `json:"machine_id"`
	MachineName  string     `json:"machine_name"`
	BatchID      string     `json:"batch_id"`
	ProductName  string     `json:"product_name"`
	TargetQty    int        `json:"target_qty"`
	GoodQty      int        `json:"good_qty"`
	BadQty       int        `json:"bad_qty"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

type DowntimeEvent struct {
	ID              int        `json:"id"`
	MachineID       int        `json:"machine_id"`
	MachineName     string     `json:"machine_name"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	Reason          string     `json:"reason"`
	Category        string     `json:"category"`
	DurationSeconds int        `json:"duration_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
}

type OEETargets struct {
	ID                           int     `json:"id"`
	MachineID                    int     `json:"machine_id"`
	AvailabilityTarget           float64 `json:"availability_target"`
	PerformanceTarget            float64 `json:"performance_target"`
	QualityTarget                float64 `json:"quality_target"`
	IdealCycleTimeSeconds        float64 `json:"ideal_cycle_time_seconds"`
	PlannedProductionTimeSeconds int     `json:"planned_production_time_seconds"`
}

type OEEValues struct {
	Availability float64 `json:"availability"`
	Performance  float64 `json:"performance"`
	Quality      float64 `json:"quality"`
	Overall      float64 `json:"overall"`
	TargetOEE    float64 `json:"target_oee"`
}

type OEEResponse struct {
	MachineID   int       `json:"machine_id"`
	MachineName string    `json:"machine_name"`
	OEE         OEEValues `json:"oee"`
	DowntimeSec int       `json:"downtime_seconds"`
	RunTimeSec  int       `json:"run_time_seconds"`
	GoodParts   int       `json:"good_parts"`
	BadParts    int       `json:"bad_parts"`
	TotalParts  int       `json:"total_parts"`
}

type ControlAction struct {
	Action  string  `json:"action"`
	Value   float64 `json:"value,omitempty"`
}

type MachineControlState struct {
	MachineID   int     `json:"machine_id"`
	Running     bool    `json:"running"`
	Speed       float64 `json:"speed"`
	Setpoint    float64 `json:"setpoint"`
	Mode        string  `json:"mode"` // auto/manual
	Temperature float64 `json:"temperature"`
}

type DashboardSummary struct {
	TotalMachines   int            `json:"total_machines"`
	RunningMachines int            `json:"running_machines"`
	StoppedMachines int            `json:"stopped_machines"`
	ActiveAlarms    int            `json:"active_alarms"`
	CriticalAlarms  int            `json:"critical_alarms"`
	OverallOEE      float64        `json:"overall_oee"`
	TodayGoodParts  int            `json:"today_good_parts"`
	TodayBadParts   int            `json:"today_bad_parts"`
	MachineStates   []MachineBrief `json:"machine_states"`
}

type MachineBrief struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Status   string  `json:"status"` // running, stopped, fault
	OEEScore float64 `json:"oee_score"`
}
