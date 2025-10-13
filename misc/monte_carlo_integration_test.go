package main

import (
	"fmt"
	"testing"
)

// Test function for Monte Carlo integration
func TestMonteCarloIntegrate(t *testing.T) {
	tests := []struct {
		a, b float64
		n    int
		want float64
	}{
		{0, 1, 1000000, 1.0 / 3.0}, // Integral of x^2 from 0 to 1 is 1/3
		{0, 2, 1000000, 8.0 / 3.0}, // Integral of x^2 from 0 to 2 is 8/3
		{-1, 1, 1000000, 2.0 / 3.0}, // Integral of x^2 from -1 to 1 is 2/3
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Integrating from %.2f to %.2f with %d points", tt.a, tt.b, tt.n), func(t *testing.T) {
			got := monteCarloIntegrate(tt.a, tt.b, tt.n)
			// Monte Carlo is probabilistic: allow a 10% error tolerance for large n
			if got < tt.want*0.9 || got > tt.want*1.1 {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}