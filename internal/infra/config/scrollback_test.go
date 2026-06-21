package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
)

func newTestScrollbackStore(t *testing.T) *ScrollbackStore {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	xdg.Reload()

	store, err := NewScrollbackStore()
	if err != nil {
		t.Fatalf("NewScrollbackStore: %v", err)
	}
	return store
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

	// Banyak baris pendek; total jauh melebihi maxSize.
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
		t.Fatalf("hasil truncate tidak boleh kosong")
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
		got, _ := store.Load(id)
		if got == "" {
			t.Fatalf("snapshot valid %s seharusnya tidak terhapus", id)
		}
	}
	for _, id := range []string{"yatim-1", "yatim-2"} {
		if _, err := os.Stat(store.pathFor(id)); !os.IsNotExist(err) {
			t.Fatalf("snapshot yatim %s seharusnya terhapus", id)
		}
	}
}

func TestScrollbackPermission(t *testing.T) {
	store := newTestScrollbackStore(t)

	if err := store.Save("perm", "data"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(store.pathFor("perm"))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("permission file = %o, mau 0600", perm)
	}
}

func TestScrollbackUnsafeID(t *testing.T) {
	store := newTestScrollbackStore(t)

	// id path-traversal harus diabaikan diam-diam, tidak menulis di luar dir.
	if err := store.Save("../evil", "jahat"); err != nil {
		t.Fatalf("Save id tidak aman tidak boleh error: %v", err)
	}

	outside := filepath.Join(filepath.Dir(store.dir), "evil"+scrollbackExt)
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("file ditulis di luar direktori sessions: %s", outside)
	}

	got, err := store.Load("../evil")
	if err != nil {
		t.Fatalf("Load id tidak aman: %v", err)
	}
	if got != "" {
		t.Fatalf("Load id tidak aman harus \"\", dapat %q", got)
	}
}
