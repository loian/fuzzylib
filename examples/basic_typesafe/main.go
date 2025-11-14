package main

import (
	"fmt"

	"fuzzy/inference"
	"fuzzy/membership"
	"fuzzy/set"
	"fuzzy/variable"
)

func main() {
	fmt.Println("Type-Safe Fuzzy Logic Example")
	fmt.Println("==============================")
	fmt.Println()

	// Create inference system
	fis := inference.NewMamdaniInferenceSystem()
	if err := fis.SetDefuzzificationMethod(inference.DefuzzMOM); err != nil {
		panic(err)
	}
	if err := fis.SetResolution(500); err != nil {
		panic(err)
	}

	// Input variable: Temperature [0-50°C]
	temp, err := variable.NewFuzzyVariable("Temperature", 0, 50)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Temperature variable: %v", err))
	}

	// AddSet now returns type-safe references
	coldMF, err := membership.NewTrapezoidal(0, 0, 12, 20)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Cold MF: %v", err))
	}
	coldSet, err := temp.AddSet(set.NewFuzzySet("Cold", coldMF))
	if err != nil {
		panic(fmt.Sprintf("Failed to add Cold set: %v", err))
	}

	coolMF, err := membership.NewTrapezoidal(8, 15, 22, 28)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Cool MF: %v", err))
	}
	coolSet, err := temp.AddSet(set.NewFuzzySet("Cool", coolMF))
	if err != nil {
		panic(fmt.Sprintf("Failed to add Cool set: %v", err))
	}

	mildMF, err := membership.NewTrapezoidal(15, 20, 26, 32)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Mild MF: %v", err))
	}
	mildSet, err := temp.AddSet(set.NewFuzzySet("Mild", mildMF))
	if err != nil {
		panic(fmt.Sprintf("Failed to add Mild set: %v", err))
	}

	hotMF, err := membership.NewTrapezoidal(28, 35, 50, 51)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Hot MF: %v", err))
	}
	hotSet, err := temp.AddSet(set.NewFuzzySet("Hot", hotMF))
	if err != nil {
		panic(fmt.Sprintf("Failed to add Hot set: %v", err))
	}

	// Output variable: FanSpeed [0-100%]
	fan, err := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	if err != nil {
		panic(fmt.Sprintf("Failed to create FanSpeed variable: %v", err))
	}

	lowFanMF, err := membership.NewTrapezoidal(0, 0, 20, 40)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Low Fan MF: %v", err))
	}
	lowFan, err := fan.AddSet(set.NewFuzzySet("Low", lowFanMF))
	if err != nil {
		panic(fmt.Sprintf("Failed to add Low fan set: %v", err))
	}

	mediumFanMF, err := membership.NewTrapezoidal(30, 45, 55, 70)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Medium Fan MF: %v", err))
	}
	mediumFan, err := fan.AddSet(set.NewFuzzySet("Medium", mediumFanMF))
	if err != nil {
		panic(fmt.Sprintf("Failed to add Medium fan set: %v", err))
	}

	highFanMF, err := membership.NewTrapezoidal(70, 90, 100, 101)
	if err != nil {
		panic(fmt.Sprintf("Failed to create High Fan MF: %v", err))
	}
	highFan, err := fan.AddSet(set.NewFuzzySet("High", highFanMF))
	if err != nil {
		panic(fmt.Sprintf("Failed to add High fan set: %v", err))
	}

	if err := fis.AddInputVariable(temp); err != nil {
		panic(fmt.Sprintf("Failed to add input variable: %v", err))
	}
	if err := fis.AddOutputVariable(fan); err != nil {
		panic(fmt.Sprintf("Failed to add output variable: %v", err))
	}

	// Build rules using type-safe references
	// IDE autocomplete works! Typos become compile errors!
	rb1, _ := inference.NewRuleBuilderRef(lowFan)
	rule1, _ := rb1.IfRef(coldSet).Build()
	if err := fis.AddRule(rule1); err != nil {
		panic(fmt.Sprintf("Failed to add rule 1: %v", err))
	}
	rb2, _ := inference.NewRuleBuilderRef(lowFan)
	rule2, _ := rb2.IfRef(coolSet).Build()
	if err := fis.AddRule(rule2); err != nil {
		panic(fmt.Sprintf("Failed to add rule 2: %v", err))
	}
	rb3, _ := inference.NewRuleBuilderRef(mediumFan)
	rule3, _ := rb3.IfRef(mildSet).Build()
	if err := fis.AddRule(rule3); err != nil {
		panic(fmt.Sprintf("Failed to add rule 3: %v", err))
	}
	rb4, _ := inference.NewRuleBuilderRef(highFan)
	rule4, _ := rb4.IfRef(hotSet).Build()
	if err := fis.AddRule(rule4); err != nil {
		panic(fmt.Sprintf("Failed to add rule 4: %v", err))
	}

	// Compare: old string-based API still works
	// fis.AddRule(inference.NewRuleBuilder("FanSpeed", "High").If("Temperature", "Hot").Build())
	// But typos won't be caught: If("Temperature", "Hott") compiles but fails at runtime

	// Run inference
	fmt.Println("Testing various temperatures:")
	temps := []float64{5, 15, 25, 35, 45}

	for _, t := range temps {
		outputs, err := fis.Infer(map[string]float64{"Temperature": t})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Temperature: %5.1f°C → Fan Speed: %5.1f%%\n", t, outputs["FanSpeed"])
	}

	fmt.Println("\nBenefits of type-safe API:")
	fmt.Println("✓ Compile-time checking - typos become errors")
	fmt.Println("✓ IDE autocomplete - discover available sets")
	fmt.Println("✓ Refactoring support - rename propagates correctly")
	fmt.Println("✓ Self-documenting - references show relationships")
}
