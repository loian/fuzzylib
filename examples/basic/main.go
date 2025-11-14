package main

import (
	"fmt"

	"github.com/loian/fuzzylib/inference"
	"github.com/loian/fuzzylib/membership"
	"github.com/loian/fuzzylib/set"
	"github.com/loian/fuzzylib/variable"
)

func main() {
	// Input variable: Temperature [0..50] (Celsius)
	temp, err := variable.NewFuzzyVariable("Temperature", 0, 50)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Temperature variable: %v", err))
	}

	// Four fuzzy sets across the 0..50°C range using trapezoids for realistic plateaus
	// Cold: plateau from 0..12, slopes to 0 at 20
	coldMF, err := membership.NewTrapezoidal(0, 0, 12, 20)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Cold MF: %v", err))
	}
	if _, err := temp.AddSet(set.NewFuzzySet("Cold", coldMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Cold set: %v", err))
	}

	// Cool: plateau from 15..22, slopes to 0 from 8 to 28
	coolMF, err := membership.NewTrapezoidal(8, 15, 22, 28)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Cool MF: %v", err))
	}
	if _, err := temp.AddSet(set.NewFuzzySet("Cool", coolMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Cool set: %v", err))
	}

	// Mild: plateau from 20..26, slopes to 0 from 15 to 32
	mildMF, err := membership.NewTrapezoidal(15, 20, 26, 32)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Mild MF: %v", err))
	}
	if _, err := temp.AddSet(set.NewFuzzySet("Mild", mildMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Mild set: %v", err))
	}

	// Hot: plateau from 35..50, slopes to 0 from 28 onwards
	hotMF, err := membership.NewTrapezoidal(28, 35, 50, 51)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Hot MF: %v", err))
	}
	if _, err := temp.AddSet(set.NewFuzzySet("Hot", hotMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Hot set: %v", err))
	}

	// Output variable: FanSpeed [0..100]
	fan, err := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	if err != nil {
		panic(fmt.Sprintf("Failed to create FanSpeed variable: %v", err))
	}

	// Use three output sets for smoother gradation
	lowMF, err := membership.NewTrapezoidal(0, 0, 20, 40)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Low MF: %v", err))
	}
	if _, err := fan.AddSet(set.NewFuzzySet("Low", lowMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Low set: %v", err))
	}

	mediumMF, err := membership.NewTrapezoidal(30, 45, 55, 70)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Medium MF: %v", err))
	}
	if _, err := fan.AddSet(set.NewFuzzySet("Medium", mediumMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Medium set: %v", err))
	}

	highMF, err := membership.NewTrapezoidal(70, 90, 100, 101)
	if err != nil {
		panic(fmt.Sprintf("Failed to create High MF: %v", err))
	}
	if _, err := fan.AddSet(set.NewFuzzySet("High", highMF)); err != nil {
		panic(fmt.Sprintf("Failed to add High set: %v", err))
	}

	// Create inference system with MOM defuzzification
	fis := inference.NewMamdaniInferenceSystem()
	if err := fis.SetDefuzzificationMethod(inference.DefuzzMOM); err != nil {
		panic(err)
	}
	if err := fis.SetResolution(500); err != nil {
		panic(err)
	}

	if err := fis.AddInputVariable(temp); err != nil {
		panic(fmt.Sprintf("Failed to add input variable: %v", err))
	}
	if err := fis.AddOutputVariable(fan); err != nil {
		panic(fmt.Sprintf("Failed to add output variable: %v", err))
	}

	// Rules

	r1Builder, _ := inference.NewRuleBuilder("FanSpeed", "High")
	r1, _ := r1Builder.If("Temperature", "Hot").Build()
	r2Builder, _ := inference.NewRuleBuilder("FanSpeed", "Low")
	r2, _ := r2Builder.If("Temperature", "Cold").Build()
	r3Builder, _ := inference.NewRuleBuilder("FanSpeed", "Medium")
	r3, _ := r3Builder.If("Temperature", "Mild").Build()
	r4Builder, _ := inference.NewRuleBuilder("FanSpeed", "Low")
	r4, _ := r4Builder.If("Temperature", "Cool").Build()

	if err := fis.AddRule(r1); err != nil {
		panic(fmt.Sprintf("Failed to add rule 1: %v", err))
	}
	if err := fis.AddRule(r2); err != nil {
		panic(fmt.Sprintf("Failed to add rule 2: %v", err))
	}
	if err := fis.AddRule(r3); err != nil {
		panic(fmt.Sprintf("Failed to add rule 3: %v", err))
	}
	if err := fis.AddRule(r4); err != nil {
		panic(fmt.Sprintf("Failed to add rule 4: %v", err))
	}

	// Example inputs across 0..50°C
	cases := []float64{0, 5, 12, 17, 22, 28, 35, 42, 50}

	for _, v := range cases {
		inputs := map[string]float64{"Temperature": v}

		// Debug: show fuzzified memberships for the input
		mem := temp.Fuzzify(v)
		fmt.Printf("\nInput: %v\n", inputs)
		fmt.Println("  Temperature memberships:")
		for name, deg := range mem {
			fmt.Printf("    %s: %.4f\n", name, deg)
		}

		// Debug: show each rule's firing strength
		fmt.Println("  Rule firing strengths:")
		for idx, r := range fis.Rules {
			strength, _ := r.Evaluate(map[string]map[string]float64{"Temperature": mem})
			fmt.Printf("    rule %d -> output %s:%s = %.4f\n", idx+1, r.Output.Variable, r.Output.Set, strength)
		}

		outputs, _ := fis.Infer(inputs)
		fmt.Printf("  Defuzzified Output: %v\n", outputs)
	}
}
