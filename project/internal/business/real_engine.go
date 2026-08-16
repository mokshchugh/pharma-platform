package business

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"pharma-platform/internal/models"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"
)

type RealEngine struct {
	productionStore *store.ProductionStore
	machineStore    *store.MachineStore
	tagStore        *store.TagStore
	reader          *questdb.Reader
	alarmAcks       *store.AlarmAckStore
	collectorPaused func() bool
}

func NewRealEngine(cfg RealEngineConfig) *RealEngine {
	return &RealEngine{
		productionStore: cfg.ProductionStore,
		machineStore:    cfg.MachineStore,
		tagStore:        cfg.TagStore,
		reader:          cfg.Reader,
		alarmAcks:       cfg.AlarmAckStore,
		collectorPaused: cfg.CollectorPaused,
	}
}

type alarmRecord struct {
	MachineID string
	TagName   string
	Severity  string
	Active    bool
	Timestamp time.Time
}

func (e *RealEngine) listAlarms() []alarmRecord {
	if e.reader == nil {
		return nil
	}

	rows, err := e.reader.ListAlarms(context.Background(), 1000)
	if err != nil {
		return nil
	}

	var acked map[string]time.Time
	if e.alarmAcks != nil {
		acked = e.alarmAcks.AckedSet()
	}

	result := make([]alarmRecord, 0, len(rows))
	for _, r := range rows {
		key := r.MachineID + "|" + r.TagName + "|" + r.Timestamp.UTC().Format(time.RFC3339Nano)
		active := true
		if _, ok := acked[key]; ok {
			active = false
		}
		result = append(result, alarmRecord{
			MachineID: r.MachineID,
			TagName:   r.TagName,
			Severity:  r.Severity,
			Active:    active,
			Timestamp: r.Timestamp,
		})
	}

	return result
}

func activeCount(alarms []alarmRecord) int {
	n := 0
	for _, a := range alarms {
		if a.Active {
			n++
		}
	}
	return n
}

func criticalCount(alarms []alarmRecord) int {
	n := 0
	for _, a := range alarms {
		if a.Active && a.Severity == "critical" {
			n++
		}
	}
	return n
}

func alarmCountByMachine(alarms []alarmRecord) map[string]int {
	counts := make(map[string]int)
	for _, a := range alarms {
		if a.Active {
			counts[a.MachineID]++
		}
	}
	return counts
}

var tempTagCandidates = []string{"Inlet_Air_Temp", "Outlet_Air_Temp", "Product_Temp", "Exhaust_Air_Temp"}
var pressureTagCandidates = []string{"Differential_Pressure", "Gun_Air_Pressure", "Atomization_Air_Pressure"}
var alarmTagCandidates = []string{"Alarm_Status", "AlarmStatus"}

func findTag(tags []models.Tag, candidates []string) *models.Tag {
	for _, c := range candidates {
		for i := range tags {
			if tags[i].Name == c {
				return &tags[i]
			}
		}
	}
	return nil
}

