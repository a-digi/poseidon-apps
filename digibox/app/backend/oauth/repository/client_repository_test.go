package repository

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"digibox-plugin/oauth/model"
)

func newTestRepo(t *testing.T, builtins []model.OAuthClient) *OauthClientsRepository {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo, err := NewOauthClientsRepository(db, builtins)
	if err != nil {
		t.Fatalf("NewOauthClientsRepository: %v", err)
	}
	return repo
}

func TestFindAll_DBFirstThenBuiltins(t *testing.T) {
	builtins := []model.OAuthClient{
		{ID: "builtin:foo", ClientId: "cid-builtin", Secret: "real-secret", Name: "Builtin Foo", Provider: "dropbox", Builtin: true},
	}
	repo := newTestRepo(t, builtins)

	dbClient := model.OAuthClient{
		ClientId: "db-client-id", Secret: "abcdefgh", Name: "DB Client", Provider: "dropbox",
	}
	inserted, err := repo.Insert(dbClient)
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 results, got %d", len(all))
	}

	db := all[0]
	if db.ID != inserted.ID {
		t.Errorf("first result should be DB row, got ID=%s", db.ID)
	}
	if db.Builtin {
		t.Error("DB row should have Builtin=false")
	}
	if db.Secret == "abcdefgh" {
		t.Error("DB secret should be redacted")
	}

	b := all[1]
	if !b.Builtin {
		t.Error("builtin row should have Builtin=true")
	}
	if b.ClientId != "" {
		t.Errorf("builtin ClientId should be empty in FindAll, got %q", b.ClientId)
	}
	if b.Secret != "" {
		t.Errorf("builtin Secret should be empty in FindAll, got %q", b.Secret)
	}
}

func TestFindByID_Builtin(t *testing.T) {
	builtins := []model.OAuthClient{
		{ID: "builtin:foo", ClientId: "real-cid", Secret: "real-secret", Name: "Foo", Provider: "dropbox", Builtin: true},
	}
	repo := newTestRepo(t, builtins)

	c, err := repo.FindByID("builtin:foo")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if c.ClientId != "real-cid" {
		t.Errorf("expected real ClientId, got %q", c.ClientId)
	}
	if c.Secret != "real-secret" {
		t.Errorf("expected real Secret, got %q", c.Secret)
	}
}

func TestDelete_BuiltinReturnsError(t *testing.T) {
	builtins := []model.OAuthClient{
		{ID: "builtin:foo", ClientId: "cid", Secret: "sec", Name: "Foo", Provider: "dropbox", Builtin: true},
	}
	repo := newTestRepo(t, builtins)

	err := repo.Delete("builtin:foo")
	if err == nil {
		t.Fatal("expected error when deleting builtin client")
	}

	// Verify the builtin is still accessible (no DB row was touched).
	c, err := repo.FindByID("builtin:foo")
	if err != nil || c == nil {
		t.Errorf("builtin should still exist after failed delete: %v", err)
	}
}

func TestFindByClientId_DBWinsOnCollision(t *testing.T) {
	sharedClientId := "shared-cid"
	builtins := []model.OAuthClient{
		{ID: "builtin:bar", ClientId: sharedClientId, Secret: "builtin-secret", Name: "Builtin", Provider: "dropbox", Builtin: true},
	}
	repo := newTestRepo(t, builtins)

	dbClient := model.OAuthClient{
		ClientId: sharedClientId, Secret: "db-secret", Name: "DB", Provider: "dropbox",
	}
	if _, err := repo.Insert(dbClient); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	c, err := repo.FindByClientId(sharedClientId)
	if err != nil {
		t.Fatalf("FindByClientId: %v", err)
	}
	if c.Builtin {
		t.Error("expected DB row to win, got builtin")
	}
	if c.Secret != "db-secret" {
		t.Errorf("expected DB secret, got %q", c.Secret)
	}
}
