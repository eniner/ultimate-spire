package lantern

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func detectDefaultRoot() string {
	if env := strings.TrimSpace(os.Getenv("SPIRE_LANTERN_ROOT")); env != "" {
		if info, err := os.Stat(env); err == nil && info.IsDir() {
			return filepath.Clean(env)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	candidates := []string{
		filepath.Join(home, "Downloads", "LanternExtractor-main", "LanternExtractor-main", "LanternExtractor", "Exports"),
		filepath.Join(home, "Desktop", "LanternExtractor-main", "LanternExtractor-main", "LanternExtractor", "Exports"),
		filepath.Join(home, "Documents", "LanternExtractor-main", "LanternExtractor-main", "LanternExtractor", "Exports"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return filepath.Clean(candidate)
		}
	}
	return ""
}

func resolveRoot() (string, error) {
	root := detectDefaultRoot()
	if root == "" {
		return "", fmt.Errorf("lantern export root not found; set SPIRE_LANTERN_ROOT to your LanternExtractor/Exports folder")
	}
	return root, nil
}

func sanitizeZone(raw string) (string, error) {
	zone := strings.ToLower(strings.TrimSpace(raw))
	if zone == "" {
		return "", fmt.Errorf("zone is required")
	}
	for _, r := range zone {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "", fmt.Errorf("zone name contains invalid characters")
	}
	return zone, nil
}

func isWithinRoot(rootDir, filePath string) bool {
	rel, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	return !filepath.IsAbs(rel)
}

func resolveModelPath(rootDir, rawRel string) (string, error) {
	rel := strings.TrimSpace(strings.ReplaceAll(rawRel, "\\", "/"))
	if rel == "" {
		return "", fmt.Errorf("model path is required")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("model path must be relative to the lantern root")
	}
	if !strings.EqualFold(filepath.Ext(rel), ".glb") {
		return "", fmt.Errorf("only .glb models are supported")
	}

	full := filepath.Clean(filepath.Join(rootDir, filepath.FromSlash(rel)))
	if !isWithinRoot(rootDir, full) {
		return "", fmt.Errorf("model path resolves outside the lantern root")
	}
	evalRoot, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		evalRoot = rootDir
	}
	evalFull, err := filepath.EvalSymlinks(full)
	if err != nil {
		return "", fmt.Errorf("model file not found")
	}
	if !isWithinRoot(evalRoot, evalFull) {
		return "", fmt.Errorf("model path resolves outside the lantern root")
	}
	info, err := os.Stat(evalFull)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("model file not found")
	}
	return evalFull, nil
}

type ZoneInfo struct {
	Zone         string `json:"zone"`
	ModelRelPath string `json:"modelRelPath"`
	HasMesh      bool   `json:"hasMesh"`
}

func listZones(rootDir string) ([]ZoneInfo, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}
	zones := make([]ZoneInfo, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, err := sanitizeZone(name); err != nil {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(name, "Zone", name+".glb"))
		abs := filepath.Join(rootDir, name, "Zone", name+".glb")
		hasMesh := false
		if info, err := os.Stat(abs); err == nil && info.Mode().IsRegular() {
			hasMesh = true
		}
		if !hasMesh {
			continue
		}
		zones = append(zones, ZoneInfo{
			Zone:         strings.ToLower(name),
			ModelRelPath: rel,
			HasMesh:      true,
		})
	}
	return zones, nil
}

func zoneModelRel(zone string) string {
	return zone + "/Zone/" + zone + ".glb"
}

type ObjectModel struct {
	ModelName    string `json:"modelName"`
	ModelRelPath string `json:"modelRelPath"`
}

func listObjectModels(rootDir, zone string) ([]ObjectModel, error) {
	dir := filepath.Join(rootDir, zone, "Objects")
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return []ObjectModel{}, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	models := make([]ObjectModel, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.EqualFold(filepath.Ext(name), ".glb") {
			continue
		}
		modelName := strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))
		rel := filepath.ToSlash(filepath.Join(zone, "Objects", name))
		abs := filepath.Join(rootDir, zone, "Objects", name)
		if !isWithinRoot(rootDir, abs) {
			continue
		}
		models = append(models, ObjectModel{
			ModelName:    modelName,
			ModelRelPath: rel,
		})
	}
	return models, nil
}
