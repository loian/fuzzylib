package inference

import (
	"fmt"
	"github.com/loian/fuzzylib/operators"
	"github.com/loian/fuzzylib/rule"
	"github.com/loian/fuzzylib/variable"
	"math"
)

// DefaultResolution is the default sampling resolution used for defuzzification.
// It is exported so callers and tests can adjust global default if needed.
var DefaultResolution = 1000

// epsilon is the tolerance for floating point comparisons
const epsilon = 1e-9

// Defuzzification method constants
const (
	DefuzzCOG = "centroid" // Center of Gravity (default)
	DefuzzMOM = "mom"      // Mean of Maximum
	DefuzzFOM = "fom"      // First of Maximum
	DefuzzLOM = "lom"      // Last of Maximum (mapped to FOM)
	DefuzzSOM = "som"      // Smallest of Maximum (mapped to FOM)
)

// MamdaniInferenceSystem represents a complete Mamdani FIS
type MamdaniInferenceSystem struct {
	InputVariables  map[string]*variable.FuzzyVariable
	OutputVariables map[string]*variable.FuzzyVariable
	Rules           []*rule.Rule
	// Resolution controls the number of sample points used during defuzzification.
	// Higher values increase accuracy but also CPU cost.
	Resolution int
	// DefuzzMethod specifies which defuzzification method to use: "centroid", "mom", "fom"
	DefuzzMethod string
}

// NewMamdaniInferenceSystem creates a new inference system
func NewMamdaniInferenceSystem() *MamdaniInferenceSystem {
	return &MamdaniInferenceSystem{
		InputVariables:  make(map[string]*variable.FuzzyVariable),
		OutputVariables: make(map[string]*variable.FuzzyVariable),
		Rules:           make([]*rule.Rule, 0),
		Resolution:      DefaultResolution,
		DefuzzMethod:    DefuzzMOM, // Default to MOM (current behavior)
	}
}

// SetResolution sets the sampling resolution used for defuzzification.
// Resolution must be > 0. Returns error if resolution is invalid.
func (fis *MamdaniInferenceSystem) SetResolution(res int) error {
	if res <= 0 {
		return fmt.Errorf("resolution must be > 0, got %d", res)
	}
	fis.Resolution = res
	return nil
}

// SetDefuzzificationMethod sets the defuzzification method.
// Valid methods: "centroid", "mom", "fom", "lom", "som"
// Returns error if method is not recognized.
func (fis *MamdaniInferenceSystem) SetDefuzzificationMethod(method string) error {
	switch method {
	case DefuzzCOG, DefuzzMOM, DefuzzFOM, DefuzzLOM, DefuzzSOM:
		fis.DefuzzMethod = method
		return nil
	default:
		return fmt.Errorf("invalid defuzzification method '%s': must be one of: centroid, mom, fom, lom, som", method)
	}
}

// AddInputVariable adds an input variable.
// Returns error if a variable with the same name already exists.
func (fis *MamdaniInferenceSystem) AddInputVariable(v *variable.FuzzyVariable) error {
	if _, exists := fis.InputVariables[v.Name]; exists {
		return fmt.Errorf("input variable '%s' already exists", v.Name)
	}
	fis.InputVariables[v.Name] = v
	return nil
}

// AddOutputVariable adds an output variable.
// Returns error if a variable with the same name already exists.
func (fis *MamdaniInferenceSystem) AddOutputVariable(v *variable.FuzzyVariable) error {
	if _, exists := fis.OutputVariables[v.Name]; exists {
		return fmt.Errorf("output variable '%s' already exists", v.Name)
	}
	fis.OutputVariables[v.Name] = v
	return nil
}

// AddRule adds a rule to the system.
// Returns error if the rule references non-existent variables or sets, or if the rule has no conditions.
func (fis *MamdaniInferenceSystem) AddRule(r *rule.Rule) error {
	// Validate rule has at least one condition
	if len(r.Conditions) == 0 {
		return fmt.Errorf("rule must have at least one condition")
	}

	// Validate output variable and set exist
	outputVar, exists := fis.OutputVariables[r.Output.Variable]
	if !exists {
		return fmt.Errorf("rule references non-existent output variable '%s'", r.Output.Variable)
	}
	if _, exists := outputVar.Sets[r.Output.Set]; !exists {
		return fmt.Errorf("rule references non-existent output set '%s' in variable '%s'", r.Output.Set, r.Output.Variable)
	}

	// Validate all input conditions
	for i, cond := range r.Conditions {
		inputVar, exists := fis.InputVariables[cond.Variable]
		if !exists {
			return fmt.Errorf("rule condition %d references non-existent input variable '%s'", i+1, cond.Variable)
		}
		if _, exists := inputVar.Sets[cond.Set]; !exists {
			return fmt.Errorf("rule condition %d references non-existent input set '%s' in variable '%s'", i+1, cond.Set, cond.Variable)
		}
	}

	fis.Rules = append(fis.Rules, r)
	return nil
}

