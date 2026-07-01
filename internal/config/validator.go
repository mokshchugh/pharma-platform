package config

import "fmt"

// Validate verifies that the loaded application configuration is valid.
func Validate(cfg *Config) error {
	if err := validatePlant(cfg); err != nil {
		return err
	}

	if err := validateAPI(cfg); err != nil {
		return err
	}

	if err := validatePLCs(cfg); err != nil {
		return err
	}

	if err := validateTags(cfg); err != nil {
		return err
	}

	if err := validateRelationships(cfg); err != nil {
		return err
	}

	return nil
}

// validatePlant verifies the plant configuration.
func validatePlant(cfg *Config) error {
	if cfg.Plant.Name == "" {
		return fmt.Errorf("plant name is required")
	}

	if cfg.Plant.Location == "" {
		return fmt.Errorf("plant location is required")
	}

	if cfg.Plant.TimeZone == "" {
		return fmt.Errorf("plant timezone is required")
	}

	return nil
}

// validateAPI verifies the API configuration.
func validateAPI(cfg *Config) error {
	if cfg.API.Host == "" {
		return fmt.Errorf("api host is required")
	}

	if cfg.API.Port <= 0 || cfg.API.Port > 65535 {
		return fmt.Errorf("api port must be between 1 and 65535")
	}

	return nil
}

// validatePLCs verifies all configured PLCs.
func validatePLCs(cfg *Config) error {
	plcIDs := make(map[string]struct{})

	for _, plc := range cfg.PLCs {
		if plc.ID == "" {
			return fmt.Errorf("plc id is required")
		}

		if _, exists := plcIDs[plc.ID]; exists {
			return fmt.Errorf("duplicate plc id %q", plc.ID)
		}

		plcIDs[plc.ID] = struct{}{}

		if plc.Name == "" {
			return fmt.Errorf("plc %q: name is required", plc.ID)
		}

		if plc.Driver == "" {
			return fmt.Errorf("plc %q: driver is required", plc.ID)
		}

		if plc.IPAddress == "" {
			return fmt.Errorf("plc %q: ip address is required", plc.ID)
		}

		if plc.Port == 0 {
			return fmt.Errorf("plc %q: invalid port", plc.ID)
		}
	}

	return nil
}

// validateTags verifies all configured tags.
func validateTags(cfg *Config) error {
	tagIDs := make(map[string]struct{})

	for _, tag := range cfg.Tags {
		if tag.ID == "" {
			return fmt.Errorf("tag id is required")
		}

		if _, exists := tagIDs[tag.ID]; exists {
			return fmt.Errorf("duplicate tag id %q", tag.ID)
		}

		tagIDs[tag.ID] = struct{}{}

		if tag.Name == "" {
			return fmt.Errorf("tag %q: name is required", tag.ID)
		}

		if tag.PLCID == "" {
			return fmt.Errorf("tag %q: plc id is required", tag.ID)
		}

		if tag.Address == "" {
			return fmt.Errorf("tag %q: address is required", tag.ID)
		}

		if tag.PollInterval <= 0 {
			return fmt.Errorf("tag %q: poll interval must be greater than zero", tag.ID)
		}
	}

	return nil
}

// validateRelationships verifies relationships between configuration objects.
func validateRelationships(cfg *Config) error {
	plcIDs := make(map[string]struct{})

	for _, plc := range cfg.PLCs {
		plcIDs[plc.ID] = struct{}{}
	}

	for _, tag := range cfg.Tags {
		if _, exists := plcIDs[tag.PLCID]; !exists {
			return fmt.Errorf(
				"tag %q references unknown plc %q",
				tag.ID,
				tag.PLCID,
			)
		}
	}

	return nil
}
