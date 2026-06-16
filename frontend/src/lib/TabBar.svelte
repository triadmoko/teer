<script lang="ts">
  import type { SessionDef } from "../../bindings/teer/internal/domain/models";
  import {
    activeSessionId,
    running,
    selectSession,
    addSession,
    renameSession,
    closeSession,
  } from "./app";
  import { promptDialog, confirmDialog } from "./dialog";

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

<div class="tabbar">
  <div class="tabs">
    {#each sessions as s (s.id)}
      <div
        class="tab"
        class:active={s.id === $activeSessionId}
        role="button"
        tabindex="0"
        onclick={() => selectSession(s.id)}
        onkeydown={(e) => e.key === "Enter" && selectSession(s.id)}
        ondblclick={(e) => onRename(s, e)}
      >
        <span class="status" class:run={$running[s.id]}></span>
        <span class="label" title="dobel-klik untuk ganti nama">{s.name}</span>
        <button class="x" title="Tutup" onclick={(e) => onClose(s, e)}>×</button>
      </div>
    {/each}
    <button class="add" title="Terminal baru" onclick={() => addSession(workspaceId)}
      >+</button
    >
  </div>
</div>

<style>
  .tabbar {
    height: 38px;
    background: #161618;
    border-bottom: 1px solid #232327;
    display: flex;
    align-items: stretch;
    overflow: hidden;
  }
  .tabs {
    display: flex;
    align-items: stretch;
    overflow-x: auto;
    width: 100%;
  }
  .tab {
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 0 10px 0 12px;
    border-right: 1px solid #232327;
    cursor: pointer;
    color: #a1a1aa;
    font-size: 13px;
    white-space: nowrap;
    user-select: none;
  }
  .tab:hover {
    background: #1f1f23;
    color: #e4e4e7;
  }
  .tab.active {
    background: #121214;
    color: #fafafa;
  }
  .status {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #52525b;
    flex-shrink: 0;
  }
  .status.run {
    background: #4ade80;
    box-shadow: 0 0 6px #4ade80;
  }
  .label {
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .x {
    background: none;
    border: none;
    color: #71717a;
    cursor: pointer;
    font-size: 16px;
    line-height: 1;
    padding: 0 2px;
    border-radius: 4px;
  }
  .x:hover {
    color: #fafafa;
    background: #3f3f46;
  }
  .add {
    background: none;
    border: none;
    color: #a1a1aa;
    cursor: pointer;
    font-size: 18px;
    padding: 0 14px;
  }
  .add:hover {
    color: #fafafa;
    background: #1f1f23;
  }
</style>
