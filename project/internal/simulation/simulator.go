package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"pharma-platform/internal/models"
)

var (
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
	rngMu sync.Mutex
)

func randFloat(min, max float64) float64 {
	rngMu.Lock()
	defer rngMu.Unlock()
	return min + rng.Float64()*(max-min)
}

func randInt(min, max int) int {
	rngMu.Lock()
	defer rngMu.Unlock()
	return min + rng.Intn(max-min+1)
}

type MachineSim struct {
	MachineID   int
	MachineName string
	running     bool
	faulted     bool
	phaseStart  time.Time
	speed       float64
}

type Simulator struct {
	mu          sync.RWMutex
	machines    map[int]*MachineSim
	samplesChan chan<- models.Sample
	cycle       int
}

func New(samplesChan chan<- models.Sample) *Simulator {
	return &Simulator{
		machines:    make(map[int]*MachineSim),
		samplesChan: samplesChan,
	}
}

func (s *Simulator) Tick() {
	now := time.Now()
	s.cycle++

	for _, tag := range allTags() {
		sim := s.getOrCreateMachine(tag.MachineID, tag.MachineName)
		val := s.simulateValue(tag, sim, now)
		s.samplesChan <- models.Sample{
			Timestamp:   now,
			MachineID:   fmt.Sprintf("%d", tag.MachineID),
			MachineName: tag.MachineName,
			TagName:     tag.Name,
			Value:       val,
			Quality:     models.QualityGood,
		}
	}
}

func (s *Simulator) getOrCreateMachine(id int, name string) *MachineSim {
	s.mu.Lock()
	defer s.mu.Unlock()

	sim, ok := s.machines[id]
	if !ok {
		sim = &MachineSim{
			MachineID:   id,
			MachineName: name,
			running:     true,
			phaseStart:  time.Now(),
			speed:       75 + randFloat(0, 25),
		}
		s.machines[id] = sim
	}

	if sim.faulted && rand.Intn(200) == 0 {
		sim.faulted = false
		sim.running = true
	}

	return sim
}