// Infer performs Mamdani inference
// inputs: map[variableName]crispValue
// returns: map[variableName]crispOutput, error
// Returns error if:
//   - System is not properly configured (no inputs, outputs, or rules)
//   - Required input variables are missing
//   - Input values are outside variable bounds
//   - No rules fired (all membership degrees are zero)
func (fis *MamdaniInferenceSystem) Infer(inputs map[string]float64) (map[string]float64, error) {
	// Validate system is configured
	if len(fis.InputVariables) == 0 {
		return nil, fmt.Errorf("inference system has no input variables")
	}
	if len(fis.OutputVariables) == 0 {
		return nil, fmt.Errorf("inference system has no output variables")
	}
	if len(fis.Rules) == 0 {
		return nil, fmt.Errorf("inference system has no rules")
	}

	// Validate that all required inputs are provided
	for varName, inputVar := range fis.InputVariables {
		value, exists := inputs[varName]
		if !exists {
			return nil, fmt.Errorf("missing required input variable: %s", varName)
		}
		// Validate bounds
		if value < inputVar.MinValue || value > inputVar.MaxValue {
			return nil, fmt.Errorf("input value %.2f for variable '%s' is out of bounds [%.2f, %.2f]",
				value, varName, inputVar.MinValue, inputVar.MaxValue)
		}
	}

	// Step 1: Fuzzification - convert crisp inputs to membership degrees
	membershipMap := make(map[string]map[string]float64)
	for varName, crispValue := range inputs {
		if inputVar, ok := fis.InputVariables[varName]; ok {
			membershipMap[varName] = inputVar.Fuzzify(crispValue)
		}
	}

	// Step 2: Rule evaluation - fire rules and collect outputs
	outputMemberships := make(map[string]map[string]float64)
	for outputName := range fis.OutputVariables {
		outputMemberships[outputName] = make(map[string]float64)
	}

	for _, r := range fis.Rules {
		firingStrength, err := r.Evaluate(membershipMap)
		if err != nil {
			return nil, fmt.Errorf("error evaluating rule: %w", err)
		}
		// Each rule contributes to its output set
		if _, ok := outputMemberships[r.Output.Variable]; ok {
			// Use MAX aggregation for multiple rules firing to same set
			if current, exists := outputMemberships[r.Output.Variable][r.Output.Set]; exists {
				if firingStrength > current {
					outputMemberships[r.Output.Variable][r.Output.Set] = firingStrength
				}
			} else {
				outputMemberships[r.Output.Variable][r.Output.Set] = firingStrength
			}
		}
	}

	// Step 3: Defuzzification - convert fuzzy outputs to crisp values
	results := make(map[string]float64)
	for varName, outputVar := range fis.OutputVariables {
		var result float64
		var err error
		switch fis.DefuzzMethod {
		case DefuzzCOG:
			result, err = defuzzifyCOGWithResolution(outputVar, outputMemberships[varName], fis.Resolution)
		case DefuzzMOM:
			result, err = defuzzifyMOMWithResolution(outputVar, outputMemberships[varName], fis.Resolution)
		case DefuzzFOM, DefuzzLOM, DefuzzSOM:
			result, err = defuzzifyFOMWithResolution(outputVar, outputMemberships[varName], fis.Resolution)
		default:
			// Default to MOM if unknown method
			result, err = defuzzifyMOMWithResolution(outputVar, outputMemberships[varName], fis.Resolution)
		}
		if err != nil {
			return nil, fmt.Errorf("defuzzification failed for variable '%s': %w", varName, err)
		}
		results[varName] = result
	}

	return results, nil
}

// defuzzifyCOG uses Center of Gravity method for defuzzification
// defuzzifyCOG is a wrapper that calls the resolution-aware implementation
func defuzzifyCOG(outputVar *variable.FuzzyVariable, memberships map[string]float64) (float64, error) {
	return defuzzifyCOGWithResolution(outputVar, memberships, DefaultResolution)
}

