# PRD — teer

**Product Requirements Document**

| Field | Value |
|-------|-------|
| Produk | **teer** — Terminal Workspace Manager |
| Versi dokumen | 1.0 (draft) |
| Tanggal | 2026-06-16 |
| Status | Draft |
| Tech stack | Wails v3 (Go) + Svelte + TypeScript + xterm.js |
| Platform target | Linux (utama), macOS & Windows (fase lanjutan) |

---

## 1. Ringkasan (Overview)

**teer** adalah aplikasi desktop untuk developer yang berfungsi sebagai *manajemen terminal CLI*. Pengguna dapat membuat beberapa **workspace**, dan di dalam tiap workspace menjalankan beberapa **terminal/sesi CLI** sekaligus. Tujuannya: menyatukan banyak sesi terminal yang tersebar (tab terminal OS, tmux, jendela berbeda-beda) ke dalam satu aplikasi yang terorganisir per-proyek.

Analoginya: gabungan antara *terminal emulator* (seperti Tabby/Wave) dengan konsep *workspace* (seperti VS Code workspace), khusus untuk mengelola proses CLI yang sedang berjalan.

---

## 2. Latar Belakang & Masalah

Developer modern menjalankan banyak proses CLI bersamaan: dev server, watcher, log tailer, database client, SSH session, build tools, dll. Masalah yang muncul:

- **Tersebar** — sesi terminal berserakan di banyak tab/jendela tanpa konteks proyek.
- **Hilang konteks** — saat pindah proyek, harus buka ulang & jalankan ulang semua perintah dari nol.
- **Tidak ada state** — terminal OS tidak ingat "proyek A butuh 4 terminal: server, worker, db, logs".
- **Sulit dimonitor** — tidak ada gambaran cepat mana proses yang masih hidup/mati.

**teer** menyelesaikan ini dengan model **workspace berisi sesi terminal yang persisten dan terdefinisi**.

---

## 3. Tujuan & Bukan Tujuan

### 3.1 Tujuan (Goals)
- G1 — Pengguna dapat membuat, menamai, dan mengelola banyak workspace.
- G2 — Di tiap workspace, pengguna dapat menjalankan banyak terminal CLI penuh (interaktif, mendukung TUI seperti vim/htop).
- G3 — Layout & definisi sesi terminal tersimpan (persisten) antar restart aplikasi.
- G4 — Perpindahan antar workspace & terminal cepat (keyboard-driven).
- G5 — Performa ringan (memori & CPU rendah) karena dipakai seharian.

### 3.2 Bukan Tujuan (Non-Goals) — untuk v1
- Bukan pengganti tmux/screen di sisi remote server (multiplexing terjadi di sisi aplikasi, bukan di host).
- Bukan IDE / code editor.
- Belum ada sinkronisasi cloud / multi-device (dipertimbangkan untuk masa depan).
- Belum ada kolaborasi real-time / shared session.
- Belum ada plugin marketplace.

---

## 4. Target Pengguna (Personas)

| Persona | Kebutuhan utama |
|---------|-----------------|
| **Backend/Fullstack Developer** | Menjalankan beberapa service + DB + log per proyek, ingin satu klik untuk "bangkitkan semua". |
| **DevOps / SRE** | Banyak sesi SSH ke server berbeda, dikelompokkan per environment/cluster. |
| **Data/ML Engineer** | Notebook server, training job, monitoring GPU dalam satu workspace. |

---

## 5. Konsep Inti & Model Data

### 5.1 Hierarki

```
Application
└── Workspace (mis. "Proyek Codemi", "Server Produksi")
    ├── metadata: nama, ikon/warna, working directory default, env vars
    └── Session[]  (terminal CLI)
        ├── metadata: nama, command awal, cwd, env, shell
        ├── layout: posisi tab/split pane
        └── runtime: PTY, status (running/exited), PID, riwayat output
```

### 5.2 Entitas

**Workspace**
| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | string (uuid) | Identifier unik |
| name | string | Nama tampilan |
| color | string | Warna/label untuk identifikasi cepat |
| defaultCwd | string | Direktori kerja default untuk sesi baru |
| env | map[string]string | Environment variable level workspace |
| sessions | Session[] | Daftar sesi |
| createdAt / updatedAt | timestamp | Audit |

**Session**
| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | string (uuid) | Identifier unik |
| workspaceId | string | Relasi ke workspace |
| name | string | Nama tampilan tab |
| shell | string | Path shell (mis. /bin/bash, /bin/zsh) |
| startupCommand | string | Perintah dijalankan saat sesi dibuka (opsional) |
| cwd | string | Direktori kerja |
| env | map[string]string | Env tambahan (override workspace) |
| autoStart | bool | Jalankan otomatis saat workspace dibuka |
| layout | object | Info posisi (tab index / split arrangement) |

