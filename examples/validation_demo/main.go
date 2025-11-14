package main

import (
	"fmt"
	"github.com/loian/fuzzylib/inference"
	"github.com/loian/fuzzylib/membership"
	"github.com/loian/fuzzylib/set"
	"github.com/loian/fuzzylib/variable"
)

func main() {
	fmt.Println("Validation Demo")
	fmt.Println("===============")

	// Create a simple temperature control system
	temp, err := variable.NewFuzzyVariable("Temperature", 0, 50)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Temperature variable: %v", err))
	}

	coldMF, err := membership.NewTriangular(0, 0, 25)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Cold membership function: %v", err))
	}
	if _, err := temp.AddSet(set.NewFuzzySet("Cold", coldMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Cold set: %v", err))
	}

	hotMF, err := membership.NewTriangular(25, 50, 50)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Hot membership function: %v", err))
	}
	if _, err := temp.AddSet(set.NewFuzzySet("Hot", hotMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Hot set: %v", err))
	}

	fan, err := variable.NewFuzzyVariable("FanSpeed", 0, 100)
	if err != nil {
		panic(fmt.Sprintf("Failed to create FanSpeed variable: %v", err))
	}

	lowMF, err := membership.NewTriangular(0, 0, 50)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Low membership function: %v", err))
	}
	if _, err := fan.AddSet(set.NewFuzzySet("Low", lowMF)); err != nil {
		panic(fmt.Sprintf("Failed to add Low set: %v", err))
	}

	highMF, err := membership.NewTriangular(50, 100, 100)
	if err != nil {
		panic(fmt.Sprintf("Failed to create High membership function: %v", err))
	}
	if _, err := fan.AddSet(set.NewFuzzySet("High", highMF)); err != nil {
		panic(fmt.Sprintf("Failed to add High set: %v", err))
	}

	fis := inference.NewMamdaniInferenceSystem()
	if err := fis.AddInputVariable(temp); err != nil {
		panic(fmt.Sprintf("Failed to add input variable: %v", err))
	}
	if err := fis.AddOutputVariable(fan); err != nil {
		panic(fmt.Sprintf("Failed to add output variable: %v", err))
	}
	rule1Builder, _ := inference.NewRuleBuilder("FanSpeed", "Low")
	rule1, _ := rule1Builder.If("Temperature", "Cold").Build()
	if err := fis.AddRule(rule1); err != nil {
		panic(fmt.Sprintf("Failed to add rule: %v", err))
	}
	rule2Builder, _ := inference.NewRuleBuilder("FanSpeed", "High")
	rule2, _ := rule2Builder.If("Temperature", "Hot").Build()
	if err := fis.AddRule(rule2); err != nil {
		panic(fmt.Sprintf("Failed to add rule: %v", err))
	}

	// Test 1: Valid input
	fmt.Println("Test 1: Valid input (Temperature=25)")
	result, err := fis.Infer(map[string]float64{"Temperature": 25})
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  ✓ Success: FanSpeed = %.1f%%\n", result["FanSpeed"])
	}
	fmt.Println()

	// Test 2: Missing required input
	fmt.Println("Test 2: Missing required input")
	result, err = fis.Infer(map[string]float64{})
	if err != nil {
		fmt.Printf("  ✓ Correctly caught error: %v\n", err)
	} else {
		fmt.Printf("  ❌ Should have failed, got: %v\n", result)
	}
	fmt.Println()

	// Test 3: Input below minimum bound
	fmt.Println("Test 3: Input below minimum bound (Temperature=-10)")
	result, err = fis.Infer(map[string]float64{"Temperature": -10})
	if err != nil {
		fmt.Printf("  ✓ Correctly caught error: %v\n", err)
	} else {
		fmt.Printf("  ❌ Should have failed, got: %v\n", result)
	}
	fmt.Println()

	// Test 4: Input above maximum bound
	fmt.Println("Test 4: Input above maximum bound (Temperature=100)")
	result, err = fis.Infer(map[string]float64{"Temperature": 100})
	if err != nil {
		fmt.Printf("  ✓ Correctly caught error: %v\n", err)
	} else {
		fmt.Printf("  ❌ Should have failed, got: %v\n", result)
	}
	fmt.Println()

	// Test 5: Boundary values (should be valid)
	fmt.Println("Test 5: Boundary values")
	result, err = fis.Infer(map[string]float64{"Temperature": 0})
	if err != nil {
		fmt.Printf("  ❌ Minimum boundary failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Minimum boundary (0): FanSpeed = %.1f%%\n", result["FanSpeed"])
	}

	result, err = fis.Infer(map[string]float64{"Temperature": 50})
	if err != nil {
		fmt.Printf("  ❌ Maximum boundary failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Maximum boundary (50): FanSpeed = %.1f%%\n", result["FanSpeed"])
	}
	fmt.Println()

	// Test 6: Resolution validation
	fmt.Println("Test 6: Resolution validation")
	err = fis.SetResolution(500)
	if err != nil {
		fmt.Printf("  ❌ Unexpected error for valid resolution: %v\n", err)
	} else {
		fmt.Printf("  ✓ Set resolution to 500: actual = %d\n", fis.Resolution)
	}

	err = fis.SetResolution(0)
	if err != nil {
		fmt.Printf("  ✓ Set resolution to 0: correctly rejected with error: %v\n", err)
	} else {
		fmt.Printf("  ❌ Should have rejected resolution 0\n")
	}

	err = fis.SetResolution(-100)
	if err != nil {
		fmt.Printf("  ✓ Set resolution to -100: correctly rejected with error: %v\n", err)
	} else {
		fmt.Printf("  ❌ Should have rejected negative resolution\n")
	}
	fmt.Println()

	// Test 7: Defuzzification method validation
	fmt.Println("Test 7: Defuzzification method validation")
	err = fis.SetDefuzzificationMethod(inference.DefuzzMOM)
	if err != nil {
		fmt.Printf("  ❌ Unexpected error for valid method: %v\n", err)
	} else {
		fmt.Printf("  ✓ Set valid method 'mom'\n")
	}

	err = fis.SetDefuzzificationMethod("invalid")
	if err != nil {
		fmt.Printf("  ✓ Invalid method correctly rejected: %v\n", err)
	} else {
		fmt.Printf("  ❌ Should have rejected invalid method\n")
	}
	fmt.Println()

	fmt.Println("All validation tests completed!")
}
