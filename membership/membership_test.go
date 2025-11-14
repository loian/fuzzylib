package membership

import (
	"math"
	"testing"
)

const epsilon = 1e-9

// Helper function to compare floats
func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// ===== Triangular Tests =====

func TestTriangular_Peak(t *testing.T) {
	tri, _ := NewTriangular(0, 5, 10)
	result := tri.Evaluate(5)
	if !floatEqual(result, 1.0) {
		t.Errorf("Expected 1.0 at peak, got %f", result)
	}
}

func TestTriangular_LeftSlope(t *testing.T) {
	tri, _ := NewTriangular(0, 5, 10)
	result := tri.Evaluate(2.5)
	expected := 0.5 // (2.5 - 0) / (5 - 0)
	if !floatEqual(result, expected) {
		t.Errorf("Expected %f on left slope, got %f", expected, result)
	}
}

func TestTriangular_RightSlope(t *testing.T) {
	tri, _ := NewTriangular(0, 5, 10)
	result := tri.Evaluate(7.5)
	expected := 0.5 // (10 - 7.5) / (10 - 5)
	if !floatEqual(result, expected) {
		t.Errorf("Expected %f on right slope, got %f", expected, result)
	}
}

func TestTriangular_OutsideRange(t *testing.T) {
	tri, _ := NewTriangular(0, 5, 10)

	tests := []float64{-1, 0, 10, 15}
	for _, val := range tests {
		result := tri.Evaluate(val)
		if result != 0.0 {
			t.Errorf("Expected 0.0 outside range at %f, got %f", val, result)
		}
	}
}

func TestTriangular_Asymmetric(t *testing.T) {
	tri, _ := NewTriangular(0, 2, 10)

	// Test left side (steeper)
	left := tri.Evaluate(1)
	expected := 0.5 // (1 - 0) / (2 - 0)
	if !floatEqual(left, expected) {
		t.Errorf("Expected %f, got %f", expected, left)
	}

	// Test right side (gentler)
	right := tri.Evaluate(6)
	expected = 0.5 // (10 - 6) / (10 - 2)
	if !floatEqual(right, expected) {
		t.Errorf("Expected %f, got %f", expected, right)
	}
}

// ===== Trapezoidal Tests =====

func TestTrapezoidal_Peak(t *testing.T) {
	trap, _ := NewTrapezoidal(0, 2, 8, 10)

	tests := []float64{2, 5, 8}
	for _, val := range tests {
		result := trap.Evaluate(val)
		if !floatEqual(result, 1.0) {
			t.Errorf("Expected 1.0 at %f, got %f", val, result)
		}
	}
}

func TestTrapezoidal_LeftSlope(t *testing.T) {
	trap, _ := NewTrapezoidal(0, 2, 8, 10)
	result := trap.Evaluate(1)
	expected := 0.5 // (1 - 0) / (2 - 0)
	if !floatEqual(result, expected) {
		t.Errorf("Expected %f on left slope, got %f", expected, result)
	}
}

func TestTrapezoidal_RightSlope(t *testing.T) {
	trap, _ := NewTrapezoidal(0, 2, 8, 10)
	result := trap.Evaluate(9)
	expected := 0.5 // (10 - 9) / (10 - 8)
	if !floatEqual(result, expected) {
		t.Errorf("Expected %f on right slope, got %f", expected, result)
	}
}

func TestTrapezoidal_OutsideRange(t *testing.T) {
	trap, _ := NewTrapezoidal(0, 2, 8, 10)

	tests := []float64{-1, 0, 10, 15}
	for _, val := range tests {
		result := trap.Evaluate(val)
		if result != 0.0 {
			t.Errorf("Expected 0.0 outside range at %f, got %f", val, result)
		}
	}
}

// ===== Gaussian Tests =====

func TestGaussian_Center(t *testing.T) {
	gauss, _ := NewGaussian(5, 2)
	result := gauss.Evaluate(5)
	if !floatEqual(result, 1.0) {
		t.Errorf("Expected 1.0 at center, got %f", result)
	}
}

func TestGaussian_Symmetric(t *testing.T) {
	gauss, _ := NewGaussian(5, 2)

	// Test points equidistant from center
	left := gauss.Evaluate(3)
	right := gauss.Evaluate(7)

	if !floatEqual(left, right) {
		t.Errorf("Expected symmetric values, got left=%f right=%f", left, right)
	}
}

func TestGaussian_Width(t *testing.T) {
	// Wider gaussian should have higher values at distance
	narrow, _ := NewGaussian(0, 1)
	wide, _ := NewGaussian(0, 2)

	testVal := 2.0
	narrowResult := narrow.Evaluate(testVal)
	wideResult := wide.Evaluate(testVal)

	if wideResult <= narrowResult {
		t.Errorf("Wider gaussian should have higher value at distance 2: narrow=%f wide=%f", narrowResult, wideResult)
	}
}

func TestGaussian_DecaysWithDistance(t *testing.T) {
	gauss, _ := NewGaussian(5, 2)

	// Values should decrease as distance from center increases
	val1 := gauss.Evaluate(5.5)
	val2 := gauss.Evaluate(6.5)
	val3 := gauss.Evaluate(7.5)

	if val1 <= val2 || val2 <= val3 {
		t.Errorf("Expected decreasing values with distance: 5.5=%f 6.5=%f 7.5=%f", val1, val2, val3)
	}
}

// ===== Integration Tests =====

func TestConstructors(t *testing.T) {
	tri, _ := NewTriangular(1, 2, 3)
	if tri.A != 1 || tri.B != 2 || tri.C != 3 {
		t.Errorf("Triangular constructor failed")
	}

	trap, _ := NewTrapezoidal(1, 2, 3, 4)
	if trap.A != 1 || trap.B != 2 || trap.C != 3 || trap.D != 4 {
		t.Errorf("Trapezoidal constructor failed")
	}

	gauss, _ := NewGaussian(5, 2)
	if gauss.Center != 5 || gauss.Width != 2 {
		t.Errorf("Gaussian constructor failed")
	}
}

func TestMembershipFunctionInterface(t *testing.T) {
	var funcs []MembershipFunction

	tri, _ := NewTriangular(0, 5, 10)
	funcs = append(funcs, tri)
	trap, _ := NewTrapezoidal(0, 2, 8, 10)
	funcs = append(funcs, trap)
	gauss, _ := NewGaussian(5, 2)
	funcs = append(funcs, gauss)

	// All should be callable without error
	for _, f := range funcs {
		_ = f.Evaluate(5)
	}
}
