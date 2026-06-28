package sqlite

import (
	"strings"
	"testing"
)

func newTestScrollbackStore(t *testing.T) *ScrollbackStore {
	t.Helper()
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewScrollbackStore(db)
}

func TestScrollbackRoundTrip(t *testing.T) {
	store := newTestScrollbackStore(t)

	want := "baris satu\r\n\x1b[31mmerah\x1b[0m\r\nbaris tiga\r\n"
	if err := store.Save("sesi-1", want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Load("sesi-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Fatalf("round-trip tidak identik:\n got=%q\nwant=%q", got, want)
	}
}

func TestScrollbackTruncate(t *testing.T) {
	store := newTestScrollbackStore(t)
	store.maxSize = 64

	var b strings.Builder
	for range 1000 {
		b.WriteString("0123456789\n")
	}
	if err := store.Save("besar", b.String()); err != nil {
		t.Fatalf("Save besar tidak boleh error: %v", err)
	}

	got, err := store.Load("besar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) > store.maxSize {
		t.Fatalf("hasil truncate %d byte melebihi maxSize %d", len(got), store.maxSize)
	}
	if got == "" {
		t.Fatal("hasil truncate tidak boleh kosong")
	}
}

func TestScrollbackLoadMissing(t *testing.T) {
	store := newTestScrollbackStore(t)

	got, err := store.Load("tidak-ada")
	if err != nil {
		t.Fatalf("Load ID tidak ada harus nil error, dapat: %v", err)
	}
	if got != "" {
		t.Fatalf("Load ID tidak ada harus \"\", dapat %q", got)
	}
}

func TestScrollbackDelete(t *testing.T) {
	store := newTestScrollbackStore(t)

	if err := store.Save("hapus", "data"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := store.Delete("hapus"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, err := store.Load("hapus")
	if err != nil {
		t.Fatalf("Load setelah Delete: %v", err)
	}
	if got != "" {
		t.Fatalf("data seharusnya hilang setelah Delete, dapat %q", got)
	}
}

func TestScrollbackPrune(t *testing.T) {
	store := newTestScrollbackStore(t)

	for _, id := range []string{"valid-1", "valid-2", "yatim-1", "yatim-2"} {
		if err := store.Save(id, "data "+id); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}

	if err := store.Prune([]string{"valid-1", "valid-2"}); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	for _, id := range []string{"valid-1", "valid-2"} {
		got, err := store.Load(id)
		if err != nil {
			t.Fatalf("Load %s: %v", id, err)
		}
		if got == "" {
			t.Fatalf("snapshot valid %s seharusnya tidak terhapus", id)
		}
	}
	for _, id := range []string{"yatim-1", "yatim-2"} {
		got, err := store.Load(id)
		if err != nil {
			t.Fatalf("Load %s: %v", id, err)
		}
		if got != "" {
			t.Fatalf("snapshot yatim %s seharusnya terhapus", id)
		}
	}
}

func TestScrollbackUnsafeID(t *testing.T) {
	store := newTestScrollbackStore(t)

	if err := store.Save("../evil", "jahat"); err != nil {
		t.Fatalf("Save id tidak aman tidak boleh error: %v", err)
	}

	got, err := store.Load("../evil")
	if err != nil {
		t.Fatalf("Load id tidak aman: %v", err)
	}
	if got != "" {
		t.Fatalf("Load id tidak aman harus \"\", dapat %q", got)
	}
}

func TestScrollbackPruneEmpty(t *testing.T) {
	store := newTestScrollbackStore(t)

	if err := store.Save("s1", "data"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Prune dengan validIDs kosong harus menghapus semua.
	if err := store.Prune([]string{}); err != nil {
		t.Fatalf("Prune empty: %v", err)
	}

	got, _ := store.Load("s1")
	if got != "" {
		t.Fatal("semua scrollback harus terhapus saat Prune dengan validIDs kosong")
	}
}
