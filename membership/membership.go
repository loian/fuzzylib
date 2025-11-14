package membership

import (
	"fmt"
	"math"
)

// MembershipFunction is the interface all membership functions must implement
type MembershipFunction interface {
	Evaluate(x float64) float64 // Returns degree of membership [0, 1]
}

// Triangular membership function: a (left foot), b (peak), c (right foot)
type Triangular struct {
	A float64
	B float64
	C float64
}

// NewTriangular creates a new triangular membership function.
// Parameters must satisfy: a <= b <= c
// Returns error if parameters are invalid.
func NewTriangular(a, b, c float64) (*Triangular, error) {
	if a > b || b > c {
		return nil, fmt.Errorf("triangular parameters must satisfy a <= b <= c, got a=%.2f, b=%.2f, c=%.2f", a, b, c)
	}
	return &Triangular{A: a, B: b, C: c}, nil
}

// Evaluate returns the membership degree for value x
func (t *Triangular) Evaluate(x float64) float64 {
	// Handle degenerate cases
	if t.A == t.B && t.B == t.C {
		// All points are the same - impulse function
		if x == t.A {
			return 1.0
		}
		return 0.0
	}

	if x <= t.A || x >= t.C {
		return 0.0
	}

	if x == t.B {
		return 1.0
	}

	if x < t.B {
		// Avoid division by zero
		if t.B == t.A {
			return 1.0
		}
		return (x - t.A) / (t.B - t.A)
	}

	// x > t.B
	// Avoid division by zero
	if t.C == t.B {
		return 1.0
	}
	return (t.C - x) / (t.C - t.B)
}

// Trapezoidal membership function: a, b (left plateau), c, d (right plateau)
type Trapezoidal struct {
	A float64
	B float64
	C float64
	D float64
}

// NewTrapezoidal creates a new trapezoidal membership function.
// Parameters must satisfy: a <= b <= c <= d
// Returns error if parameters are invalid.
func NewTrapezoidal(a, b, c, d float64) (*Trapezoidal, error) {
	if a > b || b > c || c > d {
		return nil, fmt.Errorf("trapezoidal parameters must satisfy a <= b <= c <= d, got a=%.2f, b=%.2f, c=%.2f, d=%.2f", a, b, c, d)
	}
	return &Trapezoidal{A: a, B: b, C: c, D: d}, nil
}

// Evaluate returns the membership degree for value x
func (t *Trapezoidal) Evaluate(x float64) float64 {
	// Handle degenerate cases
	if t.A == t.B && t.B == t.C && t.C == t.D {
		// All points are the same - impulse function
		if x == t.A {
			return 1.0
		}
		return 0.0
	}

	if x <= t.A || x >= t.D {
		return 0.0
	}

	if x >= t.B && x <= t.C {
		return 1.0
	}

	if x < t.B {
		// Avoid division by zero
		if t.B == t.A {
			return 1.0
		}
		return (x - t.A) / (t.B - t.A)
	}

	// x > t.C
	// Avoid division by zero
	if t.D == t.C {
		return 1.0
	}
	return (t.D - x) / (t.D - t.C)
}

// Gaussian membership function: center (μ) and width (σ)
type Gaussian struct {
	Center float64 // μ
	Width  float64 // σ
}

// NewGaussian creates a new Gaussian membership function.
// Width must be > 0.
// Returns error if width is invalid.
func NewGaussian(center, width float64) (*Gaussian, error) {
	if width <= 0 {
		return nil, fmt.Errorf("gaussian width must be > 0, got %.2f", width)
	}
	return &Gaussian{Center: center, Width: width}, nil
}

// Evaluate returns the membership degree for value x
func (g *Gaussian) Evaluate(x float64) float64 {
	exponent := -((x - g.Center) * (x - g.Center)) / (2 * g.Width * g.Width)
	return math.Exp(exponent)
}
