package contentfactory

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type DraftDoc struct {
	OK      bool            `json:"ok"`
	Error   string          `json:"error,omitempty"`
	RelPath string          `json:"relPath"`
	Body    json.RawMessage `json:"body,omitempty"`
}

func (s *Service) draftAbs() string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(draftRel))
}

func (s *Service) LoadDraft() DraftDoc {
	out := DraftDoc{RelPath: draftRel}
	if s.questsDirPath() == "" {
		out.Error = "quests directory is not set"
		return out
	}
	raw, err := os.ReadFile(s.draftAbs())
	if err != nil {
		if os.IsNotExist(err) {
			out.OK = true
			out.Body = json.RawMessage("{}")
			return out
		}
		out.Error = err.Error()
		return out
	}
	if !json.Valid(raw) {
		out.Error = "draft JSON is invalid"
		return out
	}
	out.Body = json.RawMessage(raw)
	out.OK = true
	return out
}

func (s *Service) SaveDraft(body json.RawMessage) DraftDoc {
	out := DraftDoc{RelPath: draftRel}
	if s.questsDirPath() == "" {
		out.Error = "quests directory is not set"
		return out
	}
	if len(body) == 0 {
		body = json.RawMessage("{}")
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		out.Error = "draft JSON is invalid"
		return out
	}
	if err := writeJSON(s.draftAbs(), v); err != nil {
		out.Error = err.Error()
		return out
	}
	out.Body = body
	out.OK = true
	return out
}
