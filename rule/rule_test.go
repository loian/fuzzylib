package rule

import (
	"fuzzy/operators"
	"testing"
)

func TestNewRule(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, err := NewRule(output, operators.AND)
	if err != nil {
		t.Fatalf("NewRule failed: %v", err)
	}

	if rule.Output.Variable != "FanSpeed" {
		t.Errorf("Expected output variable FanSpeed, got %s", rule.Output.Variable)
	}
	if rule.Output.Set != "High" {
		t.Errorf("Expected output set High, got %s", rule.Output.Set)
	}
	if rule.Weight != 1.0 {
		t.Errorf("Expected default weight 1.0, got %f", rule.Weight)
	}
	if len(rule.Conditions) != 0 {
		t.Errorf("Expected empty conditions initially, got %d", len(rule.Conditions))
	}
}

func TestRule_AddCondition(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	if err := rule.AddCondition("Temperature", "Hot"); err != nil {
		t.Fatalf("AddCondition failed: %v", err)
	}
	if err := rule.AddCondition("Humidity", "High"); err != nil {
		t.Fatalf("AddCondition failed: %v", err)
	}

	if len(rule.Conditions) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(rule.Conditions))
	}

	if rule.Conditions[0].Variable != "Temperature" || rule.Conditions[0].Set != "Hot" {
		t.Error("First condition not added correctly")
	}

	if rule.Conditions[1].Variable != "Humidity" || rule.Conditions[1].Set != "High" {
		t.Error("Second condition not added correctly")
	}
}

func TestRule_SetWeight(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	tests := []struct {
		input       float64
		expectError bool
		expected    float64
	}{
		{0.5, false, 0.5},
		{1.5, true, 1.0},  // Should error
		{-0.5, true, 1.0}, // Should error
		{0.0, false, 0.0},
		{1.0, false, 1.0},
	}

	for _, tt := range tests {
		err := rule.SetWeight(tt.input)
		if tt.expectError {
			if err == nil {
				t.Errorf("SetWeight(%f): expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("SetWeight(%f): unexpected error: %v", tt.input, err)
			}
			if rule.Weight != tt.expected {
				t.Errorf("SetWeight(%f): expected %f, got %f", tt.input, tt.expected, rule.Weight)
			}
		}
	}
}

func TestRule_Evaluate_AND(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)
	_ = rule.AddCondition("Temperature", "Hot")
	_ = rule.AddCondition("Humidity", "High")

	membershipMap := map[string]map[string]float64{
		"Temperature": {"Hot": 0.8},
		"Humidity":    {"High": 0.6},
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	expected := 0.6 // MIN(0.8, 0.6) * 1.0
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRule_Evaluate_OR(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.OR)
	_ = rule.AddCondition("Temperature", "Hot")
	_ = rule.AddCondition("Humidity", "High")

	membershipMap := map[string]map[string]float64{
		"Temperature": {"Hot": 0.8},
		"Humidity":    {"High": 0.6},
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	expected := 0.8 // MAX(0.8, 0.6) * 1.0
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRule_Evaluate_WithWeight(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)
	_ = rule.AddCondition("Temperature", "Hot")
	if err := rule.SetWeight(0.5); err != nil {
		t.Fatalf("Failed to set weight: %v", err)
	}

	membershipMap := map[string]map[string]float64{
		"Temperature": {"Hot": 0.8},
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	expected := 0.4 // 0.8 * 0.5
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRule_Evaluate_MissingVariable(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)
	_ = rule.AddCondition("Temperature", "Hot")

	membershipMap := map[string]map[string]float64{
		// Temperature is missing
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	// Should return 0 when condition variables not found
	if result != 0.0 {
		t.Errorf("Expected 0.0 for missing variable, got %f", result)
	}
}

func TestRule_Builder_Pattern(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)
	_ = rule.AddCondition("Temperature", "Hot")
	_ = rule.AddCondition("Humidity", "High")

	if err := rule.SetWeight(0.8); err != nil {
		t.Fatalf("Failed to set weight: %v", err)
	}

	if len(rule.Conditions) != 2 {
		t.Errorf("Expected 2 conditions after builder pattern, got %d", len(rule.Conditions))
	}
	if rule.Weight != 0.8 {
		t.Errorf("Expected weight 0.8, got %f", rule.Weight)
	}
}
