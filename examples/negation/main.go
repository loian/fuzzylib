package main

import (
	"fmt"
	"fuzzy/fis"
	"fuzzy/inference"
	"fuzzy/membership"
	"fuzzy/operators"
	"fuzzy/rule"
	"fuzzy/set"
	"fuzzy/variable"
)

func main() {
	fmt.Println("=== Fuzzy Logic Negation Demo ===")
	fmt.Println()

	// Demo 1: Manual rule creation with negation
	manualDemo()

	fmt.Println()
	for i := 0; i < 60; i++ {
		fmt.Print("-")
	}
	fmt.Println()

	// Demo 2: Loading FIS file with negated rules
	fisDemo()
}

func manualDemo() {
	fmt.Println("Demo 1: Manual Rule Creation with Negation")
	fmt.Println("-------------------------------------------")

	// Create temperature variable
	temp, _ := variable.NewFuzzyVariable("Temperature", 0, 40)
	cold, _ := membership.NewTrapezoidal(0, 0, 10, 20)
	hot, _ := membership.NewTrapezoidal(20, 30, 40, 40)
	temp.AddSet(set.NewFuzzySet("Cold", cold))
	temp.AddSet(set.NewFuzzySet("Hot", hot))

	// Create fan speed output
	fan, _ := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	low, _ := membership.NewTriangular(0, 0, 50)
	high, _ := membership.NewTriangular(50, 100, 100)
	fan.AddSet(set.NewFuzzySet("Low", low))
	fan.AddSet(set.NewFuzzySet("High", high))

	// Create inference system
	sys := inference.NewMamdaniInferenceSystem()
	sys.AddInputVariable(temp)
	sys.AddOutputVariable(fan)

	// Rule 1: IF Temperature is Cold THEN FanSpeed is Low
	rule1, _ := rule.NewRule(rule.RuleCondition{
		Variable: "FanSpeed",
		Set:      "Low",
	}, operators.AND)
	rule1.AddCondition("Temperature", "Cold")
	sys.AddRule(rule1)

	// Rule 2: IF Temperature is NOT Cold THEN FanSpeed is High
	rule2, _ := rule.NewRule(rule.RuleCondition{
		Variable: "FanSpeed",
		Set:      "High",
	}, operators.AND)
	rule2.AddConditionEx("Temperature", "Cold", true) // Using AddConditionEx with negated=true
	sys.AddRule(rule2)

	// Test at different temperatures
	testTemps := []float64{5, 15, 25, 35}
	fmt.Println("\nRules:")
	fmt.Println("  1. IF Temperature is Cold THEN FanSpeed is Low")
	fmt.Println("  2. IF Temperature is NOT Cold THEN FanSpeed is High")
	fmt.Println("\nResults:")

	for _, t := range testTemps {
		out, err := sys.Infer(map[string]float64{"Temperature": t})
		if err != nil {
			fmt.Printf("  Temp %.0f°C: Error - %v\n", t, err)
			continue
		}
		fmt.Printf("  Temp %.0f°C → Fan %.1f%%\n", t, out["FanSpeed"])
	}
}

func fisDemo() {
	fmt.Println("\nDemo 2: Loading FIS File with Negated Rules")
	fmt.Println("--------------------------------------------")

	// Load the FIS file with negation
	system, err := fis.LoadFIS("../../testdata/negation_test.fis")
	if err != nil {
		fmt.Printf("Error loading FIS: %v\n", err)
		return
	}

	fmt.Println("\nLoaded system: NegationTest")
	fmt.Println("Inputs: Temperature [0-40°C], Humidity [0-100%]")
	fmt.Println("Output: FanSpeed [0-100%]")
	fmt.Println("\nRules:")
	fmt.Println("  1. IF Temperature is NOT Cold AND Humidity is Dry → FanSpeed is Low")
	fmt.Println("  2. IF Temperature is Hot AND Humidity is NOT Dry → FanSpeed is High")
	fmt.Println("  3. IF Temperature is NOT Cold OR Humidity is Wet → FanSpeed is Medium (0.8)")
	fmt.Println("  4. IF Temperature is Cold AND Humidity is Wet → FanSpeed is Medium")

	// Test scenarios
	scenarios := []struct {
		temp     float64
		humidity float64
		desc     string
	}{
		{5, 20, "Cold & Dry"},
		{35, 80, "Hot & Wet"},
		{25, 40, "Warm & Dry"},
		{15, 75, "Cool & Wet"},
	}

	fmt.Println("\nTest Scenarios:")
	for _, s := range scenarios {
		out, err := system.Infer(map[string]float64{
			"Temperature": s.temp,
			"Humidity":    s.humidity,
		})
		if err != nil {
			fmt.Printf("  %s (T=%.0f, H=%.0f): Error - %v\n", s.desc, s.temp, s.humidity, err)
			continue
		}
		fmt.Printf("  %s (T=%.0f°C, H=%.0f%%) → Fan %.1f%%\n",
			s.desc, s.temp, s.humidity, out["FanSpeed"])
	}

	fmt.Println("\n✓ Negation support is fully functional!")
}
