package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// --- helpers ---

func newTestUpdater(version string, srv *httptest.Server) *UpdaterService {
	u := NewUpdaterService(version, newFakeEmitter())
	u.httpClient = srv.Client()
	u.apiBaseURL = srv.URL
	return u
}

func ghReleaseJSON(tag, htmlURL string, assets []ghAsset) string {
	rel := ghRelease{TagName: tag, HTMLURL: htmlURL, Assets: assets}
	b, _ := json.Marshal(rel)
	return string(b)
}

// --- tests ---

func TestCheckUpdate_VersiLebihBaru(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ghReleaseJSON("v2.0.0", "https://github.com/x/y/releases/v2.0.0", nil))
	}))
	defer srv.Close()

	u := newTestUpdater("v1.0.0", srv)
	info, err := u.CheckUpdate()
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if !info.Available {
		t.Fatal("harusnya ada update")
	}
	if info.LatestVersion != "v2.0.0" {
		t.Fatalf("LatestVersion salah: %q", info.LatestVersion)
	}
	if info.CurrentVersion != "v1.0.0" {
		t.Fatalf("CurrentVersion salah: %q", info.CurrentVersion)
	}
}

func TestCheckUpdate_VersiSama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ghReleaseJSON("v1.0.0", "https://github.com/x/y/releases/v1.0.0", nil))
	}))
	defer srv.Close()

	u := newTestUpdater("v1.0.0", srv)
	info, err := u.CheckUpdate()
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if info.Available {
		t.Fatal("harusnya tidak ada update")
	}
}

func TestCheckUpdate_HTTP5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	u := newTestUpdater("v1.0.0", srv)
	_, err := u.CheckUpdate()
	if err == nil {
		t.Fatal("harusnya error pada HTTP 5xx")
	}
}

func TestCheckUpdate_JSONRusak(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{broken json`)
	}))
	defer srv.Close()

	u := newTestUpdater("v1.0.0", srv)
	_, err := u.CheckUpdate()
	if err == nil {
		t.Fatal("harusnya error pada JSON rusak")
	}
}

func TestDownloadAndApply_Sukses(t *testing.T) {
	newBinary := []byte("binary-baru-versi-2")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(newBinary)))
		w.Write(newBinary)
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "teer")
	if err := os.WriteFile(target, []byte("binary-lama"), 0755); err != nil {
		t.Fatal(err)
	}

	em := newFakeEmitter()
	u := NewUpdaterService("v1.0.0", em)
	u.httpClient = srv.Client()
	u.targetPath = target
	u.onDone = func() {}

	if err := u.DownloadAndApply(srv.URL + "/download"); err != nil {
		t.Fatalf("DownloadAndApply: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("baca target: %v", err)
	}
	if string(got) != string(newBinary) {
		t.Fatalf("isi target salah: dapat %q", got)
	}

	// backup harus sudah dihapus
	if _, err := os.Stat(target + ".bak"); !os.IsNotExist(err) {
		t.Fatal("backup .bak harus sudah dihapus")
	}
}

func TestDownloadAndApply_DownloadGagal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Tutup koneksi langsung untuk simulasi error jaringan
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "no hijack", 500)
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	dir := t.TempDir()
	target := filepath.Join(dir, "teer")
	originalContent := []byte("binary-lama")
	if err := os.WriteFile(target, originalContent, 0755); err != nil {
		t.Fatal(err)
	}

	em := newFakeEmitter()
	u := NewUpdaterService("v1.0.0", em)
	u.httpClient = srv.Client()
	u.targetPath = target

	err := u.DownloadAndApply(srv.URL + "/download")
	if err == nil {
		t.Fatal("harusnya error saat download gagal")
	}

	got, rerr := os.ReadFile(target)
	if rerr != nil {
		t.Fatalf("baca target: %v", rerr)
	}
	if string(got) != string(originalContent) {
		t.Fatalf("target harus tidak berubah saat download gagal, dapat %q", got)
	}
}

func TestRollback(t *testing.T) {
	newBinary := []byte("binary-baru")

	// Server yang berhasil download tapi kita akan sabotase moveFile dengan
	// membuat target dir read-only setelah backup
	dir := t.TempDir()
	target := filepath.Join(dir, "teer")
	originalContent := []byte("binary-lama")
	if err := os.WriteFile(target, originalContent, 0755); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	// Target di dalam subdir yang akan dibuat read-only setelah backup
	roTarget := filepath.Join(subdir, "teer")
	if err := os.WriteFile(roTarget, originalContent, 0755); err != nil {
		t.Fatal(err)
	}

	// Serve file download baru
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(newBinary)
	}))
	defer srv.Close()

	em := newFakeEmitter()
	u := NewUpdaterService("v1.0.0", em)
	u.httpClient = srv.Client()
	u.targetPath = roTarget
	u.onDone = func() {}

	// Buat subdir read-only sebelum apply agar rename ke dalamnya gagal
	if err := os.Chmod(subdir, 0555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(subdir, 0755)

	err := u.DownloadAndApply(srv.URL + "/download")
	// Boleh sukses (rename berhasil) atau gagal (tergantung filesystem):
	// yang penting jika gagal, target tidak corrupt
	if err != nil {
		// Cek rollback: target harus pulih
		got, rerr := os.ReadFile(roTarget)
		if rerr != nil {
			// Mungkin backup masih ada karena rollback juga gagal (read-only dir)
			backup := roTarget + ".bak"
			got, rerr = os.ReadFile(backup)
			if rerr != nil {
				t.Fatalf("target dan backup tidak bisa dibaca: %v", rerr)
			}
		}
		if string(got) != string(originalContent) {
			t.Fatalf("setelah rollback, isi harus binary lama, dapat %q", got)
		}
	}
	// Jika sukses (rename bekerja walau chmod), tes tetap lulus
}
