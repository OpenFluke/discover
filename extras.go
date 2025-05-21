package discover

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// --------- PLANET AND SPAWN UTILITIES ---------

// 1. Generate evenly distributed points around a planet center using a Fibonacci sphere algorithm.
func FibonacciSphere(n int, radius float64, center []float64) [][]float64 {
	points := make([][]float64, n)
	if n == 0 {
		return points
	}
	if n == 1 {
		points[0] = []float64{
			center[0] + radius,
			center[1],
			center[2],
		}
		return points
	}
	phi := math.Pi * (3 - math.Sqrt(5)) // Golden angle in radians
	for i := 0; i < n; i++ {
		y := 1 - (float64(i)/float64(n-1))*2
		r := math.Sqrt(1 - y*y)
		theta := phi * float64(i)
		x := math.Cos(theta) * r
		z := math.Sin(theta) * r
		points[i] = []float64{
			center[0] + x*radius,
			center[1] + y*radius,
			center[2] + z*radius,
		}
	}
	return points
}

// 2. For a given planet, generate spawn positions on a sphere around it.
func (d *Discover) GenerateSpawnPositions(planetName string, n int, radius float64) ([][]float64, error) {
	planet, ok := d.Planets[planetName]
	if !ok {
		return nil, fmt.Errorf("planet %s not found", planetName)
	}
	return FibonacciSphere(n, radius, []float64{
		planet.Coordinates[0],
		planet.Coordinates[1],
		planet.Coordinates[2],
	}), nil
}

// 3. Calculate angle in degrees for an object at 'position' to face outward from a planet at 'center'
func CalculateRotationOutward(center, position []float64) float64 {
	dx := position[0] - center[0]
	dz := position[2] - center[2]
	angle := math.Atan2(dz, dx) * (180.0 / math.Pi)
	return angle
}

// 4. Find the closest planet to a given point (returns planet name and distance)
func (d *Discover) FindClosestPlanet(point []float64) (string, float64) {
	minDist := math.MaxFloat64
	var closest string
	for name, planet := range d.Planets {
		dx := planet.Coordinates[0] - point[0]
		dy := planet.Coordinates[1] - point[1]
		dz := planet.Coordinates[2] - point[2]
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist < minDist {
			minDist = dist
			closest = name
		}
	}
	return closest, minDist
}

// 5. Export planet table (name, x, y, z, host, port)
func (d *Discover) GetPlanetInfoTable() [][]string {
	table := [][]string{{"Name", "X", "Y", "Z", "Host", "Port"}}
	// Optional: sorted order
	names := make([]string, 0, len(d.Planets))
	for name := range d.Planets {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p := d.Planets[name]
		table = append(table, []string{
			p.Name,
			fmt.Sprintf("%.3f", p.Coordinates[0]),
			fmt.Sprintf("%.3f", p.Coordinates[1]),
			fmt.Sprintf("%.3f", p.Coordinates[2]),
			p.Host,
			fmt.Sprintf("%d", p.Port),
		})
	}
	return table
}

// 6. Test if a proposed spawn point is at least 'minDist' away from all planets.
func (d *Discover) IsSpawnPointFree(point []float64, minDist float64) bool {
	for _, planet := range d.Planets {
		dx := planet.Coordinates[0] - point[0]
		dy := planet.Coordinates[1] - point[1]
		dz := planet.Coordinates[2] - point[2]
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist < minDist {
			return false
		}
	}
	return true
}

// 7. (Optional) Get outward normal vector for a point on a sphere centered at planet
func OutwardNormal(center, point []float64) []float64 {
	v := []float64{
		point[0] - center[0],
		point[1] - center[1],
		point[2] - center[2],
	}
	mag := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if mag == 0 {
		return []float64{0, 1, 0}
	}
	return []float64{v[0] / mag, v[1] / mag, v[2] / mag}
}

func GenerateUnitID(role string, domain string, gen int, version int) string {
	domainParts := strings.Split(domain, ".")
	projectCode := ""
	for _, part := range domainParts {
		if len(part) > 0 {
			projectCode += strings.ToUpper(string(part[0]))
		}
	}
	return fmt.Sprintf("[%s]-%s-gen%d-v%d", strings.ToUpper(role), projectCode, gen, version)
}
