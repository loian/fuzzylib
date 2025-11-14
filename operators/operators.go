package operators

import (
	"errors"
	"fmt"
)

// Operator defines the interface for fuzzy logic operators
type Operator interface {
	// Apply applies the operator to one or more membership degrees
	// Returns ErrInvalidMembership if any value falls outside [0, 1].
	Apply(values ...float64) (float64, error)
}

// ErrInvalidMembership indicates that at least one membership degree
// provided to an operator call was outside the valid [0, 1] range.
var ErrInvalidMembership = errors.New("membership degree must be in range [0, 1]")

// InvalidMembershipError captures the offending value that was out of range.
type InvalidMembershipError struct {
	Value float64
}

func (e *InvalidMembershipError) Error() string {
	return fmt.Sprintf("membership degree %.4f is outside [0, 1]", e.Value)
}

func (e *InvalidMembershipError) Unwrap() error {
	return ErrInvalidMembership
}

// MinOperator implements the AND operator using minimum
type MinOperator struct{}

// Apply returns the minimum of all input values.
// All values should be in range [0, 1] for valid membership degrees.
func (m *MinOperator) Apply(values ...float64) (float64, error) {
	if len(values) == 0 {
		return 0.0, nil
	}
	min := values[0]
	var invalidErr error
	if min < 0 {
		invalidErr = &InvalidMembershipError{Value: min}
		min = 0
	} else if min > 1 {
		invalidErr = &InvalidMembershipError{Value: min}
		min = 1
	}
	for _, raw := range values[1:] {
		v := raw
		// Clamp values to valid membership degree range [0, 1]
		if v < 0 {
			if invalidErr == nil {
				invalidErr = &InvalidMembershipError{Value: raw}
			}
			v = 0
		} else if v > 1 {
			if invalidErr == nil {
				invalidErr = &InvalidMembershipError{Value: raw}
			}
			v = 1
		}
		if v < min {
			min = v
		}
	}
	// Clamp min to valid range
	if min < 0 {
		min = 0
	}
	if min > 1 {
		min = 1
	}
	if invalidErr != nil {
		return min, invalidErr
	}
	return min, nil
}

// MaxOperator implements the OR operator using maximum
type MaxOperator struct{}

// Apply returns the maximum of all input values.
// All values should be in range [0, 1] for valid membership degrees.
func (m *MaxOperator) Apply(values ...float64) (float64, error) {
	if len(values) == 0 {
		return 0.0, nil
	}
	max := values[0]
	var invalidErr error
	if max < 0 {
		invalidErr = &InvalidMembershipError{Value: max}
		max = 0
	} else if max > 1 {
		invalidErr = &InvalidMembershipError{Value: max}
		max = 1
	}
	for _, raw := range values[1:] {
		v := raw
		// Clamp values to valid membership degree range [0, 1]
		if v < 0 {
			if invalidErr == nil {
				invalidErr = &InvalidMembershipError{Value: raw}
			}
			v = 0
		} else if v > 1 {
			if invalidErr == nil {
				invalidErr = &InvalidMembershipError{Value: raw}
			}
			v = 1
		}
		if v > max {
			max = v
		}
	}
	// Clamp max to valid range
	if max < 0 {
		max = 0
	}
	if max > 1 {
		max = 1
	}
	if invalidErr != nil {
		return max, invalidErr
	}
	return max, nil
}

// NotOperator implements the NOT operator (complement)
type NotOperator struct{}

// Apply returns the complement of a single value (1 - x).
// Input value should be in range [0, 1] for valid membership degree.
func (n *NotOperator) Apply(values ...float64) (float64, error) {
	if len(values) == 0 {
		return 1.0, nil
	}
	v := values[0]
	var invalidErr error
	// Clamp input to valid membership degree range [0, 1]
	if v < 0 {
		invalidErr = &InvalidMembershipError{Value: v}
		v = 0
	} else if v > 1 {
		invalidErr = &InvalidMembershipError{Value: v}
		v = 1
	}
	result := 1.0 - v
	if invalidErr != nil {
		return result, invalidErr
	}
	return result, nil
}

// Zadeh operators (most common)

// AND is the Zadeh AND operator (MIN)
var AND = &MinOperator{}

// OR is the Zadeh OR operator (MAX)
var OR = &MaxOperator{}

// NOT is the Zadeh NOT operator (complement)
var NOT = &NotOperator{}
