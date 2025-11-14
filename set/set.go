package set

import (
	"fmt"
	"fuzzy/membership"
)

// FuzzySet represents a fuzzy set with a membership function
type FuzzySet struct {
	Name           string
	MembershipFunc membership.MembershipFunction
}

// NewFuzzySet creates a new fuzzy set.
// Returns error if name is empty or membership function is nil.
func NewFuzzySet(name string, mf membership.MembershipFunction) (*FuzzySet, error) {
	if name == "" {
		return nil, fmt.Errorf("fuzzy set name cannot be empty")
	}
	if mf == nil {
		return nil, fmt.Errorf("membership function cannot be nil")
	}
	return &FuzzySet{
		Name:           name,
		MembershipFunc: mf,
	}, nil
}

// Evaluate returns the membership degree for value x
func (fs *FuzzySet) Evaluate(x float64) float64 {
	return fs.MembershipFunc.Evaluate(x)
}
