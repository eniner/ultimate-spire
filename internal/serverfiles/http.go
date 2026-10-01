package serverfiles

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
)

var hrefRe = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)

type s3List struct {
	XMLName        xml.Name `xml:"ListBucketResult"`
	Contents       []struct {
		Key          string `xml:"Key"`
		Size         int64  `xml:"Size"`
		LastModified string `xml:"LastModified"`
	} `xml:"Contents"`
	CommonPrefixes []struct {
		Prefix string `xml:"Prefix"`
	} `xml:"CommonPrefixes"`
}

func (s *Service) fetchHTTP(raw string) ([]byte, string, error) {
	u, err := parseHTTPRoot(raw)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	res, err := s.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, "", fmt.Errorf("remote host returned HTTP %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxFileBytes+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(body)) > maxFileBytes {
		return nil, "", errors.New("remote file is too large to open")
	}
	return body, res.Header.Get("Content-Type"), nil
}

func joinHTTP(root, rel string) (string, error) {
	base, err := parseHTTPRoot(root)
	if err != nil {
		return "", err
	}
	cleanRel, err := sanitizeRel(rel)
	if err != nil {
		return "", err
	}
	joined := *base
	if cleanRel != "" {
		joined.Path = strings.TrimRight(joined.Path, "/") + "/" + cleanRel
	}
	return joined.String(), nil
}

func (s *Service) listHTTP(root, rel string) ([]Entry, error) {
	raw, err := joinHTTP(root, rel)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(raw, "/") {
		raw += "/"
	}
	body, ctype, err := s.fetchHTTP(raw)
	if err != nil {
		return nil, err
	}
	text := string(body)
	cleanRel, _ := sanitizeRel(rel)
	var entries []Entry
	if strings.Contains(ctype, "xml") || strings.Contains(text, "<ListBucketResult") {
		entries = parseS3List(text, cleanRel)
	}
	if len(entries) == 0 {
		entries = parseHTMLListing(raw, text, cleanRel)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	if len(entries) > maxListItems {
		entries = entries[:maxListItems]
	}
	return entries, nil
}

func parseS3List(body, prefix string) []Entry {
	var list s3List
	if xml.Unmarshal([]byte(body), &list) != nil {
		return nil
	}
	want := prefix
	if want != "" && !strings.HasSuffix(want, "/") {
		want += "/"
	}
	seen := map[string]bool{}
	var entries []Entry
	add := func(e Entry) {
		if e.Name == "" || seen[e.Path] {
			return
		}
		seen[e.Path] = true
		entries = append(entries, e)
	}
	for _, p := range list.CommonPrefixes {
		key := strings.Trim(strings.TrimPrefix(p.Prefix, want), "/")
		if key == "" || strings.Contains(key, "/") {
			continue
		}
		child := key
		if prefix != "" {
			child = prefix + "/" + key
		}
		add(Entry{Name: key, Path: child, IsDir: true})
	}
	for _, c := range list.Contents {
		rest := strings.TrimPrefix(c.Key, want)
		if rest == "" || strings.Contains(rest, "/") {
			parts := strings.Split(strings.Trim(rest, "/"), "/")
			if len(parts) > 1 && parts[0] != "" {
				child := parts[0]
				if prefix != "" {
					child = prefix + "/" + parts[0]
				}
				add(Entry{Name: parts[0], Path: child, IsDir: true})
			}
			continue
		}
		if !allowedFile(rest) {
			continue
		}
		child := rest
		if prefix != "" {
			child = prefix + "/" + rest
		}
		add(Entry{Name: rest, Path: child, Size: c.Size, Editable: true})
	}
	return entries
}

func parseHTMLListing(pageURL, body, prefix string) []Entry {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var entries []Entry
	for _, match := range hrefRe.FindAllStringSubmatch(body, maxListItems*2) {
		href := strings.TrimSpace(match[1])
		if href == "" || href == "/" || strings.HasPrefix(href, "?") || strings.HasPrefix(href, "#") {
			continue
		}
		if strings.HasPrefix(href, "../") || href == ".." {
			continue
		}
		u, err := base.Parse(href)
		if err != nil || u.Host != base.Host {
			continue
		}
		name := path.Base(strings.TrimRight(u.Path, "/"))
		if name == "" || name == "." || strings.HasPrefix(name, ".") {
			continue
		}
		decoded, err := url.PathUnescape(name)
		if err == nil {
			name = decoded
		}
		isDir := strings.HasSuffix(u.Path, "/")
		if !isDir && !allowedFile(name) {
			continue
		}
		child := name
		if prefix != "" {
			child = prefix + "/" + name
		}
		if seen[child] {
			continue
		}
		seen[child] = true
		entries = append(entries, Entry{
			Name:     name,
			Path:     child,
			IsDir:    isDir,
			Editable: !isDir && allowedFile(name),
		})
	}
	return entries
}

func (s *Service) readHTTP(root, rel string) (string, error) {
	if !allowedFile(rel) {
		return "", errors.New("that file type cannot be opened in Spire")
	}
	raw, err := joinHTTP(root, rel)
	if err != nil {
		return "", err
	}
	body, _, err := s.fetchHTTP(raw)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
