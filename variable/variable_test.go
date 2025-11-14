package variable

import (
	"github.com/loian/fuzzylib/membership"
	"github.com/loian/fuzzylib/set"
	"math"
	"testing"
)

const epsilon = 1e-9

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// ===== FuzzySet Tests =====

func TestFuzzySet_Creation(t *testing.T) {
	memFunc, _ := membership.NewTriangular(0, 5, 10)
	fuzzySet, err := set.NewFuzzySet("Test", memFunc)
	if err != nil {
		t.Fatalf("Failed to create fuzzy set: %v", err)
	}

	if fuzzySet.Name != "Test" {
		t.Errorf("Expected name 'Test', got %s", fuzzySet.Name)
	}

	if fuzzySet.MembershipFunc == nil {
		t.Errorf("Expected non-nil membership function")
	}
}

func TestFuzzySet_Evaluate(t *testing.T) {
	memFunc, _ := membership.NewTriangular(0, 5, 10)
	fuzzySet, err := set.NewFuzzySet("Peak", memFunc)
	if err != nil {
		t.Fatalf("Failed to create fuzzy set: %v", err)
	}

	result := fuzzySet.Evaluate(5)
	if !floatEqual(result, 1.0) {
		t.Errorf("Expected 1.0 at peak, got %f", result)
	}
}

// ===== FuzzyVariable Tests =====

func TestFuzzyVariable_Creation(t *testing.T) {
	fv, _ := NewFuzzyVariable("Temperature", 0, 50)

	if fv.Name != "Temperature" {
		t.Errorf("Expected name 'Temperature', got %s", fv.Name)
	}

	if fv.MinValue != 0 || fv.MaxValue != 50 {
		t.Errorf("Expected domain [0, 50], got [%f, %f]", fv.MinValue, fv.MaxValue)
	}

	if len(fv.Sets) != 0 {
		t.Errorf("Expected empty sets map on creation, got %d sets", len(fv.Sets))
	}
}

func TestFuzzyVariable_AddSet(t *testing.T) {
	fv, _ := NewFuzzyVariable("Temperature", 0, 50)

	mf1, _ := membership.NewTriangular(0, 0, 15)
	if _, err := fv.AddSet(set.NewFuzzySet("Cold", mf1)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}
	mf2, _ := membership.NewTriangular(10, 25, 40)
	if _, err := fv.AddSet(set.NewFuzzySet("Warm", mf2)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}
	mf3, _ := membership.NewTriangular(30, 50, 50)
	if _, err := fv.AddSet(set.NewFuzzySet("Hot", mf3)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}

	if len(fv.Sets) != 3 {
		t.Errorf("Expected 3 sets, got %d", len(fv.Sets))
	}

	if _, exists := fv.Sets["Cold"]; !exists {
		t.Errorf("Expected 'Cold' set to exist")
	}

	if _, exists := fv.Sets["Warm"]; !exists {
		t.Errorf("Expected 'Warm' set to exist")
	}

	if _, exists := fv.Sets["Hot"]; !exists {
		t.Errorf("Expected 'Hot' set to exist")
	}
}

func TestFuzzyVariable_Fuzzify(t *testing.T) {
	fv, _ := NewFuzzyVariable("Temperature", 0, 50)

	mf1, _ := membership.NewTriangular(0, 0, 15)
	fv.AddSet(set.NewFuzzySet("Cold", mf1))
	mf2, _ := membership.NewTriangular(10, 25, 40)
	fv.AddSet(set.NewFuzzySet("Warm", mf2))
	mf3, _ := membership.NewTriangular(30, 50, 50)
	fv.AddSet(set.NewFuzzySet("Hot", mf3))

	// Test at 25 degrees - should be maximum in Warm
	result := fv.Fuzzify(25)

	if _, exists := result["Cold"]; !exists {
		t.Errorf("Expected 'Cold' in fuzzified result")
	}

	if _, exists := result["Warm"]; !exists {
		t.Errorf("Expected 'Warm' in fuzzified result")
	}

	if _, exists := result["Hot"]; !exists {
		t.Errorf("Expected 'Hot' in fuzzified result")
	}

	// Check that Warm has highest membership at 25
	if !floatEqual(result["Warm"], 1.0) {
		t.Errorf("Expected Warm=1.0 at 25°C, got %f", result["Warm"])
	}

	if result["Cold"] != 0.0 {
		t.Errorf("Expected Cold=0.0 at 25°C, got %f", result["Cold"])
	}
}