func defuzzifyCOGWithResolution(outputVar *variable.FuzzyVariable, memberships map[string]float64, resolution int) (float64, error) {
	if len(memberships) == 0 {
		return 0, fmt.Errorf("no rules fired: all membership degrees are zero")
	}

	// Validate resolution
	if resolution <= 0 {
		resolution = DefaultResolution
	}

	// Calculate weighted sum and total weight
	numerator := 0.0
	denominator := 0.0

	step := (outputVar.MaxValue - outputVar.MinValue) / float64(resolution)

	for i := 0; i <= resolution; i++ {
		x := outputVar.MinValue + float64(i)*step

		// Get maximum membership degree at this point across all sets
		maxMembership := 0.0
		for setName, strength := range memberships {
			if outputSet, ok := outputVar.Sets[setName]; ok {
				degree := outputSet.Evaluate(x) * strength
				if degree > maxMembership {
					maxMembership = degree
				}
			}
		}

		numerator += x * maxMembership
		denominator += maxMembership
	}

	if denominator == 0 {
		return 0, fmt.Errorf("no rules fired: all membership degrees are zero")
	}

	return numerator / denominator, nil
}

// DefuzzifyMOM uses Mean of Maximum method
// defuzzifyMOM is a wrapper that calls the resolution-aware implementation
func defuzzifyMOM(outputVar *variable.FuzzyVariable, memberships map[string]float64) (float64, error) {
	return defuzzifyMOMWithResolution(outputVar, memberships, DefaultResolution)
}

func defuzzifyMOMWithResolution(outputVar *variable.FuzzyVariable, memberships map[string]float64, resolution int) (float64, error) {
	if len(memberships) == 0 {
		return 0, fmt.Errorf("no rules fired: all membership degrees are zero")
	}

	// Validate resolution
	if resolution <= 0 {
		resolution = DefaultResolution
	}

	maxMembership := 0.0
	var points []float64

	step := (outputVar.MaxValue - outputVar.MinValue) / float64(resolution)

	for i := 0; i <= resolution; i++ {
		x := outputVar.MinValue + float64(i)*step

		currentMax := 0.0
		for setName, strength := range memberships {
			if outputSet, ok := outputVar.Sets[setName]; ok {
				degree := outputSet.Evaluate(x) * strength
				if degree > currentMax {
					currentMax = degree
				}
			}
		}

		if i == 0 || currentMax > maxMembership {
			maxMembership = currentMax
			points = []float64{x}
		} else if math.Abs(currentMax-maxMembership) < epsilon {
			points = append(points, x)
		}
	}

	if len(points) == 0 || maxMembership == 0 {
		return 0, fmt.Errorf("no rules fired: all membership degrees are zero")
	}

	// Return average of maximum points
	sum := 0.0
	for _, p := range points {
		sum += p
	}
	return sum / float64(len(points)), nil
}

// DefuzzifyFOM uses First of Maximum method
// defuzzifyFOM is a wrapper that calls the resolution-aware implementation
func defuzzifyFOM(outputVar *variable.FuzzyVariable, memberships map[string]float64) (float64, error) {
	return defuzzifyFOMWithResolution(outputVar, memberships, DefaultResolution)
}

func defuzzifyFOMWithResolution(outputVar *variable.FuzzyVariable, memberships map[string]float64, resolution int) (float64, error) {
	if len(memberships) == 0 {
		return 0, fmt.Errorf("no rules fired: all membership degrees are zero")
	}

	// Validate resolution
	if resolution <= 0 {
		resolution = DefaultResolution
	}

	maxMembership := 0.0
	result := outputVar.MinValue

	step := (outputVar.MaxValue - outputVar.MinValue) / float64(resolution)

	for i := 0; i <= resolution; i++ {
		x := outputVar.MinValue + float64(i)*step

		currentMax := 0.0
		for setName, strength := range memberships {
			if outputSet, ok := outputVar.Sets[setName]; ok {
				degree := outputSet.Evaluate(x) * strength
				if degree > currentMax {
					currentMax = degree
				}
			}
		}

		if currentMax > maxMembership {
			maxMembership = currentMax
			result = x
		}
	}

	if maxMembership == 0 {
		return 0, fmt.Errorf("no rules fired: all membership degrees are zero")
	}

	return result, nil
}

// RuleBuilder is a helper for building rules with fluent API
type RuleBuilder struct {
	output rule.RuleCondition
	op     operators.Operator
	conds  []rule.RuleCondition
	weight float64
}

