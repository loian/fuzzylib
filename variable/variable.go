package variable

import (
	"fmt"
	"github.com/loian/fuzzylib/set"
)

// SetRef is a type-safe reference to a fuzzy set within a variable.
// It captures both the variable name and set name, enabling compile-time
// checking and IDE autocomplete when building rules.
type SetRef struct {
	Variable string // Variable name (e.g., "Temperature")
	Set      string // Set name (e.g., "Hot")
}

// FuzzyVariable represents a linguistic variable with multiple fuzzy sets
type FuzzyVariable struct {
	Name     string
	MinValue float64
	MaxValue float64
	Sets     map[string]*set.FuzzySet
}

// NewFuzzyVariable creates a new fuzzy variable.
// Returns error if minValue >= maxValue or if name is empty.
func NewFuzzyVariable(name string, minValue, maxValue float64) (*FuzzyVariable, error) {
	if name == "" {
		return nil, fmt.Errorf("variable name cannot be empty")
	}
	if minValue >= maxValue {
		return nil, fmt.Errorf("minValue (%.2f) must be less than maxValue (%.2f)", minValue, maxValue)
	}
	return &FuzzyVariable{
		Name:     name,
		MinValue: minValue,
		MaxValue: maxValue,
		Sets:     make(map[string]*set.FuzzySet),
	}, nil
}

// AddSet adds a fuzzy set to the variable and returns a type-safe reference.
// The returned SetRef can be used for compile-time safe rule construction.
// Returns error if a set with the same name already exists or if the set name is empty.
//
// Example:
//
//	temp, _ := NewFuzzyVariable("Temperature", 0, 50)
//	hotMF, _ := membership.NewTriangular(30, 50, 50)
//	hotSet, err := temp.AddSet(set.NewFuzzySet("Hot", hotMF))
//	if err != nil {
//	    // Handle duplicate set name or invalid set
//	}
func (fv *FuzzyVariable) AddSet(fuzzySet *set.FuzzySet, err error) (*SetRef, error) {
	if err != nil {
		return nil, err
	}
	if fuzzySet.Name == "" {
		return nil, fmt.Errorf("set name cannot be empty")
	}
	if _, exists := fv.Sets[fuzzySet.Name]; exists {
		return nil, fmt.Errorf("set '%s' already exists in variable '%s'", fuzzySet.Name, fv.Name)
	}
	fv.Sets[fuzzySet.Name] = fuzzySet
	return &SetRef{
		Variable: fv.Name,
		Set:      fuzzySet.Name,
	}, nil
}

// Fuzzify returns the membership degrees for all sets given a crisp value
func (fv *FuzzyVariable) Fuzzify(value float64) map[string]float64 {
	result := make(map[string]float64)
	for name, fuzzySet := range fv.Sets {
		result[name] = fuzzySet.Evaluate(value)
	}
	return result
}

// IsValid checks if a value is within the variable's domain
func (fv *FuzzyVariable) IsValid(value float64) bool {
	return value >= fv.MinValue && value <= fv.MaxValue
}
