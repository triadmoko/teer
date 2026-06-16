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
    let exited = $state(false);

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

  $effect(() => {
        const el = container;
    if (!el) return;

    const _term = new Terminal({
      fontFamily: untrack(() => get(terminalFontFamily)),
      fontSize: untrack(() => get(terminalFontSize)),
      cursorBlink: true,
      scrollback: 5000,
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

    }

            const onData = _term.onData((data) => {
      if (get(broadcastMode)) broadcastWrite(data);
      else writeSession(session.id, data);
    });

        const onResize = _term.onResize(({ cols, rows }) =>
      resizeSession(session.id, cols, rows),
    );

                const sid = untrack(() => session.id);

        const offOut = onSessionOutput(sid, (bytes) => _term.write(bytes));

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

                untrack(() => startPty(_term)).then(() => {
      setRunning(sid, true);
      untrack(() => { exited = false; });
      tick().then(() => {
                                                try { _fit.fit(); } catch {  }
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

<style>

  div :global(.xterm) {
    height: 100%;
  }
  div :global(.xterm-viewport) {

    scrollbar-width: thin;
  }
</style>
