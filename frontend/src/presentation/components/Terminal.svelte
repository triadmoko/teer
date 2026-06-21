<script lang="ts">
  import { tick, untrack } from "svelte";
  import { get } from "svelte/store";
  import { Terminal } from "@xterm/xterm";
  import { FitAddon } from "@xterm/addon-fit";
  import { SearchAddon } from "@xterm/addon-search";
  import { WebLinksAddon } from "@xterm/addon-web-links";
  import { SerializeAddon } from "@xterm/addon-serialize";
  import "@xterm/xterm/css/xterm.css";
  import type { SessionDef } from "@domain/models";
  import { isMac } from "@domain/platform";
  import {
    startSession,
    writeSession,
    resizeSession,
    onSessionOutput,
    onSessionExit,
    setRunning,
    restartCount,
    broadcastMode,
    broadcastWrite,
    terminalFontSize,
    terminalFontFamily,
    terminalTheme,
    terminalPersistScrollback,
    terminalScrollbackLines,
    copyOnSelect,
    middleClickPaste,
    persistScrollback,
    restoreScrollback,
  } from "@application";

  let {
    session,
    active = false,
    visible = true,
    connect = true,
    wsEnv = {},
    wsCwd = "",
    wsStartupCommand = "",
  }: {
    session: SessionDef;
    active?: boolean;
            visible?: boolean;
    // connect: PTY hanya dijalankan saat true. Session non-autoStart yang belum
    // dibuka tampil sebagai cell/tab "mati" (tombol Restart) sampai dipilih.
    connect?: boolean;
    wsEnv?: Record<string, string>;
    wsCwd?: string;
    wsStartupCommand?: string;
  } = $props();

  let container = $state<HTMLDivElement>();
  let term = $state<Terminal | undefined>();
  let fit = $state<FitAddon | undefined>();
  let search = $state<SearchAddon | undefined>();
  // restoreDone selesai setelah scrollback lama (jika ada) ditulis ke terminal,
  // sehingga PTY baru hanya dijalankan sesudahnya — cegah output bercampur.
  let restoreDone: Promise<void> = Promise.resolve();
  // dirty: ada output baru sejak snapshot terakhir. Snapshot periodik hanya
  // menserialisasi bila dirty (hemat untuk session verbose).
  let dirty = false;
  let cleanups = $state<Array<() => void>>([]);
  let ro = $state<ResizeObserver | undefined>();
  let fitTimer = $state<ReturnType<typeof setTimeout> | undefined>();
    let exited = $state(false);
    // started: PTY sudah pernah dijalankan. Cegah double-connect dan cegah
    // auto-reconnect setelah proses exit (restart hanya via tombol).
    let started = $state(false);
    // Context menu klik-kanan (Copy/Paste). null = tertutup.
    let menu = $state<{ x: number; y: number; hasSel: boolean } | null>(null);

  // Salin teks ke clipboard. Gagal diam-diam bila clipboard ditolak WebView.
  async function copyText(text: string) {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      // clipboard ditolak — abaikan
    }
  }

  // Tempel dari clipboard ke stdin shell. WAJIB lewat writeSession (bukan
  // term.write) agar teks benar-benar masuk ke PTY, bukan hanya dirender lokal.
  async function pasteInto(sid: string) {
    try {
      const text = await navigator.clipboard.readText();
      if (text) writeSession(sid, text);
    } catch {
      // clipboard ditolak — abaikan
    }
  }

  function menuCopy() {
    if (term?.hasSelection()) void copyText(term.getSelection());
    menu = null;
  }

  function menuPaste() {
    void pasteInto(session.id);
    menu = null;
  }

      function refit() {
    if (fitTimer) clearTimeout(fitTimer);
    fitTimer = setTimeout(() => {
      if (!fit || !term) return;
      try {
        fit.fit();
        term.refresh(0, term.rows - 1);
      } catch {

      }
    }, 50);
  }

  function startPty(_term: Terminal) {
    return startSession(session, wsEnv, wsCwd, wsStartupCommand, {
      cols: _term.cols,
      rows: _term.rows,
    });
  }

  // Pulihkan scrollback sesi sebelumnya (read-only) sebelum PTY dijalankan.
  // Best-effort: kegagalan tidak boleh menghalangi koneksi PTY.
  async function maybeRestore(_term: Terminal, sid: string) {
    if (!get(terminalPersistScrollback)) return;
    try {
      const prev = await restoreScrollback(sid);
      if (prev) {
        _term.write(prev);
        _term.write(
          "\r\n\x1b[90m─── sesi sebelumnya dipulihkan ───\x1b[0m\r\n",
        );
      }
    } catch {
      // abaikan — snapshot bersifat best-effort
    }
  }

  // Hubungkan PTY (sekali). Dipanggil saat mount jika connect=true, atau saat
  // session non-autoStart akhirnya dipilih (connect berubah jadi true).
  function connectPty(_term: Terminal) {
    const sid = untrack(() => session.id);
    untrack(() => { started = true; });
    restoreDone.then(() => untrack(() => startPty(_term))).then(() => {
      setRunning(sid, true);
      untrack(() => { exited = false; });
      tick().then(() => {
        try { fit?.fit(); } catch {  }
        const { cols, rows } = _term;
        if (cols > 0 && rows > 0) resizeSession(sid, cols, rows);
        if (untrack(() => visible && active)) _term.focus();
      });
    });
  }

  $effect(() => {
        const el = container;
    if (!el) return;

    const _term = new Terminal({
      fontFamily: untrack(() => get(terminalFontFamily)),
      fontSize: untrack(() => get(terminalFontSize)),
      cursorBlink: true,
      scrollback: untrack(() => get(terminalScrollbackLines)),
      theme: untrack(() => get(terminalTheme)),
      allowProposedApi: true,
    });
    const _fit = new FitAddon();
    const _search = new SearchAddon();
    const _serialize = new SerializeAddon();
    _term.loadAddon(_fit);
    _term.loadAddon(_search);
    _term.loadAddon(_serialize);
    _term.loadAddon(new WebLinksAddon());

    _term.open(el);
    try {
      _fit.fit();
    } catch {

    }

            const onData = _term.onData((data) => {
      if (get(broadcastMode)) broadcastWrite(data);
      else writeSession(session.id, data);
    });

        const onResize = _term.onResize(({ cols, rows }) =>
      resizeSession(session.id, cols, rows),
    );

                const sid = untrack(() => session.id);

        const offOut = onSessionOutput(sid, (bytes) => {
      _term.write(bytes);
      dirty = true;
    });

        const offExit = onSessionExit(sid, (code) => {
      _term.write(`\r\n\x1b[90m[proses berakhir — kode ${code}]\x1b[0m\r\n`);
      setRunning(sid, false);
      untrack(() => { exited = true; });
    });

    // Copy/Paste keyboard: Ctrl+Shift+C/V (Cmd+C/V di macOS). Selain itu
    // diteruskan ke PTY — termasuk Ctrl+C → SIGINT.
    _term.attachCustomKeyEventHandler((e) => {
      if (e.type !== "keydown") return true;
      const mod = isMac ? e.metaKey && !e.ctrlKey : e.ctrlKey && e.shiftKey;
      if (mod && e.code === "KeyC" && _term.hasSelection()) {
        void copyText(_term.getSelection());
        return false;
      }
      if (mod && e.code === "KeyV") {
        void pasteInto(sid);
        return false;
      }
      return true;
    });

    // Copy-on-select: salin otomatis saat seleksi berubah (bila diaktifkan).
    const onSel = _term.onSelectionChange(() => {
      if (!get(copyOnSelect)) return;
      const sel = _term.getSelection();
      if (sel) void copyText(sel);
    });

    // Middle-click paste (bila diaktifkan).
    const onAux = (ev: MouseEvent) => {
      if (ev.button !== 1 || !get(middleClickPaste)) return;
      ev.preventDefault();
      void pasteInto(sid);
    };
    el.addEventListener("auxclick", onAux);

    // Context menu klik-kanan: tampilkan menu Copy/Paste.
    const onCtx = (ev: MouseEvent) => {
      ev.preventDefault();
      menu = { x: ev.clientX, y: ev.clientY, hasSel: _term.hasSelection() };
    };
    el.addEventListener("contextmenu", onCtx);

    const _ro = new ResizeObserver(() => refit());
    _ro.observe(el);

    untrack(() => {
      term = _term;
      fit = _fit;
      search = _search;
      cleanups = [
        () => onData.dispose(),
        () => onResize.dispose(),
        offOut,
        offExit,
        () => onSel.dispose(),
        () => el.removeEventListener("auxclick", onAux),
        () => el.removeEventListener("contextmenu", onCtx),
      ];
      ro = _ro;
    });

    // Pulihkan scrollback dulu, lalu (untuk eager connect) jalankan PTY.
    // connectPty sendiri menunggu restoreDone, jadi urutan tetap benar walau
    // koneksi terjadi lewat lazy-connect effect.
    restoreDone = untrack(() => maybeRestore(_term, sid));

    if (untrack(() => connect)) connectPty(_term);

    // Snapshot periodik anti-crash: hanya menserialisasi bila ada output baru.
    const _snap = setInterval(() => {
      if (!dirty || !get(terminalPersistScrollback)) return;
      dirty = false;
      try {
        persistScrollback(sid, _serialize.serialize());
      } catch {
        // best-effort
      }
    }, 10000);

    return () => {
      // Snapshot terakhir sebelum dispose (serialize butuh terminal hidup).
      if (untrack(() => get(terminalPersistScrollback))) {
        try {
          persistScrollback(sid, _serialize.serialize());
        } catch {
          // best-effort
        }
      }
      clearInterval(_snap);
      cleanups.forEach((fn) => fn());
      _ro.disconnect();
      if (fitTimer) clearTimeout(fitTimer);
      _term.dispose();
      untrack(() => {
        term = undefined;
        fit = undefined;
        search = undefined;
        cleanups = [];
        ro = undefined;
        menu = null;
      });
    };
  });

    // Lazy-connect: session non-autoStart dirender tanpa PTY (connect=false).
    // Saat akhirnya dipilih/dibuka (connect -> true), jalankan PTY sekali.
    $effect(() => {
    if (!connect || !term || untrack(() => started)) return;
    connectPty(term);
  });

    // Saat terminal jadi visible lagi (mis. switch workspace/tab), paksa
    // fit + refresh setelah layout & paint. Pakai double-rAF: display:none ->
    // block butuh satu frame untuk relayout, frame kedua untuk dimensi final.
    // Tanpa ini WebKitGTK menampilkan paint basi (blank) sampai di-resize.
    $effect(() => {
    if (!visible || !term) return;
    const _t = term;
    const _f = fit;
    const sid = untrack(() => session.id);
    requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        try {
          _f?.fit();
        } catch {

        }
        const { cols, rows } = _t;
        if (cols > 0 && rows > 0) resizeSession(sid, cols, rows);
        _t.refresh(0, _t.rows - 1);
        if (untrack(() => active)) _t.focus();
      }),
    );
  });

    $effect(() => {
    if (active && visible && term) {
      tick().then(() => term?.focus());
    }
  });

    $effect(() => {
    const sid = untrack(() => session.id);
    const count = $restartCount[sid];
    if (!count || !term) return;
    exited = false;
    untrack(() => { started = true; });
    const _t = term;
    untrack(() => startPty(_t)).then(() => {
      setRunning(sid, true);
      tick().then(() => {
        refit();
        const { cols, rows } = _t;
        if (cols > 0 && rows > 0) resizeSession(sid, cols, rows);
      });
    });
  });

    $effect(() => {
    const size = $terminalFontSize;
    const family = $terminalFontFamily;
    const th = $terminalTheme;
    if (!term) return;
    term.options.fontSize = size;
    term.options.fontFamily = family;
    term.options.theme = th;
    tick().then(() => refit());
  });

  export function find(query: string, next = true) {
    if (!search || !query) return;
    if (next) search?.findNext(query);
    else search?.findPrevious(query);
  }

