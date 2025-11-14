package rule

import (
	"fmt"
	"github.com/loian/fuzzylib/operators"
)

// RuleCondition represents a condition in a rule (e.g., "Temperature IS Cold").
// It specifies that a particular variable should match a particular fuzzy set.
// Used in both rule antecedents (IF conditions) and consequents (THEN outputs).
type RuleCondition struct {
	Variable string // Variable name (e.g., "Temperature")
	Set      string // Fuzzy set name (e.g., "Cold")
	Negated  bool   // If true, apply NOT operator to this condition
}

// Rule represents an IF-THEN fuzzy rule
type Rule struct {
	Conditions []RuleCondition    // IF conditions (antecedents)
	Output     RuleCondition      // THEN output (consequent)
	Weight     float64            // Rule weight (0-1, default 1.0)
	Operator   operators.Operator // AND/OR operator for combining conditions
}

// NewRule creates a new fuzzy rule with default weight of 1.0 and AND operator.
// Returns error if output variable or set name is empty, or if output is negated.
func NewRule(output RuleCondition, operator operators.Operator) (*Rule, error) {
	if output.Variable == "" {
		return nil, fmt.Errorf("output variable name cannot be empty")
	}
	if output.Set == "" {
		return nil, fmt.Errorf("output set name cannot be empty")
	}
	if output.Negated {
		return nil, fmt.Errorf("output condition cannot be negated: negation is only valid for input conditions")
	}
	if operator == nil {
		operator = operators.AND
	}
	return &Rule{
		Conditions: make([]RuleCondition, 0),
		Output:     output,
		Weight:     1.0,
		Operator:   operator,
	}, nil
}

// AddCondition adds a condition to the rule.
// Returns error if variable or set name is empty.
func (r *Rule) AddCondition(variable, set string) error {
	return r.AddConditionEx(variable, set, false)
}

// AddConditionEx adds a condition to the rule with optional negation.
// If negated is true, the NOT operator will be applied to this condition.
// Returns error if variable or set name is empty.
func (r *Rule) AddConditionEx(variable, set string, negated bool) error {
	if variable == "" {
		return fmt.Errorf("condition variable name cannot be empty")
	}
	if set == "" {
		return fmt.Errorf("condition set name cannot be empty")
	}
	r.Conditions = append(r.Conditions, RuleCondition{
		Variable: variable,
		Set:      set,
		Negated:  negated,
	})
	return nil
}

// SetWeight sets the rule weight. Weight must be in range [0, 1].
// Returns error if weight is out of bounds.
func (r *Rule) SetWeight(weight float64) error {
	if weight < 0 || weight > 1 {
		return fmt.Errorf("weight must be in range [0, 1], got %.2f", weight)
	}
	r.Weight = weight
	return nil
}

// Evaluate evaluates the rule given input membership values.
// membershipMap: map[variableName][setName]membershipDegree
// Returns error if the rule has no conditions.
func (r *Rule) Evaluate(membershipMap map[string]map[string]float64) (float64, error) {
	if len(r.Conditions) == 0 {
		return 0, fmt.Errorf("cannot evaluate rule with no conditions")
	}

	// Get membership degrees for all conditions
	values := make([]float64, len(r.Conditions))
	for i, cond := range r.Conditions {
		if varMap, ok := membershipMap[cond.Variable]; ok {
			if degree, ok := varMap[cond.Set]; ok {
				if cond.Negated {
					// Apply NOT operator: 1 - membership_degree
					values[i] = 1.0 - degree
				} else {
					values[i] = degree
				}
			}
		}
	}

	// Apply operator to combine conditions
	result, err := r.Operator.Apply(values...)
	if err != nil {
		return 0, fmt.Errorf("error applying operator for rule output '%s.%s': %w", r.Output.Variable, r.Output.Set, err)
	}

	// Apply weight
	return result * r.Weight, nil
}