func (s *Simulator) simulateValue(tag simulatedTag, sim *MachineSim, now time.Time) float64 {
	running := sim.running && !sim.faulted

	switch tag.Name {
	case "Run_Status", "RunStatus":
		if sim.faulted {
			return 0
		}
		if running {
			return 1
		}
		return 0

	case "Alarm_Status", "AlarmStatus":
		if sim.faulted {
			return 1
		}
		if rand.Intn(500) == 0 {
			return 1
		}
		return 0

	case "Door_Interlock_Status", "DoorInterlockStatus":
		return float64(randInt(0, 1))

	case "Inlet_Air_Temp", "Outlet_Air_Temp", "Exhaust_Air_Temp":
		base := 80.0
		phase := math.Sin(float64(s.cycle) * 0.001)
		ramp := math.Abs(math.Sin(float64(s.cycle) * 0.0003))
		temp := base + ramp*30 + phase*5 + randFloat(-1, 1)
		if !running {
			temp = 25 + randFloat(-2, 2)
		}
		return math.Round(temp*10) / 10

	case "Product_Temp":
		base := 55.0
		ramp := math.Abs(math.Sin(float64(s.cycle) * 0.0002))
		temp := base + ramp*40 + randFloat(-0.5, 0.5)
		if !running {
			temp = 25 + randFloat(-1, 1)
		}
		return math.Round(temp*10) / 10

	case "Differential_Pressure":
		if !running {
			return randFloat(0, 2)
		}
		return math.Round((10+math.Sin(float64(s.cycle)*0.005)*5+randFloat(-0.5, 0.5))*10) / 10

	case "Fan_Speed", "Blower_Speed":
		if !running {
			return randFloat(0, 5)
		}
		return math.Round(sim.speed + math.Sin(float64(s.cycle)*0.002)*10 + randFloat(-2, 2))

	case "Damper_Position":
		if !running {
			return 0
		}
		return math.Round(50 + math.Sin(float64(s.cycle)*0.003)*30 + randFloat(-2, 2))

	case "Spray_Rate":
		if !running {
			return 0
		}
		return math.Round((200 + math.Sin(float64(s.cycle)*0.004)*50 + randFloat(-5, 5))*10) / 10

	case "Atomization_Air_Pressure":
		if !running {
			return 0
		}
		return math.Round((2.5 + math.Sin(float64(s.cycle)*0.001)*0.5 + randFloat(-0.05, 0.05))*100) / 100

	case "Peristaltic_Pump_Speed":
		if !running {
			return 0
		}
		return math.Round(30 + math.Sin(float64(s.cycle)*0.003)*15 + randFloat(-1, 1))

	case "Pan_Speed":
		if !running {
			return 0
		}
		return math.Round(8 + math.Sin(float64(s.cycle)*0.001)*2 + randFloat(-0.5, 0.5))

	case "Main_Compression_Force", "MainCompForce":
		if !running {
			return randFloat(0, 5)
		}
		return math.Round((20 + math.Sin(float64(s.cycle)*0.005)*3 + randFloat(-0.5, 0.5))*10) / 10

	case "PreCompression_Force", "PreCompForce":
		if !running {
			return randFloat(0, 2)
		}
		return math.Round((5 + math.Sin(float64(s.cycle)*0.004)*1 + randFloat(-0.3, 0.3))*10) / 10

	case "Machine_Speed", "TurretSpeed":
		if !running {
			return 0
		}
		return math.Round(sim.speed + math.Sin(float64(s.cycle)*0.001)*5 + randFloat(-1, 1))

	case "Ejection_Force", "EjectionForce":
		if !running {
			return randFloat(0, 5)
		}
		return math.Round(50 + math.Sin(float64(s.cycle)*0.003)*10 + randFloat(-2, 2))

	case "Upper_Punch_Penetration":
		return math.Round(3 + math.Sin(float64(s.cycle)*0.001)*0.5 + randFloat(-0.1, 0.1))

	case "Avg_Tablet_Weight":
		if !running {
			return 0
		}
		return math.Round(500 + math.Sin(float64(s.cycle)*0.002)*10 + randFloat(-2, 2))

	case "Good_Count", "GoodCount", "Good_Print_Count":
		if !running {
			return 0
		}
		return float64(s.cycle / 5)

	case "Reject_Count", "RejectCount":
		if !running {
			return 0
		}
		return float64(s.cycle / 100)

	case "Hopper_Level", "HopperLevel":
		if !running {
			return 0
		}
		return math.Round(70 + math.Sin(float64(s.cycle)*0.001)*20 + randFloat(-2, 2))

	case "Ink_Level":
		return math.Round(80 + math.Sin(float64(s.cycle)*0.0005)*15 + randFloat(-1, 1))

	case "CheckWeight":
		return math.Round(500 + math.Sin(float64(s.cycle)*0.002)*5 + randFloat(-3, 3))

	case "Line_Speed":
		if !running {
			return 0
		}
		return math.Round((1.5 + math.Sin(float64(s.cycle)*0.001)*0.3 + randFloat(-0.05, 0.05))*100) / 100

	case "Reject_Reason_Code":
		return float64(randInt(0, 5))

	case "Impeller_Speed":
		if !running {
			return 0
		}
		return math.Round(150 + math.Sin(float64(s.cycle)*0.002)*30 + randFloat(-5, 5))

	case "Chopper_Speed":
		if !running {
			return 0
		}
		return math.Round(3000 + math.Sin(float64(s.cycle)*0.001)*500 + randFloat(-50, 50))

	case "Binder_Addition_Rate":
		if !running {
			return 0
		}
		return math.Round((50 + math.Sin(float64(s.cycle)*0.003)*20 + randFloat(-2, 2))*10) / 10

	case "Impeller_Motor_Load":
		if !running {
			return 0
		}
		return math.Round(60 + math.Sin(float64(s.cycle)*0.002)*15 + randFloat(-2, 2))

	case "Kneading_Timer_Elapsed":
		return float64(s.cycle) * 0.1

	case "Blender_RPM":
		if !running {
			return 0
		}
		return math.Round(25 + math.Sin(float64(s.cycle)*0.003)*5 + randFloat(-1, 1))

	case "Blend_Timer_Elapsed":
		return float64(s.cycle) * 0.1

	case "Load_Cell_Weight":
		if !running {
			return randFloat(0, 50)
		}
		return math.Round(200 + math.Sin(float64(s.cycle)*0.002)*20 + randFloat(-1, 1))

	case "Batch_ID":
		return 1001

	case "Process_Timer_Elapsed":
		return float64(s.cycle) * 0.1

	case "Bag_Shake_Count":
		return float64(s.cycle / 50)

	case "Weight_Gain_Percent":
		return math.Round((3 + math.Sin(float64(s.cycle)*0.001)*1 + randFloat(-0.1, 0.1))*10) / 10

	case "Gun_Air_Pressure":
		if !running {
			return 0
		}
		return math.Round((2.0 + math.Sin(float64(s.cycle)*0.002)*0.3 + randFloat(-0.05, 0.05))*100) / 100

	case "PrintHead_Fault":
		if sim.faulted {
			return 1
		}
		return 0

	case "Maintenance_Due_Flag":
		return 0

	case "Process_Phase":
		return float64((s.cycle / 200) % 5)

	case "Filter_DP_Alarm_Setpoint":
		return 15

	default:
		if tag.DataType == models.DataTypeBool {
			if running {
				return 1
			}
			return 0
		}
		return 42 + math.Sin(float64(s.cycle)*0.001)*10 + randFloat(-2, 2)
	}
}