func (e *RealEngine) latestTagValue(machineID int, tagName string) (float64, bool) {
	if e.reader == nil {
		return 0, false
	}
	sample, err := e.reader.LatestByPLCAndTag(context.Background(), strconv.Itoa(machineID), tagName)
	if err != nil || sample == nil {
		return 0, false
	}
	switch v := sample.Value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func mtbfMttrFromHistory(rows []questdb.MachineStateRow, until time.Time) (mtbfHours, mttrHours float64) {
	if len(rows) == 0 {
		return 0, 0
	}

	var upSeconds, downSeconds float64
	var failures int

	for i := 0; i < len(rows); i++ {
		var segEnd time.Time
		if i+1 < len(rows) {
			segEnd = rows[i+1].Timestamp
		} else {
			segEnd = until
		}

		dur := segEnd.Sub(rows[i].Timestamp).Seconds()
		if dur < 0 {
			continue
		}

		if strings.EqualFold(rows[i].State, "running") {
			upSeconds += dur
		} else {
			downSeconds += dur
			if i > 0 && strings.EqualFold(rows[i-1].State, "running") {
				failures++
			}
		}
	}

	if failures == 0 {
		return 0, 0
	}

	return upSeconds / float64(failures) / 3600, downSeconds / float64(failures) / 3600
}

func (e *RealEngine) machineMTBFMTTR(machineID int) (mtbfHours, mttrHours float64) {
	if e.reader == nil {
		return 0, 0
	}
	since := time.Now().Add(-30 * 24 * time.Hour)
	rows, err := e.reader.MachineStateHistory(context.Background(), strconv.Itoa(machineID), since)
	if err != nil {
		return 0, 0
	}
	return mtbfMttrFromHistory(rows, time.Now())
}

func (e *RealEngine) buildMachineMetrics(m store.MachineRow, alarms []alarmRecord) BusinessMetrics {
	oee := e.productionStore.CalculateOEE(m.ID, 24*time.Hour)

	metrics := BusinessMetrics{
		MachineID:   m.ID,
		MachineName: m.MachineName,
	}

	if oee != nil {
		metrics.TotalProduction = float64(oee.TotalParts)
		metrics.GoodParts = float64(oee.GoodParts)
		metrics.RejectParts = float64(oee.BadParts)
		metrics.RunningTime = float64(oee.RunTimeSec)
		metrics.IdleTime = float64(oee.DowntimeSec)
		metrics.Availability = oee.OEE.Availability
		metrics.Performance = oee.OEE.Performance
		metrics.Quality = oee.OEE.Quality
		metrics.OEE = oee.OEE.Overall
		metrics.Utilization = oee.OEE.Availability
		metrics.QualityPct = oee.OEE.Quality * 100
		metrics.RejectPct = (1 - oee.OEE.Quality) * 100

		if oee.RunTimeSec > 0 {
			metrics.ProductionRate = float64(oee.TotalParts) / (float64(oee.RunTimeSec) / 3600)
		}
		if oee.TotalParts > 0 && oee.RunTimeSec > 0 {
			metrics.CycleTime = float64(oee.RunTimeSec) / float64(oee.TotalParts)
		}
	}

	metrics.Running = e.productionStore.GetActiveRun(m.ID) != nil

	tags := e.tagStore.GetTagsByMachineID(m.ID)

	if tag := findTag(tags, alarmTagCandidates); tag != nil {
		if v, ok := e.latestTagValue(m.ID, tag.Name); ok {
			metrics.Faulted = v == 1
		}
	}

	if tag := findTag(tags, tempTagCandidates); tag != nil {
		if v, ok := e.latestTagValue(m.ID, tag.Name); ok {
			metrics.Temperature = v
		}
	}

	if tag := findTag(tags, pressureTagCandidates); tag != nil {
		if v, ok := e.latestTagValue(m.ID, tag.Name); ok {
			metrics.Pressure = v
		}
	}

	idStr := strconv.Itoa(m.ID)
	for _, a := range alarms {
		if a.MachineID == idStr && a.Active {
			metrics.AlarmCount++
		}
	}

	metrics.MTBF, metrics.MTTR = e.machineMTBFMTTR(m.ID)

	return metrics
}

func (e *RealEngine) allMachines() []store.MachineRow {
	if e.machineStore == nil {
		return nil
	}
	machines, err := e.machineStore.GetAllMachines()
	if err != nil {
		return nil
	}
	return machines
}

func (e *RealEngine) allMetrics() []BusinessMetrics {
	alarms := e.listAlarms()
	machines := e.allMachines()

	metrics := make([]BusinessMetrics, 0, len(machines))
	for _, m := range machines {
		metrics = append(metrics, e.buildMachineMetrics(m, alarms))
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].MachineID < metrics[j].MachineID
	})

	return metrics
}

