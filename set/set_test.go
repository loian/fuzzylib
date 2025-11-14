package set

import (
	"github.com/loian/fuzzylib/membership"
	"testing"
)

func TestNewFuzzySet(t *testing.T) {
	memFunc, _ := membership.NewTriangular(0, 5, 10)
	fuzzySet, err := NewFuzzySet("TestSet", memFunc)
	if err != nil {
		t.Fatalf("Failed to create fuzzy set: %v", err)
	}

	if fuzzySet.Name != "TestSet" {
		t.Errorf("Expected name 'TestSet', got %s", fuzzySet.Name)
	}

	if fuzzySet.MembershipFunc == nil {
		t.Errorf("Expected non-nil membership function")
	}
}

func TestFuzzySet_Evaluate(t *testing.T) {
	memFunc, _ := membership.NewTriangular(0, 5, 10)
	fuzzySet, err := NewFuzzySet("TestSet", memFunc)
	if err != nil {
		t.Fatalf("Failed to create fuzzy set: %v", err)
	}

	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0.0},
		{2.5, 0.5},
		{5, 1.0},
		{7.5, 0.5},
		{10, 0.0},
	}

	for _, test := range tests {
		result := fuzzySet.Evaluate(test.input)
		if result != test.expected {
			t.Errorf("At %f: expected %f, got %f", test.input, test.expected, result)
		}
	}
}

func TestFuzzySet_WithDifferentMembershipFunctions(t *testing.T) {
	tri, _ := membership.NewTriangular(0, 5, 10)
	trap, _ := membership.NewTrapezoidal(0, 2, 8, 10)
	gauss, _ := membership.NewGaussian(5, 2)

	tests := []struct {
		name     string
		memFunc  membership.MembershipFunction
		testVal  float64
		testName string
	}{
		{
			name:     "Triangular",
			memFunc:  tri,
			testVal:  5,
			testName: "Peak",
		},
		{
			name:     "Trapezoidal",
			memFunc:  trap,
			testVal:  5,
			testName: "Plateau",
		},
		{
			name:     "Gaussian",
			memFunc:  gauss,
			testVal:  5,
			testName: "Center",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fuzzySet, err := NewFuzzySet(test.testName, test.memFunc)
			if err != nil {
				t.Fatalf("Failed to create fuzzy set: %v", err)
			}
			result := fuzzySet.Evaluate(test.testVal)

			// All should have high membership at their center
			if result < 0.9 {
				t.Errorf("Expected high membership (>0.9) at center, got %f", result)
			}
		})
	}
}
