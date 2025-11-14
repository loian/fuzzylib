package rule

import (
	"fuzzy/operators"
	"math"
	"testing"
)

const epsilon = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

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

func TestRule_Evaluate_Negated_Single(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	// Add a negated condition
	rule.Conditions = append(rule.Conditions, RuleCondition{
		Variable: "Temperature",
		Set:      "Cold",
		Negated:  true,
	})

	membershipMap := map[string]map[string]float64{
		"Temperature": {"Cold": 0.7},
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	expected := 0.3 // NOT(0.7) = 1.0 - 0.7 = 0.3
	if !almostEqual(result, expected) {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRule_Evaluate_Negated_AND(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	// IF Temperature is NOT Cold AND Humidity is High
	rule.Conditions = append(rule.Conditions, RuleCondition{
		Variable: "Temperature",
		Set:      "Cold",
		Negated:  true,
	})
	rule.Conditions = append(rule.Conditions, RuleCondition{
		Variable: "Humidity",
		Set:      "High",
		Negated:  false,
	})

	membershipMap := map[string]map[string]float64{
		"Temperature": {"Cold": 0.2}, // NOT(0.2) = 0.8
		"Humidity":    {"High": 0.6},
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	expected := 0.6 // MIN(NOT(0.2), 0.6) = MIN(0.8, 0.6) = 0.6
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRule_Evaluate_Negated_OR(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "Medium"}
	rule, _ := NewRule(output, operators.OR)

	// IF Temperature is NOT Cold OR Humidity is NOT Dry
	rule.Conditions = append(rule.Conditions, RuleCondition{
		Variable: "Temperature",
		Set:      "Cold",
		Negated:  true,
	})
	rule.Conditions = append(rule.Conditions, RuleCondition{
		Variable: "Humidity",
		Set:      "Dry",
		Negated:  true,
	})

	membershipMap := map[string]map[string]float64{
		"Temperature": {"Cold": 0.3}, // NOT(0.3) = 0.7
		"Humidity":    {"Dry": 0.8},  // NOT(0.8) = 0.2
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	expected := 0.7 // MAX(NOT(0.3), NOT(0.8)) = MAX(0.7, 0.2) = 0.7
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRule_Evaluate_Mixed_Negation(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	// Mix of negated and non-negated conditions
	rule.Conditions = append(rule.Conditions, RuleCondition{
		Variable: "Temperature",
		Set:      "Hot",
		Negated:  false,
	})
	rule.Conditions = append(rule.Conditions, RuleCondition{
		Variable: "Humidity",
		Set:      "Dry",
		Negated:  true,
	})

	membershipMap := map[string]map[string]float64{
		"Temperature": {"Hot": 0.9},
		"Humidity":    {"Dry": 0.3}, // NOT(0.3) = 0.7
	}

	result, err := rule.Evaluate(membershipMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	expected := 0.7 // MIN(0.9, NOT(0.3)) = MIN(0.9, 0.7) = 0.7
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRule_AddConditionEx(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	// Add normal condition
	if err := rule.AddConditionEx("Temperature", "Hot", false); err != nil {
		t.Fatalf("AddConditionEx failed: %v", err)
	}

	// Add negated condition
	if err := rule.AddConditionEx("Humidity", "Dry", true); err != nil {
		t.Fatalf("AddConditionEx failed: %v", err)
	}

	if len(rule.Conditions) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(rule.Conditions))
	}

	// Verify first condition is not negated
	if rule.Conditions[0].Negated {
		t.Errorf("First condition should not be negated")
	}
	if rule.Conditions[0].Variable != "Temperature" || rule.Conditions[0].Set != "Hot" {
		t.Error("First condition not added correctly")
	}

	// Verify second condition is negated
	if !rule.Conditions[1].Negated {
		t.Errorf("Second condition should be negated")
	}
	if rule.Conditions[1].Variable != "Humidity" || rule.Conditions[1].Set != "Dry" {
		t.Error("Second condition not added correctly")
	}
}

func TestRule_AddConditionEx_Validation(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	// Test empty variable
	if err := rule.AddConditionEx("", "Hot", false); err == nil {
		t.Error("Expected error for empty variable name")
	}

	// Test empty set
	if err := rule.AddConditionEx("Temperature", "", true); err == nil {
		t.Error("Expected error for empty set name")
	}

	// Should have no conditions after failed additions
	if len(rule.Conditions) != 0 {
		t.Errorf("Expected 0 conditions after validation failures, got %d", len(rule.Conditions))
	}
}

func TestRule_AddCondition_UsesAddConditionEx(t *testing.T) {
	output := RuleCondition{Variable: "FanSpeed", Set: "High"}
	rule, _ := NewRule(output, operators.AND)

	// AddCondition should add non-negated condition
	if err := rule.AddCondition("Temperature", "Hot"); err != nil {
		t.Fatalf("AddCondition failed: %v", err)
	}

	if len(rule.Conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %d", len(rule.Conditions))
	}

	// Verify condition is not negated
	if rule.Conditions[0].Negated {
		t.Error("AddCondition should add non-negated condition")
	}
}
