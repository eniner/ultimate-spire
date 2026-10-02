package serverfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	Editable bool   `json:"editable"`
}

func (s *Service) listLocal(root, rel string) ([]Entry, error) {
	abs, err := s.resolveAbs(root, rel)
	if err != nil {
		return nil, err
	}
	items, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("could not read folder: %w", err)
	}
	cleanRel, _ := sanitizeRel(rel)
	entries := make([]Entry, 0, len(items))
	for _, item := range items {
		name := item.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		info, err := item.Info()
		if err != nil {
			continue
		}
		child := name
		if cleanRel != "" {
			child = cleanRel + "/" + name
		}
		isDir := item.IsDir()
		if !isDir && !allowedFile(name) {
			continue
		}
		entries = append(entries, Entry{
			Name:     name,
			Path:     child,
			IsDir:    isDir,
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
			Editable: !isDir && allowedFile(name),
		})
		if len(entries) >= maxListItems {
			break
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries, nil
}

func fileHash(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

func (s *Service) readLocal(root, rel string) (string, error) {
	abs, err := s.resolveAbs(root, rel)
	if err != nil {
		return "", err
	}
	if !allowedFile(abs) {
		return "", errors.New("that file type cannot be opened in Spire")
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}
	if info.IsDir() {
		return "", errors.New("path is a folder")
	}
	if info.Size() > maxFileBytes {
		return "", fmt.Errorf("file is too large to edit (%d bytes)", info.Size())
	}
	body, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *Service) writeLocal(root, rel, content, expectedHash string, create bool) error {
	if len(content) > maxFileBytes {
		return errors.New("file is too large to save")
	}
	abs, err := s.resolveAbs(root, rel)
	if err != nil {
		return err
	}
	if !allowedFile(abs) {
		return errors.New("that file type cannot be saved from Spire")
	}
	info, err := os.Stat(abs)
	if create {
		if err == nil {
			return errors.New("file already exists")
		}
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("file not found: %w", err)
	} else if info.IsDir() {
		return errors.New("path is a folder")
	} else if expectedHash != "" {
		current, readErr := os.ReadFile(abs)
		if readErr != nil {
			return readErr
		}
		if fileHash(string(current)) != expectedHash {
			return errStaleFile
		}
	}
	dir := filepath.Dir(abs)
	tmp, err := os.CreateTemp(dir, ".spire-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write([]byte(content)); err != nil {
		tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, abs); err != nil {
		_ = os.Remove(abs)
		if err = os.Rename(tmpName, abs); err != nil {
			_ = os.Remove(tmpName)
			return err
		}
	}
	return nil
}

var errStaleFile = errors.New("file changed on disk; reload before saving")

func (s *Service) searchLocal(root, rel, q string) ([]Entry, error) {
	q = strings.ToLower(strings.TrimSpace(q))
	if len(q) < 2 {
		return nil, errors.New("search needs at least 2 characters")
	}
	abs, err := s.resolveAbs(root, rel)
	if err != nil {
		return nil, err
	}
	cleanRel, _ := sanitizeRel(rel)
	var found []Entry
	err = filepath.Walk(abs, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info == nil {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() || !allowedFile(info.Name()) {
			return nil
		}
		if !strings.Contains(strings.ToLower(info.Name()), q) && !strings.Contains(strings.ToLower(path), q) {
			return nil
		}
		relPath, err := filepath.Rel(abs, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)
		if cleanRel != "" {
			relPath = cleanRel + "/" + relPath
		}
		if _, err := sanitizeRel(relPath); err != nil {
			return nil
		}
		found = append(found, Entry{
			Name:     info.Name(),
			Path:     relPath,
			IsDir:    false,
			Size:     info.Size(),
			Modified: info.ModTime().Unix(),
			Editable: true,
		})
		if len(found) >= 200 {
			return errors.New("limit")
		}
		return nil
	})
	if err != nil && err.Error() != "limit" {
		return nil, err
	}
	sort.Slice(found, func(i, j int) bool {
		return strings.ToLower(found[i].Path) < strings.ToLower(found[j].Path)
	})
	return found, nil
}
