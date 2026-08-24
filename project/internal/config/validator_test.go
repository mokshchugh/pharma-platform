package config

import "testing"

func TestValidateDefaultsQuestDB(t *testing.T) {
	cfg := &Config{Plant: PlantConfig{Name: "Test Plant"}}

	if err := Validate(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.QuestDB.BatchSize <= 0 {
		t.Errorf("expected QuestDB.BatchSize to be defaulted, got %d", cfg.QuestDB.BatchSize)
	}
	if cfg.QuestDB.FlushInterval <= 0 {
		t.Errorf("expected QuestDB.FlushInterval to be defaulted, got %v", cfg.QuestDB.FlushInterval)
	}
}

func TestValidateRequiresPlantName(t *testing.T) {
	cfg := &Config{}
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error for missing plant name")
	}
}
