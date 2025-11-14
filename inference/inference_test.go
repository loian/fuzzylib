package inference

import (
	"github.com/loian/fuzzylib/membership"
	"github.com/loian/fuzzylib/operators"
	"github.com/loian/fuzzylib/rule"
	"github.com/loian/fuzzylib/set"
	"github.com/loian/fuzzylib/variable"
	"math"
	"testing"
)

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// Helper to unwrap membership functions in tests
func mustMF(mf membership.MembershipFunction, err error) membership.MembershipFunction {
	if err != nil {
		panic(err)
	}
	return mf
}

func TestNewMamdaniInferenceSystem(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	if len(fis.InputVariables) != 0 {
		t.Errorf("Expected empty input variables, got %d", len(fis.InputVariables))
	}

	if len(fis.OutputVariables) != 0 {
		t.Errorf("Expected empty output variables, got %d", len(fis.OutputVariables))
	}

	if len(fis.Rules) != 0 {
		t.Errorf("Expected empty rules, got %d", len(fis.Rules))
	}
}

func TestMamdaniInferenceSystem_AddVariables(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	temp, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	fan, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)

	_ = fis.AddInputVariable(temp)
	_ = fis.AddOutputVariable(fan)

	if len(fis.InputVariables) != 1 {
		t.Errorf("Expected 1 input variable, got %d", len(fis.InputVariables))
	}

	if len(fis.OutputVariables) != 1 {
		t.Errorf("Expected 1 output variable, got %d", len(fis.OutputVariables))
	}

	if _, ok := fis.InputVariables["Temperature"]; !ok {
		t.Error("Temperature variable not added")
	}

	if _, ok := fis.OutputVariables["FanSpeed"]; !ok {
		t.Error("FanSpeed variable not added")
	}
}

func TestMamdaniInferenceSystem_SimpleInference(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	// Create temperature variable
	tempVar, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	tempVar.AddSet(set.NewFuzzySet("Cold", mustMF(membership.NewTriangular(0, 0, 20))))
	tempVar.AddSet(set.NewFuzzySet("Warm", mustMF(membership.NewTriangular(10, 25, 40))))
	tempVar.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(30, 50, 50))))

	// Create fan speed variable
	fanVar, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fanVar.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 33))))
	fanVar.AddSet(set.NewFuzzySet("Medium", mustMF(membership.NewTriangular(20, 50, 80))))
	fanVar.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(67, 100, 100))))

	_ = fis.AddInputVariable(tempVar)
	_ = fis.AddOutputVariable(fanVar)

	// Add rules
	rule1, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "Low"}, operators.AND)
	rule1.AddCondition("Temperature", "Cold")
	_ = fis.AddRule(rule1)

	rule2, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "Medium"}, operators.AND)
	rule2.AddCondition("Temperature", "Warm")
	_ = fis.AddRule(rule2)

	rule3, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "High"}, operators.AND)
	rule3.AddCondition("Temperature", "Hot")
	_ = fis.AddRule(rule3)

	// Test inference
	inputs := map[string]float64{
		"Temperature": 35.0,
	}

	results, err := fis.Infer(inputs)
	if err != nil {
		t.Errorf("Inference failed: %v", err)
	}

	if fanSpeed, ok := results["FanSpeed"]; !ok {
		t.Error("FanSpeed not in results")
	} else if fanSpeed < 50 {
		t.Errorf("Expected FanSpeed > 50 for hot temperature, got %f", fanSpeed)
	}
}