func (e *RealEngine) GetOverview() *ExecutiveOverview {
	metrics := e.allMetrics()
	alarms := e.listAlarms()

	var totalProd, totalGood, totalReject float64
	var avgA, avgP, avgQ, avgOEE, avgUtil, avgM, avgMttr float64
	var totalAlarms int

	for _, m := range metrics {
		totalProd += m.TotalProduction
		totalGood += m.GoodParts
		totalReject += m.RejectParts
		totalAlarms += m.AlarmCount
		avgA += m.Availability
		avgP += m.Performance
		avgQ += m.Quality
		avgOEE += m.OEE
		avgUtil += m.Utilization
		avgM += m.MTBF
		avgMttr += m.MTTR
	}

	n := float64(len(metrics))
	if n == 0 {
		n = 1
	}

	ac := activeCount(alarms)
	cc := criticalCount(alarms)

	plcs := e.machineStore.GetPLCs()
	collecting := 0
	for _, m := range metrics {
		if m.Running {
			collecting++
		}
	}

	status := "healthy"
	if cc > 0 {
		status = "critical"
	} else if ac > 0 {
		status = "warning"
	}

	collectorStatus := "running"
	if e.collectorPaused != nil && e.collectorPaused() {
		collectorStatus = "paused"
	}

	return &ExecutiveOverview{
		PlantStatus:        status,
		CollectorStatus:    collectorStatus,
		QuestDBStatus:      "connected",
		ConfiguredMachines: len(metrics),
		CollectingMachines: collecting,
		ConfiguredPLCs:     len(plcs),
		ConfiguredTags:     len(e.tagStore.GetTags()),
		SamplesPerSec:      0,
		TelemetryToday:     int64(totalProd),
		LatestSample:       time.Now().Format(time.RFC3339),
		ActiveAlarms:       ac,
		CriticalAlarms:     cc,
		WarningAlarms:      ac - cc,
		Machines:           metrics,
		Aggregates: KPIAggregates{
			TotalProduction:  math.Round(totalProd*100) / 100,
			TotalGoodParts:   math.Round(totalGood*100) / 100,
			TotalRejectParts: math.Round(totalReject*100) / 100,
			AvgAvailability:  math.Round(avgA/n*10000) / 100,
			AvgPerformance:   math.Round(avgP/n*10000) / 100,
			AvgQuality:       math.Round(avgQ/n*10000) / 100,
			AvgOEE:           math.Round(avgOEE/n*10000) / 100,
			AvgUtilization:   math.Round(avgUtil/n*10000) / 100,
			TotalAlarms:      totalAlarms,
			TotalCritical:    cc,
			AvgMTBF:          math.Round(avgM/n*100) / 100,
			AvgMTTR:          math.Round(avgMttr/n*100) / 100,
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
		Simulated:   false,
	}
}

func (e *RealEngine) GetProductionAnalytics() *ProductionAnalytics {
	return &ProductionAnalytics{
		Hourly:     e.productionStore.ProductionByHour(),
		Daily:      e.productionStore.ProductionByDay(),
		Weekly:     e.productionStore.ProductionByWeek(),
		PerMachine: e.productionStore.ProductionByMachine(),
		Shift:      e.productionStore.ProductionByShift(),
		Batch:      e.productionStore.ProductionByBatch(),
	}
}

func (e *RealEngine) GetQualityAnalytics() *QualityAnalytics {
	good, bad := e.productionStore.QualityTotals()
	total := good + bad

	qualPct, rejPct := 0.0, 0.0
	if total > 0 {
		qualPct = good / total * 100
		rejPct = bad / total * 100
	}

	trendRows := e.productionStore.QualityTrendHourly(24)
	trend := make([]TimeSeriesPoint, 0, len(trendRows))
	for _, r := range trendRows {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: r.Timestamp.Format(time.RFC3339),
			Value:     math.Round(r.RejectPct*100) / 100,
		})
	}

	pareto := map[string]float64{
		"Temperature Deviation": 0,
		"Pressure Drop":         0,
		"Vibration Spike":       0,
		"Speed Fluctuation":     0,
	}
	for _, m := range e.allMetrics() {
		if m.Temperature > 85 {
			pareto["Temperature Deviation"] += m.RejectParts * 0.3
		}
		if m.Pressure > 15 {
			pareto["Pressure Drop"] += m.RejectParts * 0.25
		}
	}

	return &QualityAnalytics{
		QualityPct:     math.Round(qualPct*100) / 100,
		RejectPct:      math.Round(rejPct*100) / 100,
		FirstPassYield: math.Round(qualPct*100) / 100,
		RejectTrend:    trend,
		Pareto:         pareto,
		PerMachine:     e.productionStore.QualityByMachine(),
	}
}

func (e *RealEngine) GetMachineAnalytics() *MachineAnalytics {
	metrics := e.allMetrics()

	perMachine := make(map[string]BusinessMetrics, len(metrics))
	all := make([]BusinessMetrics, len(metrics))
	copy(all, metrics)
	for _, m := range metrics {
		perMachine[m.MachineName] = m
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Utilization > all[j].Utilization
	})

	top := 3
	bottom := 3
	if len(all) < top {
		top = len(all)
	}
	if len(all) < bottom {
		bottom = len(all)
	}

	return &MachineAnalytics{
		PerMachine: perMachine,
		Top:        all[:top],
		Bottom:     all[len(all)-bottom:],
	}
}

