# Draf 03 — Test untuk SessionService & UpdaterService

**Status:** Draf
**Prioritas:** 🟡 Sedang
**Effort:** Sedang

## Masalah

Dua jalur paling kritikal & berisiko justru tanpa test:

- **`SessionService`** (`internal/service/session.go`) — jalur data inti: spawn
  PTY, base64-encode output, lifecycle, mutex map `live`, shutdown. 0 test.
- **`UpdaterService`** (`internal/service/updater.go`) — download binary dari
  GitHub, **replace executable yang sedang berjalan**, rollback. 0 test.

Yang sudah ada hanya: `config/store_test.go`, `terminal/pty_test.go`
(Unix/Darwin saja), `service/workspace_test.go`.

## Tujuan

1. Test `SessionService` **tanpa PTY asli** — manfaatkan seam `Spawner` yang
   sudah ada.
2. Test `UpdaterService` **tanpa jaringan asli** — abstraksi HTTP client &
   filesystem.
3. Tidak menurunkan kemampuan: hanya menambah test + (bila perlu) seam injeksi
   minimal yang tidak mengubah perilaku.

## Bagian A — SessionService

### Seam yang sudah tersedia

`SessionService.spawn` bertipe `Spawner func(terminal.Options) (terminal.PTY, error)`
dan di-set di `NewSessionService` (`session.go:30-37`). `terminal.PTY` adalah
interface (`internal/infra/terminal/pty.go:14`). Jadi kita bisa inject **fake PTY**
+ **fake EventEmitter** tanpa proses nyata.

> `NewSessionService(emitter)` saat ini meng-hardcode `spawn: terminal.Start`.
> Untuk test, set `svc.spawn = fakeSpawn` langsung (field unexported, tapi test
> ada di package `service` yang sama). Tidak perlu ubah konstruktor.

### Fake yang dibutuhkan (`internal/service/session_test.go` baru)

```go
// fakePTY: PTY in-memory yang dikontrol test.
type fakePTY struct {
    readCh   chan []byte   // test push output ke sini
    written  bytes.Buffer  // tangkap WriteSession
    closed   chan struct{}
    waitCode int
    mu       sync.Mutex
}
// Read blok sampai ada data / closed; Write simpan ke buffer; Close tutup channel;
// Wait kembalikan error yang ExitCode()-nya = waitCode; Resize no-op.

// fakeEmitter: rekam (event, payload) untuk assertion.
type fakeEmitter struct {
    mu     sync.Mutex
    events []struct{ Name string; Data any }
}
```

### Kasus uji

| Test | Verifikasi |
|------|-----------|
| `StartSession` ID kosong | error `"session: id wajib diisi"` (`session.go:62`) |
| `StartSession` sukses | session masuk map; `IsRunning(id)==true`; `ListRunning` memuat id |
| `StartSession` dua kali ID sama | pemanggilan kedua no-op, tidak spawn ulang (`session.go:64-69`) |
| Spawner error | error dibungkus `"gagal spawn shell: ..."` (`session.go:78`) |
| `StartupCommand` | byte `cmd+"\n"` tertulis ke PTY (`session.go:88`) |
| readLoop → output | push byte ke `readCh` → emitter terima `session:<id>:out` berisi base64 yang benar (`session.go:152-153`) |
| readLoop → exit | tutup PTY → emit `session:<id>:exit` dgn `Code` sesuai `waitCode`; id terhapus dari map (`session.go:160-166`) |
| `WriteSession` ke id mati | error `"session: tidak berjalan"` (`session.go` WriteSession) |
| `ResizeSession` id mati / dimensi ≤0 | `nil`, tidak panik (`session.go:104-112`) |
| `ServiceShutdown` | semua PTY ter-`Close`; map kosong (`session.go:169-183`) |
| `buildEnv` | selalu set `TERM=xterm-256color` & `COLORTERM=truecolor`; override workspace menimpa tanpa duplikat key (`session.go` buildEnv/setEnv) |

> Catatan konkurensi: `readLoop` jalan di goroutine. Gunakan channel/`sync`
> untuk sinkronisasi assertion (mis. tunggu emit lewat channel pada fakeEmitter),
> jangan `time.Sleep`. Jalankan `go test -race ./internal/service/...`.

## Bagian B — UpdaterService

### Tantangan

`updater.go` melakukan HTTP ke GitHub dan `os.Rename` pada executable berjalan.
Keduanya perlu di-seam agar bisa dites secara hermetic.

### Refactor minimal yang diperlukan

Lihat `internal/service/updater.go` (`CheckUpdate`, `DownloadAndApply`,
rollback `updater.go:151-155`) lalu ekstrak dua seam:

1. **HTTP**: field `httpClient *http.Client` (default `http.DefaultClient`).
   Test set ke client yang menunjuk `httptest.Server`. Atau, lebih sempit,
   field `apiBaseURL string` agar bisa diarahkan ke server uji.
2. **Filesystem/target path**: parameter/field `targetPath string` (path biner
   yang akan ditimpa) alih-alih resolusi `os.Executable()` di dalam — sehingga
   test bisa pakai file temp.

Keduanya perubahan kecil yang **tidak mengubah perilaku produksi** (default
tetap sama), hanya membuka titik injeksi.

### Kasus uji (`internal/service/updater_test.go` baru)

| Test | Verifikasi |
|------|-----------|
| `CheckUpdate` versi lebih baru | parse JSON release `httptest` → kembalikan info update tersedia |
| `CheckUpdate` versi sama/older | "tidak ada update" |
| `CheckUpdate` HTTP 5xx / JSON rusak | error tertangani, tidak panik |
| `DownloadAndApply` sukses | file target berisi byte baru; backup dihapus/sesuai |
| `DownloadAndApply` download gagal | target **tidak berubah**, error dikembalikan |
| Rollback | target rusak di tengah → biner lama dipulihkan (`updater.go:151-155`) |

## Bagian C — Windows PTY (catatan, opsional)

`pty_windows.go` (ConPTY) tanpa test; `pty_test.go` Unix/Darwin saja
(`pty_test.go:1`). Unit test ConPTY butuh host Windows (CI matrix). Di luar
scope iterasi ini — catat sebagai follow-up bila CI Windows tersedia.

## Perintah

```bash
go test ./internal/service/...                 # paket service
go test -race ./internal/service/...           # wajib: readLoop goroutine
go test -run TestSessionStart ./internal/service
```

## Pertanyaan terbuka

1. Untuk Updater, pilih seam `apiBaseURL` (sempit) atau `httpClient` (fleksibel)?
   Rekomendasi: `httpClient` — lebih mudah simulasi error jaringan.
2. Seberapa jauh test rollback meniru "replace biner berjalan"? Cukup file temp
   biasa (proses test tidak benar-benar mengganti dirinya).
3. Perlu tambah CI Windows untuk ConPTY sekarang, atau tunda?

## Checklist

- [ ] `session_test.go`: fakePTY + fakeEmitter
- [ ] Cover tabel kasus SessionService (termasuk `buildEnv`/`setEnv`)
- [ ] Lulus `go test -race`
- [ ] Seam updater (`httpClient` + `targetPath`)
- [ ] `updater_test.go` dgn `httptest`
- [ ] Catat follow-up test ConPTY Windows