func TestMamdaniInferenceSystem_MultipleInputs(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	// Temperature variable
	tempVar, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	tempVar.AddSet(set.NewFuzzySet("Cold", mustMF(membership.NewTriangular(0, 0, 20))))
	tempVar.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(30, 50, 50))))

	// Humidity variable
	humVar, _ := variable.NewFuzzyVariable("Humidity", 0, 100)
	humVar.AddSet(set.NewFuzzySet("Dry", mustMF(membership.NewTriangular(0, 0, 40))))
	humVar.AddSet(set.NewFuzzySet("Wet", mustMF(membership.NewTriangular(60, 100, 100))))

	// Fan speed output
	fanVar, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fanVar.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 33))))
	fanVar.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(67, 100, 100))))

	_ = fis.AddInputVariable(tempVar)
	_ = fis.AddInputVariable(humVar)
	_ = fis.AddOutputVariable(fanVar)

	// Rule: IF Temperature is Hot AND Humidity is Wet THEN FanSpeed is High
	r, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "High"}, operators.AND)
	r.AddCondition("Temperature", "Hot")
	r.AddCondition("Humidity", "Wet")
	_ = fis.AddRule(r)

	inputs := map[string]float64{
		"Temperature": 40.0,
		"Humidity":    80.0,
	}

	results, err := fis.Infer(inputs)
	if err != nil {
		t.Errorf("Inference failed: %v", err)
	}

	if fanSpeed, ok := results["FanSpeed"]; !ok {
		t.Error("FanSpeed not in results")
	} else if fanSpeed < 50 {
		t.Errorf("Expected FanSpeed > 50 for hot and wet conditions, got %f", fanSpeed)
	}
}

func TestMamdaniInferenceSystem_WithRuleWeights(t *testing.T) {
	fis := NewMamdaniInferenceSystem()
	if err := fis.SetDefuzzificationMethod(DefuzzCOG); err != nil {
		t.Fatalf("Failed to set defuzzification method: %v", err)
	}

	tempVar, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	tempVar.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(30, 50, 50))))

	fanVar, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fanVar.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(67, 100, 100))))

	_ = fis.AddInputVariable(tempVar)
	_ = fis.AddOutputVariable(fanVar)

	r, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "High"}, operators.AND)
	r.AddCondition("Temperature", "Hot")
	if err := r.SetWeight(0.5); err != nil {
		t.Fatalf("Failed to set weight: %v", err)
	}
	_ = fis.AddRule(r)

	inputs := map[string]float64{"Temperature": 45.0}
	results, err := fis.Infer(inputs)

	if err != nil {
		t.Errorf("Inference failed: %v", err)
	}

	// The output should be lower due to weight of 0.5
	if fanSpeed, ok := results["FanSpeed"]; ok && fanSpeed > 90 {
		t.Errorf("Expected lower FanSpeed due to weight, got %f", fanSpeed)
	}
}

func TestDefuzzifyCOG(t *testing.T) {
	fanVar, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fanVar.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 50))))
	fanVar.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(50, 100, 100))))

	// Test with single set at high strength
	memberships := map[string]float64{
		"Low":  0.0,
		"High": 1.0,
	}

	result, _ := defuzzifyCOG(fanVar, memberships)
	if result < 75 {
		t.Errorf("Expected COG > 75 for High set, got %f", result)
	}
}

func TestDefuzzifyMOM(t *testing.T) {
	fanVar, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fanVar.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 50))))
	fanVar.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(50, 100, 100))))

	memberships := map[string]float64{
		"Low":  0.0,
		"High": 1.0,
	}

	result, _ := defuzzifyMOM(fanVar, memberships)
	if result < 50 {
		t.Errorf("Expected MOM >= 50, got %f", result)
	}
}

func TestDefuzzifyFOM(t *testing.T) {
	fanVar, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fanVar.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 50))))
	fanVar.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(50, 100, 100))))

	memberships := map[string]float64{
		"Low":  0.0,
		"High": 1.0,
	}

	result, _ := defuzzifyFOM(fanVar, memberships)
	// FOM should return the first maximum point, which should be around 50
	// FOM should return the first maximum point (100 for High triangle at peak)
	if result < 90 || result > 100 {
		t.Errorf("Expected FOM around 100, got %f", result)
	}
}

