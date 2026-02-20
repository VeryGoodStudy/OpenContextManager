package memory

import (
	"sort"
	"testing"

	ocm "github.com/VeryGoodStudy/OpenContextManager"
)

func TestStore_SaveAndLoad(t *testing.T) {
	s := New()
	ctx := ocm.NewContext("test-1")
	ctx.AddMessage(ocm.NewMessage(ocm.RoleUser, "hello"))

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
}

func TestStore_LoadNotFound(t *testing.T) {
	s := New()
	_, err := s.Load("nonexistent")
	if err != ocm.ErrContextNotFound {
		t.Errorf("Load error = %v, want ErrContextNotFound", err)
	}
}

func TestStore_Delete(t *testing.T) {
	s := New()
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
	s := New()
	err := s.Delete("nonexistent")
	if err != ocm.ErrContextNotFound {
		t.Errorf("Delete error = %v, want ErrContextNotFound", err)
	}
}

func TestStore_List(t *testing.T) {
	s := New()
	_ = s.Save(ocm.NewContext("a"))
	_ = s.Save(ocm.NewContext("b"))
	_ = s.Save(ocm.NewContext("c"))

	ids, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	sort.Strings(ids)
	if len(ids) != 3 || ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
		t.Errorf("List = %v, want [a b c]", ids)
	}
}

// Verify that Store implements ocm.Storage.
var _ ocm.Storage = (*Store)(nil)
