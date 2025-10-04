package main

import (
	"context"
	"flag"
	"fmt"
	"time"
)

var (
	numRandom = flag.Int("random", 50, "number of random agents")
	numMM     = flag.Int("mm", 10, "number of market maker agents")
	numTrend  = flag.Int("trend", 50, "number of trend follower agents")
	dur       = flag.Duration("dur", 30*time.Second, "simulation duration")
)

func main() {
	flag.Parse()

}