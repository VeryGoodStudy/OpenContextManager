package file

import (
	"os"
	"sort"
	"testing"

	ocm "github.com/VeryGoodStudy/OpenContextManager"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestStore_SaveAndLoad(t *testing.T) {
	s := newTestStore(t)
	ctx := ocm.NewContext("test-1")
	ctx.AddMessage(ocm.NewMessage(ocm.RoleUser, "hello"))
	ctx.SetMetadata("model", "gpt-4")

	if err := s.Save(ctx); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := s.Load("test-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != "test-1" {
		t.Errorf("ID = %q, want %q", loaded.ID, "test-1")
	}
	if len(loaded.Messages) != 1 {
		t.Fatalf("Messages len = %d, want 1", len(loaded.Messages))
	}
	if loaded.Messages[0].Content != "hello" {
		t.Errorf("Content = %q, want %q", loaded.Messages[0].Content, "hello")
	}
	if loaded.Metadata["model"] != "gpt-4" {
		t.Errorf("Metadata[model] = %v, want gpt-4", loaded.Metadata["model"])
	}
}

func TestStore_LoadNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Load("nonexistent")
	if err != ocm.ErrContextNotFound {
		t.Errorf("Load error = %v, want ErrContextNotFound", err)
	}
}

func TestStore_Delete(t *testing.T) {
	s := newTestStore(t)
	ctx := ocm.NewContext("del-1")
	_ = s.Save(ctx)

	if err := s.Delete("del-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := s.Load("del-1")
	if err != ocm.ErrContextNotFound {
		t.Errorf("Load after delete: %v, want ErrContextNotFound", err)
	}
}

func TestStore_DeleteNotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.Delete("nonexistent")
	if err != ocm.ErrContextNotFound {
		t.Errorf("Delete error = %v, want ErrContextNotFound", err)
	}
}

func TestStore_List(t *testing.T) {
	s := newTestStore(t)
	_ = s.Save(ocm.NewContext("x"))
	_ = s.Save(ocm.NewContext("y"))
	_ = s.Save(ocm.NewContext("z"))

	ids, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	sort.Strings(ids)
	if len(ids) != 3 || ids[0] != "x" || ids[1] != "y" || ids[2] != "z" {
		t.Errorf("List = %v, want [x y z]", ids)
	}
}

func TestStore_FileCreated(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := ocm.NewContext("file-check")
	_ = s.Save(ctx)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	if entries[0].Name() != "file-check.ctx.json" {
		t.Errorf("filename = %q, want %q", entries[0].Name(), "file-check.ctx.json")
	}
}

// Verify that Store implements ocm.Storage.
var _ ocm.Storage = (*Store)(nil)
