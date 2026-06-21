# Draf 02 — Copy / Paste Terminal yang Proper

**Status:** Draf
**Prioritas:** 🟡 Sedang
**Effort:** Kecil

## Masalah

Copy/paste mengandalkan perilaku default xterm.js + WebView. Di WebKitGTK
(Linux) ini sering tidak konsisten:

- Tidak ada shortcut eksplisit `Ctrl+Shift+C` / `Ctrl+Shift+V`.
- `Ctrl+C` di terminal harus tetap mengirim SIGINT, bukan copy — jadi konvensi
  terminal memakai `Ctrl+Shift+C/V`.
- Tidak ada "copy on select" / "paste on middle-click" (kebiasaan terminal Linux).
- Bug selection di WebKitGTK sudah pernah disentuh: lihat override
  `user-select` di `Terminal.svelte:246-255` (perubahan yang sedang berjalan).

Ini melengkapi kerja selection yang sudah dimulai.

## Tujuan

1. `Ctrl+Shift+C` menyalin selection aktif; `Ctrl+Shift+V` menempel dari clipboard.
2. Opsional: **copy on select** dan **middle-click paste** (toggle di settings).
3. `Ctrl+C` tanpa Shift tetap meneruskan ke PTY (SIGINT) seperti sekarang.
4. Tidak memakai shortcut yang bentrok dengan command palette (`Ctrl+Shift+P`)
   atau new/close session (`Ctrl+T` / `Ctrl+W`).

## Desain

Seluruhnya di frontend — **tidak ada perubahan backend**. xterm.js sudah
menyediakan API yang dibutuhkan:

- `term.hasSelection()`, `term.getSelection()`, `term.clearSelection()`
- `term.onSelectionChange(...)` untuk copy-on-select
- `term.attachCustomKeyEventHandler(cb)` untuk intercept keydown sebelum dikirim
  ke PTY (return `false` agar event tidak diteruskan)
- Clipboard via `navigator.clipboard.writeText/readText`

### Alur keyboard

Pasang handler di pembuatan `Terminal` (`Terminal.svelte:98`):

```ts
_term.attachCustomKeyEventHandler((e) => {
  if (e.type !== "keydown") return true;
  const mod = e.ctrlKey && e.shiftKey;        // Linux/Windows
  // (macOS: pakai e.metaKey untuk Cmd+C/Cmd+V — deteksi platform)
  if (mod && e.code === "KeyC" && _term.hasSelection()) {
    void copyText(_term.getSelection());
    return false; // jangan teruskan ke PTY
  }
  if (mod && e.code === "KeyV") {
    void pasteFromClipboard(_term);
    return false;
  }
  return true; // selain itu, biarkan PTY (termasuk Ctrl+C → SIGINT)
});
```

### Paste

```ts
async function pasteFromClipboard(term: Terminal) {
  try {
    const text = await navigator.clipboard.readText();
    if (text) writeSession(term ...session.id, text); // lewat application layer
  } catch { /* clipboard ditolak — abaikan diam-diam */ }
}
```

> Paste harus melalui `writeSession(...)` (alur existing di `Terminal.svelte:119-122`),
> bukan `term.write()` — agar teks benar-benar masuk ke stdin shell, bukan cuma
> dirender lokal.

### Copy on select (opsional, di balik setting)

```ts
const onSel = _term.onSelectionChange(() => {
  if (!get(copyOnSelect)) return;
  const sel = _term.getSelection();
  if (sel) void copyText(sel);
});
// daftarkan onSel.dispose() ke array `cleanups` (Terminal.svelte:145)
```

### Middle-click paste (opsional)

Listener `auxclick`/`mousedown` (button 1) pada `container`
(`Terminal.svelte:243`) → `pasteFromClipboard`. Hormati toggle `middleClickPaste`.

## Catatan WebKitGTK

- `navigator.clipboard` butuh konteks "secure"/user-gesture; di WebView Wails
  umumnya OK karena dipicu keypress/click. Bila `readText()` ditolak, fallback
  diam (jangan throw ke user).
- Selection sudah dibetulkan via CSS override (`Terminal.svelte:248-255`).
  Verifikasi drag-to-select tetap jalan setelah handler dipasang.

## Settings

Tambah ke terminal settings (`frontend/src/domain/terminalSettings.ts` +
panel UI-nya):

- `copyOnSelect: boolean` (default `false`)
- `middleClickPaste: boolean` (default `false` di non-Linux, `true` di Linux?)

## Rencana Test

Murni manual (perilaku UI/clipboard, sulit di-unit-test):

1. Select teks → `Ctrl+Shift+C` → tempel di app lain → cocok.
2. Copy di app lain → `Ctrl+Shift+V` di terminal → teks masuk ke shell.
3. `Ctrl+C` saat proses jalan (mis. `ping`) → tetap SIGINT (berhenti), **tidak** copy.
4. Toggle copy-on-select & middle-click → sesuai setting.
5. Multi-platform bila memungkinkan (Linux WebKitGTK + Windows WebView2).

## Pertanyaan terbuka

1. macOS: pakai `Cmd+C/Cmd+V` (metaKey). Perlu deteksi platform — apakah ada
   util platform yang sudah dipakai di frontend? Cek `application/`.
2. Default `middleClickPaste` per-OS atau seragam `false`?
3. Apakah perlu menu konteks klik-kanan (Copy/Paste) selain shortcut?

## Checklist

- [ ] `attachCustomKeyEventHandler` (Ctrl+Shift+C/V, hormati Ctrl+C→SIGINT)
- [ ] `copyText` / `pasteFromClipboard` (paste via `writeSession`)
- [ ] Deteksi platform untuk Cmd di macOS
- [ ] `onSelectionChange` copy-on-select (+ dispose di cleanup)
- [ ] Middle-click paste listener
- [ ] Settings `copyOnSelect` + `middleClickPaste`
- [ ] Uji manual 5 skenario