func (e *RealEngine) GetEnergyAnalytics() *EnergyAnalytics {
	metrics := e.allMetrics()

	perMachine := make(map[string]BusinessMetrics, len(metrics))
	var totalAvg, totalMax, totalCur, totalVol float64
	n := float64(len(metrics))
	if n == 0 {
		n = 1
	}

	for _, m := range metrics {
		runtimeHours := m.RunningTime / 3600
		profile := syntheticEnergyProfile(m.MachineID, m.Running, m.TotalProduction, runtimeHours)

		m.AvgPower = profile.AvgPower
		m.MaxPower = profile.MaxPower
		m.Current = profile.Current
		m.Voltage = profile.Voltage
		m.EnergyPerPart = profile.EnergyPerPart

		perMachine[m.MachineName] = m
		totalAvg += profile.AvgPower
		if profile.MaxPower > totalMax {
			totalMax = profile.MaxPower
		}
		totalCur += profile.Current
		totalVol += profile.Voltage
	}

	avgPower := totalAvg / n

	return &EnergyAnalytics{
		AvgPower:      math.Round(avgPower*100) / 100,
		MaxPower:      math.Round(totalMax*100) / 100,
		Current:       math.Round(totalCur/n*100) / 100,
		Voltage:       math.Round(totalVol/n*100) / 100,
		PowerTrend:    syntheticTrend(avgPower, 24, 1),
		EnergyTrend:   syntheticTrend(avgPower*0.9, 24, 2),
		EnergyPerPart: math.Round(avgPowerPerPart(perMachine)*100) / 100,
		PeakDemand:    math.Round(totalMax*1.2*100) / 100,
		PerMachine:    perMachine,
		Simulated:     true,
	}
}

func avgPowerPerPart(perMachine map[string]BusinessMetrics) float64 {
	var total float64
	n := float64(len(perMachine))
	if n == 0 {
		return 0
	}
	for _, m := range perMachine {
		total += m.EnergyPerPart
	}
	return total / n
}

func (e *RealEngine) GetAlarmAnalytics() *AlarmAnalytics {
	alarms := e.listAlarms()

	perMachine := make(map[string]int)
	machines := e.allMachines()
	nameByID := make(map[string]string, len(machines))
	for _, m := range machines {
		nameByID[strconv.Itoa(m.ID)] = m.MachineName
	}
	for machineID, count := range alarmCountByMachine(alarms) {
		name := nameByID[machineID]
		if name == "" {
			name = fmt.Sprintf("Machine %s", machineID)
		}
		perMachine[name] = count
	}

	trend := make([]TimeSeriesPoint, 24)
	criticalTrend := make([]TimeSeriesPoint, 24)
	now := time.Now()
	for i := 0; i < 24; i++ {
		bucketStart := now.Add(-time.Duration(23-i) * time.Hour).Truncate(time.Hour)
		trend[i] = TimeSeriesPoint{Timestamp: bucketStart.Format(time.RFC3339), Value: 0}
		criticalTrend[i] = TimeSeriesPoint{Timestamp: bucketStart.Format(time.RFC3339), Value: 0}
	}
	for _, a := range alarms {
		hoursAgo := int(now.Sub(a.Timestamp).Hours())
		idx := 23 - hoursAgo
		if idx < 0 || idx >= 24 {
			continue
		}
		trend[idx].Value++
		if a.Severity == "critical" {
			criticalTrend[idx].Value++
		}
	}

	faultFreq := 0.0
	if len(machines) > 0 {
		faultFreq = float64(len(alarms)) / float64(len(machines))
	}

	return &AlarmAnalytics{
		ActiveCount:    activeCount(alarms),
		TotalCount:     len(alarms),
		Trend:          trend,
		PerMachine:     perMachine,
		CriticalTrend:  criticalTrend,
		FaultFrequency: math.Round(faultFreq*100) / 100,
	}
}

func getMetricValue(m *BusinessMetrics, field string) float64 {
	switch field {
	case "temperature":
		return m.Temperature
	case "pressure":
		return m.Pressure
	case "vibration":
		return m.Vibration
	case "cycle_time":
		return m.CycleTime
	case "production":
		return m.TotalProduction
	case "power":
		return m.AvgPower
	case "speed":
		return m.ProductionRate
	case "reject_rate":
		return m.RejectPct
	case "oee":
		return m.OEE
	case "utilization":
		return m.Utilization
	case "downtime":
		return m.IdleTime
	case "fault_count":
		return float64(m.AlarmCount)
	default:
		return 0
	}
}

