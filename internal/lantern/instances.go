package lantern

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ObjectInstance struct {
	ModelName    string     `json:"modelName"`
	ModelRelPath string     `json:"modelRelPath"`
	Pos          [3]float64 `json:"pos"`
	Rot          [3]float64 `json:"rot"`
	Scale        [3]float64 `json:"scale"`
	ColorIndex   int        `json:"colorIndex"`
}

func parseObjectInstances(rootDir, zone string) ([]ObjectInstance, error) {
	filePath := filepath.Join(rootDir, zone, "Zone", "object_instances.txt")
	info, err := os.Stat(filePath)
	if err != nil || !info.Mode().IsRegular() {
		return []ObjectInstance{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	instances := make([]ObjectInstance, 0, 256)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 11 {
			continue
		}
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		modelName := strings.ToLower(parts[0])
		if modelName == "" {
			continue
		}

		nums := make([]float64, 9)
		ok := true
		for i := 0; i < 9; i++ {
			n, err := strconv.ParseFloat(parts[i+1], 64)
			if err != nil {
				ok = false
				break
			}
			nums[i] = n
		}
		if !ok {
			continue
		}
		colorIndex, _ := strconv.Atoi(parts[10])

		rel := zone + "/Objects/" + modelName + ".glb"
		abs := filepath.Join(rootDir, filepath.FromSlash(rel))
		if !isWithinRoot(rootDir, abs) {
			continue
		}
		if st, err := os.Stat(abs); err != nil || !st.Mode().IsRegular() {
			continue
		}

		instances = append(instances, ObjectInstance{
			ModelName:    modelName,
			ModelRelPath: rel,
			Pos:          [3]float64{nums[0], nums[1], nums[2]},
			Rot:          [3]float64{nums[3], nums[4], nums[5]},
			Scale:        [3]float64{nums[6], nums[7], nums[8]},
			ColorIndex:   colorIndex,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return instances, nil
}
