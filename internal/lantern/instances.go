package lantern

import (
	"bufio"
	"fmt"
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

const maxObjectInstances = 5000

func sanitizeModelName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if name == "" {
		return "", fmt.Errorf("model name is required")
	}
	if len(name) > 64 {
		return "", fmt.Errorf("model name is too long")
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "", fmt.Errorf("model name contains invalid characters")
	}
	return name, nil
}

func objectInstancesPath(rootDir, zone string) string {
	return filepath.Join(rootDir, zone, "Zone", "object_instances.txt")
}

func modelExists(rootDir, zone, modelName string) bool {
	abs := filepath.Join(rootDir, zone, "Objects", modelName+".glb")
	if !isWithinRoot(rootDir, abs) {
		return false
	}
	st, err := os.Stat(abs)
	return err == nil && st.Mode().IsRegular()
}

func validateInstances(rootDir, zone string, instances []ObjectInstance) error {
	if len(instances) > maxObjectInstances {
		return fmt.Errorf("too many instances (max %d)", maxObjectInstances)
	}
	for i := range instances {
		name, err := sanitizeModelName(instances[i].ModelName)
		if err != nil {
			return fmt.Errorf("instances[%d]: %v", i, err)
		}
		if !modelExists(rootDir, zone, name) {
			return fmt.Errorf("instances[%d]: model %q is not in %s/Objects", i, name, zone)
		}
		instances[i].ModelName = name
		instances[i].ModelRelPath = zone + "/Objects/" + name + ".glb"
		if instances[i].Scale[0] == 0 && instances[i].Scale[1] == 0 && instances[i].Scale[2] == 0 {
			instances[i].Scale = [3]float64{1, 1, 1}
		}
	}
	return nil
}

func formatFloat(n float64) string {
	s := strconv.FormatFloat(n, 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

func formatObjectInstances(instances []ObjectInstance) string {
	var b strings.Builder
	b.WriteString("# Lantern object instances — authored by Spire Atlas\n")
	b.WriteString("# model,posx,posy,posz,rotx,roty,rotz,sx,sy,sz,colorIndex\n")
	for _, inst := range instances {
		b.WriteString(inst.ModelName)
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Pos[0]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Pos[1]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Pos[2]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Rot[0]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Rot[1]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Rot[2]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Scale[0]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Scale[1]))
		b.WriteByte(',')
		b.WriteString(formatFloat(inst.Scale[2]))
		b.WriteByte(',')
		b.WriteString(strconv.Itoa(inst.ColorIndex))
		b.WriteByte('\n')
	}
	return b.String()
}

func writeObjectInstances(rootDir, zone string, instances []ObjectInstance) error {
	if err := validateInstances(rootDir, zone, instances); err != nil {
		return err
	}
	zoneDir := filepath.Join(rootDir, zone, "Zone")
	if !isWithinRoot(rootDir, zoneDir) {
		return fmt.Errorf("zone path resolves outside the lantern root")
	}
	if err := os.MkdirAll(zoneDir, 0755); err != nil {
		return err
	}
	dest := objectInstancesPath(rootDir, zone)
	if !isWithinRoot(rootDir, dest) {
		return fmt.Errorf("instance path resolves outside the lantern root")
	}
	body := []byte(formatObjectInstances(instances))
	if st, err := os.Stat(dest); err == nil && st.Mode().IsRegular() {
		bak := dest + ".bak"
		if current, readErr := os.ReadFile(dest); readErr == nil {
			_ = os.WriteFile(bak, current, 0644)
		}
	}
	tmp, err := os.CreateTemp(zoneDir, ".spire-instances-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(dest)
		if err = os.Rename(tmpName, dest); err != nil {
			_ = os.Remove(tmpName)
			return err
		}
	}
	return nil
}
