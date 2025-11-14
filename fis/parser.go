package fis

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseFIS parses a .fis file and returns a FISModel
func ParseFIS(filename string) (*FISModel, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseFISReader(bufio.NewScanner(file))
}

// ParseFISString parses FIS content from a string
func ParseFISString(content string) (*FISModel, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	return ParseFISReader(scanner)
}

// ParseFISReader parses FIS content from a scanner
func ParseFISReader(scanner *bufio.Scanner) (*FISModel, error) {
	model := &FISModel{
		Inputs:  make([]VariableSection, 0),
		Outputs: make([]VariableSection, 0),
		Rules:   make([]RuleSpec, 0),
	}

	var currentSection string
	var currentVariable *VariableSection
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "%") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous variable section if any
			if currentVariable != nil {
				if strings.HasPrefix(currentSection, "Input") {
					model.Inputs = append(model.Inputs, *currentVariable)
				} else if strings.HasPrefix(currentSection, "Output") {
					model.Outputs = append(model.Outputs, *currentVariable)
				}
				currentVariable = nil
			}

			currentSection = strings.Trim(line, "[]")

			// Initialize new variable section
			if strings.HasPrefix(currentSection, "Input") || strings.HasPrefix(currentSection, "Output") {
				currentVariable = &VariableSection{
					MFs: make([]MembershipFunctionSpec, 0),
				}
			}
			continue
		}

		// Parse based on current section
		switch {
		case currentSection == "System":
			if err := parseSystemLine(&model.System, line); err != nil {
				return nil, fmt.Errorf("line %d: error parsing system line '%s': %w", lineNum, line, err)
			}
		case strings.HasPrefix(currentSection, "Input") || strings.HasPrefix(currentSection, "Output"):
			if currentVariable != nil {
				if err := parseVariableLine(currentVariable, line); err != nil {
					return nil, fmt.Errorf("line %d: error parsing variable line '%s': %w", lineNum, line, err)
				}
			}
		case currentSection == "Rules":
			rule, err := parseRuleLine(line, model.System.NumInputs, model.System.NumOutputs)
			if err != nil {
				return nil, fmt.Errorf("line %d: error parsing rule line '%s': %w", lineNum, line, err)
			}
			model.Rules = append(model.Rules, *rule)
		}
	}

	// Save last variable section
	if currentVariable != nil {
		if strings.HasPrefix(currentSection, "Input") {
			model.Inputs = append(model.Inputs, *currentVariable)
		} else if strings.HasPrefix(currentSection, "Output") {
			model.Outputs = append(model.Outputs, *currentVariable)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return model, nil
}

// parseSystemLine parses a line from the [System] section
func parseSystemLine(sys *SystemSection, line string) error {
	key, value, err := parseKeyValue(line)
	if err != nil {
		return err
	}

	switch key {
	case "Name":
		sys.Name = value
	case "Type":
		sys.Type = value
	case "Version":
		sys.Version = value
	case "NumInputs":
		sys.NumInputs, _ = strconv.Atoi(value)
	case "NumOutputs":
		sys.NumOutputs, _ = strconv.Atoi(value)
	case "NumRules":
		sys.NumRules, _ = strconv.Atoi(value)
	case "AndMethod":
		sys.AndMethod = value
	case "OrMethod":
		sys.OrMethod = value
	case "ImpMethod":
		sys.ImpMethod = value
	case "AggMethod":
		sys.AggMethod = value
	case "DefuzzMethod":
		sys.DefuzzMethod = value
	}

	return nil
}

// parseVariableLine parses a line from an [Input#] or [Output#] section
func parseVariableLine(v *VariableSection, line string) error {
	key, value, err := parseKeyValue(line)
	if err != nil {
		return err
	}

	switch key {
	case "Name":
		v.Name = value
	case "Range":
		rangeVals, err := parseArray(value)
		if err != nil || len(rangeVals) != 2 {
			return fmt.Errorf("invalid range format: %s", value)
		}
		v.Range = [2]float64{rangeVals[0], rangeVals[1]}
	case "NumMFs":
		v.NumMFs, _ = strconv.Atoi(value)
	default:
		// Check if it's a membership function definition (MF1, MF2, etc.)
		if strings.HasPrefix(key, "MF") {
			mf, err := parseMF(value)
			if err != nil {
				return err
			}
			v.MFs = append(v.MFs, *mf)
		}
	}

	return nil
}

// parseRuleLine parses a rule line: "1 2 0, 3 (1.0) : 1"
func parseRuleLine(line string, numInputs, numOutputs int) (*RuleSpec, error) {
	// Split by comma
	parts := strings.Split(line, ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid rule format")
	}

	// Parse antecedents
	antecedents, err := parseIndices(strings.TrimSpace(parts[0]), numInputs)
	if err != nil {
		return nil, err
	}

	// Parse consequents, weight, and connection
	rest := strings.TrimSpace(parts[1])

	// Extract weight if present: (1.0)
	weight := 1.0
	if idx := strings.Index(rest, "("); idx >= 0 {
		endIdx := strings.Index(rest, ")")
		if endIdx > idx {
			weightStr := rest[idx+1 : endIdx]
			weight, _ = strconv.ParseFloat(strings.TrimSpace(weightStr), 64)
			rest = strings.TrimSpace(rest[:idx] + rest[endIdx+1:])
		}
	}

	// Parse consequents and connection operator
	connection := 1 // AND default
	consequentPart := rest
	if idx := strings.Index(rest, ":"); idx >= 0 {
		consequentPart = strings.TrimSpace(rest[:idx])
		connectionStr := strings.TrimSpace(rest[idx+1:])
		connection, _ = strconv.Atoi(connectionStr)
	}

	consequents, err := parseIndices(consequentPart, numOutputs)
	if err != nil {
		return nil, err
	}

	return &RuleSpec{
		Antecedents: antecedents,
		Consequents: consequents,
		Weight:      weight,
		Connection:  connection,
	}, nil
}

// parseKeyValue parses a "Key=Value" or "Key='Value'" line
func parseKeyValue(line string) (key, value string, err error) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid key=value format: %s", line)
	}
	key = strings.TrimSpace(parts[0])
	value = strings.Trim(strings.TrimSpace(parts[1]), "'\"")
	return
}

