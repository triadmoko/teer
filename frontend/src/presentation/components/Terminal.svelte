<script lang="ts">
  import { tick, untrack } from "svelte";
  import { get } from "svelte/store";
  import { Terminal } from "@xterm/xterm";
  import { FitAddon } from "@xterm/addon-fit";
  import { SearchAddon } from "@xterm/addon-search";
  import { WebLinksAddon } from "@xterm/addon-web-links";
  import "@xterm/xterm/css/xterm.css";
  import type { SessionDef } from "@domain/models";
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
  } from "@application";

  let {
    session,
    active = false,
    visible = true,
    wsEnv = {},
    wsCwd = "",
    wsStartupCommand = "",
  }: {
    session: SessionDef;
    active?: boolean;
    // Apakah terminal sedang ditampilkan (mode tab: hanya yang aktif;
    // mode grid: semua sesi terbuka tampil sekaligus).
    visible?: boolean;
    wsEnv?: Record<string, string>;
    wsCwd?: string;
    wsStartupCommand?: string;
  } = $props();

  let container = $state<HTMLDivElement>();
  let term = $state<Terminal | undefined>();
  let fit = $state<FitAddon | undefined>();
  let search = $state<SearchAddon | undefined>();
  let cleanups = $state<Array<() => void>>([]);
  let ro = $state<ResizeObserver | undefined>();
  let fitTimer = $state<ReturnType<typeof setTimeout> | undefined>();
  // Lacak apakah sesi masih exited (belum di-restart) untuk tampilkan tombol restart.
  let exited = $state(false);

  // Debounce fit untuk menghindari resize race saat container berubah cepat
  // (PRD §11: resize PTY saat output streaming).
  function refit() {
    if (fitTimer) clearTimeout(fitTimer);
    fitTimer = setTimeout(() => {
      if (!visible || !fit) return;
      try {
        fit.fit();
      } catch {
        /* container belum punya dimensi */
      }
    }, 50);
  }

  function startPty(_term: Terminal) {
    return startSession(session, wsEnv, wsCwd, wsStartupCommand, {
      cols: _term.cols,
      rows: _term.rows,
    });
  }

  $effect(() => {
    // Akses container agar effect re-run saat container siap.
    const el = container;
    if (!el) return;

    const _term = new Terminal({
      fontFamily: untrack(() => get(terminalFontFamily)),
      fontSize: untrack(() => get(terminalFontSize)),
      cursorBlink: true,
      scrollback: 5000, // batas sesuai PRD NFR-1 / §11
      theme: untrack(() => get(terminalTheme)),
      allowProposedApi: true,
    });
    const _fit = new FitAddon();
    const _search = new SearchAddon();
    _term.loadAddon(_fit);
    _term.loadAddon(_search);
    _term.loadAddon(new WebLinksAddon());

    _term.open(el);
    try {
      _fit.fit();
    } catch {
      /* noop */
    }

    // Input keyboard/paste → stdin shell (FR-18).
    // Bila broadcast aktif (FR-16), kirim ke semua sesi yang terbuka.
    const onData = _term.onData((data) => {
      if (get(broadcastMode)) broadcastWrite(data);
      else writeSession(session.id, data);
    });

    // xterm resize → PTY resize (FR-17).
    const onResize = _term.onResize(({ cols, rows }) =>
      resizeSession(session.id, cols, rows),
    );

    // Snapshot session.id sekali — id tidak pernah berubah untuk instance ini
    // (keyed by s.id). Untrack mencegah refresh() workspace men-trigger ulang
    // effect ini hanya karena object session baru dibuat dengan id yang sama.
    const sid = untrack(() => session.id);

    // Output PTY (byte mentah) → tulis ke xterm.
    const offOut = onSessionOutput(sid, (bytes) => _term.write(bytes));

    // Sesi berakhir → tampilkan penanda + lapor status + aktifkan tombol restart (FR-14,15).
    const offExit = onSessionExit(sid, (code) => {
      _term.write(`\r\n\x1b[90m[proses berakhir — kode ${code}]\x1b[0m\r\n`);
      setRunning(sid, false);
      untrack(() => { exited = true; });
    });

    const _ro = new ResizeObserver(() => refit());
    _ro.observe(el);

    untrack(() => {
      term = _term;
      fit = _fit;
      search = _search;
      cleanups = [() => onData.dispose(), () => onResize.dispose(), offOut, offExit];
      ro = _ro;
    });

    // Untrack startPty agar session/wsEnv/wsCwd/wsStartupCommand tidak menjadi
    // dependensi effect ini — perubahan prop saat workspace refresh tidak boleh
    // menghancurkan dan membuat ulang xterm.
    untrack(() => startPty(_term)).then(() => {
      setRunning(sid, true);
      untrack(() => { exited = false; });
      tick().then(() => {
        // Fit xterm ke container, lalu eksplisit kirim dimensi ke PTY.
        // Ini memaksa SIGWINCH ke proses PTY sehingga aplikasi (shell, vim, dll.)
        // menggambar ulang layarnya ke xterm baru — penting saat reconnect ke PTY
        // yang sudah berjalan (contoh: pindah workspace lalu kembali), karena PTY
        // tidak mengirim ulang buffer layarnya secara otomatis ke subscriber baru.
        try { _fit.fit(); } catch { /* container belum punya dimensi */ }
        const { cols, rows } = _term;
        if (cols > 0 && rows > 0) resizeSession(sid, cols, rows);
        if (visible && active) _term.focus();
      });
    });

    return () => {
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
      });
    };
  });

  // Saat terminal tampil (tab aktif / masuk grid), refit ukurannya.
  $effect(() => {
    if (visible && term) {
      tick().then(() => refit());
    }
  });

  // Hanya terminal aktif yang menerima fokus keyboard.
  $effect(() => {
    if (active && visible && term) {
      tick().then(() => term?.focus());
    }
  });

  // Restart PTY saat restartCount[id] berubah (FR-15).
  $effect(() => {
    const sid = untrack(() => session.id);
    const count = $restartCount[sid];
    if (!count || !term) return;
    exited = false;
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

  // Terapkan perubahan pengaturan font/tema ke instance xterm yang sudah ada (FR-20).
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

  /** Dipanggil parent untuk mencari teks di scrollback (FR-19). */
  export function find(query: string, next = true) {
    if (!search || !query) return;
    if (next) search?.findNext(query);
    else search?.findPrevious(query);
  }

</script>

<div
  class="absolute inset-0 px-2 py-[6px] {visible ? 'block' : 'hidden'}"
  bind:this={container}
></div>

<style>
  /* xterm meng-inject DOM-nya sendiri, jadi tetap pakai selektor global. */
  div :global(.xterm) {
    height: 100%;
  }
  div :global(.xterm-viewport) {
    /* scrollbar tipis agar bersih */
    scrollbar-width: thin;
  }
</style>
