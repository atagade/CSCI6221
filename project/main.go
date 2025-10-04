package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"cda-simulator/agent"
	"cda-simulator/simulation"
)

var (
	numRandom = flag.Int("random", 50, "number of random agents")
	numMM     = flag.Int("mm", 10, "number of market maker agents")
	numTrend  = flag.Int("trend", 50, "number of trend follower agents")
	dur       = flag.Duration("dur", 30*time.Second, "simulation duration")
)

func main() {
	flag.Parse()

	sim := simulation.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < *numRandom; i++ {
		id := fmt.Sprintf("rand%d", i)
		a := agent.NewRandom(id, 100000, 100)
		sim.AddAgent(id, &a.BaseAgent)
		go a.Run(ctx, sim)
	}

	for i := 0; i < *numTrend; i++ {
		id := fmt.Sprintf("trend%d", i)
		a := agent.NewTrendFollower(id, 100000, 100, 0.1)
		sim.AddAgent(id, &a.BaseAgent)
		go a.Run(ctx, sim)
	}


	throughput := float64(sim.GetTradeCount()) / dur.Seconds()
	fmt.Printf("Simulation completed. Total trades: %d, Throughput: %.2f trades/sec\n", sim.GetTradeCount(), throughput)
	fmt.Printf("Final last price: %.2f\n", sim.Book.GetLastPrice())
	fmt.Printf("Goroutines used: many lightweight ones for agents, showcasing Go's concurrency efficiency compared to threads in other languages like Python (GIL-limited) or Java (heavier threads). For higher loads, try -random=500 -mm=50 -trend=500 -dur=1m\n")
}