// parseArray parses "[a b c]" or "[a, b, c]" into []float64
func parseArray(s string) ([]float64, error) {
	s = strings.Trim(s, "[]")
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)

	result := make([]float64, len(parts))
	for i, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", p)
		}
		result[i] = val
	}
	return result, nil
}

// parseMF parses a membership function definition: "'Cold':'trimf',[0 10 20]"
func parseMF(s string) (*MembershipFunctionSpec, error) {
	// Find the last colon before the bracket (params section)
	bracketIdx := strings.Index(s, "[")
	if bracketIdx < 0 {
		return nil, fmt.Errorf("invalid MF format, missing params: %s", s)
	}

	// Split the part before params by ':'
	beforeParams := s[:bracketIdx]
	paramsStr := s[bracketIdx:]

	parts := strings.Split(beforeParams, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid MF format: %s", s)
	}

	name := strings.Trim(strings.TrimSpace(parts[0]), "'\"")
	// Remove trailing comma and quotes from type
	mfType := strings.Trim(strings.TrimSpace(parts[1]), ",'\"")

	params, err := parseArray(paramsStr)
	if err != nil {
		return nil, err
	}

	return &MembershipFunctionSpec{
		Name:   name,
		Type:   mfType,
		Params: params,
	}, nil
}

// parseIndices parses space-separated integers: "1 2 0" -> []int{1, 2, 0}
func parseIndices(s string, expectedCount int) ([]int, error) {
	parts := strings.Fields(s)
	result := make([]int, len(parts))

	for i, p := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid index: %s", p)
		}
		result[i] = val
	}

	return result, nil
}