**Status runtime** (tidak dipersist): `idle | running | exited(code)`.

### 5.3 Penyimpanan
- Konfigurasi disimpan sebagai file lokal (mis. JSON/SQLite) di direktori konfig OS (`~/.config/teer/` di Linux via `adrg/xdg`).
- **Output/scrollback terminal tidak dipersist** di v1 (hanya sesi & definisinya). Persistensi scrollback masuk backlog.

---

## 6. Kebutuhan Fungsional (Functional Requirements)

Prioritas: **P0** = wajib v1 (MVP), **P1** = penting, **P2** = nice-to-have.

### 6.1 Manajemen Workspace
- **FR-1 (P0)** — Buat workspace baru (nama, warna, cwd default).
- **FR-2 (P0)** — Lihat daftar workspace di sidebar.
- **FR-3 (P0)** — Pindah/aktifkan workspace.
- **FR-4 (P0)** — Edit & hapus workspace (dengan konfirmasi).
- **FR-5 (P1)** — Duplikat workspace (beserta definisi sesinya).
- **FR-6 (P2)** — Urutkan / drag-reorder workspace.

### 6.2 Manajemen Terminal/Sesi
- **FR-7 (P0)** — Buka terminal baru di workspace aktif.
- **FR-8 (P0)** — Terminal interaktif penuh: mendukung input keyboard, warna ANSI, dan aplikasi TUI (vim, htop, top).
- **FR-9 (P0)** — Beberapa terminal dalam satu workspace, ditampilkan sebagai **tab**.
- **FR-10 (P0)** — Tutup terminal (kill proses + konfirmasi bila masih running).
- **FR-11 (P0)** — Rename terminal.
- **FR-12 (P1)** — **Split pane** (horizontal/vertical) untuk lihat >1 terminal bersamaan.
- **FR-13 (P1)** — `startupCommand` & `autoStart`: workspace bisa otomatis membangkitkan set sesinya.
- **FR-14 (P1)** — Indikator status sesi (running/exited) di tab.
- **FR-15 (P2)** — Restart sesi yang sudah exited (re-run command).
- **FR-16 (P2)** — Broadcast input ke beberapa terminal sekaligus.

### 6.3 Interaksi Terminal
- **FR-17 (P0)** — Resize PTY mengikuti ukuran window/pane (responsive).
- **FR-18 (P0)** — Copy/paste.
- **FR-19 (P1)** — Cari teks di scrollback.
- **FR-20 (P1)** — Atur font family & size, tema warna terminal.
- **FR-21 (P2)** — Clickable links (URL) di output.

### 6.4 Navigasi & UX
- **FR-22 (P0)** — Persistensi: saat aplikasi dibuka ulang, semua workspace & definisi sesi kembali.
- **FR-23 (P1)** — Keyboard shortcut: pindah workspace, pindah/buka/tutup tab terminal.
- **FR-24 (P1)** — Command palette (cari & jalankan aksi cepat).
- **FR-25 (P2)** — Multi-window: buka workspace di window OS terpisah (memanfaatkan multi-window native Wails v3).

---

## 7. Kebutuhan Non-Fungsional

- **NFR-1 Performa** — Idle RAM < 150 MB; mendukung ≥ 20 terminal aktif tanpa lag pada mesin 16 GB.
- **NFR-2 Latency** — Input keyboard → tampil di terminal < 30 ms (terasa instan).
- **NFR-3 Footprint** — Ukuran binary terdistribusi kecil (manfaat Wails: pakai webview OS, bukan bundle Chromium).
- **NFR-4 Stabilitas** — Crash satu terminal tidak menjatuhkan aplikasi (isolasi per-PTY via goroutine).
- **NFR-5 Keamanan** — File konfig disimpan dengan permission ketat (**0600**). Env (termasuk yang sensitif) disimpan plaintext untuk v1 — keputusan sadar, lihat §13.4. Integrasi OS keyring dipertimbangkan pasca-v1.
- **NFR-6 Cross-platform readiness** — Arsitektur PTY diabstraksi agar Windows (ConPTY) bisa ditambahkan tanpa rombak besar.

---

## 8. Arsitektur Teknis

