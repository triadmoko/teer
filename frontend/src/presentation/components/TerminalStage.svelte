<script lang="ts">
  import { tick } from "svelte";
  import {
    IconMaximize,
    IconMinimize,
    IconX,
    IconSearch,
    IconChevronUp,
    IconChevronDown,
    IconRefresh,
  } from "@tabler/icons-svelte";
  import type { SessionDef } from "@domain/models";
  import {
    activeSessionId,
    running,
    layoutMode,
    gridCols,
    gridRowH,
    setRowH,
    fullscreenSessionId,
    selectSession,
    renameSession,
    closeSession,
    restartSession,
    promptDialog,
    confirmDialog,
  } from "@application";
  import Terminal from "./Terminal.svelte";

  type OpenEntry = {
    s: SessionDef;
    wsEnv: Record<string, string>;
    wsCwd: string;
    wsStartupCommand: string;
  };

  let {
    openSessions,
    allOpenSessions,
    allSessionsCount,
    awEnv,
    awCwd,
    awStartupCommand = "",
  }: {
    openSessions: SessionDef[];
    allOpenSessions: OpenEntry[];
    allSessionsCount: number;
    awEnv: Record<string, string>;
    awCwd: string;
    awStartupCommand?: string;
  } = $props();

    let terminalRefs = $state<Record<string, { find: (q: string, next?: boolean) => void }>>({});

    let searchOpen = $state(false);
  let searchQuery = $state("");
  let searchInput = $state<HTMLInputElement | undefined>();

  function openSearch() {
    searchOpen = true;
    tick().then(() => searchInput?.focus());
  }

  function closeSearch() {
    searchOpen = false;
    searchQuery = "";
  }

  function doFind(next = true) {
    if (!searchQuery) return;
    const ref = terminalRefs[$activeSessionId ?? ""];
    ref?.find(searchQuery, next);
  }

  function onSearchKey(e: KeyboardEvent) {
    if (e.key === "Escape") {
      e.preventDefault();
      closeSearch();
    } else if (e.key === "Enter") {
      e.preventDefault();
      doFind(!e.shiftKey);
    }
  }

      function onWindowKey(e: KeyboardEvent) {
    if (e.ctrlKey && e.key === "f") {
      e.preventDefault();
      if (searchOpen) closeSearch();
      else openSearch();
    }
  }

        let resizing = $state(false);

  function startResize(e: PointerEvent) {
    e.preventDefault();
    e.stopPropagation();
    const el = e.currentTarget as HTMLElement;
    const startY = e.clientY;
    const startH = $gridRowH;
    el.setPointerCapture(e.pointerId);
    resizing = true;
    const move = (ev: PointerEvent) => {
      setRowH(startH + (ev.clientY - startY));
    };
    const up = (ev: PointerEvent) => {
      resizing = false;
      el.releasePointerCapture(ev.pointerId);
      el.removeEventListener("pointermove", move);
      el.removeEventListener("pointerup", up);
    };
    el.addEventListener("pointermove", move);
    el.addEventListener("pointerup", up);
  }

  async function renameFromCell(s: SessionDef, e: Event) {
    e.stopPropagation();
    const name = await promptDialog("Ganti nama terminal", s.name);
    if (!name || name === s.name) return;
    await renameSession(s, name);
  }

  async function closeFromCell(s: SessionDef, e: MouseEvent) {
    e.stopPropagation();
    if ($running[s.id]) {
      const ok = await confirmDialog(
        `Sesi "${s.name}" masih berjalan. Tutup & hentikan prosesnya?`,
        { confirmLabel: "Tutup", danger: true },
      );
      if (!ok) return;
    }
    await closeSession(s);
  }

  function toggleFullscreen(s: SessionDef, e: MouseEvent) {
    e.stopPropagation();
    fullscreenSessionId.update((cur) => (cur === s.id ? null : s.id));
    selectSession(s.id);
  }

  $effect(() => {
    if ($layoutMode !== "grid") fullscreenSessionId.set(null);
  });
</script>

<svelte:window onkeydown={onWindowKey} />

<div
  class="relative min-h-0 flex-1 {$layoutMode === 'grid'
    ? 'grid auto-rows-[var(--rowH,320px)] grid-cols-[repeat(var(--cols),minmax(0,1fr))] gap-1 overflow-y-auto p-1'
    : ''}"
  style:--cols={$gridCols}
  style:--rowH={`${$gridRowH}px`}