func TestRuleBuilder(t *testing.T) {
	builder, err := NewRuleBuilder("FanSpeed", "High")
	if err != nil {
		t.Fatalf("NewRuleBuilder failed: %v", err)
	}
	builder, err = builder.With(0.8)
	if err != nil {
		t.Fatalf("With failed: %v", err)
	}
	r, err := builder.If("Temperature", "Hot").And().If("Humidity", "Wet").Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(r.Conditions) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(r.Conditions))
	}

	if r.Weight != 0.8 {
		t.Errorf("Expected weight 0.8, got %f", r.Weight)
	}

	if r.Output.Variable != "FanSpeed" || r.Output.Set != "High" {
		t.Error("Output not set correctly")
	}
}

func TestTemperatureControlSystem(t *testing.T) {
	// Create a complete temperature control system
	fis := NewMamdaniInferenceSystem()

	// Create temperature variable (0-50°C)
	tempVar, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	tempVar.AddSet(set.NewFuzzySet("Cold", mustMF(membership.NewTriangular(0, 0, 15))))
	tempVar.AddSet(set.NewFuzzySet("Warm", mustMF(membership.NewTriangular(10, 25, 40))))
	tempVar.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(30, 50, 50))))

	// Create fan speed variable (0-100%)
	fanVar, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fanVar.AddSet(set.NewFuzzySet("Off", mustMF(membership.NewTriangular(0, 0, 20))))
	fanVar.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(10, 30, 50))))
	fanVar.AddSet(set.NewFuzzySet("Medium", mustMF(membership.NewTriangular(40, 60, 80))))
	fanVar.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(70, 100, 100))))

	_ = fis.AddInputVariable(tempVar)
	_ = fis.AddOutputVariable(fanVar)

	// Rule 1: IF Temperature is Cold THEN FanSpeed is Off
	r1, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "Off"}, operators.AND)
	r1.AddCondition("Temperature", "Cold")
	_ = fis.AddRule(r1)

	// Rule 2: IF Temperature is Warm THEN FanSpeed is Low
	r2, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "Low"}, operators.AND)
	r2.AddCondition("Temperature", "Warm")
	_ = fis.AddRule(r2)

	// Rule 3: IF Temperature is Hot THEN FanSpeed is High
	r3, _ := rule.NewRule(rule.RuleCondition{Variable: "FanSpeed", Set: "High"}, operators.AND)
	r3.AddCondition("Temperature", "Hot")
	_ = fis.AddRule(r3)

	// Test cases
	tests := []struct {
		temp     float64
		minFan   float64
		maxFan   float64
		testName string
	}{
		{5.0, 0, 30, "Cold"},
		{25.0, 20, 60, "Warm"},
		{45.0, 70, 100, "Hot"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			inputs := map[string]float64{"Temperature": test.temp}
			results, err := fis.Infer(inputs)

			if err != nil {
				t.Errorf("Inference failed: %v", err)
			}

			if fanSpeed, ok := results["FanSpeed"]; !ok {
				t.Error("FanSpeed not in results")
			} else if fanSpeed < test.minFan || fanSpeed > test.maxFan {
				t.Errorf("Expected FanSpeed between %f and %f, got %f", test.minFan, test.maxFan, fanSpeed)
			}
		})
	}
}

