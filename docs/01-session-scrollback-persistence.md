# Draf 01 — Persistensi & Recovery Scrollback Sesi

**Status:** Draf
**Prioritas:** 🔴 Tinggi
**Effort:** Sedang–Besar

## Masalah

Saat ini hanya **definisi** session (`SessionDef`) yang dipersist. Proses PTY
dan seluruh scrollback bersifat ephemeral:

- `ServiceShutdown` mematikan semua PTY saat quit (`internal/service/session.go:182`).
- Scrollback hidup hanya di memori xterm.js, hardcoded 5000 baris
  (`frontend/src/presentation/components/Terminal.svelte:102`).
- Tidak ada jalur untuk menyimpan/memulihkan output.

Akibatnya: tutup app (atau crash) → seluruh log dev server / watcher / build
hilang. Untuk sebuah *Workspace Manager*, ini gap fungsional terbesar.

## Tujuan

1. Saat app dibuka kembali, setiap session menampilkan **scrollback terakhir**
   (read-only) sebelum proses di-restart.
2. Tidak mengubah model "session PTY tidak persisten lintas restart" — kita
   memulihkan **teks**, bukan proses hidup.
3. Hemat disk: ada batas ukuran & bisa dimatikan.

### Non-tujuan (untuk iterasi ini)

- Reattach ke proses PTY yang masih hidup (butuh daemon terpisah — di luar scope).
- Sinkronisasi scrollback real-time ke disk setiap byte (terlalu mahal).

## Desain

### Pendekatan: snapshot buffer saat shutdown + restore saat mount

Alih-alih streaming ke disk terus-menerus, kita ambil **snapshot** scrollback
pada dua titik: graceful shutdown dan interval throttled (untuk lindungi dari crash).

```
xterm.js buffer ──serialize──> SessionService.SaveSnapshot(id, data)
                                      │
                                      ▼
                         ~/.config/teer/sessions/<id>.scrollback   (0600, gzip)
                                      │
        Terminal.svelte mount ◀──RestoreSnapshot(id)──┘  (tulis ke term sebelum connectPty)
```

Serialisasi buffer pakai addon resmi **`@xterm/addon-serialize`** (belum
terpasang) — menghasilkan string berisi teks + escape sequence ANSI, sehingga
warna/format ikut terpulihkan.

### Kenapa snapshot, bukan tee setiap byte

`readLoop` (`session.go:145`) sudah memegang setiap chunk PTY. Tergoda untuk
nge-tee ke file di situ. Tapi:

- Itu menyimpan **byte mentah** tanpa tahu state terminal (clear, alt-screen,
  resize) → replay bisa berantakan.
- `addon-serialize` menserialisasi **state akhir** buffer xterm secara benar
  (sudah memproses semua escape). Hasil restore jauh lebih akurat.

Jadi sumber kebenaran snapshot = sisi frontend (xterm), bukan byte PTY.

## Perubahan Backend (Go)

### 1. Storage scrollback baru — `internal/infra/config/scrollback.go` (baru)

Pisahkan dari `store.go` (yang khusus `config.json`). Pola atomic-write +
`0600` ikuti `store.go:57-71`.

```go
// ScrollbackStore menyimpan snapshot scrollback per-session sebagai file
// gzip di ~/.config/teer/sessions/<id>.scrollback. Snapshot bersifat
// best-effort: kegagalan I/O tidak boleh menggagalkan operasi utama.
type ScrollbackStore struct {
    dir     string // ~/.config/teer/sessions
    maxSize int    // batas byte per snapshot (mis. 1<<20 = 1 MiB)
}

func (s *ScrollbackStore) Save(id, data string) error  // gzip, truncate jika > maxSize, atomic tmp+rename
func (s *ScrollbackStore) Load(id string) (string, error) // "" + nil jika tidak ada
func (s *ScrollbackStore) Delete(id string) error
func (s *ScrollbackStore) Prune(validIDs []string) error // hapus snapshot yatim
```

### 2. Method baru di SessionService — `internal/service/session.go`

```go
func (s *SessionService) SaveScrollback(id string, data string) error
func (s *SessionService) LoadScrollback(id string) (string, error)
```

