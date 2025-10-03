package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Function to integrate
func f(x float64) float64 {
	return x * x // Example: f(x) = x^2
}

// Monte Carlo integration function
func monteCarloIntegrate(a, b float64, n int) float64 {
	rand.Seed(time.Now().UnixNano())
	sum := 0.0

	for i := 0; i < n; i++ {
		x := a + (b-a)*rand.Float64() // Generate random x in [a, b]
		sum += f(x)                   // Evaluate f at x
	}

	average := sum / float64(n)     // Calculate average
	return average * (b - a)         // Scale by the width of the interval
}

// Main function to run the integration
func main() {
	a := 0.0  // Lower bound
	b := 1.0  // Upper bound
	n := 1000000 // Number of random points

	result := monteCarloIntegrate(a, b, n)
	fmt.Printf("Estimated integral of f(x) from %.2f to %.2f is %.5f\n", a, b, result)
}