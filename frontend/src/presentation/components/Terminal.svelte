<script lang="ts">
  import { tick, untrack } from "svelte";
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
  } from "@application";

  let {
    session,
    active = false,
    visible = true,
    wsEnv = {},
    wsCwd = "",
  }: {
    session: SessionDef;
    active?: boolean;
    // Apakah terminal sedang ditampilkan (mode tab: hanya yang aktif;
    // mode grid: semua sesi terbuka tampil sekaligus).
    visible?: boolean;
    wsEnv?: Record<string, string>;
    wsCwd?: string;
  } = $props();

  let container = $state<HTMLDivElement>();
  let term = $state<Terminal | undefined>();
  let fit = $state<FitAddon | undefined>();
  let search = $state<SearchAddon | undefined>();
  let cleanups = $state<Array<() => void>>([]);
  let ro = $state<ResizeObserver | undefined>();
  let fitTimer = $state<ReturnType<typeof setTimeout> | undefined>();

  // Tema gelap default mengikuti estetika developer tool (PRD §9).
  const theme = {
    background: "#121214",
    foreground: "#d4d4d8",
    cursor: "#e4e4e7",
    selectionBackground: "#3f3f46",
    black: "#18181b",
    red: "#f87171",
    green: "#4ade80",
    yellow: "#facc15",
    blue: "#60a5fa",
    magenta: "#c084fc",
    cyan: "#22d3ee",
    white: "#d4d4d8",
    brightBlack: "#52525b",
  };

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
    return startSession(session, wsEnv, wsCwd, {
      cols: _term.cols,
      rows: _term.rows,
    });
  }

  $effect(() => {
    // Akses container agar effect re-run saat container siap.
    const el = container;
    if (!el) return;

    const _term = new Terminal({
      fontFamily: 'ui-monospace, "Cascadia Code", "JetBrains Mono", Menlo, monospace',
      fontSize: 13,
      cursorBlink: true,
      scrollback: 5000, // batas sesuai PRD NFR-1 / §11
      theme,
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
    const onData = _term.onData((data) => writeSession(session.id, data));

    // xterm resize → PTY resize (FR-17).
    const onResize = _term.onResize(({ cols, rows }) =>
      resizeSession(session.id, cols, rows),
    );

    // Output PTY (byte mentah) → tulis ke xterm.
    const offOut = onSessionOutput(session.id, (bytes) => _term.write(bytes));

    // Sesi berakhir → tampilkan penanda + lapor status (FR-14).
    const offExit = onSessionExit(session.id, (code) => {
      _term.write(`\r\n\x1b[90m[proses berakhir — kode ${code}]\x1b[0m\r\n`);
      setRunning(session.id, false);
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

    startPty(_term).then(() => {
      setRunning(session.id, true);
      if (visible) {
        tick().then(() => {
          refit();
          if (active) _term.focus();
        });
      }
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
