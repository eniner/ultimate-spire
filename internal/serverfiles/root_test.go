package serverfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeRel(t *testing.T) {
	got, err := sanitizeRel(`cazicthule\player.pl`)
	if err != nil || got != "cazicthule/player.pl" {
		t.Fatalf("got %q %v", got, err)
	}
	if _, err := sanitizeRel("../etc/passwd"); err == nil {
		t.Fatal("expected traversal to fail")
	}
	if _, err := sanitizeRel(".git/config"); err == nil {
		t.Fatal("expected hidden path to fail")
	}
}

func TestDetectSource(t *testing.T) {
	if detectSource("https://s3.amazonaws.com/bucket/quests") != SourceHTTP {
		t.Fatal("expected http")
	}
	if detectSource(`C:\eqemu\quests`) != SourceLocal {
		t.Fatal("expected local")
	}
}

func TestResolveAbsStaysInRoot(t *testing.T) {
	root := t.TempDir()
	s := &Service{}
	if err := os.WriteFile(filepath.Join(root, "player.pl"), []byte("say(1)"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := s.resolveAbs(root, "player.pl")
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(root, "player.pl") {
		t.Fatalf("got %s", got)
	}
	if _, err := s.resolveAbs(root, "../player.pl"); err == nil {
		t.Fatal("expected traversal to fail")
	}
}

func TestLocalListAndWrite(t *testing.T) {
	root := t.TempDir()
	zone := filepath.Join(root, "cazicthule")
	if err := os.MkdirAll(zone, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zone, "player.pl"), []byte("sub EVENT_SAY {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	s := &Service{}
	entries, err := s.listLocal(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !entries[0].IsDir || entries[0].Name != "cazicthule" {
		t.Fatalf("unexpected listing %#v", entries)
	}
	body, err := s.readLocal(root, "cazicthule/player.pl")
	if err != nil || !strings.Contains(body, "EVENT_SAY") {
		t.Fatalf("read failed %q %v", body, err)
	}
	if err := s.writeLocal(root, "cazicthule/player.pl", "sub EVENT_SPAWN {}\n", "", false); err != nil {
		t.Fatal(err)
	}
	body, err = s.readLocal(root, "cazicthule/player.pl")
	if err != nil || !strings.Contains(body, "EVENT_SPAWN") {
		t.Fatalf("write did not persist %q %v", body, err)
	}
	if err := s.writeLocal(root, "cazicthule/new.lua", "function event_say() end\n", "", true); err != nil {
		t.Fatal(err)
	}
}

func TestParseHTTPRootBlocksInternal(t *testing.T) {
	if _, err := parseHTTPRoot("http://127.0.0.1/"); err == nil {
		t.Fatal("expected loopback blocked")
	}
	if _, err := parseHTTPRoot("http://user:pass@example.com/"); err == nil {
		t.Fatal("expected userinfo blocked")
	}
	if _, err := parseHTTPRoot("ftp://files.example.com/"); err == nil {
		t.Fatal("expected scheme blocked")
	}
}

func TestWriteLocalRejectsStaleHash(t *testing.T) {
	root := t.TempDir()
	s := &Service{}
	if err := os.WriteFile(filepath.Join(root, "player.pl"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := s.writeLocal(root, "player.pl", "new\n", fileHash("not-old"), false); err == nil {
		t.Fatal("expected stale hash")
	}
	if err := s.writeLocal(root, "player.pl", "new\n", fileHash("old"), false); err != nil {
		t.Fatal(err)
	}
}

func TestResolveAbsRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.pl")
	if err := os.WriteFile(secret, []byte("nope"), 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape.pl")
	if err := os.Symlink(secret, link); err != nil {
		t.Skip("symlink not permitted")
	}
	s := &Service{}
	if _, err := s.resolveAbs(root, "escape.pl"); err == nil {
		t.Fatal("expected symlink escape to fail")
	}
}

func TestParseHTMLListing(t *testing.T) {
	html := `<html><a href="../">..</a><a href="cazicthule/">cazicthule/</a><a href="global.pl">global.pl</a><a href="skip.bin">skip.bin</a></html>`
	got := parseHTMLListing("https://files.example.com/quests/", html, "")
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %#v", got)
	}
}