- Inject `*config.ScrollbackStore` via `NewSessionService` (`session.go:30`),
  selaras pola dependency injection yang sudah ada.
- `SaveScrollback` truncate diam-diam bila melebihi `maxSize` (jangan error).
- Panggil `Prune` di `ServiceShutdown` (`session.go:182`) memakai daftar ID
  session valid dari config — cegah file yatim menumpuk.

> Setelah ubah signature service: jalankan `wails3 generate bindings`
> (otomatis saat `task build`/`task dev`). Jangan edit `frontend/bindings/` manual.

### 3. main.go

Bangun `ScrollbackStore` di samping `config.Store` lalu suntikkan ke
`NewSessionService` (lihat `main.go` bagian wiring service).

## Perubahan Frontend

### 4. Infra wrapper — `frontend/src/infrastructure/wails/session-gateway.ts`

Tambah `saveScrollback(id, data)` dan `loadScrollback(id)` membungkus binding
baru. Re-export lewat `infrastructure/wails/index.ts`.

### 5. Use-case — `frontend/src/application/session.ts`

Tambah `persistScrollback(id, data)` & `restoreScrollback(id)`; barrel di `@application`.

### 6. Terminal.svelte

Dependency: `npm i @xterm/addon-serialize`.

- Muat `SerializeAddon` bersama addon lain (`Terminal.svelte:106-110`).
- **Restore**: di `$effect` mount, **sebelum** `connectPty(_term)`
  (`Terminal.svelte:149`), `await restoreScrollback(sid)`; jika ada, tulis ke
  terminal lalu cetak garis pemisah, mis.
  `\r\n\x1b[90m─── sesi sebelumnya dipulihkan ───\x1b[0m\r\n`, baru PTY start.
- **Snapshot saat unmount**: di cleanup `$effect` (`Terminal.svelte:151`),
  panggil `persistScrollback(sid, serialize.serialize())` **sebelum** `_term.dispose()`.
- **Snapshot periodik (anti-crash)**: `setInterval` throttled (mis. 10 dtk)
  hanya saat session running; clear di cleanup.

## Konfigurasi & Settings

Tambah toggle di terminal settings (lihat `frontend/src/domain/terminalSettings.ts`):

- `persistScrollback: boolean` (default `true`)
- `scrollbackLines: number` (default `5000`) — sekalian gantikan angka hardcoded
  di `Terminal.svelte:102` dengan nilai dari store. Ini menutup gap "scrollback
  tidak bisa dikonfigurasi" sekaligus.

## Rencana Test

Backend (`internal/infra/config/scrollback_test.go` baru), pola `store_test.go`:

- Save→Load round-trip identik (setelah gunzip).
- Snapshot > `maxSize` → ter-truncate, tidak error.
- Load ID tidak ada → `("", nil)`.
- `Prune` menghapus hanya file yatim, menyisakan ID valid.
- Permission file `0600`.

Frontend: minimal manual — buka app, jalankan `ls -la --color`, tutup, buka
lagi → output + warna muncul kembali sebelum prompt baru.

## Pertanyaan terbuka

1. Batas ukuran per snapshot — 1 MiB cukup? Atau ikut `scrollbackLines`?
2. Snapshot periodik 10 dtk: terlalu sering untuk session yang sangat verbose?
   Pertimbangkan throttle berbasis "hanya jika ada perubahan".
3. Apakah perlu indikator visual bahwa bagian atas adalah scrollback lama
   (read-only) vs. output sesi baru? (Garis pemisah sudah cukup?)

## Checklist

- [ ] `scrollback.go` + test
- [ ] Method `SaveScrollback`/`LoadScrollback` + inject store
- [ ] `Prune` di `ServiceShutdown`
- [ ] Wiring `main.go`
- [ ] Regenerate bindings
- [ ] Wrapper gateway + use-case + barrel
- [ ] `addon-serialize` di Terminal.svelte (restore + snapshot + interval)
- [ ] Settings `persistScrollback` + `scrollbackLines` (ganti hardcode 5000)
- [ ] Uji manual round-trip restart
