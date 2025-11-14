package fis

import (
	"testing"
)

func TestParseFIS(t *testing.T) {
	model, err := ParseFIS("../testdata/temp_control.fis")
	if err != nil {
		t.Fatalf("Failed to parse FIS: %v", err)
	}

	if model.System.Name != "TemperatureFanControl" {
		t.Errorf("Expected name 'TemperatureFanControl', got '%s'", model.System.Name)
	}
	if model.System.Type != "mamdani" {
		t.Errorf("Expected type 'mamdani', got '%s'", model.System.Type)
	}
	if model.System.NumInputs != 1 {
		t.Errorf("Expected 1 input, got %d", model.System.NumInputs)
	}
	if len(model.Inputs) != 1 {
		t.Fatalf("Expected 1 input variable, got %d", len(model.Inputs))
	}
	if len(model.Outputs) != 1 {
		t.Fatalf("Expected 1 output variable, got %d", len(model.Outputs))
	}
	if len(model.Rules) != 4 {
		t.Fatalf("Expected 4 rules, got %d", len(model.Rules))
	}
}

func TestLoadFIS(t *testing.T) {
	fis, err := LoadFIS("../testdata/temp_control.fis")
	if err != nil {
		t.Fatalf("Failed to load FIS: %v", err)
	}

	if len(fis.InputVariables) != 1 {
		t.Errorf("Expected 1 input variable, got %d", len(fis.InputVariables))
	}
	if len(fis.OutputVariables) != 1 {
		t.Errorf("Expected 1 output variable, got %d", len(fis.OutputVariables))
	}
	if len(fis.Rules) != 4 {
		t.Errorf("Expected 4 rules, got %d", len(fis.Rules))
	}

	// Test inference
	outputs, err := fis.Infer(map[string]float64{"Temperature": 40})
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	fanSpeed := outputs["FanSpeed"]
	if fanSpeed < 70 {
		t.Errorf("Expected High fan speed (>70) for temp 40, got %f", fanSpeed)
	}
}
