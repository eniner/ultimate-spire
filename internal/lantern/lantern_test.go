package lantern

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeZone(t *testing.T) {
	if _, err := sanitizeZone("overthere"); err != nil {
		t.Fatalf("expected valid zone, got %v", err)
	}
	if _, err := sanitizeZone("../etc"); err == nil {
		t.Fatal("expected invalid zone")
	}
	if _, err := sanitizeZone("over there"); err == nil {
		t.Fatal("expected invalid zone")
	}
}

func TestResolveModelPathStaysInRoot(t *testing.T) {
	root := t.TempDir()
	zoneDir := filepath.Join(root, "crushbone", "Zone")
	if err := os.MkdirAll(zoneDir, 0755); err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(zoneDir, "crushbone.glb")
	if err := os.WriteFile(good, []byte("glb"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveModelPath(root, "crushbone/Zone/crushbone.glb")
	if err != nil {
		t.Fatalf("expected model path, got %v", err)
	}
	if got != good {
		t.Fatalf("got %s want %s", got, good)
	}

	if _, err := resolveModelPath(root, "../crushbone/Zone/crushbone.glb"); err == nil {
		t.Fatal("expected traversal to fail")
	}
	if _, err := resolveModelPath(root, "crushbone/Zone/crushbone.obj"); err == nil {
		t.Fatal("expected non-glb to fail")
	}
}

func TestParseObjectInstances(t *testing.T) {
	root := t.TempDir()
	zone := "crushbone"
	if err := os.MkdirAll(filepath.Join(root, zone, "Zone"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, zone, "Objects"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, zone, "Objects", "tree.glb"), []byte("glb"), 0644); err != nil {
		t.Fatal(err)
	}
	txt := strings.Join([]string{
		"# Lantern Extractor 0.1.7 - Object Instances",
		"tree,1,2,3,0,45,0,1,1,1,-1",
		"missing,1,2,3,0,0,0,1,1,1,0",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, zone, "Zone", "object_instances.txt"), []byte(txt), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := parseObjectInstances(root, zone)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(got))
	}
	if got[0].ModelName != "tree" || got[0].Pos[0] != 1 || got[0].Rot[1] != 45 {
		t.Fatalf("unexpected instance %#v", got[0])
	}
}

func TestWriteObjectInstancesRoundTrip(t *testing.T) {
	root := t.TempDir()
	zone := "crushbone"
	if err := os.MkdirAll(filepath.Join(root, zone, "Zone"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, zone, "Objects"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, zone, "Objects", "tree.glb"), []byte("glb"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, zone, "Zone", "object_instances.txt"), []byte("tree,0,0,0,0,0,0,1,1,1,0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := writeObjectInstances(root, zone, []ObjectInstance{{
		ModelName:  "tree",
		Pos:        [3]float64{10, 20, 30},
		Rot:        [3]float64{0, 90, 0},
		Scale:      [3]float64{1, 1, 1},
		ColorIndex: -1,
	}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseObjectInstances(root, zone)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Pos[0] != 10 || got[0].Rot[1] != 90 || got[0].ColorIndex != -1 {
		t.Fatalf("unexpected rewrite %#v", got)
	}
	if _, err := os.Stat(filepath.Join(root, zone, "Zone", "object_instances.txt.bak")); err != nil {
		t.Fatalf("expected backup: %v", err)
	}
}

func TestWriteObjectInstancesRejectsMissingModel(t *testing.T) {
	root := t.TempDir()
	zone := "crushbone"
	if err := os.MkdirAll(filepath.Join(root, zone, "Zone"), 0755); err != nil {
		t.Fatal(err)
	}
	err := writeObjectInstances(root, zone, []ObjectInstance{{
		ModelName: "missing",
		Scale:     [3]float64{1, 1, 1},
	}})
	if err == nil {
		t.Fatal("expected missing model to fail")
	}
}

func TestListZonesRequiresCanonicalMesh(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "abh", "Characters"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "abh", "Characters", "abh.glb"), []byte("glb"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "crushbone", "Zone"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "crushbone", "Zone", "crushbone.glb"), []byte("glb"), 0644); err != nil {
		t.Fatal(err)
	}

	zones, err := listZones(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(zones) != 1 || zones[0].Zone != "crushbone" {
		t.Fatalf("expected only crushbone, got %#v", zones)
	}
}
