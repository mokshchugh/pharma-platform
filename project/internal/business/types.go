package business

type BusinessMetrics struct {
	MachineID      int     `json:"machine_id"`
	MachineName    string  `json:"machine_name"`
	Running        bool    `json:"running"`
	Faulted        bool    `json:"faulted"`
	TotalProduction float64 `json:"total_production"`
	GoodParts      float64 `json:"good_parts"`
	RejectParts    float64 `json:"reject_parts"`
	ProductionRate float64 `json:"production_rate"`
	RunningTime    float64 `json:"running_time_sec"`
	IdleTime       float64 `json:"idle_time_sec"`
	Availability   float64 `json:"availability"`
	Performance    float64 `json:"performance"`
	Quality        float64 `json:"quality"`
	OEE            float64 `json:"oee"`
	Utilization    float64 `json:"utilization"`
	QualityPct     float64 `json:"quality_pct"`
	RejectPct      float64 `json:"reject_pct"`
	AvgPower       float64 `json:"avg_power"`
	MaxPower       float64 `json:"max_power"`
	EnergyPerPart  float64 `json:"energy_per_part"`
	Current        float64 `json:"current"`
	Voltage        float64 `json:"voltage"`
	AlarmCount     int     `json:"alarm_count"`
	FaultFrequency float64 `json:"fault_frequency"`
	MTBF           float64 `json:"mtbf_hours"`
	MTTR           float64 `json:"mttr_hours"`
	Temperature    float64 `json:"temperature"`
	Vibration      float64 `json:"vibration"`
	Pressure       float64 `json:"pressure"`
	CycleTime      float64 `json:"cycle_time_sec"`
}

type ExecutiveOverview struct {
	PlantStatus       string            `json:"plant_status"`
	CollectorStatus   string            `json:"collector_status"`
	QuestDBStatus     string            `json:"questdb_status"`
	ConfiguredMachines int              `json:"configured_machines"`
	CollectingMachines int              `json:"collecting_machines"`
	ConfiguredPLCs    int               `json:"configured_plcs"`
	ConfiguredTags    int               `json:"configured_tags"`
	SamplesPerSec     float64           `json:"samples_per_sec"`
	TelemetryToday    int64             `json:"telemetry_today"`
	LatestSample      string            `json:"latest_sample"`
	ActiveAlarms      int               `json:"active_alarms"`
	CriticalAlarms    int               `json:"critical_alarms"`
	WarningAlarms     int               `json:"warning_alarms"`
	Machines          []BusinessMetrics `json:"machines"`
	Aggregates        KPIAggregates     `json:"aggregates"`
	GeneratedAt       string            `json:"generated_at"`
	Simulated         bool              `json:"simulated"`
}

type KPIAggregates struct {
	TotalProduction  float64 `json:"total_production"`
	TotalGoodParts   float64 `json:"total_good_parts"`
	TotalRejectParts float64 `json:"total_reject_parts"`
	AvgAvailability  float64 `json:"avg_availability"`
	AvgPerformance   float64 `json:"avg_performance"`
	AvgQuality       float64 `json:"avg_quality"`
	AvgOEE           float64 `json:"avg_oee"`
	AvgUtilization   float64 `json:"avg_utilization"`
	TotalAlarms      int     `json:"total_alarms"`
	TotalCritical    int     `json:"total_critical"`
	AvgPower         float64 `json:"avg_power"`
	PeakPower        float64 `json:"peak_power"`
	AvgEnergyPerPart float64 `json:"avg_energy_per_part"`
	AvgMTBF          float64 `json:"avg_mtbf_hours"`
	AvgMTTR          float64 `json:"avg_mttr_hours"`
}

type TimeSeriesPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

type ProductionAnalytics struct {
	Hourly      map[string]float64            `json:"hourly"`
	Daily       map[string]float64            `json:"daily"`
	Weekly      map[string]float64            `json:"weekly"`
	PerMachine  map[string]float64            `json:"per_machine"`
	Shift       map[string]float64            `json:"shift"`
	Batch       map[string]float64            `json:"batch"`
}

type QualityAnalytics struct {
	QualityPct        float64                        `json:"quality_pct"`
	RejectPct         float64                        `json:"reject_pct"`
	FirstPassYield    float64                        `json:"first_pass_yield"`
	RejectTrend       []TimeSeriesPoint              `json:"reject_trend"`
	Pareto            map[string]float64             `json:"pareto"`
	PerMachine        map[string]float64             `json:"per_machine"`
}

type MachineAnalytics struct {
	PerMachine map[string]BusinessMetrics `json:"per_machine"`
	Top        []BusinessMetrics          `json:"top_utilized"`
	Bottom     []BusinessMetrics          `json:"least_utilized"`
}

type EnergyAnalytics struct {
	AvgPower        float64              `json:"avg_power"`
	MaxPower        float64              `json:"max_power"`
	Current         float64              `json:"current"`
	Voltage         float64              `json:"voltage"`
	PowerTrend      []TimeSeriesPoint    `json:"power_trend"`
	EnergyTrend     []TimeSeriesPoint    `json:"energy_trend"`
	EnergyPerPart   float64              `json:"energy_per_part"`
	PeakDemand      float64              `json:"peak_demand"`
	PerMachine      map[string]BusinessMetrics `json:"per_machine"`
}

type AlarmAnalytics struct {
	ActiveCount      int                `json:"active_count"`
	TotalCount       int                `json:"total_count"`
	Trend            []TimeSeriesPoint  `json:"trend"`
	PerMachine       map[string]int     `json:"per_machine"`
	CriticalTrend    []TimeSeriesPoint  `json:"critical_trend"`
	FaultFrequency   float64            `json:"fault_frequency"`
}

type CorrelationPair struct {
	X          string  `json:"x"`
	Y          string  `json:"y"`
	Correlation float64 `json:"correlation"`
	Available  bool    `json:"available"`
	Label      string  `json:"label"`
}

type CorrelationAnalysis struct {
	Pairs    []CorrelationPair        `json:"pairs"`
	Heatmap  map[string]map[string]float64 `json:"heatmap"`
	Available bool                    `json:"available"`
}

type MaintenanceRecommendation struct {
	MachineID   int    `json:"machine_id"`
	MachineName string `json:"machine_name"`
	Issue       string `json:"issue"`
	Severity    string `json:"severity"`
	Recommendation string `json:"recommendation"`
	Metric      string `json:"metric"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
}

type MaintenanceAnalysis struct {
	Recommendations []MaintenanceRecommendation `json:"recommendations"`
	HighTemp       []MaintenanceRecommendation  `json:"high_temperature"`
	HighVibration  []MaintenanceRecommendation  `json:"high_vibration"`
	HighCurrent    []MaintenanceRecommendation  `json:"high_current"`
	FrequentFaults []MaintenanceRecommendation  `json:"frequent_faults"`
	Outliers       []MaintenanceRecommendation  `json:"outliers"`
}

type Insight struct {
	Category    string `json:"category"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Metric      string `json:"metric"`
	Value       string `json:"value"`
}

type InsightsAnalysis struct {
	Insights                []Insight `json:"insights"`
	TopObservations         []Insight `json:"top_observations"`
	ProductionBottlenecks   []Insight `json:"production_bottlenecks"`
	UnderperformingMachines []Insight `json:"underperforming_machines"`
	HighEnergyConsumers     []Insight `json:"high_energy_consumers"`
	QualityConcerns         []Insight `json:"quality_concerns"`
	OperationalRecs         []Insight `json:"operational_recommendations"`
	BusinessRecs            []Insight `json:"business_recommendations"`
}