var realDataFields = map[string]bool{
	"oee": true, "utilization": true, "downtime": true, "fault_count": true,
	"cycle_time": true, "production": true, "reject_rate": true,
}

func hasRealData(m BusinessMetrics, field string) bool {
	switch field {
	case "temperature":
		return m.Temperature != 0
	case "pressure":
		return m.Pressure != 0
	case "vibration", "power", "speed":
		return false
	default:
		return realDataFields[field]
	}
}

func pearsonCorrelation(x, y []float64) float64 {
	n := float64(len(x))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	num := n*sumXY - sumX*sumY
	den := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
	if den == 0 {
		return 0
	}
	return num / den
}

func (e *RealEngine) GetCorrelationAnalysis() *CorrelationAnalysis {
	metrics := e.allMetrics()

	pairs := []CorrelationPair{
		{Label: "Temperature vs Pressure", X: "temperature", Y: "pressure"},
		{Label: "Temperature vs Vibration", X: "temperature", Y: "vibration"},
		{Label: "Pressure vs Cycle Time", X: "pressure", Y: "cycle_time"},
		{Label: "Cycle Time vs Production", X: "cycle_time", Y: "production"},
		{Label: "Power vs Load", X: "power", Y: "speed"},
		{Label: "Reject Rate vs Temperature", X: "temperature", Y: "reject_rate"},
		{Label: "OEE vs Utilization", X: "oee", Y: "utilization"},
		{Label: "Downtime vs Fault Count", X: "downtime", Y: "fault_count"},
		{Label: "Temperature vs Power", X: "temperature", Y: "power"},
		{Label: "Vibration vs Power", X: "vibration", Y: "power"},
	}

	heatmap := make(map[string]map[string]float64)

	for i, p := range pairs {
		var xVals, yVals []float64
		for _, m := range metrics {
			if !hasRealData(m, p.X) || !hasRealData(m, p.Y) {
				continue
			}
			x := getMetricValue(&m, p.X)
			y := getMetricValue(&m, p.Y)
			if x == 0 && y == 0 {
				continue
			}
			xVals = append(xVals, x)
			yVals = append(yVals, y)
		}

		if len(xVals) > 1 {
			corr := pearsonCorrelation(xVals, yVals)
			if !math.IsNaN(corr) {
				pairs[i].Correlation = math.Round(corr*1000) / 1000
				pairs[i].Available = true
			}
		}

		if heatmap[p.X] == nil {
			heatmap[p.X] = make(map[string]float64)
		}
		heatmap[p.X][p.Y] = pairs[i].Correlation
		if heatmap[p.Y] == nil {
			heatmap[p.Y] = make(map[string]float64)
		}
		heatmap[p.Y][p.X] = pairs[i].Correlation
	}

	return &CorrelationAnalysis{
		Pairs:     pairs,
		Heatmap:   heatmap,
		Available: len(metrics) > 1,
	}
}

func (e *RealEngine) GetMaintenanceAnalysis() *MaintenanceAnalysis {
	recs := make([]MaintenanceRecommendation, 0)
	var highTemp, highVib, highCur, freqFaults, outliers []MaintenanceRecommendation

	for _, m := range e.allMetrics() {
		if m.Temperature > 85 {
			r := MaintenanceRecommendation{
				MachineID: m.MachineID, MachineName: m.MachineName,
				Issue: "High Temperature", Severity: "warning",
				Recommendation: "Inspect cooling system and check for blockages",
				Metric:         "temperature", Value: math.Round(m.Temperature*10) / 10, Threshold: 85,
			}
			recs = append(recs, r)
			highTemp = append(highTemp, r)
		}

		if m.AlarmCount > 5 {
			r := MaintenanceRecommendation{
				MachineID: m.MachineID, MachineName: m.MachineName,
				Issue: "Frequent Faults", Severity: "critical",
				Recommendation: "Investigate recurring fault patterns",
				Metric:         "alarm_count", Value: float64(m.AlarmCount), Threshold: 5,
			}
			recs = append(recs, r)
			freqFaults = append(freqFaults, r)
		}

		if m.Temperature > 90 {
			r := MaintenanceRecommendation{
				MachineID: m.MachineID, MachineName: m.MachineName,
				Issue: "Abnormal Operating Condition", Severity: "critical",
				Recommendation: "Schedule immediate inspection",
				Metric:         "temperature", Value: m.Temperature, Threshold: 90,
			}
			recs = append(recs, r)
			outliers = append(outliers, r)
		}
	}

	return &MaintenanceAnalysis{
		Recommendations: recs,
		HighTemp:        highTemp,
		HighVibration:   highVib,
		HighCurrent:     highCur,
		FrequentFaults:  freqFaults,
		Outliers:        outliers,
	}
}

