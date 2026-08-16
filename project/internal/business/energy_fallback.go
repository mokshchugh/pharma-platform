package business

import (
	"math"
	"math/rand"
	"time"
)

// energyFallback generates a small, clearly-labeled synthetic power/energy
// profile per machine since no power/current/voltage/vibration tags exist
// anywhere in the schema. Values are seeded deterministically per machine ID
// so repeated calls within a process are stable-ish but still vary run to run.
type machineEnergyProfile struct {
	AvgPower      float64
	MaxPower      float64
	Current       float64
	Voltage       float64
	EnergyPerPart float64
}

func syntheticEnergyProfile(machineID int, running bool, totalParts float64, runtimeHours float64) machineEnergyProfile {
	rng := rand.New(rand.NewSource(int64(machineID)*99991 + time.Now().Unix()/3600))

	if !running {
		return machineEnergyProfile{
			AvgPower: rng.Float64() * 2,
			MaxPower: rng.Float64() * 3,
			Current:  rng.Float64() * 0.5,
			Voltage:  440 + rng.Float64()*5 - 2.5,
		}
	}

	loadFactor := 0.6 + rng.Float64()*0.4

	avgPower := 40*loadFactor + rng.Float64()*5 - 2.5
	maxPower := avgPower * (1.1 + rng.Float64()*0.2)
	current := 18*loadFactor + rng.Float64()*2 - 1
	voltage := 440 + rng.Float64()*5 - 2.5

	energyPerPart := 0.0
	if totalParts > 0 && runtimeHours > 0 {
		energyPerPart = (avgPower * runtimeHours) / totalParts
	}

	return machineEnergyProfile{
		AvgPower:      math.Round(avgPower*100) / 100,
		MaxPower:      math.Round(maxPower*100) / 100,
		Current:       math.Round(current*100) / 100,
		Voltage:       math.Round(voltage*100) / 100,
		EnergyPerPart: math.Round(energyPerPart*10000) / 10000,
	}
}

func syntheticTrend(base float64, points int, seed int64) []TimeSeriesPoint {
	rng := rand.New(rand.NewSource(seed))
	trend := make([]TimeSeriesPoint, 0, points)
	for i := 0; i < points; i++ {
		ts := time.Now().Add(-time.Duration(points-1-i) * time.Hour).Format(time.RFC3339)
		trend = append(trend, TimeSeriesPoint{
			Timestamp: ts,
			Value:     math.Round((base+rng.Float64()*base*0.3-base*0.15)*100) / 100,
		})
	}
	return trend
}
