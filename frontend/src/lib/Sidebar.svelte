<script lang="ts">
  import type { Workspace } from "../../bindings/teer/internal/domain/models";
  import {
    workspaces,
    activeWorkspaceId,
    running,
    selectWorkspace,
    createWorkspace,
    renameWorkspace,
    deleteWorkspace,
    duplicateWorkspace,
  } from "./app";
  import { promptDialog, confirmDialog } from "./dialog";

  const palette = ["#60a5fa", "#4ade80", "#facc15", "#f87171", "#c084fc", "#22d3ee"];

  // Apakah workspace punya minimal satu sesi yang sedang berjalan (indikator titik).
  function hasRunning(ws: Workspace, run: Record<string, boolean>): boolean {
    return (ws.sessions ?? []).some((s) => s && run[s.id]);
  }

  async function onNew() {
    const name = await promptDialog("Nama workspace baru", "", "mis. Proyek Codemi");
    if (!name) return;
    const color = palette[$workspaces.length % palette.length];
    await createWorkspace(name, color, "");
  }

  async function onRename(ws: Workspace, e: MouseEvent) {
    e.stopPropagation();
    const name = await promptDialog("Ganti nama workspace", ws.name);
    if (!name || name === ws.name) return;
    await renameWorkspace(ws, name);
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

<aside class="sidebar">
  <div class="brand">teer</div>

  <div class="ws-list">
    {#each $workspaces as ws (ws.id)}
      <div
        class="ws"
        class:active={ws.id === $activeWorkspaceId}
        role="button"
        tabindex="0"
        onclick={() => selectWorkspace(ws.id)}
        onkeydown={(e) => e.key === "Enter" && selectWorkspace(ws.id)}
      >
        <span class="dot" style="background:{ws.color || '#52525b'}"></span>
        <span class="name" title={ws.name}>{ws.name}</span>
        {#if hasRunning(ws, $running)}
          <span class="live" title="ada sesi berjalan"></span>
        {/if}
        <span class="actions">
          <button title="Duplikat" onclick={(e) => onDuplicate(ws, e)}>⧉</button>
          <button title="Ganti nama" onclick={(e) => onRename(ws, e)}>✎</button>
          <button title="Hapus" onclick={(e) => onDelete(ws, e)}>🗑</button>
        </span>
      </div>
    {/each}

    {#if $workspaces.length === 0}
      <div class="empty">Belum ada workspace.</div>
    {/if}
  </div>

  <button class="new" onclick={onNew}>+ Workspace baru</button>
</aside>

<style>
  .sidebar {
    width: 220px;
    min-width: 220px;
    background: #161618;
    border-right: 1px solid #232327;
    display: flex;
    flex-direction: column;
    height: 100%;
    box-sizing: border-box;
  }
  .brand {
    font-weight: 700;
    letter-spacing: 0.5px;
    padding: 14px 16px;
    color: #e4e4e7;
    border-bottom: 1px solid #232327;
  }
  .ws-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }
  .ws {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border-radius: 8px;
    cursor: pointer;
    color: #a1a1aa;
    user-select: none;
  }
  .ws:hover {
    background: #1f1f23;
    color: #e4e4e7;
  }
  .ws.active {
    background: #26262b;
    color: #fafafa;
  }
  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
  }
  .live {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #4ade80;
    box-shadow: 0 0 6px #4ade80;
    flex-shrink: 0;
  }
  .actions {
    display: none;
    gap: 2px;
  }
  .ws:hover .actions {
    display: flex;
  }
  .actions button {
    background: none;
    border: none;
    color: #71717a;
    cursor: pointer;
    font-size: 12px;
    padding: 2px 3px;
    border-radius: 4px;
  }
  .actions button:hover {
    color: #fafafa;
    background: #34343a;
  }
  .empty {
    color: #52525b;
    font-size: 12px;
    padding: 12px 10px;
  }
  .new {
    margin: 8px;
    padding: 9px;
    background: #1f1f23;
    border: 1px solid #2c2c31;
    color: #d4d4d8;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
  }
  .new:hover {
    background: #26262b;
    color: #fff;
  }
</style>
