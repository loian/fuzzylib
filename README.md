# fuzzy

The very first public release of **fuzzy**, a lightweight Mamdani-style fuzzy inference toolkit for Go. It ships with everything you need to build a small FIS, evaluate it, and experiment with control logic without pulling in heavy external dependencies.

## Highlights

- **Pure Go core**: no CGO, no external binaries.
- **Classic membership functions**: triangular, trapezoidal, and Gaussian with parameter validation.
- **Rule engine**: weighted IF/THEN rules, fluent builder helpers, AND/OR operators, and safe evaluation.
- **Defuzzification trio**: Center of Gravity (COG), Mean of Maximum (MOM), and First/Last/Smallest of Maximum via a shared sampler.
- **`.fis` importer**: load basic MATLAB/scikit-fuzzy-compatible FIS files for quick prototyping.

## Requirements

- Go 1.21+ (the module currently targets Go `1.25.4`).

## Installation

```bash
go get github.com/lorenzo-iannone/fuzzy
```

Alternatively, clone the repository if you prefer local development:

```bash
git clone https://github.com/lorenzo-iannone/fuzzy.git
cd fuzzy
```

## Quick Example

```go
package main

import (
    "fmt"
    "fuzzy/inference"
    "fuzzy/membership"
    "fuzzy/set"
    "fuzzy/variable"
)

func main() {
    temp, _ := variable.NewFuzzyVariable("Temperature", 0, 40)
    cool, _ := membership.NewTrapezoidal(0, 0, 15, 22)
    warm, _ := membership.NewTrapezoidal(18, 24, 40, 40)
    temp.AddSet(set.NewFuzzySet("Cool", cool))
    temp.AddSet(set.NewFuzzySet("Warm", warm))

    fan, _ := variable.NewFuzzyVariable("Fan", 0, 100)
    low, _ := membership.NewTriangular(0, 0, 50)
    high, _ := membership.NewTriangular(50, 100, 100)
    fan.AddSet(set.NewFuzzySet("Low", low))
    fan.AddSet(set.NewFuzzySet("High", high))

    fis := inference.NewMamdaniInferenceSystem()
    fis.AddInputVariable(temp)
    fis.AddOutputVariable(fan)

    ruleCool, _ := inference.NewRuleBuilder("Fan", "Low").If("Temperature", "Cool").Build()
    ruleWarm, _ := inference.NewRuleBuilder("Fan", "High").If("Temperature", "Warm").Build()
    fis.AddRule(ruleCool)
    fis.AddRule(ruleWarm)

    out, err := fis.Infer(map[string]float64{"Temperature": 26})
    if err != nil {
        panic(err)
    }
    fmt.Printf("Fan output: %.1f%%\n", out["Fan"])
}
```

Run it straight from the repo:

```bash
go run examples/basic/main.go
```

## Testing

```bash
go test ./...
```

## Project Layout

- `membership/` – Gaussian, triangular, and trapezoidal membership functions.
- `set/` – `FuzzySet` wrapper around membership functions.
- `variable/` – Linguistic variables, fuzzification helpers, and typed references.
- `operators/` – Zadeh AND/OR/NOT implementations with input validation.
- `rule/` – Rule definition plus the fluent builder API.
- `inference/` – Mamdani inference engine and defuzzification routines.
- `fis/` – `.fis` parser + converter to the runtime engine.
- `examples/` – Small runnable demos (`basic`, `basic_typesafe`, `brake_control`, `fis`, `validation_demo`).
- `testdata/` – Supporting files used by the importer tests.

## Roadmap

Early priorities include richer `.fis` coverage (negation support, Sugeno files), additional membership shapes, and higher-level diagnostics for debugging rule bases.
3, 2 (1.0) : 1  # IF Temperature is Mild THEN FanSpeed is Medium
4, 3 (1.0) : 1  # IF Temperature is Hot THEN FanSpeed is High
```

## API Reference

### Defuzzification Methods

```go
fis := inference.NewMamdaniInferenceSystem()
fis.SetDefuzzificationMethod(inference.DefuzzCOG)  // Center of Gravity (default)
fis.SetDefuzzificationMethod(inference.DefuzzMOM)  // Mean of Maximum
fis.SetDefuzzificationMethod(inference.DefuzzFOM)  // First of Maximum
```

### Resolution Tuning

```go
if err := fis.SetResolution(2000); err != nil {  // Default: 1000
    panic(err)
}
```

Higher resolution improves numeric accuracy but increases CPU cost. Typical range: `500-2000`.

## Error Handling

The library follows Go best practices with explicit error handling. Key operations that return errors:

- **Variable creation**: `NewFuzzyVariable(name, min, max)` validates that min < max
- **Membership functions**: `NewTriangular(a, b, c)`, `NewTrapezoidal(a, b, c, d)`, `NewGaussian(center, width)` validate parameters
- **System configuration**: `AddSet()`, `AddInputVariable()`, `AddOutputVariable()`, `AddRule()` check for duplicates and invalid references
- **Inference**: `Infer()` validates inputs are within bounds and returns errors if no rules fire
- **Defuzzification**: Returns errors when no rules fire instead of silently using default values

Example with comprehensive error handling:

```go
// Create variable with validation
temp, err := variable.NewFuzzyVariable("Temperature", 0, 50)
if err != nil {
    return fmt.Errorf("failed to create variable: %w", err)
}

// Create membership function with validation
mf, err := membership.NewTriangular(0, 25, 50)
if err != nil {
    return fmt.Errorf("invalid membership function: %w", err)
}

// Add set with duplicate checking
if _, err := temp.AddSet(set.NewFuzzySet("Medium", mf)); err != nil {
    return fmt.Errorf("failed to add set: %w", err)
}

// Infer with bounds checking and rule validation
outputs, err := fis.Infer(map[string]float64{"Temperature": 25})
if err != nil {
    return fmt.Errorf("inference failed: %w", err)
}
```

See `MIGRATION.md` for a complete guide on updating from version 1.x to 2.x.

### FIS File Format

Supported membership functions in `.fis` files:
- `trimf` → Triangular
- `trapmf` → Trapezoidal  
- `gaussmf` → Gaussian

Rule format: `input1_idx input2_idx, output_idx (weight) : connection`
- Connection: `1` = AND, `2` = OR
- Weight: `0.0-1.0` (rule strength multiplier)

