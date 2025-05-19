package main

import (
	"discover"
	"fmt"
)

func main() {
	cfg := discover.Config{
		Hosts:      []string{"localhost"},
		StartPort:  14000,
		PortStep:   3,
		NumPods:    1,
		AuthPass:   "my_secure_password",
		Delimiter:  "<???DONE???---",
		TimeoutSec: 10,
	}

	disco := discover.NewDiscover(cfg)
	disco.ScanAll()
	disco.PrintSummary()

	fmt.Println("\n-- Discovered Planets --")
	for name, p := range disco.Planets {
		fmt.Printf("%s at %v (host: %s)\n", name, p.Coordinates, p.Host)
	}

	fmt.Println("\n-- Discovered Cubes --")
	for cube, host := range disco.Cubes {
		fmt.Printf("%s at %s\n", cube, host)
	}

	// --- Use the new extras.go features ---
	fmt.Println("\n== Planet Info Table ==")
	table := disco.GetPlanetInfoTable()
	for _, row := range table {
		fmt.Println(row)
	}

	// Pick the first planet as an example
	var firstPlanetName string
	for name := range disco.Planets {
		firstPlanetName = name
		break
	}
	if firstPlanetName == "" {
		fmt.Println("No planets found.")
		return
	}
	planet := disco.Planets[firstPlanetName]
	center := planet.Coordinates[:] // <-- Fix: Copy to slice via variable, not directly from map

	// Generate 5 spawn positions at radius 120 around the first planet
	fmt.Printf("\n== Spawning 5 units around %s ==\n", firstPlanetName)
	spawnPositions, _ := disco.GenerateSpawnPositions(firstPlanetName, 5, 120.0)
	for i, pos := range spawnPositions {
		angle := discover.CalculateRotationOutward(center, pos)
		normal := discover.OutwardNormal(center, pos)
		free := disco.IsSpawnPointFree(pos, 50.0)
		fmt.Printf("Spawn %d: pos %v | Face angle: %.2f | Normal: %v | Free? %v\n", i+1, pos, angle, normal, free)
	}

	// Example: Find closest planet to a random point
	testPoint := []float64{500, 1000, 0}
	closest, dist := disco.FindClosestPlanet(testPoint)
	fmt.Printf("\nClosest planet to %v is %s (distance %.2f)\n", testPoint, closest, dist)
}
