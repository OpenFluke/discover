package discover

import (
	"fmt"
	"sync"
)

// Only use PodResult and PlanetRecord from pod.go!

type Discover struct {
	Config  Config
	Results []PodResult
	Planets map[string]PlanetRecord
	Cubes   map[string]string // cubeName -> host
	mu      sync.Mutex
}

type Config struct {
	Hosts      []string
	StartPort  int
	PortStep   int
	NumPods    int
	AuthPass   string
	Delimiter  string // Now part of the config
	TimeoutSec int
}

func NewDiscover(cfg Config) *Discover {
	return &Discover{
		Config:  cfg,
		Planets: make(map[string]PlanetRecord),
		Cubes:   make(map[string]string),
	}
}

func (d *Discover) ScanAll() {
	var wg sync.WaitGroup
	resultsChan := make(chan PodResult, d.Config.NumPods*len(d.Config.Hosts))

	for _, host := range d.Config.Hosts {
		for i := 0; i < d.Config.NumPods; i++ {
			port := d.Config.StartPort + i*d.Config.PortStep
			wg.Add(1)
			go func(host string, port int) {
				defer wg.Done()
				result := ScanPod(host, port, d.Config.AuthPass, d.Config.Delimiter, d.Config.TimeoutSec)
				resultsChan <- result
			}(host, port)
		}
	}

	wg.Wait()
	close(resultsChan)

	for result := range resultsChan {
		d.mu.Lock()
		d.Results = append(d.Results, result)
		if result.Success {
			for _, planet := range result.Planets {
				d.Planets[planet.Name] = planet
			}
			for _, cube := range result.Cubes {
				d.Cubes[cube] = result.Host
			}
		}
		d.mu.Unlock()
	}
}

func (d *Discover) PrintSummary() {
	totalCubes, totalPlanets, successCount := 0, 0, 0
	fmt.Println("\n=== D.I.S.C.O.V.E.R.™ SUMMARY ===")
	for _, res := range d.Results {
		if res.Success {
			successCount++
			totalCubes += len(res.Cubes)
			totalPlanets += len(res.Planets)
			fmt.Printf("[%s:%d] ✅ Cubes=%d Planets=%d\n", res.Host, res.Port, len(res.Cubes), len(res.Planets))
		} else {
			fmt.Printf("[%s:%d] ❌ %s\n", res.Host, res.Port, res.Error)
		}
	}
	fmt.Printf("\nSuccessful pods: %d / %d\n", successCount, d.Config.NumPods*len(d.Config.Hosts))
	fmt.Printf("Total Cubes: %d\n", totalCubes)
	fmt.Printf("Total Planets: %d\n", totalPlanets)
	fmt.Printf("Unique Planets: %d\n", len(d.Planets))
}

// ExtractPlanetCenters returns a slice of [x, y, z] float64 slices for each planet discovered.
func (d *Discover) ExtractPlanetCenters() [][]float64 {
	centers := [][]float64{}
	for _, planet := range d.Planets {
		// Copy the [3]float64 array to a slice so users can easily use it
		c := make([]float64, 3)
		copy(c, planet.Coordinates[:])
		centers = append(centers, c)
	}
	return centers
}
