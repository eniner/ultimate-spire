package serverfiles

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/EQEmu/spire/internal/eqemuserverconfig"
	"github.com/EQEmu/spire/internal/pathmgmt"
)

const (
	SourceLocal = "local"
	SourceHTTP  = "http"
	maxFileBytes = 1_500_000
	maxListItems = 4000
)

var allowedExt = map[string]bool{
	".pl": true, ".lua": true, ".py": true, ".inc": true,
	".txt": true, ".md": true, ".json": true, ".sql": true,
	".xml": true, ".cfg": true, ".ini": true, ".csv": true,
	".perl": true, ".pcc": true,
}

type Service struct {
	pathmgmt *pathmgmt.PathManagement
	config   *eqemuserverconfig.Config
	http     *http.Client
}

func NewService(pathmgmt *pathmgmt.PathManagement, config *eqemuserverconfig.Config) *Service {
	return &Service{
		pathmgmt: pathmgmt,
		config:   config,
		http: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
					return errors.New("redirect blocked")
				}
				if len(via) >= 5 {
					return errors.New("too many redirects")
				}
				if err := rejectBlockedURL(req.URL); err != nil {
					return err
				}
				return nil
			},
		},
	}
}

type Connection struct {
	Source   string `json:"source"`
	Root     string `json:"root"`
	Writable bool   `json:"writable"`
}

type Detected struct {
	ServerPath   string   `json:"serverPath"`
	QuestsDir    string   `json:"questsDir"`
	ConfigQuests string   `json:"configQuests"`
	EnvRoot      string   `json:"envRoot"`
	Suggestions  []string `json:"suggestions"`
}

func detectSource(root string) string {
	r := strings.ToLower(strings.TrimSpace(root))
	if strings.HasPrefix(r, "http://") || strings.HasPrefix(r, "https://") {
		return SourceHTTP
	}
	return SourceLocal
}

func sanitizeRel(rel string) (string, error) {
	rel = strings.ReplaceAll(rel, "\\", "/")
	rel = strings.Trim(rel, "/")
	if rel == "" || rel == "." {
		return "", nil
	}
	parts := strings.Split(rel, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		if p == ".." || strings.ContainsRune(p, 0) {
			return "", errors.New("invalid path")
		}
		if strings.HasPrefix(p, ".") {
			return "", errors.New("hidden paths are not allowed")
		}
		for _, r := range p {
			if r < 32 || !unicode.IsPrint(r) {
				return "", errors.New("invalid path")
			}
		}
		out = append(out, p)
	}
	return strings.Join(out, "/"), nil
}

func allowedFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return allowedExt[ext]
}

func (s *Service) detected() Detected {
	d := Detected{
		ServerPath: s.pathmgmt.GetEQEmuServerPath(),
		QuestsDir:  s.pathmgmt.GetQuestsDir(),
		EnvRoot:    strings.TrimSpace(os.Getenv("SPIRE_QUESTS_ROOT")),
	}
	if d.EnvRoot == "" {
		d.EnvRoot = strings.TrimSpace(os.Getenv("SPIRE_SERVER_FILES_ROOT"))
	}
	if cfg, err := s.config.Get(); err == nil {
		d.ConfigQuests = strings.TrimSpace(cfg.Server.Directories.Quests)
	}
	seen := map[string]bool{}
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			return
		}
		if detectSource(p) == SourceLocal {
			if info, err := os.Stat(p); err != nil || !info.IsDir() {
				return
			}
		}
		seen[p] = true
		d.Suggestions = append(d.Suggestions, p)
	}
	add(d.EnvRoot)
	add(d.ConfigQuests)
	add(d.QuestsDir)
	add(d.ServerPath)
	if d.ServerPath != "" {
		add(filepath.Join(d.ServerPath, "quests"))
		add(filepath.Join(d.ServerPath, "plugins"))
		add(filepath.Join(d.ServerPath, "lua_modules"))
	}
	return d
}

func (s *Service) savedConnection() (Connection, bool) {
	cfg, err := s.config.Get()
	if err != nil || cfg.Spire.Files == nil {
		return Connection{}, false
	}
	root := strings.TrimSpace(cfg.Spire.Files.Root)
	if root == "" {
		return Connection{}, false
	}
	src := strings.TrimSpace(cfg.Spire.Files.Source)
	if src == "" {
		src = detectSource(root)
	}
	return Connection{Source: src, Root: root, Writable: src == SourceLocal}, true
}

