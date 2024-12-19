package discovery

import (
	"encoding/json"
	"os"
	"testing"
)

func TestConfiguration(t *testing.T) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	_ = enc.Encode(NewConfig(&Required{}, nil))
}