</script>

<div
  class="absolute inset-0 px-2 py-[6px]"
  class:hidden={!visible}
  class:pointer-events-none={!visible}
  bind:this={container}
></div>

{#if menu}
  <!-- Backdrop transparan menutup menu saat klik di luar. -->
  <button
    class="ctx-backdrop"
    aria-label="Tutup menu"
    onclick={() => (menu = null)}
    oncontextmenu={(e) => {
      e.preventDefault();
      menu = null;
    }}
  ></button>
  <div
    class="ctx-menu"
    style="left: {menu.x}px; top: {menu.y}px;"
    role="menu"
    tabindex="-1"
  >
    <button
      class="ctx-item"
      role="menuitem"
      disabled={!menu.hasSel}
      onclick={menuCopy}>Copy</button
    >
    <button class="ctx-item" role="menuitem" onclick={menuPaste}>Paste</button>
  </div>
{/if}

<svelte:window
  onkeydown={(e) => {
    if (menu && e.key === "Escape") menu = null;
  }}
/>

<style>

  div {
    /* WebKitGTK menekan mousemove saat drag pada elemen user-select:none.
       Override di sini agar drag-to-select sampai ke xterm; xterm.css
       sendiri menimpa kembali ke none pada .xterm sehingga native
       browser selection tetap tidak terlihat. */
    -webkit-user-select: text;
    user-select: text;
  }

  div :global(.xterm) {
    height: 100%;
  }
  div :global(.xterm-viewport) {

    scrollbar-width: thin;
  }

  .ctx-backdrop {
    position: fixed;
    inset: 0;
    z-index: 40;
    background: transparent;
    border: none;
    padding: 0;
    cursor: default;
  }

  .ctx-menu {
    position: fixed;
    z-index: 50;
    min-width: 120px;
    padding: 4px;
    border: 1px solid var(--color-line, #333);
    border-radius: 7px;
    background: var(--color-raise, #1a1a1a);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  }

  .ctx-item {
    display: block;
    width: 100%;
    padding: 5px 10px;
    border: none;
    border-radius: 5px;
    background: transparent;
    color: #d4d4d8;
    font-size: 13px;
    text-align: left;
    cursor: pointer;
  }
  .ctx-item:hover:not(:disabled) {
    background: var(--color-active, #2a2a2a);
    color: #fff;
  }
  .ctx-item:disabled {
    color: #52525b;
    cursor: default;
  }
</style>
