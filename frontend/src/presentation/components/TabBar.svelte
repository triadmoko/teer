<script lang="ts">
  import {
    IconX, IconPlus, IconRectangle, IconLayoutGrid, IconRefresh,
    IconBroadcast, IconSettings,
  } from "@tabler/icons-svelte";
  import type { SessionDef } from "@domain/models";
  import { COL_CHOICES } from "@domain/layout";
  import {
    activeSessionId,
    running,
    selectSession,
    addSession,
    renameSession,
    closeSession,
    restartSession,
    layoutMode,
    gridCols,
    broadcastMode,
    openTerminalSettings,
    promptDialog,
    confirmDialog,
  } from "@application";

  let { workspaceId, sessions = [] }: { workspaceId: string; sessions?: SessionDef[] } =
    $props();

  async function onClose(s: SessionDef, e: MouseEvent) {
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

  async function onRename(s: SessionDef, e: MouseEvent) {
    e.stopPropagation();
    const name = await promptDialog("Ganti nama terminal", s.name);
    if (!name || name === s.name) return;
    await renameSession(s, name);
  }
</script>

<div
  class="flex h-[38px] items-stretch overflow-hidden border-b border-line bg-surface"
>
  <div class="flex min-w-0 flex-1 items-stretch overflow-x-auto">
    {#if $layoutMode === "tabs"}
      {#each sessions as s (s.id)}
        <div
          class="flex cursor-pointer select-none items-center gap-[7px] whitespace-nowrap border-r border-line pl-3 pr-[10px] text-[13px] {s.id ===
          $activeSessionId
            ? 'bg-base text-zinc-50'
            : 'text-zinc-400 hover:bg-raise hover:text-zinc-200'}"
          role="button"
          tabindex="0"
          onclick={() => selectSession(s.id)}
          onkeydown={(e) => e.key === "Enter" && selectSession(s.id)}
          ondblclick={(e) => onRename(s, e)}
        >
          <span
            class="h-[7px] w-[7px] shrink-0 rounded-full {$running[s.id]
              ? 'bg-green-400 shadow-[0_0_6px_#4ade80]'
              : 'bg-zinc-600'}"
          ></span>
          <span
            class="max-w-40 overflow-hidden text-ellipsis"
            title="dobel-klik untuk ganti nama">{s.name}</span
          >
          {#if !$running[s.id]}
            <button
              class="flex cursor-pointer items-center rounded border-none bg-transparent px-[2px] text-zinc-500 hover:bg-zinc-700 hover:text-green-400"
              title="Restart"
              onclick={(e) => { e.stopPropagation(); restartSession(s.id); }}
            ><IconRefresh size={12} /></button>
          {/if}
          <button
            class="flex cursor-pointer items-center rounded border-none bg-transparent px-[2px] text-zinc-500 hover:bg-zinc-700 hover:text-zinc-50"
            title="Tutup"
            onclick={(e) => onClose(s, e)}><IconX size={13} /></button
          >
        </div>
      {/each}
    {/if}
    <button
      class="flex cursor-pointer items-center border-none bg-transparent px-[14px] text-zinc-400 hover:bg-raise hover:text-zinc-50"
      title="Terminal baru"
      onclick={() => addSession(workspaceId)}><IconPlus size={16} /></button
    >
  </div>

  <div class="flex shrink-0 items-center gap-[2px] border-l border-line px-2">
    <!-- Broadcast toggle (FR-16) -->
    <button
      class="flex min-w-[26px] cursor-pointer items-center justify-center rounded-[5px] border px-[7px] py-1 {$broadcastMode
        ? 'border-orange-500 bg-orange-500/20 text-orange-300'
        : 'border-transparent bg-transparent text-zinc-400 hover:bg-raise hover:text-zinc-50'}"
      title={$broadcastMode ? 'Broadcast aktif — klik untuk nonaktifkan' : 'Broadcast input ke semua terminal (FR-16)'}
      aria-label="Toggle broadcast input"
      onclick={() => broadcastMode.update((v) => !v)}
    ><IconBroadcast size={14} /></button>

    <span class="mx-1 h-[18px] w-px bg-line"></span>

    <!-- Layout: tabs / grid -->
    <button
      class="flex min-w-[26px] cursor-pointer items-center justify-center rounded-[5px] border px-[7px] py-1 {$layoutMode ===
      'tabs'
        ? 'border-zinc-700 bg-active text-zinc-50'
        : 'border-transparent bg-transparent text-zinc-400 hover:bg-raise hover:text-zinc-50'}"
      title="Mode tab"
      aria-label="Mode tab"
      onclick={() => layoutMode.set("tabs")}><IconRectangle size={14} /></button
    >
    <button
      class="flex min-w-[26px] cursor-pointer items-center justify-center rounded-[5px] border px-[7px] py-1 {$layoutMode ===
      'grid'
        ? 'border-zinc-700 bg-active text-zinc-50'
        : 'border-transparent bg-transparent text-zinc-400 hover:bg-raise hover:text-zinc-50'}"
      title="Mode grid"
      aria-label="Mode grid"
      onclick={() => layoutMode.set("grid")}><IconLayoutGrid size={14} /></button
    >
    {#if $layoutMode === "grid"}
      <span class="mx-1 h-[18px] w-px bg-line"></span>
      {#each COL_CHOICES as c (c)}
        <button
          class="min-w-[26px] cursor-pointer rounded-[5px] border px-[7px] py-1 text-[13px] leading-none {$gridCols ===
          c
            ? 'border-zinc-700 bg-active text-zinc-50'
            : 'border-transparent bg-transparent text-zinc-400 hover:bg-raise hover:text-zinc-50'}"
          title={`${c} kolom`}
          onclick={() => gridCols.set(c)}>{c}</button
        >
      {/each}
    {/if}

    <span class="mx-1 h-[18px] w-px bg-line"></span>

    <!-- Terminal settings (FR-20) -->
    <button
      class="flex min-w-[26px] cursor-pointer items-center justify-center rounded-[5px] border border-transparent bg-transparent px-[7px] py-1 text-zinc-400 hover:bg-raise hover:text-zinc-50"
      title="Pengaturan terminal (font, tema)"
      aria-label="Pengaturan terminal"
      onclick={openTerminalSettings}
    ><IconSettings size={14} /></button>
  </div>
</div>
