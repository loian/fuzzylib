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

func TestParseFIS_Negation(t *testing.T) {
	model, err := ParseFIS("../testdata/negation_test.fis")
	if err != nil {
		t.Fatalf("Failed to parse FIS with negation: %v", err)
	}

	if model.System.Name != "NegationTest" {
		t.Errorf("Expected name 'NegationTest', got '%s'", model.System.Name)
	}
	if model.System.NumInputs != 2 {
		t.Errorf("Expected 2 inputs, got %d", model.System.NumInputs)
	}
	if len(model.Rules) != 4 {
		t.Fatalf("Expected 4 rules, got %d", len(model.Rules))
	}

	// Check first rule: -1 1, 1 (1.0) : 1
	// IF Temperature is NOT Cold AND Humidity is Dry THEN FanSpeed is Low
	rule1 := model.Rules[0]
	if len(rule1.Antecedents) != 2 {
		t.Errorf("Rule 1: expected 2 antecedents, got %d", len(rule1.Antecedents))
	}
	if rule1.Antecedents[0] != -1 {
		t.Errorf("Rule 1: expected first antecedent -1 (NOT), got %d", rule1.Antecedents[0])
	}
	if rule1.Antecedents[1] != 1 {
		t.Errorf("Rule 1: expected second antecedent 1, got %d", rule1.Antecedents[1])
	}
	if rule1.Connection != 1 {
		t.Errorf("Rule 1: expected AND (1), got %d", rule1.Connection)
	}

	// Check second rule: 2 -1, 3 (1.0) : 1
	// IF Temperature is Hot AND Humidity is NOT Dry THEN FanSpeed is High
	rule2 := model.Rules[1]
	if rule2.Antecedents[0] != 2 {
		t.Errorf("Rule 2: expected first antecedent 2, got %d", rule2.Antecedents[0])
	}
	if rule2.Antecedents[1] != -1 {
		t.Errorf("Rule 2: expected second antecedent -1 (NOT), got %d", rule2.Antecedents[1])
	}

	// Check third rule: -1 2, 2 (0.8) : 2
	// IF Temperature is NOT Cold OR Humidity is Wet THEN FanSpeed is Medium
	rule3 := model.Rules[2]
	if rule3.Connection != 2 {
		t.Errorf("Rule 3: expected OR (2), got %d", rule3.Connection)
	}
	if rule3.Weight != 0.8 {
		t.Errorf("Rule 3: expected weight 0.8, got %f", rule3.Weight)
	}
}

func TestLoadFIS_Negation(t *testing.T) {
	fis, err := LoadFIS("../testdata/negation_test.fis")
	if err != nil {
		t.Fatalf("Failed to load FIS with negation: %v", err)
	}

	if len(fis.InputVariables) != 2 {
		t.Errorf("Expected 2 input variables, got %d", len(fis.InputVariables))
	}
	if len(fis.OutputVariables) != 1 {
		t.Errorf("Expected 1 output variable, got %d", len(fis.OutputVariables))
	}
	if len(fis.Rules) != 4 {
		t.Errorf("Expected 4 rules, got %d", len(fis.Rules))
	}

	// Check that negated conditions are properly converted
	// Rule 1: IF Temperature is NOT Cold AND Humidity is Dry THEN FanSpeed is Low
	rule1 := fis.Rules[0]
	if len(rule1.Conditions) != 2 {
		t.Fatalf("Rule 1: expected 2 conditions, got %d", len(rule1.Conditions))
	}

	// First condition should be negated (NOT Cold)
	cond1 := rule1.Conditions[0]
	if !cond1.Negated {
		t.Errorf("Rule 1, condition 1: expected Negated=true, got false")
	}
	if cond1.Set != "Cold" {
		t.Errorf("Rule 1, condition 1: expected set 'Cold', got '%s'", cond1.Set)
	}

	// Second condition should not be negated (Dry)
	cond2 := rule1.Conditions[1]
	if cond2.Negated {
		t.Errorf("Rule 1, condition 2: expected Negated=false, got true")
	}
	if cond2.Set != "Dry" {
		t.Errorf("Rule 1, condition 2: expected set 'Dry', got '%s'", cond2.Set)
	}

	// Test inference with negated rules
	// Using values that ensure rules will fire
	// Temp=25 (between Cold and Hot), Humidity=40 (Dry region)
	// Rule 1: NOT Cold AND Dry should fire (NOT Cold will be moderate/high at temp=25)
	outputs, err := fis.Infer(map[string]float64{"Temperature": 25, "Humidity": 40})
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	fanSpeed := outputs["FanSpeed"]
	// Just verify we got a valid output
	if fanSpeed < 0 || fanSpeed > 100 {
		t.Errorf("Expected valid fan speed [0-100], got %f", fanSpeed)
	}

	// Temp=35 (Hot region), Humidity=80 (Wet region)
	// Rule 2: Hot AND NOT Dry should fire strongly
	outputs2, err := fis.Infer(map[string]float64{"Temperature": 35, "Humidity": 80})
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	fanSpeed2 := outputs2["FanSpeed"]
	// Should result in higher fan speed due to hot+wet
	if fanSpeed2 < 50 {
		t.Errorf("Expected medium-high fan speed (>50) for hot+wet conditions, got %f", fanSpeed2)
	}
}
