<script lang="ts">
  import { IconMaximize, IconMinimize, IconX } from "@tabler/icons-svelte";
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
    promptDialog,
    confirmDialog,
  } from "@application";
  import Terminal from "./Terminal.svelte";

  let {
    openSessions,
    allSessionsCount,
    awEnv,
    awCwd,
  }: {
    openSessions: SessionDef[];
    allSessionsCount: number;
    awEnv: Record<string, string>;
    awCwd: string;
  } = $props();

  // Drag tepi bawah window grid untuk mengubah tinggi baris (semua baris grid
  // seragam, jadi menarik satu window menyetel tinggi grid). Pointer capture
  // memastikan event tetap diterima walau kursor melintas di atas xterm.
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

  // Rename / tutup terminal langsung dari header window grid (mirror TabBar).
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

  // Reset fullscreen saat keluar dari mode grid.
  $effect(() => {
    if ($layoutMode !== "grid") fullscreenSessionId.set(null);
  });
</script>

<div
  class="relative min-h-0 flex-1 {$layoutMode === 'grid'
    ? 'grid auto-rows-[var(--rowH,320px)] grid-cols-[repeat(var(--cols),minmax(0,1fr))] gap-1 overflow-y-auto p-1'
    : ''}"
  style:--cols={$gridCols}
  style:--rowH={`${$gridRowH}px`}
>
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
          <button
            class="flex shrink-0 cursor-pointer items-center rounded border-none bg-transparent px-1 leading-none text-zinc-500 hover:bg-zinc-700 hover:text-zinc-50"
            title={isFs ? "Kembalikan ukuran" : "Layar penuh"}
            aria-label={isFs ? "Kembalikan ukuran terminal" : "Perluas terminal layar penuh"}
            onclick={(e) => toggleFullscreen(s, e)}>
            {#if isFs}
              <IconMinimize size={12} />
            {:else}
              <IconMaximize size={12} />
            {/if}
          </button
          >
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
    {#each openSessions as s (s.id)}
      <Terminal
        session={s}
        active={s.id === $activeSessionId}
        visible={s.id === $activeSessionId}
        wsEnv={awEnv}
        wsCwd={awCwd}
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
