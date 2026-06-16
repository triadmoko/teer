<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import { Terminal } from "@xterm/xterm";
  import { FitAddon } from "@xterm/addon-fit";
  import { SearchAddon } from "@xterm/addon-search";
  import { WebLinksAddon } from "@xterm/addon-web-links";
  import "@xterm/xterm/css/xterm.css";
  import { Events } from "@wailsio/runtime";
  import { SessionService } from "../../bindings/teer/internal/service";
  import type { SessionDef } from "../../bindings/teer/internal/domain/models";
  import { setRunning } from "./app";

  let {
    session,
    active = false,
    wsEnv = {},
    wsCwd = "",
  }: {
    session: SessionDef;
    active?: boolean;
    wsEnv?: Record<string, string>;
    wsCwd?: string;
  } = $props();

  let container = $state<HTMLDivElement>();
  let term: Terminal;
  let fit: FitAddon;
  let search: SearchAddon;
  let cleanups: Array<() => void> = [];
  let ro: ResizeObserver | undefined;
  let fitTimer: ReturnType<typeof setTimeout> | undefined;

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

  function b64ToBytes(b64: string): Uint8Array {
    const bin = atob(b64);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return bytes;
  }

  // Debounce fit untuk menghindari resize race saat container berubah cepat
  // (PRD §11: resize PTY saat output streaming).
  function refit() {
    if (fitTimer) clearTimeout(fitTimer);
    fitTimer = setTimeout(() => {
      if (!active) return;
      try {
        fit.fit();
      } catch {
        /* container belum punya dimensi */
      }
    }, 50);
  }

  async function startPty() {
    const env = { ...wsEnv, ...(session.env ?? {}) } as Record<string, string>;
    await SessionService.StartSession({
      id: session.id,
      shell: session.shell ?? "",
      cwd: session.cwd || wsCwd,
      env,
      startupCommand: session.startupCommand ?? "",
      cols: term.cols,
      rows: term.rows,
    } as any);
  }

  onMount(async () => {
    term = new Terminal({
      fontFamily: 'ui-monospace, "Cascadia Code", "JetBrains Mono", Menlo, monospace',
      fontSize: 13,
      cursorBlink: true,
      scrollback: 5000, // batas sesuai PRD NFR-1 / §11
      theme,
      allowProposedApi: true,
    });
    fit = new FitAddon();
    search = new SearchAddon();
    term.loadAddon(fit);
    term.loadAddon(search);
    term.loadAddon(new WebLinksAddon());

    term.open(container!);
    try {
      fit.fit();
    } catch {
      /* noop */
    }

    // Input keyboard/paste → stdin shell (FR-18).
    const onData = term.onData((data) =>
      SessionService.WriteSession(session.id, data),
    );
    cleanups.push(() => onData.dispose());

    // xterm resize → PTY resize (FR-17).
    const onResize = term.onResize(({ cols, rows }) =>
      SessionService.ResizeSession(session.id, cols, rows),
    );
    cleanups.push(() => onResize.dispose());

    // Output PTY (base64) → tulis byte mentah ke xterm.
    const offOut = Events.On(`session:${session.id}:out`, (ev) => {
      term.write(b64ToBytes(ev.data as string));
    });
    cleanups.push(offOut);

    // Sesi berakhir → tampilkan penanda + lapor status (FR-14).
    const offExit = Events.On(`session:${session.id}:exit`, (ev) => {
      const code = (ev.data as { code?: number })?.code ?? 0;
      term.write(`\r\n\x1b[90m[proses berakhir — kode ${code}]\x1b[0m\r\n`);
      setRunning(session.id, false);
    });
    cleanups.push(offExit);

    ro = new ResizeObserver(() => refit());
    ro.observe(container!);

    await startPty();
    setRunning(session.id, true);
    if (active) {
      await tick();
      refit();
      term.focus();
    }
  });

  // Saat tab menjadi aktif, refit & fokus.
  $effect(() => {
    if (active && term) {
      tick().then(() => {
        refit();
        term.focus();
      });
    }
  });

  /** Dipanggil parent untuk mencari teks di scrollback (FR-19). */
  export function find(query: string, next = true) {
    if (!search || !query) return;
    if (next) search.findNext(query);
    else search.findPrevious(query);
  }

  onDestroy(() => {
    cleanups.forEach((fn) => fn());
    ro?.disconnect();
    if (fitTimer) clearTimeout(fitTimer);
    term?.dispose();
  });
</script>

<div class="term" class:active bind:this={container}></div>

<style>
  .term {
    position: absolute;
    inset: 0;
    padding: 6px 8px;
    box-sizing: border-box;
    display: none;
  }
  .term.active {
    display: block;
  }
  .term :global(.xterm) {
    height: 100%;
  }
  .term :global(.xterm-viewport) {
    /* sembunyikan scrollbar default agar bersih */
    scrollbar-width: thin;
  }
</style>