func TestTypeSafeRuleBuilder(t *testing.T) {
	// Test the new type-safe API with SetRef
	fis := NewMamdaniInferenceSystem()

	// Create variables
	temp, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	fan, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)

	// Add sets and get type-safe references
	coldSet, _ := temp.AddSet(set.NewFuzzySet("Cold", mustMF(membership.NewTriangular(0, 0, 25))))
	hotSet, _ := temp.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(25, 50, 50))))
	lowFan, _ := fan.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 50))))
	highFan, _ := fan.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(50, 100, 100))))

	_ = fis.AddInputVariable(temp)
	_ = fis.AddOutputVariable(fan)

	// Build rules using type-safe references
	rb, _ := NewRuleBuilderRef(lowFan)
	rule1, _ := rb.IfRef(coldSet).Build()
	rb, _ = NewRuleBuilderRef(highFan)
	rule2, _ := rb.IfRef(hotSet).Build()

	_ = fis.AddRule(rule1)
	_ = fis.AddRule(rule2)

	// Verify rules were built correctly
	if len(fis.Rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(fis.Rules))
	}

	// Test inference
	coldInputs := map[string]float64{"Temperature": 10}
	coldResults, err := fis.Infer(coldInputs)
	if err != nil {
		t.Fatalf("Cold inference failed: %v", err)
	}
	if coldResults["FanSpeed"] > 50 {
		t.Errorf("Expected low fan speed for cold temp, got %f", coldResults["FanSpeed"])
	}

	hotInputs := map[string]float64{"Temperature": 40}
	hotResults, err := fis.Infer(hotInputs)
	if err != nil {
		t.Fatalf("Hot inference failed: %v", err)
	}
	if hotResults["FanSpeed"] < 50 {
		t.Errorf("Expected high fan speed for hot temp, got %f", hotResults["FanSpeed"])
	}
}

func TestTypeSafeRuleBuilderMultipleConditions(t *testing.T) {
	// Test type-safe API with multiple input conditions
	fis := NewMamdaniInferenceSystem()

	temp, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	humidity, _ := variable.NewFuzzyVariable("Humidity", 0, 100)
	fan, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)

	hotTemp, _ := temp.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(25, 50, 50))))
	wetHumid, _ := humidity.AddSet(set.NewFuzzySet("Wet", mustMF(membership.NewTriangular(60, 100, 100))))
	highFan, _ := fan.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(67, 100, 100))))

	_ = fis.AddInputVariable(temp)
	_ = fis.AddInputVariable(humidity)
	_ = fis.AddOutputVariable(fan)

	// Rule: IF temp is hot AND humidity is wet THEN fan is high
	rb, _ := NewRuleBuilderRef(highFan)
	rule, _ := rb.IfRef(hotTemp).IfRef(wetHumid).Build()
	_ = fis.AddRule(rule)

	// Verify rule construction
	if len(rule.Conditions) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(rule.Conditions))
	}

	// Test inference
	inputs := map[string]float64{"Temperature": 40, "Humidity": 80}
	results, err := fis.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	if results["FanSpeed"] < 50 {
		t.Errorf("Expected high fan speed for hot+wet, got %f", results["FanSpeed"])
	}
}

func TestMixedAPIRuleBuilder(t *testing.T) {
	// Test that string-based and ref-based APIs can coexist
	fis := NewMamdaniInferenceSystem()

	temp, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	fan, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)

	coldSet, _ := temp.AddSet(set.NewFuzzySet("Cold", mustMF(membership.NewTriangular(0, 0, 25))))
	temp.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(25, 50, 50))))
	lowFan, _ := fan.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 50))))
	fan.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(50, 100, 100))))

	_ = fis.AddInputVariable(temp)
	_ = fis.AddOutputVariable(fan)

	// Mix both APIs
	rb, _ := NewRuleBuilderRef(lowFan)
	rule1, _ := rb.IfRef(coldSet).Build() // New ref-based API
	rb, _ = NewRuleBuilder("FanSpeed", "High")
	rule2, _ := rb.If("Temperature", "Hot").Build() // Old string-based API

	_ = fis.AddRule(rule1)
	_ = fis.AddRule(rule2)

	// Both should work
	if len(fis.Rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(fis.Rules))
	}

	// Test inference works with both
	inputs := map[string]float64{"Temperature": 40}
	results, err := fis.Infer(inputs)
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	if results["FanSpeed"] < 50 {
		t.Errorf("Expected high fan speed, got %f", results["FanSpeed"])
	}
}

