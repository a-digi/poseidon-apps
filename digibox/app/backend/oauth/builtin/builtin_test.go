package builtin

import (
	"strings"
	"testing"
)

func TestLoad_EmbeddedFile(t *testing.T) {
	clients, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	for _, c := range clients {
		if !strings.HasPrefix(c.ID, "builtin:") {
			t.Errorf("client id %q must start with \"builtin:\"", c.ID)
		}
		if !c.Builtin {
			t.Errorf("client %q: expected Builtin=true", c.ID)
		}
	}
}

func TestLoad_ValidBuiltinPrefix(t *testing.T) {
	data := []byte(`[{"id":"builtin:foo","clientId":"cid","secret":"sec","name":"Foo","provider":"dropbox"}]`)
	clients, err := load(data)
	if err != nil {
		t.Fatalf("load() error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}
	if !clients[0].Builtin {
		t.Error("expected Builtin=true")
	}
	if clients[0].ID != "builtin:foo" {
		t.Errorf("unexpected ID: %s", clients[0].ID)
	}
}

func TestLoad_MissingBuiltinPrefix(t *testing.T) {
	data := []byte(`[{"id":"no-prefix","clientId":"cid","secret":"sec"}]`)
	_, err := load(data)
	if err == nil {
		t.Fatal("expected error for missing builtin: prefix")
	}
	if !strings.Contains(err.Error(), "builtin:") {
		t.Errorf("unexpected error: %v", err)
	}
}