### 8.1 Tech Stack
| Lapisan | Teknologi |
|---------|-----------|
| Desktop framework | **Wails v3** (alpha) |
| Backend | **Go 1.25+** |
| PTY / shell spawning | **`creack/pty`** (Linux/macOS), ConPTY untuk Windows (fase lanjutan) |
| Frontend framework | **Svelte + TypeScript** |
| Terminal renderer | **xterm.js** + addon: `@xterm/addon-fit`, `@xterm/addon-search`, `@xterm/addon-web-links` |
| Build tool | Vite |
| Penyimpanan | File JSON atau SQLite di `~/.config/teer/` |
| Komunikasi FE↔BE | Wails bindings (RPC) + Events (streaming I/O terminal) |

### 8.2 Alur Data Terminal (I/O streaming)

```
[xterm.js di Svelte]  --(keystroke)-->  Wails binding: WriteToSession(id, data)
        ^                                          |
        |                                          v
   Events.On(output)  <--(stream)--  Go: baca PTY  -->  shell proses (bash/zsh)
```

- Tiap sesi = 1 goroutine yang membaca dari PTY dan **emit event** (`session:<id>:output`) ke frontend.
- Frontend menulis input via binding `Session.Write(id, data)`.
- Resize via binding `Session.Resize(id, cols, rows)`.

### 8.3 Service Go (rancangan)

```
SessionService
  - Create(workspaceId, opts) -> sessionId   // spawn PTY
  - Write(sessionId, data)                    // stdin
  - Resize(sessionId, cols, rows)
  - Close(sessionId)                          // kill PTY
  - List(workspaceId) -> []SessionInfo

WorkspaceService
  - Create / Update / Delete / List
  - Persist / Load                            // ke file konfig
```

### 8.4 Struktur Proyek (saat ini)

```
teer/
├── main.go                 # entrypoint Wails, registrasi service & window
├── go.mod                  # module: teer
├── *service.go             # service Go (akan: sessionservice.go, workspaceservice.go)
├── frontend/
│   ├── src/
│   │   ├── App.svelte
│   │   ├── lib/            # (akan) komponen: Terminal.svelte, Sidebar.svelte, TabBar.svelte
│   │   └── bindings/       # auto-generated Go->TS bindings
│   └── package.json
├── docs/PRD.md             # dokumen ini
└── Taskfile.yml            # task dev/build
```

---

## 9. Gambaran UI/UX

Layout utama (single window, v1):

```
┌───────────┬──────────────────────────────────────────────┐
│           │  [Tab: server] [Tab: worker] [Tab: db] [ + ]  │
│ Workspace │ ─────────────────────────────────────────────│
│  sidebar  │                                               │
│           │                                               │
│ • Codemi  │            xterm.js terminal area             │
│ • Prod ●  │            (bisa di-split jadi pane)           │
│ • Sandbox │                                               │
│           │                                               │
│  [+ new]  │                                               │
└───────────┴──────────────────────────────────────────────┘
```

- **Kiri**: daftar workspace (klik untuk aktifkan, indikator titik = ada sesi running).
- **Atas-kanan**: tab bar untuk terminal dalam workspace aktif, tombol `+` untuk terminal baru.
- **Tengah**: area xterm.js, bisa di-split (P1).
- Tema gelap default, mengikuti estetika developer tool.

---

## 10. Roadmap / Milestone

### M0 — Scaffolding ✅ (selesai)
Setup Wails v3 + Svelte TS, struktur proyek, build berjalan.

### M1 — Terminal Tunggal (MVP inti PTY)
- Integrasi xterm.js di Svelte.
- `SessionService` Go: spawn PTY (`creack/pty`), streaming I/O via Events, resize.
- Satu terminal fungsional penuh (bisa jalankan vim/htop). **(FR-7,8,17,18)**

### M2 — Multi-terminal & Tabs
- Banyak sesi, tab bar, buka/tutup/rename. **(FR-9,10,11,14)**

### M3 — Workspace
- Sidebar workspace, CRUD, aktivasi, persistensi konfig. **(FR-1..4, 22)**

### M4 — Otomasi & Layout
- `startupCommand` + `autoStart`, split pane. **(FR-12,13)**

### M5 — Polish & UX
- Shortcut, command palette, pengaturan font/tema, search. **(FR-19,20,23,24)**

### M6 — Cross-platform
- Dukungan Windows (ConPTY), packaging & distribusi. **(NFR-6, FR-25)**

---

