package main

import (
	"fmt"
	"fuzzy/fis"
)

func main() {
	// Load FIS file
	system, err := fis.LoadFIS("testdata/temp_control.fis")
	if err != nil {
		fmt.Printf("Error loading FIS: %v\n", err)
		return
	}

	fmt.Println("Temperature Control System - Loaded from FIS file")
	fmt.Println("=================================================")

	// Test various temperatures
	temps := []float64{5, 15, 25, 35, 45}
	for _, temp := range temps {
		outputs, err := system.Infer(map[string]float64{"Temperature": temp})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Temperature: %.0f°C → Fan Speed: %.1f%%\n", temp, outputs["FanSpeed"])
	}
}
