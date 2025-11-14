package operators

import (
	"errors"
	"math"
	"testing"
)

const epsilon = 1e-9

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}
func TestMinOperator_AND(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"both high", []float64{0.8, 0.9}, 0.8},
		{"one low", []float64{0.2, 0.9}, 0.2},
		{"three values", []float64{0.5, 0.7, 0.3}, 0.3},
		{"empty", []float64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AND.Apply(tt.values...)
			if err != nil {
				t.Fatalf("AND.Apply returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("AND.Apply(%v) = %f, expected %f", tt.values, result, tt.expected)
			}
		})
	}
}

func TestMaxOperator_OR(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"both low", []float64{0.2, 0.1}, 0.2},
		{"one high", []float64{0.2, 0.9}, 0.9},
		{"three values", []float64{0.5, 0.7, 0.3}, 0.7},
		{"empty", []float64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OR.Apply(tt.values...)
			if err != nil {
				t.Fatalf("OR.Apply returned error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("OR.Apply(%v) = %f, expected %f", tt.values, result, tt.expected)
			}
		})
	}
}

func TestNotOperator(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected float64
	}{
		{"high", 0.8, 0.2},
		{"low", 0.2, 0.8},
		{"zero", 0.0, 1.0},
		{"one", 1.0, 0.0},
		{"mid", 0.5, 0.5},
		{"empty", 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NOT.Apply(tt.value)
			if err != nil {
				t.Fatalf("NOT.Apply returned error: %v", err)
			}
			if !floatEqual(result, tt.expected) {
				t.Errorf("NOT.Apply(%f) = %f, expected %f", tt.value, result, tt.expected)
			}
		})
	}
}

func TestOperatorInterface(t *testing.T) {
	var op Operator

	op = AND
	if result, err := op.Apply(0.5, 0.7); err != nil || result != 0.5 {
		t.Error("AND doesn't implement Operator interface correctly")
	}

	op = OR
	if result, err := op.Apply(0.5, 0.7); err != nil || result != 0.7 {
		t.Error("OR doesn't implement Operator interface correctly")
	}

	op = NOT
	if result, err := op.Apply(0.5); err != nil || result != 0.5 {
		t.Error("NOT doesn't implement Operator interface correctly")
	}
}

func TestOperatorInvalidInputs(t *testing.T) {
	if _, err := AND.Apply(-0.2, 0.4); err == nil || !errors.Is(err, ErrInvalidMembership) {
		t.Fatalf("expected ErrInvalidMembership from AND, got %v", err)
	}
	if _, err := OR.Apply(0.2, 1.5); err == nil || !errors.Is(err, ErrInvalidMembership) {
		t.Fatalf("expected ErrInvalidMembership from OR, got %v", err)
	}
	if _, err := NOT.Apply(1.2); err == nil || !errors.Is(err, ErrInvalidMembership) {
		t.Fatalf("expected ErrInvalidMembership from NOT, got %v", err)
	}
}
