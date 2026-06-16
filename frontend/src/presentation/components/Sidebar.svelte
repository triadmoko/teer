<script lang="ts">
  import { IconCopy, IconSettings2, IconTrash, IconPlus } from "@tabler/icons-svelte";
  import type { Workspace } from "@domain/models";
  import {
    workspaces,
    activeWorkspaceId,
    running,
    selectWorkspace,
    createWorkspace,
    editWorkspace,
    deleteWorkspace,
    duplicateWorkspace,
    reorderWorkspaces,
    confirmDialog,
    openWorkspaceSettings,
  } from "@application";

  const palette = ["#60a5fa", "#4ade80", "#facc15", "#f87171", "#c084fc", "#22d3ee"];

  // ---- Drag-and-drop reorder (FR-6) ----
  let dragId = $state<string | null>(null);
  let dragOverId = $state<string | null>(null);

  function onDragStart(e: DragEvent, ws: Workspace) {
    dragId = ws.id;
    e.dataTransfer!.effectAllowed = "move";
  }

  function onDragOver(e: DragEvent, ws: Workspace) {
    if (dragId === ws.id) return;
    e.preventDefault();
    e.dataTransfer!.dropEffect = "move";
    dragOverId = ws.id;
  }

  function onDragLeave() {
    dragOverId = null;
  }

  async function onDrop(e: DragEvent, ws: Workspace) {
    e.preventDefault();
    dragOverId = null;
    if (!dragId || dragId === ws.id) { dragId = null; return; }
    const list = $workspaces;
    const fromIdx = list.findIndex((w) => w?.id === dragId);
    const toIdx = list.findIndex((w) => w?.id === ws.id);
    if (fromIdx === -1 || toIdx === -1) { dragId = null; return; }
    const reordered = [...list];
    const [item] = reordered.splice(fromIdx, 1);
    reordered.splice(toIdx, 0, item);
    dragId = null;
    await reorderWorkspaces(reordered.map((w) => w.id));
  }

  function onDragEnd() {
    dragId = null;
    dragOverId = null;
  }

  // Apakah workspace punya minimal satu sesi yang sedang berjalan (indikator titik).
  function hasRunning(ws: Workspace, run: Record<string, boolean>): boolean {
    return (ws.sessions ?? []).some((s) => s && run[s.id]);
  }

  async function onNew() {
    const color = palette[$workspaces.length % palette.length];
    const empty = { id: "", name: "", color, defaultCwd: "", env: {}, sessions: [], createdAt: null, updatedAt: null } as unknown;
    const result = await openWorkspaceSettings(empty as Workspace);
    if (!result || !result.name) return;
    await createWorkspace(result.name, result.color || color, result.defaultCwd);
  }

  async function onEdit(ws: Workspace, e: MouseEvent) {
    e.stopPropagation();
    await editWorkspace(ws);
  }

  async function onDelete(ws: Workspace, e: MouseEvent) {
    e.stopPropagation();
    const ok = await confirmDialog(
      `Hapus workspace "${ws.name}" beserta semua sesinya?`,
      { confirmLabel: "Hapus", danger: true },
    );
    if (!ok) return;
    await deleteWorkspace(ws.id);
  }

  async function onDuplicate(ws: Workspace, e: MouseEvent) {
    e.stopPropagation();
    await duplicateWorkspace(ws.id);
  }
</script>

<aside
  class="flex h-full w-[220px] min-w-[220px] flex-col border-r border-line bg-surface"
>
  <div
    class="border-b border-line px-4 py-[14px] font-bold tracking-[0.5px] text-zinc-200"
  >
    teer
  </div>

  <div class="flex-1 overflow-y-auto p-2">
    {#each $workspaces as ws (ws.id)}
      <div
        class="group flex cursor-pointer select-none items-center gap-2 rounded-lg px-[10px] py-2 {ws.id ===
        $activeWorkspaceId
          ? 'bg-active text-zinc-50'
          : 'text-zinc-400 hover:bg-raise hover:text-zinc-200'} {dragOverId === ws.id
          ? 'ring-1 ring-blue-400'
          : ''}"
        role="button"
        tabindex="0"
        draggable="true"
        onclick={() => selectWorkspace(ws.id)}
        onkeydown={(e) => e.key === "Enter" && selectWorkspace(ws.id)}
        ondragstart={(e) => onDragStart(e, ws)}
        ondragover={(e) => onDragOver(e, ws)}
        ondragleave={onDragLeave}
        ondrop={(e) => onDrop(e, ws)}
        ondragend={onDragEnd}
      >
        <span
          class="h-[9px] w-[9px] shrink-0 rounded-full"
          style="background:{ws.color || '#52525b'}"
        ></span>
        <span
          class="flex-1 overflow-hidden text-ellipsis whitespace-nowrap text-[13px]"
          title={ws.name}>{ws.name}</span
        >
        {#if hasRunning(ws, $running)}
          <span
            class="h-[7px] w-[7px] shrink-0 rounded-full bg-green-400 shadow-[0_0_6px_#4ade80]"
            title="ada sesi berjalan"
          ></span>
        {/if}
        <span class="hidden gap-[2px] group-hover:flex">
          <button
            class="cursor-pointer rounded border-none bg-transparent px-[3px] py-[2px] text-zinc-500 hover:bg-line-3 hover:text-zinc-50"
            title="Duplikat"
            onclick={(e) => onDuplicate(ws, e)}><IconCopy size={13} /></button
          >
          <button
            class="cursor-pointer rounded border-none bg-transparent px-[3px] py-[2px] text-zinc-500 hover:bg-line-3 hover:text-zinc-50"
            title="Pengaturan"
            onclick={(e) => onEdit(ws, e)}><IconSettings2 size={13} /></button
          >
          <button
            class="cursor-pointer rounded border-none bg-transparent px-[3px] py-[2px] text-zinc-500 hover:bg-line-3 hover:text-zinc-50"
            title="Hapus"
            onclick={(e) => onDelete(ws, e)}><IconTrash size={13} /></button
          >
        </span>
      </div>
    {/each}

    {#if $workspaces.length === 0}
      <div class="px-[10px] py-3 text-xs text-zinc-600">Belum ada workspace.</div>
    {/if}
  </div>

  <button
    class="m-2 flex cursor-pointer items-center justify-center gap-2 rounded-lg border border-line-2 bg-raise p-[9px] text-[13px] text-zinc-300 hover:bg-active hover:text-white"
    onclick={onNew}><IconPlus size={14} /> Workspace baru</button
  >
</aside>