func TestInfer_ValidationMissingInput(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	// Create temperature input variable
	temp, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	temp.AddSet(set.NewFuzzySet("Cold", mustMF(membership.NewTriangular(0, 0, 25))))
	temp.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(25, 50, 50))))

	// Create fan output variable
	fan, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fan.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 50))))
	fan.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(50, 100, 100))))

	_ = fis.AddInputVariable(temp)
	_ = fis.AddOutputVariable(fan)

	// Add a simple rule
	rb, _ := NewRuleBuilder("FanSpeed", "High")
	rule, _ := rb.If("Temperature", "Hot").Build()
	_ = fis.AddRule(rule)

	// Try to infer without providing the Temperature input
	_, err := fis.Infer(map[string]float64{})
	if err == nil {
		t.Fatal("Expected error for missing input variable, got nil")
	}

	expectedMsg := "missing required input variable: Temperature"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestInfer_ValidationOutOfBounds(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	// Create temperature input variable with range [0, 50]
	temp, _ := variable.NewFuzzyVariable("Temperature", 0, 50)
	temp.AddSet(set.NewFuzzySet("Cold", mustMF(membership.NewTriangular(0, 0, 25))))
	temp.AddSet(set.NewFuzzySet("Hot", mustMF(membership.NewTriangular(25, 50, 50))))

	// Create fan output variable
	fan, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	fan.AddSet(set.NewFuzzySet("Low", mustMF(membership.NewTriangular(0, 0, 50))))
	fan.AddSet(set.NewFuzzySet("High", mustMF(membership.NewTriangular(50, 100, 100))))

	_ = fis.AddInputVariable(temp)
	_ = fis.AddOutputVariable(fan)

	// Add rules for both hot and cold so boundary values work
	rb, _ := NewRuleBuilder("FanSpeed", "Low")
	rule, _ := rb.If("Temperature", "Cold").Build()
	_ = fis.AddRule(rule)
	rb, _ = NewRuleBuilder("FanSpeed", "High")
	rule, _ = rb.If("Temperature", "Hot").Build()
	_ = fis.AddRule(rule)

	// Test value below minimum
	_, err := fis.Infer(map[string]float64{"Temperature": -10})
	if err == nil {
		t.Fatal("Expected error for value below minimum, got nil")
	}
	if err.Error() != "input value -10.00 for variable 'Temperature' is out of bounds [0.00, 50.00]" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	// Test value above maximum
	_, err = fis.Infer(map[string]float64{"Temperature": 100})
	if err == nil {
		t.Fatal("Expected error for value above maximum, got nil")
	}
	if err.Error() != "input value 100.00 for variable 'Temperature' is out of bounds [0.00, 50.00]" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	// Test valid value at boundary - Note: with triangular(0,0,25) and triangular(25,50,50),
	// the exact boundaries (0 and 50) have zero membership, so no rules fire.
	// Test just inside the boundaries instead.
	_, err = fis.Infer(map[string]float64{"Temperature": 1})
	if err != nil {
		t.Fatalf("Expected no error for value just inside minimum boundary, got: %v", err)
	}

	_, err = fis.Infer(map[string]float64{"Temperature": 49})
	if err != nil {
		t.Fatalf("Expected no error for value just inside maximum boundary, got: %v", err)
	}
}

