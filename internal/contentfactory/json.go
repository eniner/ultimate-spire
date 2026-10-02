package contentfactory

import (
	"bytes"
	"encoding/json"
)

func marshalPretty(v interface{}) ([]byte, error) {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(bytes.TrimSpace(body), '\n'), nil
}
