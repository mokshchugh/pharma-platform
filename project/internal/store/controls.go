package store

import (
	"pharma-platform/internal/models"
	"pharma-platform/internal/postgres"
)

type ControlStore struct {
	client *postgres.Client
}

func NewControlStore(client *postgres.Client) *ControlStore {
	return &ControlStore{client: client}
}

func (s *ControlStore) EnsureDefaults(machineIDs []int) {
	db := s.client.DB()
	if db == nil {
		return
	}

	for _, id := range machineIDs {
		db.Exec(
			`INSERT INTO machine_control_state (machine_id, running, speed, setpoint, mode, temperature)
			 VALUES ($1, false, 0, 100, 'auto', 25)
			 ON CONFLICT (machine_id) DO NOTHING`,
			id,
		)
	}
}

func (s *ControlStore) List() []*models.MachineControlState {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	rows, err := db.Query(
		`SELECT machine_id, running, speed, setpoint, mode, temperature
		 FROM machine_control_state
		 ORDER BY machine_id`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var states []*models.MachineControlState
	for rows.Next() {
		var st models.MachineControlState
		if err := rows.Scan(&st.MachineID, &st.Running, &st.Speed, &st.Setpoint, &st.Mode, &st.Temperature); err != nil {
			continue
		}
		states = append(states, &st)
	}

	return states
}

func (s *ControlStore) Get(machineID int) *models.MachineControlState {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	var st models.MachineControlState
	err := db.QueryRow(
		`SELECT machine_id, running, speed, setpoint, mode, temperature
		 FROM machine_control_state
		 WHERE machine_id = $1`,
		machineID,
	).Scan(&st.MachineID, &st.Running, &st.Speed, &st.Setpoint, &st.Mode, &st.Temperature)
	if err != nil {
		return nil
	}

	return &st
}

func (s *ControlStore) Start(machineID int) error {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	_, err := db.Exec(
		`UPDATE machine_control_state
		 SET running = true, speed = setpoint, updated_at = now()
		 WHERE machine_id = $1`,
		machineID,
	)
	return err
}

func (s *ControlStore) Stop(machineID int) error {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	_, err := db.Exec(
		`UPDATE machine_control_state
		 SET running = false, speed = 0, updated_at = now()
		 WHERE machine_id = $1`,
		machineID,
	)
	return err
}

func (s *ControlStore) SetSetpoint(machineID int, value float64) error {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	_, err := db.Exec(
		`UPDATE machine_control_state
		 SET setpoint = $1,
		     speed = CASE WHEN running THEN $1 ELSE speed END,
		     updated_at = now()
		 WHERE machine_id = $2`,
		value, machineID,
	)
	return err
}

func (s *ControlStore) SetMode(machineID int, mode string) error {
	db := s.client.DB()
	if db == nil {
		return nil
	}

	_, err := db.Exec(
		`UPDATE machine_control_state
		 SET mode = $1, updated_at = now()
		 WHERE machine_id = $2`,
		mode, machineID,
	)
	return err
}
