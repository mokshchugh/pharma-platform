package business

import (
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"
)

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
}

type RealEngineConfig struct {
	ProductionStore *store.ProductionStore
	MachineStore    *store.MachineStore
	TagStore        *store.TagStore
	Reader          *questdb.Reader
	AlarmAckStore   *store.AlarmAckStore
	CollectorPaused func() bool
}

func NewEngine(cfg RealEngineConfig) Engine {
	return NewRealEngine(cfg)
}
