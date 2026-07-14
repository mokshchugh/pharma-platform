package business

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// AlarmEvent is a business-level alarm representation used by the simulator.
type AlarmEvent struct {
	ID        string    `json:"id"`
	PLCID     string    `json:"plc_id"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type BusinessSimulator struct {
	mu       sync.RWMutex
	metrics  map[int]*BusinessMetrics
	history  map[int][]BusinessMetrics
	alarms   []AlarmEvent
	started  time.Time

	rng      *rand.Rand
	machineSims map[int]*machineSimState

	config SimulatorConfig
}

type SimulatorConfig struct {
	MachineCount    int
	AlarmStore      interface{ ActiveCount() int; CriticalCount() int }
	CollectorPaused func() bool
}

type machineSimState struct {
	running       bool
	faulted       bool
	runningSince  time.Time
	faultCount    int
	totalFaults   int
	lastFaultTime time.Time
	totalDowntime float64
	totalRuntime  float64
	goodParts     float64
	badParts      float64
	speed         float64
	temperature   float64
	pressure      float64
	vibration     float64
	power         float64
	current       float64
	voltage       float64
	cycleTime     float64
	faultRate     float64

	hourlyProd    map[int]float64
	dailyProd     map[string]float64
	shiftProd     map[string]float64
	lastHour      int
	lastDay       string
	lastShift     string

	powerHistory   []float64
	productionHistory []float64
	rejectHistory  []float64

	corrTempPress  []struct{ t, p float64 }
}

func NewBusinessSimulator(cfg SimulatorConfig) *BusinessSimulator {
	s := &BusinessSimulator{
		metrics:     make(map[int]*BusinessMetrics),
		history:     make(map[int][]BusinessMetrics),
		alarms:      make([]AlarmEvent, 0),
		started:     time.Now(),
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		machineSims: make(map[int]*machineSimState),
		config:      cfg,
	}

	for i := 1; i <= cfg.MachineCount; i++ {
		names := map[int]string{
			1: "Fluid Bed Dryer", 2: "Fluid Bed Processor", 3: "Fluid Bed Equipment",
			4: "Tablet Coating Machine", 5: "Compression Machine", 6: "Compression Machine 2",
			7: "Tablet Printing Machine", 8: "Capsule Checkweigher", 9: "Rapid Mixer Granulator",
			10: "Blender 1", 11: "Blender 2",
		}
		name := names[i]
		if name == "" {
			name = fmt.Sprintf("Machine %d", i)
		}
		s.machineSims[i] = &machineSimState{
			running:    true,
			speed:      75 + s.rng.Float64()*25,
			temperature: 25 + s.rng.Float64()*60,
			pressure:   5 + s.rng.Float64()*15,
			vibration:  s.rng.Float64() * 3,
			power:      10 + s.rng.Float64()*40,
			current:    5 + s.rng.Float64()*20,
			voltage:    440 + s.rng.Float64()*20,
			cycleTime:  0.5 + s.rng.Float64()*2,
			faultRate:  s.rng.Float64() * 0.05,
			hourlyProd: make(map[int]float64),
			dailyProd:  make(map[string]float64),
			shiftProd:  make(map[string]float64),
		}
		s.metrics[i] = &BusinessMetrics{
			MachineID:   i,
			MachineName: name,
			Running:     true,
		}
	}

	return s
}

func (s *BusinessSimulator) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	currentHour := now.Hour()
	currentDay := now.Format("2006-01-02")
	currentShift := s.shiftFor(now)

	for id, sim := range s.machineSims {
		s.tickMachine(id, sim, now, currentHour, currentDay, currentShift)
	}

	s.pruneHistory()
}

func (s *BusinessSimulator) tickMachine(id int, sim *machineSimState, now time.Time, hour int, day string, shift string) {
	paused := s.config.CollectorPaused != nil && s.config.CollectorPaused()
	if paused {
		return
	}

	if sim.running && sim.faulted {
		sim.faulted = false
	}

	faultRoll := s.rng.Float64()
	if sim.running && !sim.faulted && faultRoll < sim.faultRate*0.01 {
		sim.faulted = true
		sim.faultCount++
		sim.totalFaults++
		sim.lastFaultTime = now
	}

	if sim.faulted {
		sim.running = false
		recoveryRoll := s.rng.Float64()
		if recoveryRoll < 0.05 {
			sim.faulted = false
			sim.running = true
			sim.runningSince = now
		}
	}

	tickDuration := 5.0 / 3600.0

	if sim.running {
		sim.totalRuntime += tickDuration

		loadFactor := sim.speed / 100.0
		sim.temperature = 80*loadFactor + s.rng.Float64()*5 - 2.5
		sim.pressure = 12*loadFactor + s.rng.Float64()*2 - 1
		sim.vibration = s.rng.Float64()*2 + loadFactor*1.5
		if sim.faulted {
			sim.vibration += 3
		}

		sim.power = 40*loadFactor + s.rng.Float64()*5 - 2.5
		sim.current = 18*loadFactor + s.rng.Float64()*2 - 1
		sim.voltage = 440 + s.rng.Float64()*5 - 2.5
		sim.cycleTime = 1.5 - loadFactor*0.8 + s.rng.Float64()*0.2 - 0.1

		prodRate := sim.speed * 0.1
		goodRate := prodRate * (1 - 0.02 - loadFactor*0.02)
		badRate := prodRate * (0.02 + loadFactor*0.02)

		if sim.faulted {
			goodRate = 0
			badRate = 0
		}

		newGood := goodRate * tickDuration
		newBad := badRate * tickDuration
		sim.goodParts += newGood
		sim.badParts += newBad

		sim.hourlyProd[hour] += newGood + newBad
		sim.dailyProd[day] += newGood + newBad
		sim.shiftProd[shift] += newGood + newBad

		sim.productionHistory = append(sim.productionHistory, newGood+newBad)
		sim.rejectHistory = append(sim.rejectHistory, newBad)
	} else {
		sim.totalDowntime += tickDuration

		sim.temperature = 25 + s.rng.Float64()*3
		sim.pressure = s.rng.Float64() * 2
		sim.vibration = s.rng.Float64() * 0.3
		sim.power = s.rng.Float64() * 2
		sim.current = s.rng.Float64() * 0.5
	}

	sim.powerHistory = append(sim.powerHistory, sim.power)

	totalTime := sim.totalRuntime + sim.totalDowntime
	avail := 1.0
	if totalTime > 0 {
		avail = sim.totalRuntime / totalTime
	}

	perf := 1.0
	totalParts := sim.goodParts + sim.badParts
	if totalParts > 0 && sim.totalRuntime > 0 {
		actualRate := totalParts / sim.totalRuntime
		idealRate := 100.0
		perf = actualRate / idealRate
		if perf > 1 {
			perf = 1
		}
	}

	qual := 1.0
	if totalParts > 0 {
		qual = sim.goodParts / totalParts
	}

	oee := avail * perf * qual

	util := avail

	mtbf := 168.0
	if sim.faultCount > 0 {
		mtbf = sim.totalRuntime / float64(sim.faultCount) * 3600
	}
	mttr := 1.0
	if sim.faultCount > 0 {
		mttr = sim.totalDowntime / float64(sim.faultCount) * 3600
	}

	count := s.metrics[id]
	count.Running = sim.running
	count.Faulted = sim.faulted
	count.TotalProduction = totalParts
	count.GoodParts = sim.goodParts
	count.RejectParts = sim.badParts
	count.ProductionRate = sim.speed * 0.1
	count.RunningTime = sim.totalRuntime * 3600
	count.IdleTime = sim.totalDowntime * 3600
	count.Availability = avail
	count.Performance = perf
	count.Quality = qual
	count.OEE = oee
	count.Utilization = util
	count.QualityPct = qual * 100
	count.RejectPct = (1 - qual) * 100
	count.AvgPower = sim.power
	count.MaxPower = sim.maxPower()
	count.EnergyPerPart = s.energyPerPart(sim)
	count.Current = sim.current
	count.Voltage = sim.voltage
	count.Temperature = sim.temperature
	count.Vibration = sim.vibration
	count.Pressure = sim.pressure
	count.CycleTime = sim.cycleTime
	count.AlarmCount = sim.faultCount
	count.FaultFrequency = sim.faultRate * 100
	count.MTBF = mtbf
	count.MTTR = mttr

	s.history[id] = append(s.history[id], *count)
}

func (s *BusinessSimulator) shiftFor(t time.Time) string {
	h := t.Hour()
	switch {
	case h >= 6 && h < 14:
		return "Shift A (06-14)"
	case h >= 14 && h < 22:
		return "Shift B (14-22)"
	default:
		return "Shift C (22-06)"
	}
}

func (s *machineSimState) maxPower() float64 {
	max := 0.0
	for _, p := range s.powerHistory {
		if p > max {
			max = p
		}
	}
	return max
}

func (s *BusinessSimulator) energyPerPart(sim *machineSimState) float64 {
	total := sim.goodParts + sim.badParts
	if total == 0 {
		return 0
	}
	avgP := sim.power
	hours := sim.totalRuntime
	if hours == 0 {
		return 0
	}
	energyKWh := avgP * hours
	return energyKWh / total
}

func (s *BusinessSimulator) pruneHistory() {
	for id := range s.history {
		if len(s.history[id]) > 100 {
			s.history[id] = s.history[id][len(s.history[id])-100:]
		}
	}
}

func (s *BusinessSimulator) GetOverview() *ExecutiveOverview {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalProd, totalGood, totalReject float64
	var avgA, avgP, avgQ, avgOEE, avgUtil, avgPwr, avgEP, avgM float64
	var totalAlarms, totalCritical int
	var peakPwr float64

	for _, m := range s.metrics {
		totalProd += m.TotalProduction
		totalGood += m.GoodParts
		totalReject += m.RejectParts
		totalAlarms += m.AlarmCount
		avgA += m.Availability
		avgP += m.Performance
		avgQ += m.Quality
		avgOEE += m.OEE
		avgUtil += m.Utilization
		avgPwr += m.AvgPower
		avgEP += m.EnergyPerPart
		avgM += m.MTBF
		if m.MaxPower > peakPwr {
			peakPwr = m.MaxPower
		}
	}

	n := float64(len(s.metrics))
	ac := 0
	cc := 0
	if s.config.AlarmStore != nil {
		ac = s.config.AlarmStore.ActiveCount()
		cc = s.config.AlarmStore.CriticalCount()
	}

	machines := make([]BusinessMetrics, 0, len(s.metrics))
	for _, m := range s.metrics {
		machines = append(machines, *m)
	}
	sort.Slice(machines, func(i, j int) bool {
		return machines[i].MachineID < machines[j].MachineID
	})

	return &ExecutiveOverview{
		PlantStatus:       s.plantStatus(ac, cc),
		CollectorStatus:   s.collectorStatus(),
		QuestDBStatus:     "connected",
		ConfiguredMachines: len(s.metrics),
		CollectingMachines: s.collectingCount(),
		ConfiguredPLCs:    len(s.metrics),
		ConfiguredTags:    100,
		SamplesPerSec:     2000,
		TelemetryToday:    int64(totalProd),
		LatestSample:      time.Now().Format(time.RFC3339),
		ActiveAlarms:      ac,
		CriticalAlarms:    cc,
		WarningAlarms:     ac - cc,
		Machines:          machines,
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
			TotalCritical:    totalCritical,
			AvgPower:         math.Round(avgPwr/n*100) / 100,
			PeakPower:        math.Round(peakPwr*100) / 100,
			AvgEnergyPerPart: math.Round(avgEP/n*10000) / 100,
			AvgMTBF:          math.Round(avgM/n*100) / 100,
			AvgMTTR:          math.Round(mttrAvg(s.metrics)*100) / 100,
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
		Simulated:   true,
	}
}

func (s *BusinessSimulator) plantStatus(ac, cc int) string {
	if cc > 0 {
		return "critical"
	}
	if ac > 0 {
		return "warning"
	}
	return "healthy"
}

func (s *BusinessSimulator) collectorStatus() string {
	if s.config.CollectorPaused != nil && s.config.CollectorPaused() {
		return "paused"
	}
	return "running"
}

func (s *BusinessSimulator) collectingCount() int {
	count := 0
	for _, m := range s.machineSims {
		if m.running {
			count++
		}
	}
	return count
}

func (s *BusinessSimulator) GetProductionAnalytics() *ProductionAnalytics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hourly := make(map[string]float64)
	daily := make(map[string]float64)
	perMachine := make(map[string]float64)
	shift := make(map[string]float64)
	batch := make(map[string]float64)

	for id, sim := range s.machineSims {
		metrics := s.metrics[id]
		perMachine[metrics.MachineName] = metrics.TotalProduction

		for h, v := range sim.hourlyProd {
			key := fmt.Sprintf("%02d:00", h)
			hourly[key] += v
		}
		for d, v := range sim.dailyProd {
			daily[d] += v
		}
		for sh, v := range sim.shiftProd {
			shift[sh] += v
		}
	}

	batch["Batch-2401"] = hourlyTotal(s.machineSims)

	return &ProductionAnalytics{
		Hourly:     hourly,
		Daily:      daily,
		PerMachine: perMachine,
		Shift:      shift,
		Batch:      batch,
	}
}

func (s *BusinessSimulator) GetQualityAnalytics() *QualityAnalytics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalGood, totalReject float64
	perMachine := make(map[string]float64)
	pareto := map[string]float64{
		"Temperature Deviation": 0,
		"Pressure Drop":        0,
		"Vibration Spike":      0,
		"Speed Fluctuation":    0,
	}

	for _, m := range s.metrics {
		totalGood += m.GoodParts
		totalReject += m.RejectParts
		perMachine[m.MachineName] = m.RejectPct

		if m.Temperature > 85 {
			pareto["Temperature Deviation"] += m.RejectParts * 0.3
		}
		if m.Pressure > 15 {
			pareto["Pressure Drop"] += m.RejectParts * 0.25
		}
		if m.Vibration > 4 {
			pareto["Vibration Spike"] += m.RejectParts * 0.25
		}
		if m.CycleTime < 0.5 || m.CycleTime > 2.5 {
			pareto["Speed Fluctuation"] += m.RejectParts * 0.2
		}
	}

	total := totalGood + totalReject
	qualPct := 0.0
	rejPct := 0.0
	fpy := 0.0
	if total > 0 {
		qualPct = (totalGood / total) * 100
		rejPct = (totalReject / total) * 100
		fpy = qualPct
	}

	trend := make([]TimeSeriesPoint, 0)
	for i := 0; i < 24; i++ {
		trend = append(trend, TimeSeriesPoint{
			Timestamp: time.Now().Add(-time.Duration(23-i)*time.Hour).Format(time.RFC3339),
			Value:     math.Round(rejPct*0.8+s.rng.Float64()*rejPct*0.4) / 100,
		})
	}

	return &QualityAnalytics{
		QualityPct:     math.Round(qualPct*100) / 100,
		RejectPct:      math.Round(rejPct*100) / 100,
		FirstPassYield: math.Round(fpy*100) / 100,
		RejectTrend:    trend,
		Pareto:         pareto,
		PerMachine:     perMachine,
	}
}

func (s *BusinessSimulator) GetMachineAnalytics() *MachineAnalytics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	perMachine := make(map[string]BusinessMetrics)
	var all []BusinessMetrics

	for _, m := range s.metrics {
		perMachine[m.MachineName] = *m
		all = append(all, *m)
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

func (s *BusinessSimulator) GetEnergyAnalytics() *EnergyAnalytics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	perMachine := make(map[string]BusinessMetrics)
	var totalAvg, totalMax, totalCur, totalVol float64
	n := float64(len(s.metrics))

	for _, m := range s.metrics {
		perMachine[m.MachineName] = *m
		totalAvg += m.AvgPower
		if m.MaxPower > totalMax {
			totalMax = m.MaxPower
		}
		totalCur += m.Current
		totalVol += m.Voltage
	}

	powerTrend := make([]TimeSeriesPoint, 0)
	energyTrend := make([]TimeSeriesPoint, 0)
	for i := 0; i < 24; i++ {
		ts := time.Now().Add(-time.Duration(23-i)*time.Hour).Format(time.RFC3339)
		basePwr := totalAvg / n
		powerTrend = append(powerTrend, TimeSeriesPoint{
			Timestamp: ts,
			Value:     math.Round((basePwr + s.rng.Float64()*10 - 5) * 100) / 100,
		})
		energyTrend = append(energyTrend, TimeSeriesPoint{
			Timestamp: ts,
			Value:     math.Round((basePwr*0.8+s.rng.Float64()*basePwr*0.4)*100) / 100,
		})
	}

	return &EnergyAnalytics{
		AvgPower:      math.Round(totalAvg/n*100) / 100,
		MaxPower:      math.Round(totalMax*100) / 100,
		Current:       math.Round(totalCur/n*100) / 100,
		Voltage:       math.Round(totalVol/n*100) / 100,
		PowerTrend:    powerTrend,
		EnergyTrend:   energyTrend,
		EnergyPerPart: math.Round(s.metrics[1].EnergyPerPart*100) / 100,
		PeakDemand:    math.Round(totalMax*1.2*100) / 100,
		PerMachine:    perMachine,
	}
}

func (s *BusinessSimulator) GetAlarmAnalytics() *AlarmAnalytics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ac := 0
	cc := 0
	if s.config.AlarmStore != nil {
		ac = s.config.AlarmStore.ActiveCount()
		cc = s.config.AlarmStore.CriticalCount()
	}

	perMachine := make(map[string]int)
	for _, m := range s.metrics {
		perMachine[m.MachineName] = m.AlarmCount
	}

	trend := make([]TimeSeriesPoint, 0)
	criticalTrend := make([]TimeSeriesPoint, 0)
	for i := 0; i < 24; i++ {
		ts := time.Now().Add(-time.Duration(23-i)*time.Hour).Format(time.RFC3339)
		trend = append(trend, TimeSeriesPoint{
			Timestamp: ts,
			Value:     float64(s.rng.Intn(5)),
		})
		criticalTrend = append(criticalTrend, TimeSeriesPoint{
			Timestamp: ts,
			Value:     float64(s.rng.Intn(2)),
		})
	}

	return &AlarmAnalytics{
		ActiveCount:    ac,
		TotalCount:     ac + cc,
		Trend:          trend,
		PerMachine:     perMachine,
		CriticalTrend:  criticalTrend,
		FaultFrequency: avgFaultFreq(s.metrics),
	}
}

func (s *BusinessSimulator) GetCorrelationAnalysis() *CorrelationAnalysis {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	var ids []int
	for id := range s.metrics {
		ids = append(ids, id)
	}

	for i, p := range pairs {
		vals := make([]float64, 0)
		var xVals, yVals []float64

		for _, id := range ids {
			m := s.metrics[id]
			x := getMetricValue(m, p.X)
			y := getMetricValue(m, p.Y)
			if x != 0 || y != 0 {
				xVals = append(xVals, x)
				yVals = append(yVals, y)
			}
		}

		if len(xVals) > 1 {
			corr := pearsonCorrelation(xVals, yVals)
			pairs[i].Correlation = math.Round(corr*1000) / 1000
			pairs[i].Available = !math.IsNaN(corr)
			vals = xVals
		}

		_ = vals

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
		Available: len(ids) > 1,
	}
}

func (s *BusinessSimulator) GetMaintenanceAnalysis() *MaintenanceAnalysis {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recs := make([]MaintenanceRecommendation, 0)
	var highTemp, highVib, highCur, freqFaults, outliers []MaintenanceRecommendation

	for _, m := range s.metrics {
		if m.Temperature > 85 {
			r := MaintenanceRecommendation{
				MachineID:   m.MachineID,
				MachineName: m.MachineName,
				Issue:       "High Temperature",
				Severity:    "warning",
				Recommendation: "Inspect cooling system and check for blockages",
				Metric:      "temperature",
				Value:       math.Round(m.Temperature*10) / 10,
				Threshold:   85,
			}
			recs = append(recs, r)
			highTemp = append(highTemp, r)
		}

		if m.Vibration > 4 {
			r := MaintenanceRecommendation{
				MachineID:   m.MachineID,
				MachineName: m.MachineName,
				Issue:       "High Vibration",
				Severity:    "warning",
				Recommendation: "Inspect bearings and alignment",
				Metric:      "vibration",
				Value:       math.Round(m.Vibration*10) / 10,
				Threshold:   4,
			}
			recs = append(recs, r)
			highVib = append(highVib, r)
		}

		if m.Current > 30 {
			r := MaintenanceRecommendation{
				MachineID:   m.MachineID,
				MachineName: m.MachineName,
				Issue:       "High Current Draw",
				Severity:    "warning",
				Recommendation: "Inspect motor and electrical system",
				Metric:      "current",
				Value:       math.Round(m.Current*10) / 10,
				Threshold:   30,
			}
			recs = append(recs, r)
			highCur = append(highCur, r)
		}

		if m.FaultFrequency > 2 {
			r := MaintenanceRecommendation{
				MachineID:   m.MachineID,
				MachineName: m.MachineName,
				Issue:       "Frequent Faults",
				Severity:    "critical",
				Recommendation: "Investigate recurring fault patterns",
				Metric:      "fault_frequency",
				Value:       math.Round(m.FaultFrequency*10) / 10,
				Threshold:   2,
			}
			recs = append(recs, r)
			freqFaults = append(freqFaults, r)
		}

		if m.Temperature > 90 || m.Vibration > 5 {
			r := MaintenanceRecommendation{
				MachineID:   m.MachineID,
				MachineName: m.MachineName,
				Issue:       "Abnormal Operating Condition",
				Severity:    "critical",
				Recommendation: "Schedule immediate inspection",
				Metric:      "combined",
				Value:       math.Max(m.Temperature, m.Vibration*10),
				Threshold:   90,
			}
			outliers = append(outliers, r)
			if m.Temperature > 90 || m.Vibration > 5 {
				recs = append(recs, r)
			}
		}
	}

	return &MaintenanceAnalysis{
		Recommendations: recs,
		HighTemp:       highTemp,
		HighVibration:  highVib,
		HighCurrent:    highCur,
		FrequentFaults: freqFaults,
		Outliers:       outliers,
	}
}

func (s *BusinessSimulator) GetInsights() *InsightsAnalysis {
	s.mu.RLock()
	defer s.mu.RUnlock()

	insights := make([]Insight, 0)
	var topObs, bottlenecks, underperf, highEnergy, qualityConcerns, opRecs, bizRecs []Insight

	var totalProd, totalGood, totalRej float64
	var minOEE, maxOEE float64 = 1, 0
	var minName, maxName string

	for _, m := range s.metrics {
		totalProd += m.TotalProduction
		totalGood += m.GoodParts
		totalRej += m.RejectParts

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
				Severity: "warning",
				Metric:   "utilization",
				Value:    fmt.Sprintf("%.0f%%", m.Utilization*100),
			})
		}

		if m.AvgPower > 30 {
			highEnergy = append(highEnergy, Insight{
				Category: "High Energy Consumer",
				Message:  fmt.Sprintf("%s consuming %.1f kW average", m.MachineName, m.AvgPower),
				Severity: "info",
				Metric:   "avg_power",
				Value:    fmt.Sprintf("%.1f kW", m.AvgPower),
			})
		}
	}

	topObs = append(topObs, Insight{
		Category: "Top Observation",
		Message:  fmt.Sprintf("%s has highest OEE at %.1f%%", maxName, maxOEE*100),
		Severity: "success",
		Metric:   "oee",
		Value:    fmt.Sprintf("%.1f%%", maxOEE*100),
	})

	topObs = append(topObs, Insight{
		Category: "Top Observation",
		Message:  fmt.Sprintf("%s needs attention with lowest OEE at %.1f%%", minName, minOEE*100),
		Severity: "warning",
		Metric:   "oee",
		Value:    fmt.Sprintf("%.1f%%", minOEE*100),
	})

	overallQual := (totalGood / (totalProd + 1)) * 100
	if overallQual < 95 {
		qualityConcerns = append(qualityConcerns, Insight{
			Category: "Quality Concern",
			Message:  fmt.Sprintf("Overall quality at %.1f%% — below 95%% threshold", overallQual),
			Severity: "critical",
			Metric:   "quality",
			Value:    fmt.Sprintf("%.1f%%", overallQual),
		})
	}

	opRecs = append(opRecs, Insight{
		Category: "Operational Recommendation",
		Message:  fmt.Sprintf("Review maintenance schedule for %s", minName),
		Severity: "info",
		Metric:   "maintenance",
		Value:    fmt.Sprintf("%.0f%% OEE", minOEE*100),
	})

	bizRecs = append(bizRecs, Insight{
		Category: "Business Recommendation",
		Message:  fmt.Sprintf("Total production: %.0f units — review demand forecasting", totalProd),
		Severity: "info",
		Metric:   "production",
		Value:    fmt.Sprintf("%.0f units", totalProd),
	})

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

func hourlyTotal(sims map[int]*machineSimState) float64 {
	var t float64
	for _, s := range sims {
		for _, v := range s.hourlyProd {
			t += v
		}
	}
	return t
}

func avgFaultFreq(metrics map[int]*BusinessMetrics) float64 {
	var total float64
	n := float64(len(metrics))
	if n == 0 {
		return 0
	}
	for _, m := range metrics {
		total += m.FaultFrequency
	}
	return math.Round(total/n*100) / 100
}

func mttrAvg(metrics map[int]*BusinessMetrics) float64 {
	var total float64
	n := float64(len(metrics))
	if n == 0 {
		return 0
	}
	for _, m := range metrics {
		total += m.MTTR
	}
	return total / n
}
