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
  import { park, checkout } from "@infrastructure/terminalRegistry";

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

  // wrapper: div Svelte-managed (bind:this). Datang dan pergi sesuai lifecycle komponen.
  // _el (di dalam $effect): div yang xterm.open() dipanggil — dikelola manual lewat
  // registry agar bisa di-park/checkout tanpa dispose saat komponen unmount.
  let wrapper = $state<HTMLDivElement>();
  let term = $state<Terminal | undefined>();
  let fit = $state<FitAddon | undefined>();
  let search = $state<SearchAddon | undefined>();
  let restoreDone: Promise<void> = Promise.resolve();
  let fitTimer = $state<ReturnType<typeof setTimeout> | undefined>();
  let exited = $state(false);
  // started: PTY sudah pernah dijalankan. Cegah double-connect dan cegah
  // auto-reconnect setelah proses exit (restart hanya via tombol).
  let started = $state(false);
  // Context menu klik-kanan (Copy/Paste). null = tertutup.
  let menu = $state<{ x: number; y: number; hasSel: boolean } | null>(null);

  async function copyText(text: string) {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      // clipboard ditolak — abaikan
    }
  }

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
        // abaikan — fit gagal saat container tersembunyi
      }
    }, 50);
  }

  function startPty(_term: Terminal) {
    return startSession(session, wsEnv, wsCwd, wsStartupCommand, {
      cols: _term.cols,
      rows: _term.rows,
    });
  }

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

  function connectPty(_term: Terminal) {
    const sid = untrack(() => session.id);
    untrack(() => { started = true; });
    restoreDone.then(() => untrack(() => startPty(_term))).then(() => {
      setRunning(sid, true);
      untrack(() => { exited = false; });
      tick().then(() => {
        try { fit?.fit(); } catch {}
        const { cols, rows } = _term;
        if (cols > 0 && rows > 0) resizeSession(sid, cols, rows);
        if (untrack(() => visible && active)) _term.focus();
      });
    });
  }

  $effect(() => {
    const _wrapper = wrapper;
    if (!_wrapper) return;

    const sid = untrack(() => session.id);
    const cached = checkout(sid);

    let _term: Terminal;
    let _fit: FitAddon;
    let _search: SearchAddon;
    let _serialize: SerializeAddon;
    let _el: HTMLDivElement;
    let dirtyRef: { value: boolean };

    if (cached) {
      // Checkout: reuse xterm instance + DOM element dari registry.
      // Semua listener persistent (onData, onSessionOutput, dll.) sudah aktif.
      _term = cached.term;
      _fit = cached.fit;
      _search = cached.search;
      _serialize = cached.serialize;
      _el = cached.el;
      dirtyRef = cached.dirtyRef;
      untrack(() => { started = cached.started; });
      _wrapper.appendChild(_el);
    } else {
      // Fresh: buat xterm baru beserta semua listener persistent-nya.
      dirtyRef = { value: false };

      _el = document.createElement("div");
      // Styling inline — tidak bisa pakai Tailwind pada elemen dinamis.
      _el.style.cssText =
        "position:absolute;inset:0;padding:6px 8px;box-sizing:border-box;" +
        "-webkit-user-select:text;user-select:text;";
      _wrapper.appendChild(_el);

      _term = new Terminal({
        fontFamily: untrack(() => get(terminalFontFamily)),
        fontSize: untrack(() => get(terminalFontSize)),
        cursorBlink: true,
        scrollback: untrack(() => get(terminalScrollbackLines)),
        theme: untrack(() => get(terminalTheme)),
        allowProposedApi: true,
      });
      _fit = new FitAddon();
      _search = new SearchAddon();
      _serialize = new SerializeAddon();
      _term.loadAddon(_fit);
      _term.loadAddon(_search);
      _term.loadAddon(_serialize);
      _term.loadAddon(new WebLinksAddon());
      _term.open(_el);
      try { _fit.fit(); } catch {}

      // Listener persistent — tetap aktif selama sesi hidup, melewati park/checkout.
      // onSessionOutput sengaja tidak diputus saat park supaya buffer xterm tetap
      // menerima output meskipun terminal sedang tersembunyi di parking div.
      _term.onData((data) => {
        if (get(broadcastMode)) broadcastWrite(data);
        else writeSession(sid, data);
      });

      _term.onResize(({ cols, rows }) => resizeSession(sid, cols, rows));

      onSessionOutput(sid, (bytes) => {
        _term.write(bytes);
        dirtyRef.value = true;
      });

      onSessionExit(sid, (code) => {
        _term.write(`\r\n\x1b[90m[proses berakhir — kode ${code}]\x1b[0m\r\n`);
        setRunning(sid, false);
        untrack(() => { exited = true; });
      });

      _term.onSelectionChange(() => {
        if (!get(copyOnSelect)) return;
        const sel = _term.getSelection();
        if (sel) void copyText(sel);
      });

      _el.addEventListener("auxclick", (ev: MouseEvent) => {
        if (ev.button !== 1 || !get(middleClickPaste)) return;
        ev.preventDefault();
        void pasteInto(sid);
      });

      _el.addEventListener("contextmenu", (ev: MouseEvent) => {
        ev.preventDefault();
        menu = { x: ev.clientX, y: ev.clientY, hasSel: _term.hasSelection() };
      });

      restoreDone = untrack(() => maybeRestore(_term, sid));
      if (untrack(() => connect)) connectPty(_term);
    }

    // Sync Svelte state — untrack agar tidak memicu re-run effect.
    untrack(() => {
      term = _term;
      fit = _fit;
      search = _search;
    });

    // ResizeObserver: ephemeral per-mount. Diputus saat park agar tidak
    // memicu refit ke ukuran 0×0 di parking div.
    const _ro = new ResizeObserver(() => refit());
    _ro.observe(_wrapper);

    if (cached) {
      // Setelah checkout: paksa refit supaya canvas menyesuaikan ukuran wrapper baru.
      tick().then(() => {
        try { _fit.fit(); } catch {}
        const { cols, rows } = _term;
        if (cols > 0 && rows > 0) resizeSession(sid, cols, rows);
        if (untrack(() => visible && active)) _term.focus();
      });
    }

    // Snapshot periodik: ephemeral, hanya saat terminal visible/aktif.
    const _snap = setInterval(() => {
      if (!dirtyRef.value || !get(terminalPersistScrollback)) return;
      dirtyRef.value = false;
      try {
        persistScrollback(sid, _serialize.serialize());
      } catch {
        // best-effort
      }
    }, 10000);

    return () => {
      // Snapshot terakhir sebelum park.
      if (untrack(() => get(terminalPersistScrollback))) {
        try {
          persistScrollback(sid, _serialize.serialize());
        } catch {
          // best-effort
        }
      }
      clearInterval(_snap);
      _ro.disconnect();
      if (fitTimer) clearTimeout(fitTimer);

      // Park: pindahkan _el ke parking div, simpan entry ke registry.
      // Listener persistent (onData, onSessionOutput, dll.) TIDAK diputus —
      // onSessionOutput khususnya harus tetap aktif agar buffer ter-update
      // saat terminal berada di background.
      park(sid, {
        term: _term,
        fit: _fit,
        search: _search,
        serialize: _serialize,
        el: _el,
        dirtyRef,
        started: untrack(() => started),
      });

      untrack(() => {
        term = undefined;
        fit = undefined;
        search = undefined;
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
        try { _f?.fit(); } catch {}
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
  class="absolute inset-0"
  class:hidden={!visible}
  class:pointer-events-none={!visible}
  bind:this={wrapper}
></div>

{#if menu}
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