func TestFuzzyVariable_FuzzifyWithOverlap(t *testing.T) {
	fv, _ := NewFuzzyVariable("Temperature", 0, 50)

	mf1, _ := membership.NewTriangular(0, 0, 15)
	fv.AddSet(set.NewFuzzySet("Cold", mf1))
	mf2, _ := membership.NewTriangular(10, 25, 40)
	fv.AddSet(set.NewFuzzySet("Warm", mf2))

	// Test at 12 - should have membership in both Cold and Warm
	result := fv.Fuzzify(12)

	coldMembership := result["Cold"]
	warmMembership := result["Warm"]

	// Both should be non-zero
	if coldMembership <= 0 {
		t.Errorf("Expected Cold > 0 at 12°C, got %f", coldMembership)
	}

	if warmMembership <= 0 {
		t.Errorf("Expected Warm > 0 at 12°C, got %f", warmMembership)
	}

	// At 12, Cold is descending from peak at 0, Warm is ascending toward peak at 25
	// They should both be positive values
	sum := coldMembership + warmMembership
	if sum <= 0 {
		t.Errorf("Expected positive sum of memberships at overlap, got %f", sum)
	}
}

func TestFuzzyVariable_IsValid(t *testing.T) {
	fv, _ := NewFuzzyVariable("Temperature", 0, 50)

	tests := []struct {
		value    float64
		expected bool
	}{
		{-1, false},
		{0, true},
		{25, true},
		{50, true},
		{51, false},
		{100, false},
	}

	for _, test := range tests {
		result := fv.IsValid(test.value)
		if result != test.expected {
			t.Errorf("IsValid(%f): expected %v, got %v", test.value, test.expected, result)
		}
	}
}

func TestFuzzyVariable_ReplaceSet(t *testing.T) {
	fv, _ := NewFuzzyVariable("Temperature", 0, 50)

	mf1, _ := membership.NewTriangular(0, 5, 15)
	if _, err := fv.AddSet(set.NewFuzzySet("Cold", mf1)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}

	result1 := fv.Fuzzify(5)
	if !floatEqual(result1["Cold"], 1.0) {
		t.Errorf("Initial set not working correctly, got %f", result1["Cold"])
	}

	// Try to add a set with the same name - should return error
	mf2, _ := membership.NewTriangular(0, 0, 20)
	_, err := fv.AddSet(set.NewFuzzySet("Cold", mf2))

	// Should get error for duplicate set name
	if err == nil {
		t.Error("Expected error when adding duplicate set name, got nil")
	}

	// Original set should still be in use
	result2 := fv.Fuzzify(5)
	if !floatEqual(result2["Cold"], 1.0) {
		t.Errorf("Original set should still be active, expected 1.0 at peak, got %f", result2["Cold"])
	}
}

// ===== Integration Tests =====

func TestTemperatureControlExample(t *testing.T) {
	// Create a temperature variable with three linguistic categories
	tempVar, _ := NewFuzzyVariable("Temperature", 0, 50)

	mf1, _ := membership.NewTriangular(0, 0, 15)
	if _, err := tempVar.AddSet(set.NewFuzzySet("Cold", mf1)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}
	mf2, _ := membership.NewTriangular(10, 25, 40)
	if _, err := tempVar.AddSet(set.NewFuzzySet("Warm", mf2)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}
	mf3, _ := membership.NewTriangular(30, 50, 50)
	if _, err := tempVar.AddSet(set.NewFuzzySet("Hot", mf3)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}

	// Test case 1: Very cold
	result1 := tempVar.Fuzzify(5)
	if result1["Cold"] == 0 {
		t.Errorf("Expected non-zero Cold membership at 5°C")
	}

	// Test case 2: Comfortable
	result2 := tempVar.Fuzzify(25)
	if !floatEqual(result2["Warm"], 1.0) {
		t.Errorf("Expected maximum Warm membership at 25°C")
	}

	// Test case 3: Very hot
	result3 := tempVar.Fuzzify(45)
	if result3["Hot"] == 0 {
		t.Errorf("Expected non-zero Hot membership at 45°C")
	}
}

func TestMultipleVariables(t *testing.T) {
	// Create temperature variable
	temp, _ := NewFuzzyVariable("Temperature", 0, 50)
	mf1, _ := membership.NewTriangular(0, 0, 25)
	temp.AddSet(set.NewFuzzySet("Low", mf1))
	mf2, _ := membership.NewTriangular(20, 50, 50)
	temp.AddSet(set.NewFuzzySet("High", mf2))

	// Create humidity variable
	humidity, _ := NewFuzzyVariable("Humidity", 0, 100)
	mf3, _ := membership.NewTriangular(0, 0, 40)
	humidity.AddSet(set.NewFuzzySet("Dry", mf3))
	mf4, _ := membership.NewTriangular(30, 100, 100)
	humidity.AddSet(set.NewFuzzySet("Wet", mf4))

	// Fuzzify both
	tempResult := temp.Fuzzify(25)
	humidityResult := humidity.Fuzzify(60)

	// Both should have valid results
	if len(tempResult) != 2 {
		t.Errorf("Expected 2 sets for temperature, got %d", len(tempResult))
	}

	if len(humidityResult) != 2 {
		t.Errorf("Expected 2 sets for humidity, got %d", len(humidityResult))
	}
}