>

  {#if searchOpen && $layoutMode !== "grid"}
    <div
      class="absolute right-3 top-2 z-50 flex items-center gap-1 rounded-lg border border-zinc-700 bg-elevated px-2 py-1 shadow-lg"
    >
      <IconSearch size={13} class="shrink-0 text-zinc-400" />
      <input
        bind:this={searchInput}
        bind:value={searchQuery}
        class="w-[200px] bg-transparent text-[13px] text-zinc-50 outline-none placeholder:text-zinc-600"
        placeholder="Cari di terminal…"
        type="text"
        onkeydown={onSearchKey}
        oninput={() => doFind(true)}
      />
      <button
        class="flex cursor-pointer items-center rounded border-none bg-transparent px-1 text-zinc-400 hover:text-zinc-50"
        title="Sebelumnya (Shift+Enter)"
        onclick={() => doFind(false)}><IconChevronUp size={13} /></button
      >
      <button
        class="flex cursor-pointer items-center rounded border-none bg-transparent px-1 text-zinc-400 hover:text-zinc-50"
        title="Berikutnya (Enter)"
        onclick={() => doFind(true)}><IconChevronDown size={13} /></button
      >
      <button
        class="flex cursor-pointer items-center rounded border-none bg-transparent px-1 text-zinc-500 hover:text-red-400"
        title="Tutup (Escape)"
        onclick={closeSearch}><IconX size={13} /></button
      >
    </div>
  {/if}

  {#if $layoutMode === "grid"}
    {#each openSessions as s (s.id)}
      {@const isFs = $fullscreenSessionId === s.id}
      <div
        class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-md border bg-base {s.id ===
        $activeSessionId
          ? 'border-blue-500 shadow-[0_0_0_1px_#3b82f6]'
          : 'border-line'} {isFs ? 'absolute inset-0 z-50' : ''}"
      >
        <div
          class="flex h-7 shrink-0 cursor-pointer select-none items-center gap-[7px] border-b border-line pl-[10px] pr-[6px] text-xs {s.id ===
          $activeSessionId
            ? 'bg-elevated text-zinc-50'
            : 'bg-surface text-zinc-400'}"
          role="button"
          tabindex="0"
          title="dobel-klik untuk ganti nama"
          onclick={() => selectSession(s.id)}
          onkeydown={(e) => e.key === "Enter" && selectSession(s.id)}
          ondblclick={(e) => renameFromCell(s, e)}
        >
          <span
            class="h-[7px] w-[7px] shrink-0 rounded-full {$running[s.id]
              ? 'bg-green-400 shadow-[0_0_6px_#4ade80]'
              : 'bg-zinc-600'}"
          ></span>
          <span
            class="min-w-0 flex-1 overflow-hidden text-ellipsis whitespace-nowrap"
            >{s.name}</span
          >
          {#if !$running[s.id]}
            <button
              class="flex shrink-0 cursor-pointer items-center rounded border-none bg-transparent px-1 leading-none text-zinc-500 hover:bg-zinc-700 hover:text-green-400"
              title="Restart"
              onclick={(e) => { e.stopPropagation(); restartSession(s.id); }}
            ><IconRefresh size={12} /></button>
          {/if}
          <button
            class="flex shrink-0 cursor-pointer items-center rounded border-none bg-transparent px-1 leading-none text-zinc-500 hover:bg-zinc-700 hover:text-zinc-50"
            title={isFs ? "Kembalikan ukuran" : "Layar penuh"}
            aria-label={isFs
              ? "Kembalikan ukuran terminal"
              : "Perluas terminal layar penuh"}
            onclick={(e) => toggleFullscreen(s, e)}
          >
            {#if isFs}
              <IconMinimize size={12} />
            {:else}
              <IconMaximize size={12} />
            {/if}
          </button>
          <button
            class="flex shrink-0 cursor-pointer items-center rounded border-none bg-transparent px-1 leading-none text-zinc-500 hover:bg-zinc-700 hover:text-zinc-50"
            title="Tutup"
            aria-label="Tutup terminal"
            onclick={(e) => closeFromCell(s, e)}><IconX size={12} /></button
          >
        </div>
        <div
          class="relative min-h-0 flex-1"
          role="presentation"
          onclick={() => selectSession(s.id)}
        >
          <Terminal
            session={s}
            active={s.id === $activeSessionId}
            visible={true}
            wsEnv={awEnv}
            wsCwd={awCwd}
            wsStartupCommand={awStartupCommand}
            bind:this={terminalRefs[s.id]}
          />
        </div>
        <div
          class="h-[7px] shrink-0 cursor-ns-resize touch-none border-t border-line {resizing
            ? 'bg-blue-500'
            : 'bg-surface hover:bg-blue-500'}"
          role="separator"
          aria-label="Tarik untuk ubah tinggi terminal"
          title="Tarik untuk ubah tinggi"
          onpointerdown={startResize}
        ></div>
      </div>
    {/each}
  {:else}
    {#each allOpenSessions as entry (entry.s.id)}
      <Terminal
        session={entry.s}
        active={entry.s.id === $activeSessionId}
        visible={entry.s.id === $activeSessionId}
        wsEnv={entry.wsEnv}
        wsCwd={entry.wsCwd}
        wsStartupCommand={entry.wsStartupCommand}
        bind:this={terminalRefs[entry.s.id]}
      />
    {/each}
  {/if}
  {#if allSessionsCount === 0}
    <div
      class="absolute inset-0 flex items-center justify-center text-sm text-zinc-600"
    >
      Belum ada terminal. Klik <b>+</b> di atas (atau <kbd
        class="rounded border border-line-3 bg-active px-[5px] py-px text-xs text-zinc-300"
        >Ctrl</kbd
      >+<kbd
        class="rounded border border-line-3 bg-active px-[5px] py-px text-xs text-zinc-300"
        >T</kbd
      >).
    </div>
  {/if}
</div>
