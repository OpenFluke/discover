package discover

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// --- Result Types ---

type PlanetRecord struct {
	Name        string
	Coordinates [3]float64
	Host        string
	Port        int
}

type PodResult struct {
	Host    string
	Port    int
	Success bool
	Error   string
	Cubes   []string
	Planets []PlanetRecord
}

// --- Full planet struct for server JSON ---

type Planet struct {
	Position          map[string]float64   `json:"Position"`
	Seed              int                  `json:"Seed"`
	Name              string               `json:"Name"`
	ResourceLocations []map[string]float64 `json:"ResourceLocations"`
	TreeLocations     []map[string]float64 `json:"TreeLocations"`
	BiomeType         int                  `json:"BiomeType"`
}

// --- Main scan logic ---

func ScanPod(host string, port int, auth string, delim string, timeout int) PodResult {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, time.Duration(timeout)*time.Second)
	if err != nil {
		return PodResult{Host: host, Port: port, Success: false, Error: err.Error()}
	}
	defer conn.Close()

	// Authenticate
	if err := sendMsg(conn, auth, delim); err != nil {
		return PodResult{Host: host, Port: port, Success: false, Error: "Auth failed"}
	}
	if !strings.Contains(readMsg(conn, timeout, delim), "auth_success") {
		return PodResult{Host: host, Port: port, Success: false, Error: "Bad password"}
	}

	// Get Cubes
	if err := sendMsg(conn, `{"type":"get_cube_list"}`, delim); err != nil {
		return PodResult{Host: host, Port: port, Success: false, Error: "Cube req fail"}
	}
	var cubesData map[string]interface{}
	if err := json.Unmarshal([]byte(readMsg(conn, timeout, delim)), &cubesData); err != nil {
		return PodResult{Host: host, Port: port, Success: false, Error: "Cube parse fail"}
	}
	cubes := toStringSlice(cubesData["cubes"])

	// Get Planets (server returns: map[string][]Planet)
	if err := sendMsg(conn, `{"type":"get_planets"}`, delim); err != nil {
		return PodResult{Host: host, Port: port, Success: false, Error: "Planet req fail"}
	}
	raw := readMsg(conn, timeout, delim)
	var planetsData map[string][]Planet
	if err := json.Unmarshal([]byte(raw), &planetsData); err != nil {
		return PodResult{Host: host, Port: port, Success: false, Error: "Planet parse fail"}
	}
	var planetRecords []PlanetRecord
	for _, ps := range planetsData {
		for _, p := range ps {
			coords := [3]float64{0, 0, 0}
			if p.Position != nil {
				coords[0] = p.Position["x"]
				coords[1] = p.Position["y"]
				coords[2] = p.Position["z"]
			}
			planetRecords = append(planetRecords, PlanetRecord{
				Name:        p.Name,
				Coordinates: coords,
				Host:        host,
				Port:        port,
			})
		}
	}
	return PodResult{Host: host, Port: port, Success: true, Cubes: cubes, Planets: planetRecords}
}

// --- helpers ---

func sendMsg(conn net.Conn, msg string, delim string) error {
	_, err := conn.Write([]byte(msg + delim))
	return err
}

func readMsg(conn net.Conn, timeout int, delim string) string {
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	reader := bufio.NewReader(conn)
	var buf bytes.Buffer

	for {
		chunk, err := reader.ReadString(delim[len(delim)-1]) // read up to possible delim ending char
		if err != nil && err != io.EOF {
			break
		}
		buf.WriteString(chunk)
		if strings.Contains(buf.String(), delim) {
			break
		}
		if err == io.EOF {
			break
		}
	}
	full := buf.String()
	// Remove delimiter and any trailing/leading whitespace
	return strings.TrimSpace(strings.ReplaceAll(full, delim, ""))
}

func toStringSlice(v interface{}) []string {
	if arr, ok := v.([]interface{}); ok {
		out := make([]string, 0, len(arr))
		for _, el := range arr {
			if str, ok := el.(string); ok {
				out = append(out, str)
			}
		}
		return out
	}
	return nil
}