func TestFuzzifyAllMembershipFunctions(t *testing.T) {
	fv, _ := NewFuzzyVariable("Value", 0, 100)

	// Test with different membership function types
	mf1, _ := membership.NewTriangular(40, 50, 60)
	if _, err := fv.AddSet(set.NewFuzzySet("Triangular", mf1)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}
	mf2, _ := membership.NewTrapezoidal(30, 45, 55, 70)
	if _, err := fv.AddSet(set.NewFuzzySet("Trapezoidal", mf2)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}
	mf3, _ := membership.NewGaussian(50, 15)
	if _, err := fv.AddSet(set.NewFuzzySet("Gaussian", mf3)); err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}

	result := fv.Fuzzify(50)

	// All should have been evaluated with non-zero membership at 50
	if result["Triangular"] != 1.0 {
		t.Errorf("Triangular should have membership 1.0 at peak (50), got %f", result["Triangular"])
	}
	if result["Trapezoidal"] <= 0.9 {
		t.Errorf("Trapezoidal should have high membership at 50, got %f", result["Trapezoidal"])
	}
	if result["Gaussian"] != 1.0 {
		t.Errorf("Gaussian should have membership 1.0 at center (50), got %f", result["Gaussian"])
	}
}

// ===== SetRef Tests =====

func TestSetRef_AddSetReturnsReference(t *testing.T) {
	fv, _ := NewFuzzyVariable("Temperature", 0, 100)
	mf, _ := membership.NewTriangular(50, 100, 100)

	// AddSet should return a SetRef
	ref, err := fv.AddSet(set.NewFuzzySet("Hot", mf))
	if err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}

	if ref == nil {
		t.Fatal("AddSet returned nil SetRef")
	}

	if ref.Variable != "Temperature" {
		t.Errorf("Expected SetRef.Variable 'Temperature', got '%s'", ref.Variable)
	}

	if ref.Set != "Hot" {
		t.Errorf("Expected SetRef.Set 'Hot', got '%s'", ref.Set)
	}
}

func TestSetRef_MultipleVariables(t *testing.T) {
	temp, _ := NewFuzzyVariable("Temperature", 0, 100)
	humidity, _ := NewFuzzyVariable("Humidity", 0, 100)

	mf1, _ := membership.NewTriangular(0, 0, 30)
	coldTemp, _ := temp.AddSet(set.NewFuzzySet("Cold", mf1))
	mf2, _ := membership.NewTriangular(70, 100, 100)
	hotTemp, _ := temp.AddSet(set.NewFuzzySet("Hot", mf2))
	mf3, _ := membership.NewTriangular(0, 0, 40)
	dryHumidity, _ := humidity.AddSet(set.NewFuzzySet("Dry", mf3))

	// Verify each reference points to the correct variable and set
	if coldTemp.Variable != "Temperature" || coldTemp.Set != "Cold" {
		t.Error("coldTemp reference incorrect")
	}

	if hotTemp.Variable != "Temperature" || hotTemp.Set != "Hot" {
		t.Error("hotTemp reference incorrect")
	}

	if dryHumidity.Variable != "Humidity" || dryHumidity.Set != "Dry" {
		t.Error("dryHumidity reference incorrect")
	}
}

func TestSetRef_TypeSafeRuleConstruction(t *testing.T) {
	// Demonstrate type-safe pattern
	temp, _ := NewFuzzyVariable("Temperature", 0, 50)
	fan, _ := NewFuzzyVariable("FanSpeed", 0, 100)

	// These references can be used for compile-time safe rule building
	mf1, _ := membership.NewTrapezoidal(0, 0, 12, 20)
	coldSet, _ := temp.AddSet(set.NewFuzzySet("Cold", mf1))
	mf2, _ := membership.NewTrapezoidal(28, 35, 50, 51)
	hotSet, _ := temp.AddSet(set.NewFuzzySet("Hot", mf2))
	mf3, _ := membership.NewTriangular(0, 0, 50)
	lowFan, _ := fan.AddSet(set.NewFuzzySet("Low", mf3))
	mf4, _ := membership.NewTriangular(50, 100, 100)
	highFan, _ := fan.AddSet(set.NewFuzzySet("High", mf4))

	// Verify all references are correct
	refs := []*SetRef{coldSet, hotSet, lowFan, highFan}
	expectedVars := []string{"Temperature", "Temperature", "FanSpeed", "FanSpeed"}
	expectedSets := []string{"Cold", "Hot", "Low", "High"}

	for i, ref := range refs {
		if ref.Variable != expectedVars[i] {
			t.Errorf("Reference %d: expected variable %s, got %s", i, expectedVars[i], ref.Variable)
		}
		if ref.Set != expectedSets[i] {
			t.Errorf("Reference %d: expected set %s, got %s", i, expectedSets[i], ref.Set)
		}
	}
}
