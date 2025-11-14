package main

import (
	"fmt"

	"fuzzy/inference"
	"fuzzy/membership"
	"fuzzy/set"
	"fuzzy/variable"
)

func main() {
	fmt.Println("Brake Control System - Three Input Example")
	fmt.Println("===========================================")
	fmt.Println()

	// Create inference system
	fis := inference.NewMamdaniInferenceSystem()
	if err := fis.SetDefuzzificationMethod(inference.DefuzzMOM); err != nil {
		panic(err)
	}
	if err := fis.SetResolution(1000); err != nil {
		panic(err)
	}

	// Input 1: Current Speed [0-120 km/h]
	speed, err := variable.NewFuzzyVariable("Speed", 0, 120)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Speed variable: %v", err))
	}
	slowSpeedMF, _ := membership.NewTrapezoidal(0, 0, 20, 40)
	slowSpeed, _ := speed.AddSet(set.NewFuzzySet("Slow", slowSpeedMF))
	moderateSpeedMF, _ := membership.NewTrapezoidal(30, 50, 70, 90)
	moderateSpeed, _ := speed.AddSet(set.NewFuzzySet("Moderate", moderateSpeedMF))
	fastSpeedMF, _ := membership.NewTrapezoidal(80, 100, 120, 120)
	fastSpeed, _ := speed.AddSet(set.NewFuzzySet("Fast", fastSpeedMF))

	// Input 2: Desired Speed Decrease [0-60 km/h]
	decel, err := variable.NewFuzzyVariable("Deceleration", 0, 60)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Deceleration variable: %v", err))
	}
	gentleDecelMF, _ := membership.NewTrapezoidal(0, 0, 10, 20)
	gentleDecel, _ := decel.AddSet(set.NewFuzzySet("Gentle", gentleDecelMF))
	moderateDecelMF, _ := membership.NewTrapezoidal(15, 25, 35, 45)
	moderateDecel, _ := decel.AddSet(set.NewFuzzySet("Moderate", moderateDecelMF))
	urgentDecelMF, _ := membership.NewTrapezoidal(40, 50, 60, 60)
	urgentDecel, _ := decel.AddSet(set.NewFuzzySet("Urgent", urgentDecelMF))

	// Input 3: Road Wetness [0-100%]
	wetness, err := variable.NewFuzzyVariable("Wetness", 0, 100)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Wetness variable: %v", err))
	}
	dryRoadMF, _ := membership.NewTrapezoidal(0, 0, 20, 40)
	dryRoad, _ := wetness.AddSet(set.NewFuzzySet("Dry", dryRoadMF))
	dampRoadMF, _ := membership.NewTrapezoidal(30, 45, 55, 70)
	dampRoad, _ := wetness.AddSet(set.NewFuzzySet("Damp", dampRoadMF))
	wetRoadMF, _ := membership.NewTrapezoidal(60, 80, 100, 100)
	wetRoad, _ := wetness.AddSet(set.NewFuzzySet("Wet", wetRoadMF))

	// Output 1: Brake Pressure [0-100%]
	brake, err := variable.NewFuzzyVariable("BrakePressure", 0, 100)
	if err != nil {
		panic(fmt.Sprintf("Failed to create BrakePressure variable: %v", err))
	}
	lightBrakeMF, _ := membership.NewTrapezoidal(0, 0, 20, 35)
	lightBrake, _ := brake.AddSet(set.NewFuzzySet("Light", lightBrakeMF))
	moderateBrakeMF, _ := membership.NewTrapezoidal(25, 40, 60, 75)
	moderateBrake, _ := brake.AddSet(set.NewFuzzySet("Moderate", moderateBrakeMF))
	hardBrakeMF, _ := membership.NewTrapezoidal(65, 85, 100, 100)
	hardBrake, _ := brake.AddSet(set.NewFuzzySet("Hard", hardBrakeMF))

	// Output 2: Braking Time [0-10 seconds]
	brakingTime, err := variable.NewFuzzyVariable("BrakingTime", 0, 10)
	if err != nil {
		panic(fmt.Sprintf("Failed to create BrakingTime variable: %v", err))
	}
	shortTimeMF, _ := membership.NewTrapezoidal(0, 0, 2, 3)
	shortTime, _ := brakingTime.AddSet(set.NewFuzzySet("Short", shortTimeMF))
	moderateTimeMF, _ := membership.NewTrapezoidal(2.5, 4, 6, 7.5)
	moderateTime, _ := brakingTime.AddSet(set.NewFuzzySet("Moderate", moderateTimeMF))
	longTimeMF, _ := membership.NewTrapezoidal(7, 8.5, 10, 10)
	longTime, _ := brakingTime.AddSet(set.NewFuzzySet("Long", longTimeMF))

	if err := fis.AddInputVariable(speed); err != nil {
		panic(err)
	}
	if err := fis.AddInputVariable(decel); err != nil {
		panic(err)
	}
	if err := fis.AddInputVariable(wetness); err != nil {
		panic(err)
	}
	if err := fis.AddOutputVariable(brake); err != nil {
		panic(err)
	}
	if err := fis.AddOutputVariable(brakingTime); err != nil {
		panic(err)
	}

	// Fuzzy Rules using type-safe API
	// Rule logic: Wet roads need LOWER brake force but LONGER time
	//             Dry roads can use HIGHER brake force for SHORTER time

	// Fast speed + Urgent decel + Dry → Hard brake, Moderate time
	rb1, _ := inference.NewRuleBuilderRef(hardBrake)
	rule1, _ := rb1.
		IfRef(fastSpeed).
		IfRef(urgentDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule1); err != nil {
		panic(err)
	}
	rb2, _ := inference.NewRuleBuilderRef(moderateTime)
	rule2, _ := rb2.
		IfRef(fastSpeed).
		IfRef(urgentDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule2); err != nil {
		panic(err)
	}

	// Fast speed + Urgent decel + Wet → Moderate brake, Long time
	rb3, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule3, _ := rb3.
		IfRef(fastSpeed).
		IfRef(urgentDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule3); err != nil {
		panic(err)
	}
	rb4, _ := inference.NewRuleBuilderRef(longTime)
	rule4, _ := rb4.
		IfRef(fastSpeed).
		IfRef(urgentDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule4); err != nil {
		panic(err)
	}

	// Fast speed + Moderate decel + Dry → Hard brake, Short time
	rb5, _ := inference.NewRuleBuilderRef(hardBrake)
	rule5, _ := rb5.
		IfRef(fastSpeed).
		IfRef(moderateDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule5); err != nil {
		panic(err)
	}
	rb6, _ := inference.NewRuleBuilderRef(shortTime)
	rule6, _ := rb6.
		IfRef(fastSpeed).
		IfRef(moderateDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule6); err != nil {
		panic(err)
	}

	// Fast speed + Moderate decel + Wet → Moderate brake, Moderate time
	rb7, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule7, _ := rb7.
		IfRef(fastSpeed).
		IfRef(moderateDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule7); err != nil {
		panic(err)
	}
	rb8, _ := inference.NewRuleBuilderRef(moderateTime)
	rule8, _ := rb8.
		IfRef(fastSpeed).
		IfRef(moderateDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule8); err != nil {
		panic(err)
	}

	// Fast speed + Gentle decel + Dry → Moderate brake, Short time
	rb9, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule9, _ := rb9.
		IfRef(fastSpeed).
		IfRef(gentleDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule9); err != nil {
		panic(err)
	}
	rb10, _ := inference.NewRuleBuilderRef(shortTime)
	rule10, _ := rb10.
		IfRef(fastSpeed).
		IfRef(gentleDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule10); err != nil {
		panic(err)
	}

	// Fast speed + Gentle decel + Wet → Light brake, Moderate time
	rb11, _ := inference.NewRuleBuilderRef(lightBrake)
	rule11, _ := rb11.
		IfRef(fastSpeed).
		IfRef(gentleDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule11); err != nil {
		panic(err)
	}
	rb12, _ := inference.NewRuleBuilderRef(moderateTime)
	rule12, _ := rb12.
		IfRef(fastSpeed).
		IfRef(gentleDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule12); err != nil {
		panic(err)
	}

	// Moderate speed + Urgent decel + Dry → Hard brake, Short time
	rb13, _ := inference.NewRuleBuilderRef(hardBrake)
	rule13, _ := rb13.
		IfRef(moderateSpeed).
		IfRef(urgentDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule13); err != nil {
		panic(err)
	}
	rb14, _ := inference.NewRuleBuilderRef(shortTime)
	rule14, _ := rb14.
		IfRef(moderateSpeed).
		IfRef(urgentDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule14); err != nil {
		panic(err)
	}

	// Moderate speed + Urgent decel + Wet → Moderate brake, Moderate time
	rb15, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule15, _ := rb15.
		IfRef(moderateSpeed).
		IfRef(urgentDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule15); err != nil {
		panic(err)
	}
	rb16, _ := inference.NewRuleBuilderRef(moderateTime)
	rule16, _ := rb16.
		IfRef(moderateSpeed).
		IfRef(urgentDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule16); err != nil {
		panic(err)
	}

	// Moderate speed + Moderate decel + Dry → Moderate brake, Short time
	rb17, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule17, _ := rb17.
		IfRef(moderateSpeed).
		IfRef(moderateDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule17); err != nil {
		panic(err)
	}
	rb18, _ := inference.NewRuleBuilderRef(shortTime)
	rule18, _ := rb18.
		IfRef(moderateSpeed).
		IfRef(moderateDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule18); err != nil {
		panic(err)
	}

	// Moderate speed + Moderate decel + Wet → Light brake, Long time
	rb19, _ := inference.NewRuleBuilderRef(lightBrake)
	rule19, _ := rb19.
		IfRef(moderateSpeed).
		IfRef(moderateDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule19); err != nil {
		panic(err)
	}
	rb20, _ := inference.NewRuleBuilderRef(longTime)
	rule20, _ := rb20.
		IfRef(moderateSpeed).
		IfRef(moderateDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule20); err != nil {
		panic(err)
	}

	// Moderate speed + Gentle decel + Dry → Light brake, Short time
	rb21, _ := inference.NewRuleBuilderRef(lightBrake)
	rule21, _ := rb21.
		IfRef(moderateSpeed).
		IfRef(gentleDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule21); err != nil {
		panic(err)
	}
	rb22, _ := inference.NewRuleBuilderRef(shortTime)
	rule22, _ := rb22.
		IfRef(moderateSpeed).
		IfRef(gentleDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule22); err != nil {
		panic(err)
	}

	// Moderate speed + Gentle decel + Wet → Light brake, Moderate time
	rb23, _ := inference.NewRuleBuilderRef(lightBrake)
	rule23, _ := rb23.
		IfRef(moderateSpeed).
		IfRef(gentleDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule23); err != nil {
		panic(err)
	}
	rb24, _ := inference.NewRuleBuilderRef(moderateTime)
	rule24, _ := rb24.
		IfRef(moderateSpeed).
		IfRef(gentleDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule24); err != nil {
		panic(err)
	}

	// Slow speed + Urgent decel + Dry → Moderate brake, Short time
	rb25, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule25, _ := rb25.
		IfRef(slowSpeed).
		IfRef(urgentDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule25); err != nil {
		panic(err)
	}
	rb26, _ := inference.NewRuleBuilderRef(shortTime)
	rule26, _ := rb26.
		IfRef(slowSpeed).
		IfRef(urgentDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule26); err != nil {
		panic(err)
	}

	// Slow speed + Urgent decel + Wet → Light brake, Moderate time
	rb27, _ := inference.NewRuleBuilderRef(lightBrake)
	rule27, _ := rb27.
		IfRef(slowSpeed).
		IfRef(urgentDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule27); err != nil {
		panic(err)
	}
	rb28, _ := inference.NewRuleBuilderRef(moderateTime)
	rule28, _ := rb28.
		IfRef(slowSpeed).
		IfRef(urgentDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule28); err != nil {
		panic(err)
	}

	// Slow speed + Gentle/Moderate decel + Dry → Light brake, Short time
	rb29, _ := inference.NewRuleBuilderRef(lightBrake)
	rule29, _ := rb29.
		IfRef(slowSpeed).
		IfRef(gentleDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule29); err != nil {
		panic(err)
	}
	rb30, _ := inference.NewRuleBuilderRef(shortTime)
	rule30, _ := rb30.
		IfRef(slowSpeed).
		IfRef(gentleDecel).
		IfRef(dryRoad).
		Build()
	if err := fis.AddRule(rule30); err != nil {
		panic(err)
	}

	// Slow speed + Gentle/Moderate decel + Wet → Light brake, Moderate time
	rb31, _ := inference.NewRuleBuilderRef(lightBrake)
	rule31, _ := rb31.
		IfRef(slowSpeed).
		IfRef(gentleDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule31); err != nil {
		panic(err)
	}
	rb32, _ := inference.NewRuleBuilderRef(moderateTime)
	rule32, _ := rb32.
		IfRef(slowSpeed).
		IfRef(gentleDecel).
		IfRef(wetRoad).
		Build()
	if err := fis.AddRule(rule32); err != nil {
		panic(err)
	}

	// Damp road - intermediate cases
	rb33, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule33, _ := rb33.
		IfRef(fastSpeed).
		IfRef(urgentDecel).
		IfRef(dampRoad).
		Build()
	if err := fis.AddRule(rule33); err != nil {
		panic(err)
	}
	rb34, _ := inference.NewRuleBuilderRef(moderateTime)
	rule34, _ := rb34.
		IfRef(fastSpeed).
		IfRef(urgentDecel).
		IfRef(dampRoad).
		Build()
	if err := fis.AddRule(rule34); err != nil {
		panic(err)
	}

	rb35, _ := inference.NewRuleBuilderRef(moderateBrake)
	rule35, _ := rb35.
		IfRef(moderateSpeed).
		IfRef(moderateDecel).
		IfRef(dampRoad).
		Build()
	if err := fis.AddRule(rule35); err != nil {
		panic(err)
	}
	rb36, _ := inference.NewRuleBuilderRef(moderateTime)
	rule36, _ := rb36.
		IfRef(moderateSpeed).
		IfRef(moderateDecel).
		IfRef(dampRoad).
		Build()
	if err := fis.AddRule(rule36); err != nil {
		panic(err)
	}

	// Test scenarios
	fmt.Println("Brake Control Scenarios:")
	fmt.Println("------------------------")

	scenarios := []struct {
		speed   float64
		decel   float64
		wetness float64
		desc    string
	}{
		{30, 5, 10, "Slow, gentle decel, dry"},
		{30, 50, 10, "Slow, urgent decel, dry"},
		{30, 50, 85, "Slow, urgent decel, wet"},
		{60, 8, 15, "Moderate, gentle decel, dry"},
		{60, 30, 15, "Moderate, moderate decel, dry"},
		{60, 55, 15, "Moderate, urgent decel, dry"},
		{60, 30, 85, "Moderate, moderate decel, wet"},
		{100, 30, 15, "Fast, moderate decel, dry"},
		{100, 55, 15, "Fast, urgent decel, dry"},
		{100, 10, 85, "Fast, gentle decel, wet"},
		{50, 25, 50, "Mixed conditions"},
	}

	for _, s := range scenarios {
		inputs := map[string]float64{
			"Speed":        s.speed,
			"Deceleration": s.decel,
			"Wetness":      s.wetness,
		}

		outputs, err := fis.Infer(inputs)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("%-32s | Speed: %3.0f | Decel: %2.0f | Wet: %2.0f%% → Brake: %5.1f%% | Time: %4.1fs\n",
			s.desc, s.speed, s.decel, s.wetness, outputs["BrakePressure"], outputs["BrakingTime"])
	}

	fmt.Println()
	fmt.Println("Key Insights:")
	fmt.Println("• Dry roads: Higher brake force, shorter braking time")
	fmt.Println("• Wet roads: Lower brake force, longer braking time (safety)")
	fmt.Println("• Higher deceleration needs: More force and/or more time")
	fmt.Println("• The system balances TWO outputs to achieve safe braking")
	fmt.Println("• Type-safe SetRef with multiple outputs demonstrates flexibility")
}