func (e *RealEngine) GetInsights() *InsightsAnalysis {
	metrics := e.allMetrics()

	insights := make([]Insight, 0)
	var topObs, bottlenecks, underperf, highEnergy, qualityConcerns, opRecs, bizRecs []Insight

	var totalProd, totalGood float64
	minOEE, maxOEE := 1.0, 0.0
	var minName, maxName string

	for _, m := range metrics {
		totalProd += m.TotalProduction
		totalGood += m.GoodParts

		if m.OEE < minOEE {
			minOEE = m.OEE
			minName = m.MachineName
		}
		if m.OEE > maxOEE {
			maxOEE = m.OEE
			maxName = m.MachineName
		}

		if m.Utilization < 0.5 {
			bottlenecks = append(bottlenecks, Insight{
				Category: "Production Bottleneck",
				Message:  fmt.Sprintf("%s has low utilization (%.0f%%)", m.MachineName, m.Utilization*100),
				Severity: "warning", Metric: "utilization", Value: fmt.Sprintf("%.0f%%", m.Utilization*100),
			})
		}

		if m.OEE < 0.6 && m.MachineName != "" {
			underperf = append(underperf, Insight{
				Category: "Underperforming Machine",
				Message:  fmt.Sprintf("%s running at %.0f%% OEE", m.MachineName, m.OEE*100),
				Severity: "warning", Metric: "oee", Value: fmt.Sprintf("%.0f%%", m.OEE*100),
			})
		}
	}

	if maxName != "" {
		topObs = append(topObs, Insight{
			Category: "Top Observation",
			Message:  fmt.Sprintf("%s has highest OEE at %.1f%%", maxName, maxOEE*100),
			Severity: "success", Metric: "oee", Value: fmt.Sprintf("%.1f%%", maxOEE*100),
		})
	}
	if minName != "" {
		topObs = append(topObs, Insight{
			Category: "Top Observation",
			Message:  fmt.Sprintf("%s needs attention with lowest OEE at %.1f%%", minName, minOEE*100),
			Severity: "warning", Metric: "oee", Value: fmt.Sprintf("%.1f%%", minOEE*100),
		})
	}

	overallQual := 0.0
	if totalProd > 0 {
		overallQual = totalGood / totalProd * 100
	}
	if totalProd > 0 && overallQual < 95 {
		qualityConcerns = append(qualityConcerns, Insight{
			Category: "Quality Concern",
			Message:  fmt.Sprintf("Overall quality at %.1f%% — below 95%% threshold", overallQual),
			Severity: "critical", Metric: "quality", Value: fmt.Sprintf("%.1f%%", overallQual),
		})
	}

	if minName != "" {
		opRecs = append(opRecs, Insight{
			Category: "Operational Recommendation",
			Message:  fmt.Sprintf("Review maintenance schedule for %s", minName),
			Severity: "info", Metric: "maintenance", Value: fmt.Sprintf("%.0f%% OEE", minOEE*100),
		})
	}

	if totalProd > 0 {
		bizRecs = append(bizRecs, Insight{
			Category: "Business Recommendation",
			Message:  fmt.Sprintf("Total production: %.0f units — review demand forecasting", totalProd),
			Severity: "info", Metric: "production", Value: fmt.Sprintf("%.0f units", totalProd),
		})
	}

	return &InsightsAnalysis{
		Insights:                insights,
		TopObservations:         topObs,
		ProductionBottlenecks:   bottlenecks,
		UnderperformingMachines: underperf,
		HighEnergyConsumers:     highEnergy,
		QualityConcerns:         qualityConcerns,
		OperationalRecs:         opRecs,
		BusinessRecs:            bizRecs,
	}
}