func (s *Service) currentConnection() Connection {
	if saved, ok := s.savedConnection(); ok {
		return saved
	}
	d := s.detected()
	if d.EnvRoot != "" {
		return Connection{Source: detectSource(d.EnvRoot), Root: d.EnvRoot, Writable: detectSource(d.EnvRoot) == SourceLocal}
	}
	if d.ConfigQuests != "" {
		return Connection{Source: SourceLocal, Root: d.ConfigQuests, Writable: true}
	}
	if d.QuestsDir != "" {
		if info, err := os.Stat(d.QuestsDir); err == nil && info.IsDir() {
			return Connection{Source: SourceLocal, Root: d.QuestsDir, Writable: true}
		}
	}
	return Connection{}
}

func (s *Service) saveConnection(src, root string) error {
	if s.pathmgmt.GetEQEmuServerPath() == "" {
		return errors.New("eqemu_config.json not found; run Spire from the EQ server folder or set SPIRE_QUESTS_ROOT")
	}
	cfg, err := s.config.Get()
	if err != nil {
		return err
	}
	cfg.Spire.Files = &eqemuserverconfig.SpireFilesConfig{
		Source: src,
		Root:   root,
	}
	return s.config.Save(cfg)
}

func (s *Service) probe(src, root string) (Connection, error) {
	src = strings.TrimSpace(src)
	root = strings.TrimSpace(root)
	if root == "" {
		return Connection{}, errors.New("folder or URL is required")
	}
	if src == "" {
		src = detectSource(root)
	}
	if src != SourceLocal && src != SourceHTTP {
		return Connection{}, fmt.Errorf("unknown source %q", src)
	}
	if src == SourceHTTP {
		if detectSource(root) != SourceHTTP {
			return Connection{}, errors.New("HTTP source needs an http:// or https:// URL")
		}
		if _, err := parseHTTPRoot(root); err != nil {
			return Connection{}, err
		}
		req, err := http.NewRequest(http.MethodGet, strings.TrimRight(root, "/")+"/", nil)
		if err != nil {
			return Connection{}, err
		}
		res, err := s.http.Do(req)
		if err != nil {
			return Connection{}, fmt.Errorf("could not reach URL: %w", err)
		}
		res.Body.Close()
		if res.StatusCode >= 400 {
			return Connection{}, fmt.Errorf("remote host returned HTTP %d", res.StatusCode)
		}
		return Connection{Source: SourceHTTP, Root: strings.TrimRight(root, "/"), Writable: false}, nil
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return Connection{}, errors.New("invalid folder path")
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Connection{}, fmt.Errorf("folder not found: %s", abs)
	}
	if !info.IsDir() {
		return Connection{}, errors.New("path is a file; point at the folder that contains quests")
	}
	return Connection{Source: SourceLocal, Root: abs, Writable: true}, nil
}

func parseHTTPRoot(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return nil, errors.New("invalid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("only http:// and https:// URLs are allowed")
	}
	if err := rejectBlockedURL(u); err != nil {
		return nil, err
	}
	return u, nil
}

func rejectBlockedURL(u *url.URL) error {
	if u == nil || u.Host == "" {
		return errors.New("invalid URL")
	}
	if u.User != nil {
		return errors.New("URLs must not contain usernames or passwords")
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return errors.New("internal hosts are not allowed")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("could not resolve host: %w", err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return errors.New("internal addresses are not allowed")
		}
	}
	return nil
}

func (s *Service) resolveAbs(root, rel string) (string, error) {
	cleanRel, err := sanitizeRel(rel)
	if err != nil {
		return "", err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", errors.New("invalid root")
	}
	target := filepath.Join(absRoot, filepath.FromSlash(cleanRel))
	target, err = filepath.Abs(target)
	if err != nil {
		return "", errors.New("invalid path")
	}
	relBack, err := filepath.Rel(absRoot, target)
	if err != nil || relBack == ".." || strings.HasPrefix(relBack, ".."+string(os.PathSeparator)) {
		return "", errors.New("path traversal detected")
	}
	evalRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		evalRoot = absRoot
	}
	evalTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", errors.New("invalid path")
		}
		parent, perr := filepath.EvalSymlinks(filepath.Dir(target))
		if perr != nil {
			if !os.IsNotExist(perr) {
				return "", errors.New("invalid path")
			}
			parent = filepath.Dir(target)
		}
		evalTarget = filepath.Join(parent, filepath.Base(target))
	}
	relEval, err := filepath.Rel(evalRoot, evalTarget)
	if err != nil || relEval == ".." || strings.HasPrefix(relEval, ".."+string(os.PathSeparator)) {
		return "", errors.New("path traversal detected")
	}
	return target, nil
}
