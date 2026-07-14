package business

type Engine interface {
	GetOverview() *ExecutiveOverview
	GetProductionAnalytics() *ProductionAnalytics
	GetQualityAnalytics() *QualityAnalytics
	GetMachineAnalytics() *MachineAnalytics
	GetEnergyAnalytics() *EnergyAnalytics
	GetAlarmAnalytics() *AlarmAnalytics
	GetCorrelationAnalysis() *CorrelationAnalysis
	GetMaintenanceAnalysis() *MaintenanceAnalysis
	GetInsights() *InsightsAnalysis
	Tick()
}

func NewEngine(cfg SimulatorConfig) Engine {
	return NewBusinessSimulator(cfg)
}
