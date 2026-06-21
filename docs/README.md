# Draf Implementasi Teer

Kumpulan draf desain untuk peningkatan Teer hasil audit fitur. Setiap draf
bersifat **self-contained**: berisi konteks, desain, perubahan file konkret
(dengan referensi `file:line`), rencana test, dan checklist.

Urutan prioritas (berdasarkan dampak ke pengguna vs. risiko):

| # | Draf | Dampak | Risiko | Effort |
|---|------|--------|--------|--------|
| 01 | [Persistensi & recovery scrollback sesi](./01-session-scrollback-persistence.md) | 🔴 Tinggi | Sedang | M–L |
| 02 | [Copy / paste terminal yang proper](./02-copy-paste.md) | 🟡 Sedang | Rendah | S |
| 03 | [Test untuk SessionService & UpdaterService](./03-testing-session-updater.md) | 🟡 Sedang | Rendah | M |

## Konvensi yang dipakai semua draf

- Komentar & pesan error Go dalam **Bahasa Indonesia** (mis. `errors.New("session: id wajib diisi")`) — ikuti gaya kode existing.
- Setiap fitur backend yang dipanggil UI menempuh alur: tambah method di `internal/service/` → `wails3 generate bindings` → bungkus di `frontend/src/infrastructure/wails/` → pakai di `application/`. UI tidak pernah `import @bindings` langsung.
- File platform-specific pakai build tag (`pty_unix.go` / `pty_windows.go`); kontrak bersama tetap di `pty.go`.
- `WorkspaceService` satu-satunya pemilik persistensi config: `repo.Load()` → ubah → `repo.Save()`.

## Status

Semua dokumen di sini adalah **draf** — belum disetujui untuk implementasi.
Tandai keputusan terbuka di bagian "Pertanyaan terbuka" tiap draf sebelum mulai koding.