## 11. Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|--------|--------|----------|
| Wails v3 masih **alpha**, API bisa berubah | Refactor mendadak | Kunci versi (`alpha.82`), isolasi pemakaian API Wails di lapisan tipis |
| Performa banyak PTY + streaming event | Lag/CPU tinggi | Throttle **hanya output** (lihat catatan latency di bawah), gunakan binary frame, ujicoba beban dini (gate akhir M2) |
| PTY Windows (ConPTY) berbeda dari Unix | Cross-platform tertunda | Definisikan `interface PTY` eksplisit sebagai sub-task M1, bukan implisit |
| Scrollback besar memakan memori | RAM membengkak / NFR-1 jebol | Batas default **5.000 baris/sesi** di xterm.js (≈ konsumsi terukur untuk 20 sesi), bisa dikonfigurasi |
| Mismatch Go 1.25 (CLI) vs 1.26 (sistem) | Warning generate bindings | Pantau; non-fatal saat ini |
| **Sesi mati total saat app crash** (multiplexing di sisi app) | DevOps kehilangan SSH session panjang | Trade-off sadar v1 (lihat §13.5); host-side `tmux` via `startupCommand` sebagai pelarian |
| **Resize PTY saat output streaming** → frame korup | Tampilan terminal kacau | Debounce resize (pakai `bep/debounce`, sudah ada di `go.mod`); serialize resize vs write per-sesi |
| **Latency input vs throttle output bertabrakan** | Keystroke terasa lambat | Pisahkan jalur: input echo **tidak pernah** di-throttle; throttle hanya untuk flooding output (mis. `yes`) |
| **Lifecycle proses saat window/app close** belum jelas | Proses zombie / kehilangan kerja | Keputusan eksplisit di §13.6: kill semua PTY saat quit + konfirmasi bila ada sesi `running` |

> **Catatan latency (NFR-2).** Target < 30 ms berlaku untuk **input echo** (keystroke → render), diukur terpisah dari throughput output. Throttle/batch hanya boleh diterapkan pada arah output saat banjir data, tidak pada arah input.

---

## 12. Metrik Keberhasilan

- **Adopsi**: pengguna membuat ≥ 2 workspace dan ≥ 3 sesi rata-rata.
- **Retensi**: aplikasi dibuka & dipakai harian (daily driver).
- **Performa**: memenuhi NFR-1 & NFR-2 pada beban 20 terminal.
- **Stabilitas**: 0 crash aplikasi akibat 1 terminal mati selama sesi normal.

---

## 13. Keputusan (sebelumnya Pertanyaan Terbuka)

Empat pertanyaan awal sudah diputuskan agar M1–M3 bisa berjalan tanpa ambiguitas. Keputusan dapat ditinjau ulang bila asumsi berubah.

### 13.1 Penyimpanan konfig → **JSON (v1)**
Data v1 kecil & jarang berubah (definisi workspace/sesi, bukan scrollback). JSON: human-editable, zero migration, mudah di-diff. **Pindah ke SQLite hanya jika** persistensi scrollback/riwayat naik dari backlog jadi prioritas. Tidak memilih SQLite "untuk jaga-jaga".

### 13.2 Profil shell global → **Ya, masuk v1**
Preset shell (bash/zsh/fish) ringan dibangun (~1 hari) dan langsung bernilai untuk semua persona. Sesi baru mewarisi shell default workspace, bisa di-override per-sesi (`Session.shell`).

### 13.3 SSH first-class → **Ditunda (v1 cukup `startupCommand: ssh ...`)**
Template sesi SSH menyeret kompleksitas (manajemen secret/keyring, lihat §13.4). Untuk v1, SSH cukup lewat `startupCommand`. SSH first-class dipertimbangkan pasca-v1.

### 13.4 Manajemen secret / env sensitif → **Plaintext + 0600 untuk v1**
Tidak ada integrasi OS keyring (libsecret/Keychain) di v1 — over-engineering untuk tahap ini. Env disimpan plaintext di file konfig dengan permission **0600**. NFR-5 diklarifikasi: ini keputusan sadar, bukan ambiguitas. Keyring dipertimbangkan saat SSH first-class.

### 13.5 Multi-window → **Ditunda ke M6**
Multi-window Wails v3 (alpha) + state sharing antar-window berisiko jadi sumber bug. Tidak boleh ada di jalur kritis MVP.

### 13.6 Lifecycle PTY saat window/app close → **Kill semua + konfirmasi**
Karena multiplexing terjadi di sisi app (Non-Goal §3.2), sesi **tidak** bertahan setelah app ditutup. Saat quit: bila ada sesi `running`, tampilkan konfirmasi, lalu kill seluruh PTY secara graceful (SIGTERM → SIGKILL). Konsekuensi ini wajib disadari desain `SessionService` sejak M1. User yang butuh sesi tahan-banting diarahkan ke host-side `tmux` (lihat risiko di §11).

### Masih terbuka
- Format event streaming output: teks UTF-8 vs binary frame (diputuskan setelah load-test M2).

---

*Dokumen ini hidup (living document) — akan diperbarui seiring pengembangan.*