// NewRuleBuilder creates a new rule builder using string-based variable and set names.
// Weight defaults to 1.0. Use Weight() method to set a different value.
// For type-safe construction, use NewRuleBuilderRef instead.
// Returns error if weight is not in range [0, 1].
func NewRuleBuilder(outputVar, outputSet string, weight ...float64) (*RuleBuilder, error) {
	w := 1.0
	if len(weight) > 0 {
		w = weight[0]
		if w < 0 || w > 1 {
			return nil, fmt.Errorf("rule weight must be in range [0, 1], got %.2f", w)
		}
	}
	return &RuleBuilder{
		output: rule.RuleCondition{Variable: outputVar, Set: outputSet},
		op:     operators.AND,
		conds:  make([]rule.RuleCondition, 0),
		weight: w,
	}, nil
}

// NewRuleBuilderRef creates a new rule builder using a type-safe SetRef.
// Weight defaults to 1.0. Use Weight() method to set a different value.
// This provides compile-time checking and IDE autocomplete.
// Returns error if weight is not in range [0, 1].
//
// Example:
//
//	hotSet, _ := temp.AddSet(set.NewFuzzySet("Hot", membership.NewTriangular(30, 50, 50)))
//	highFan, _ := fan.AddSet(set.NewFuzzySet("High", membership.NewTriangular(67, 100, 100)))
//	rule, _ := inference.NewRuleBuilderRef(highFan).IfRef(hotSet).Build()
func NewRuleBuilderRef(outputRef *variable.SetRef, weight ...float64) (*RuleBuilder, error) {
	w := 1.0
	if len(weight) > 0 {
		w = weight[0]
		if w < 0 || w > 1 {
			return nil, fmt.Errorf("rule weight must be in range [0, 1], got %.2f", w)
		}
	}
	return &RuleBuilder{
		output: rule.RuleCondition{Variable: outputRef.Variable, Set: outputRef.Set},
		op:     operators.AND,
		conds:  make([]rule.RuleCondition, 0),
		weight: w,
	}, nil
}

// If adds a condition to the rule using string-based variable and set names.
// For type-safe construction, use IfRef instead.
func (rb *RuleBuilder) If(variable, set string) *RuleBuilder {
	rb.conds = append(rb.conds, rule.RuleCondition{Variable: variable, Set: set})
	return rb
}

// IfRef adds a condition to the rule using a type-safe SetRef.
// This provides compile-time checking and IDE autocomplete.
//
// Example:
//
//	hotSet := temp.AddSet(set.NewFuzzySet("Hot", membership.NewTriangular(30, 50, 50)))
//	rule := inference.NewRuleBuilderRef(highFan).IfRef(hotSet).Build()
func (rb *RuleBuilder) IfRef(setRef *variable.SetRef) *RuleBuilder {
	rb.conds = append(rb.conds, rule.RuleCondition{Variable: setRef.Variable, Set: setRef.Set})
	return rb
}

// And specifies AND operator
func (rb *RuleBuilder) And() *RuleBuilder {
	rb.op = operators.AND
	return rb
}

// Or specifies OR operator
func (rb *RuleBuilder) Or() *RuleBuilder {
	rb.op = operators.OR
	return rb
}

// Weight specifies rule weight (0-1). More natural than With() for weight setting.
// Weight must be in range [0, 1].
func (rb *RuleBuilder) Weight(weight float64) (*RuleBuilder, error) {
	if weight < 0 || weight > 1 {
		return nil, fmt.Errorf("rule weight must be in range [0, 1], got %.2f", weight)
	}
	rb.weight = weight
	return rb, nil
}

// With specifies weight. Deprecated: use Weight() instead for clarity.
// Weight must be in range [0, 1].
func (rb *RuleBuilder) With(weight float64) (*RuleBuilder, error) {
	if weight < 0 || weight > 1 {
		return nil, fmt.Errorf("rule weight must be in range [0, 1], got %.2f", weight)
	}
	rb.weight = weight
	return rb, nil
}

// Build creates the rule.
// Returns error if the rule configuration is invalid.
func (rb *RuleBuilder) Build() (*rule.Rule, error) {
	r, err := rule.NewRule(rb.output, rb.op)
	if err != nil {
		return nil, err
	}
	for _, cond := range rb.conds {
		r.Conditions = append(r.Conditions, cond)
	}
	// Use SetWeight to ensure validation
	if err := r.SetWeight(rb.weight); err != nil {
		return nil, fmt.Errorf("invalid rule weight: %w", err)
	}
	return r, nil
}
