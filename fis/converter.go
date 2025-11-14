package fis

import (
	"fmt"

	"github.com/loian/fuzzylib/inference"
	"github.com/loian/fuzzylib/membership"
	"github.com/loian/fuzzylib/operators"
	"github.com/loian/fuzzylib/rule"
	"github.com/loian/fuzzylib/set"
	"github.com/loian/fuzzylib/variable"
)

// LoadFIS parses a .fis file and returns a configured MamdaniInferenceSystem
func LoadFIS(filename string) (*inference.MamdaniInferenceSystem, error) {
	model, err := ParseFIS(filename)
	if err != nil {
		return nil, err
	}

	return ConvertToInferenceSystem(model)
}

// ConvertToInferenceSystem converts a FISModel to a MamdaniInferenceSystem
func ConvertToInferenceSystem(model *FISModel) (*inference.MamdaniInferenceSystem, error) {
	// Validate system type
	if model.System.Type != "mamdani" && model.System.Type != "" {
		return nil, fmt.Errorf("only mamdani FIS supported, got: %s", model.System.Type)
	}

	fis := inference.NewMamdaniInferenceSystem()

	// Map defuzzification method
	defuzzMethod := mapDefuzzMethod(model.System.DefuzzMethod)
	if err := fis.SetDefuzzificationMethod(defuzzMethod); err != nil {
		return nil, fmt.Errorf("error setting defuzzification method: %w", err)
	}

	// Convert input variables
	for i, inputSpec := range model.Inputs {
		inputVar, err := convertVariable(inputSpec)
		if err != nil {
			return nil, fmt.Errorf("error converting input variable #%d ('%s'): %w", i+1, inputSpec.Name, err)
		}
		if err := fis.AddInputVariable(inputVar); err != nil {
			return nil, fmt.Errorf("error adding input variable #%d ('%s'): %w", i+1, inputSpec.Name, err)
		}
	}

	// Convert output variables
	for i, outputSpec := range model.Outputs {
		outputVar, err := convertVariable(outputSpec)
		if err != nil {
			return nil, fmt.Errorf("error converting output variable #%d ('%s'): %w", i+1, outputSpec.Name, err)
		}
		if err := fis.AddOutputVariable(outputVar); err != nil {
			return nil, fmt.Errorf("error adding output variable #%d ('%s'): %w", i+1, outputSpec.Name, err)
		}
	}

	// Convert rules
	for i, ruleSpec := range model.Rules {
		r, err := convertRule(ruleSpec, model.Inputs, model.Outputs)
		if err != nil {
			return nil, fmt.Errorf("error converting rule #%d: %w", i+1, err)
		}
		if err := fis.AddRule(r); err != nil {
			return nil, fmt.Errorf("error adding rule #%d: %w", i+1, err)
		}
	}

	return fis, nil
}

// convertVariable converts a VariableSection to a FuzzyVariable
func convertVariable(spec VariableSection) (*variable.FuzzyVariable, error) {
	v, err := variable.NewFuzzyVariable(spec.Name, spec.Range[0], spec.Range[1])
	if err != nil {
		return nil, fmt.Errorf("invalid variable specification: %w", err)
	}

	for i, mfSpec := range spec.MFs {
		mf, err := convertMembershipFunction(mfSpec)
		if err != nil {
			return nil, fmt.Errorf("error in membership function #%d ('%s'): %w", i+1, mfSpec.Name, err)
		}
		if _, err := v.AddSet(set.NewFuzzySet(mfSpec.Name, mf)); err != nil {
			return nil, fmt.Errorf("error adding membership function #%d ('%s'): %w", i+1, mfSpec.Name, err)
		}
	}

	return v, nil
}

