package fis

// FISModel represents the intermediate data structure for a .fis file
type FISModel struct {
	System  SystemSection
	Inputs  []VariableSection
	Outputs []VariableSection
	Rules   []RuleSpec
}

// SystemSection represents the [System] section of a .fis file
type SystemSection struct {
	Name         string
	Type         string // "mamdani" or "sugeno"
	Version      string
	NumInputs    int
	NumOutputs   int
	NumRules     int
	AndMethod    string // "min" or "prod"
	OrMethod     string // "max" or "probor"
	ImpMethod    string // "min" or "prod"
	AggMethod    string // "max", "sum", or "probor"
	DefuzzMethod string // "centroid", "mom", "fom", etc.
}

// VariableSection represents an [Input#] or [Output#] section
type VariableSection struct {
	Name   string
	Range  [2]float64
	NumMFs int
	MFs    []MembershipFunctionSpec
}

// MembershipFunctionSpec represents a membership function definition
type MembershipFunctionSpec struct {
	Name   string
	Type   string // "trimf", "trapmf", "gaussmf", etc.
	Params []float64
}

// RuleSpec represents a rule in compact numeric format
type RuleSpec struct {
	Antecedents []int   // MF indices for inputs (1-based, 0=don't care, negative=NOT)
	Consequents []int   // MF indices for outputs (1-based)
	Weight      float64 // Rule weight (default 1.0)
	Connection  int     // 1=AND, 2=OR
}