type simulatedTag struct {
	models.Tag
}

var cachedTags []simulatedTag

func allTags() []simulatedTag {
	if cachedTags != nil {
		return cachedTags
	}

	entries := []struct {
		machineID int
		name      string
		machine   string
		dtype     models.DataType
	}{
		{1, "Inlet_Air_Temp", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Outlet_Air_Temp", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Product_Temp", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Differential_Pressure", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Fan_Speed", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Damper_Position", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Bag_Shake_Count", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Batch_ID", "Fluid Bed Dryer", models.DataTypeInt16},
		{1, "Process_Timer_Elapsed", "Fluid Bed Dryer", models.DataTypeInt32},
		{1, "Run_Status", "Fluid Bed Dryer", models.DataTypeBool},
		{1, "Alarm_Status", "Fluid Bed Dryer", models.DataTypeBool},
		{1, "Door_Interlock_Status", "Fluid Bed Dryer", models.DataTypeBool},
		{2, "Inlet_Air_Temp", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Outlet_Air_Temp", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Product_Temp", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Spray_Rate", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Atomization_Air_Pressure", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Peristaltic_Pump_Speed", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Differential_Pressure", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Batch_ID", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Process_Phase", "Fluid Bed Processor", models.DataTypeInt16},
		{2, "Run_Status", "Fluid Bed Processor", models.DataTypeBool},
		{2, "Alarm_Status", "Fluid Bed Processor", models.DataTypeBool},
		{3, "Inlet_Air_Temp", "Fluid Bed Equipment", models.DataTypeInt16},
		{3, "Outlet_Air_Temp", "Fluid Bed Equipment", models.DataTypeInt16},
		{3, "Blower_Speed", "Fluid Bed Equipment", models.DataTypeInt16},
		{3, "Differential_Pressure", "Fluid Bed Equipment", models.DataTypeInt16},
		{3, "Filter_DP_Alarm_Setpoint", "Fluid Bed Equipment", models.DataTypeInt16},
		{3, "Run_Status", "Fluid Bed Equipment", models.DataTypeBool},
		{3, "Alarm_Status", "Fluid Bed Equipment", models.DataTypeBool},
		{3, "Maintenance_Due_Flag", "Fluid Bed Equipment", models.DataTypeBool},
		{4, "Pan_Speed", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Spray_Rate", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Inlet_Air_Temp", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Exhaust_Air_Temp", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Product_Temp", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Gun_Air_Pressure", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Peristaltic_Pump_Speed", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Weight_Gain_Percent", "Tablet Coating Machine", models.DataTypeInt16},
		{4, "Process_Timer_Elapsed", "Tablet Coating Machine", models.DataTypeInt32},
		{4, "Run_Status", "Tablet Coating Machine", models.DataTypeBool},
		{4, "Alarm_Status", "Tablet Coating Machine", models.DataTypeBool},
		{5, "Main_Compression_Force", "Compression Machine", models.DataTypeInt16},
		{5, "PreCompression_Force", "Compression Machine", models.DataTypeInt16},
		{5, "Machine_Speed", "Compression Machine", models.DataTypeInt16},
		{5, "Ejection_Force", "Compression Machine", models.DataTypeInt16},
		{5, "Upper_Punch_Penetration", "Compression Machine", models.DataTypeInt16},
		{5, "Avg_Tablet_Weight", "Compression Machine", models.DataTypeInt16},
		{5, "Good_Count", "Compression Machine", models.DataTypeInt32},
		{5, "Reject_Count", "Compression Machine", models.DataTypeInt32},
		{5, "Hopper_Level", "Compression Machine", models.DataTypeInt16},
		{5, "Run_Status", "Compression Machine", models.DataTypeBool},
		{5, "Alarm_Status", "Compression Machine", models.DataTypeBool},
		{6, "MainCompForce", "Compression Machine", models.DataTypeFloat32},
		{6, "PreCompForce", "Compression Machine", models.DataTypeFloat32},
		{6, "TurretSpeed", "Compression Machine", models.DataTypeFloat32},
		{6, "EjectionForce", "Compression Machine", models.DataTypeFloat32},
		{6, "GoodCount", "Compression Machine", models.DataTypeInt32},
		{6, "RejectCount", "Compression Machine", models.DataTypeInt32},
		{6, "HopperLevel", "Compression Machine", models.DataTypeFloat32},
		{6, "RunStatus", "Compression Machine", models.DataTypeBool},
		{6, "AlarmStatus", "Compression Machine", models.DataTypeBool},
		{7, "Machine_Speed", "Tablet Printing Machine", models.DataTypeInt16},
		{7, "Good_Print_Count", "Tablet Printing Machine", models.DataTypeInt16},
		{7, "Reject_Count", "Tablet Printing Machine", models.DataTypeInt16},
		{7, "Ink_Level", "Tablet Printing Machine", models.DataTypeInt16},
		{7, "Batch_ID", "Tablet Printing Machine", models.DataTypeInt16},
		{7, "Run_Status", "Tablet Printing Machine", models.DataTypeBool},
		{7, "Alarm_Status", "Tablet Printing Machine", models.DataTypeBool},
		{7, "PrintHead_Fault", "Tablet Printing Machine", models.DataTypeBool},
		{8, "CheckWeight", "Capsule Checkweigher", models.DataTypeFloat32},
		{8, "Good_Count", "Capsule Checkweigher", models.DataTypeInt32},
		{8, "Reject_Count", "Capsule Checkweigher", models.DataTypeInt32},
		{8, "Line_Speed", "Capsule Checkweigher", models.DataTypeFloat32},
		{8, "Reject_Reason_Code", "Capsule Checkweigher", models.DataTypeInt16},
		{8, "Run_Status", "Capsule Checkweigher", models.DataTypeBool},
		{8, "Alarm_Status", "Capsule Checkweigher", models.DataTypeBool},
		{9, "Impeller_Speed", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Chopper_Speed", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Binder_Addition_Rate", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Product_Temp", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Impeller_Motor_Load", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Kneading_Timer_Elapsed", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Batch_ID", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Process_Phase", "Rapid Mixer Granulator", models.DataTypeInt16},
		{9, "Run_Status", "Rapid Mixer Granulator", models.DataTypeBool},
		{9, "Alarm_Status", "Rapid Mixer Granulator", models.DataTypeBool},
		{10, "Blender_RPM", "Blender", models.DataTypeInt16},
		{10, "Blend_Timer_Elapsed", "Blender", models.DataTypeInt16},
		{10, "Batch_ID", "Blender", models.DataTypeInt16},
		{10, "Load_Cell_Weight", "Blender", models.DataTypeInt16},
		{10, "Door_Interlock_Status", "Blender", models.DataTypeBool},
		{10, "Run_Status", "Blender", models.DataTypeBool},
		{10, "Alarm_Status", "Blender", models.DataTypeBool},
		{11, "Blender_RPM", "Blender", models.DataTypeInt16},
		{11, "Blend_Timer_Elapsed", "Blender", models.DataTypeInt16},
		{11, "Batch_ID", "Blender", models.DataTypeInt16},
		{11, "Load_Cell_Weight", "Blender", models.DataTypeInt16},
		{11, "Run_Status", "Blender", models.DataTypeBool},
		{11, "Alarm_Status", "Blender", models.DataTypeBool},
	}

	for _, e := range entries {
		cachedTags = append(cachedTags, simulatedTag{
			Tag: models.Tag{
				ID:          fmt.Sprintf("tag-%d-%s", e.machineID, e.name),
				PLCID:       fmt.Sprintf("machine-%d", e.machineID),
				Name:        e.name,
				MachineID:   e.machineID,
				MachineName: e.machine,
				DataType:    e.dtype,
				Enabled:     true,
			},
		})
	}

	return cachedTags
}