// convertMembershipFunction converts a MembershipFunctionSpec to a membership.MembershipFunction
func convertMembershipFunction(spec MembershipFunctionSpec) (membership.MembershipFunction, error) {
	switch spec.Type {
	case "trimf":
		if len(spec.Params) != 3 {
			return nil, fmt.Errorf("trimf requires 3 parameters, got %d: %v", len(spec.Params), spec.Params)
		}
		mf, err := membership.NewTriangular(spec.Params[0], spec.Params[1], spec.Params[2])
		if err != nil {
			return nil, fmt.Errorf("invalid trimf parameters: %w", err)
		}
		return mf, nil

	case "trapmf":
		if len(spec.Params) != 4 {
			return nil, fmt.Errorf("trapmf requires 4 parameters, got %d: %v", len(spec.Params), spec.Params)
		}
		mf, err := membership.NewTrapezoidal(spec.Params[0], spec.Params[1], spec.Params[2], spec.Params[3])
		if err != nil {
			return nil, fmt.Errorf("invalid trapmf parameters: %w", err)
		}
		return mf, nil

	case "gaussmf":
		if len(spec.Params) != 2 {
			return nil, fmt.Errorf("gaussmf requires 2 parameters (sigma, center), got %d: %v", len(spec.Params), spec.Params)
		}
		// gaussmf params are [sigma, center]
		mf, err := membership.NewGaussian(spec.Params[1], spec.Params[0])
		if err != nil {
			return nil, fmt.Errorf("invalid gaussmf parameters: %w", err)
		}
		return mf, nil

	default:
		return nil, fmt.Errorf("unsupported membership function type '%s' (supported: trimf, trapmf, gaussmf)", spec.Type)
	}
}

// convertRule converts a RuleSpec to a Rule
func convertRule(spec RuleSpec, inputs, outputs []VariableSection) (*rule.Rule, error) {
	// Validate indices
	if len(spec.Consequents) == 0 {
		return nil, fmt.Errorf("rule must have at least one consequent")
	}

	// Get first non-zero consequent
	var outputVarIdx, outputSetIdx int
	for i, idx := range spec.Consequents {
		if idx != 0 {
			outputVarIdx = i
			outputSetIdx = idx - 1 // Convert from 1-based to 0-based
			break
		}
	}

	if outputVarIdx >= len(outputs) || outputSetIdx >= len(outputs[outputVarIdx].MFs) {
		return nil, fmt.Errorf("invalid output index in rule")
	}

	outputVar := outputs[outputVarIdx].Name
	outputSet := outputs[outputVarIdx].MFs[outputSetIdx].Name

	// Determine operator
	var op operators.Operator = operators.AND
	if spec.Connection == 2 {
		op = operators.OR
	}

	// Create rule
	r, err := rule.NewRule(rule.RuleCondition{
		Variable: outputVar,
		Set:      outputSet,
	}, op)
	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	// Add conditions
	for i, idx := range spec.Antecedents {
		if idx == 0 {
			// Don't care - skip this input
			continue
		}

		if i >= len(inputs) {
			return nil, fmt.Errorf("antecedent index %d exceeds number of inputs %d", i, len(inputs))
		}

		isNegated := idx < 0
		setIdx := idx - 1
		if isNegated {
			setIdx = -idx - 1
		}

		if setIdx >= len(inputs[i].MFs) {
			return nil, fmt.Errorf("invalid MF index %d for input %s", setIdx+1, inputs[i].Name)
		}

		condition := rule.RuleCondition{
			Variable: inputs[i].Name,
			Set:      inputs[i].MFs[setIdx].Name,
			Negated:  isNegated,
		}

		r.Conditions = append(r.Conditions, condition)
	}

	// Set weight (validate it's in valid range)
	if err := r.SetWeight(spec.Weight); err != nil {
		return nil, fmt.Errorf("invalid rule weight %.2f: %w", spec.Weight, err)
	}

	return r, nil
}

// mapDefuzzMethod maps FIS defuzzification method names to internal constants
func mapDefuzzMethod(fisMethod string) string {
	switch fisMethod {
	case "centroid", "bisector":
		return inference.DefuzzCOG
	case "mom":
		return inference.DefuzzMOM
	case "som", "lom":
		return inference.DefuzzFOM
	default:
		// Default to MOM
		return inference.DefuzzMOM
	}
}