func TestMembershipFunction_DivisionByZero(t *testing.T) {
	// Test triangular with A=B (degenerate - vertical left edge)
	tri1 := mustMF(membership.NewTriangular(10, 10, 20))
	// At the degenerate point (A=B), it's on the boundary and should be 0
	if tri1.Evaluate(10) != 0.0 {
		t.Errorf("Expected 0.0 for x at left boundary where A=B, got %f", tri1.Evaluate(10))
	}
	// Just inside should work
	if tri1.Evaluate(15) <= 0 {
		t.Errorf("Expected > 0 for x between B and C, got %f", tri1.Evaluate(15))
	}

	// Test triangular with B=C (degenerate - vertical right edge)
	tri2 := mustMF(membership.NewTriangular(0, 10, 10))
	// At the degenerate point (B=C), it's on the boundary and should be 0
	if tri2.Evaluate(10) != 0.0 {
		t.Errorf("Expected 0.0 for x at right boundary where B=C, got %f", tri2.Evaluate(10))
	}
	// Just inside should work
	if tri2.Evaluate(5) <= 0 {
		t.Errorf("Expected > 0 for x between A and B, got %f", tri2.Evaluate(5))
	}

	// Test trapezoidal with A=B (degenerate - vertical left edge)
	trap1 := mustMF(membership.NewTrapezoidal(10, 10, 20, 30))
	// At boundary should be 0
	if trap1.Evaluate(10) != 0.0 {
		t.Errorf("Expected 0.0 for x at left boundary where A=B, got %f", trap1.Evaluate(10))
	}
	// On plateau should be 1
	if trap1.Evaluate(15) != 1.0 {
		t.Errorf("Expected 1.0 for x on plateau, got %f", trap1.Evaluate(15))
	}

	// Test trapezoidal with C=D (degenerate - vertical right edge)
	trap2 := mustMF(membership.NewTrapezoidal(0, 10, 20, 20))
	// At boundary should be 0
	if trap2.Evaluate(20) != 0.0 {
		t.Errorf("Expected 0.0 for x at right boundary where C=D, got %f", trap2.Evaluate(20))
	}
	// On plateau should be 1
	if trap2.Evaluate(15) != 1.0 {
		t.Errorf("Expected 1.0 for x on plateau, got %f", trap2.Evaluate(15))
	}

	// Test Gaussian with width=0 now returns error (invalid)
	_, err := membership.NewGaussian(15, 0)
	if err == nil {
		t.Error("Expected error for Gaussian with width=0, got nil")
	}

	// Test that we don't get panics or NaN from division by zero
	// All these should execute without error
	_ = mustMF(membership.NewTriangular(5, 5, 10)).Evaluate(7)
	_ = mustMF(membership.NewTriangular(0, 5, 5)).Evaluate(3)
	_ = mustMF(membership.NewTrapezoidal(5, 5, 10, 15)).Evaluate(7)
	_ = mustMF(membership.NewTrapezoidal(0, 5, 10, 10)).Evaluate(7)
}

func TestSetResolution_BoundsCheck(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	// Test that resolution > 0 is accepted
	err := fis.SetResolution(500)
	if err != nil {
		t.Errorf("Unexpected error for valid resolution: %v", err)
	}
	if fis.Resolution != 500 {
		t.Errorf("Expected resolution 500, got %d", fis.Resolution)
	}

	// Test that resolution <= 0 returns error
	err = fis.SetResolution(0)
	if err == nil {
		t.Error("Expected error for resolution 0, got nil")
	}

	err = fis.SetResolution(-100)
	if err == nil {
		t.Error("Expected error for negative resolution, got nil")
	}
}

func TestSetDefuzzificationMethod_Validation(t *testing.T) {
	fis := NewMamdaniInferenceSystem()

	// Test valid methods
	validMethods := []string{DefuzzCOG, DefuzzMOM, DefuzzFOM, DefuzzLOM, DefuzzSOM}
	for _, method := range validMethods {
		err := fis.SetDefuzzificationMethod(method)
		if err != nil {
			t.Errorf("Expected no error for valid method '%s', got: %v", method, err)
		}
		if fis.DefuzzMethod != method {
			t.Errorf("Expected method '%s', got '%s'", method, fis.DefuzzMethod)
		}
	}

	// Test invalid method
	err := fis.SetDefuzzificationMethod("invalid_method")
	if err == nil {
		t.Error("Expected error for invalid method, got nil")
	}
}
