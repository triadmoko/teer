<script lang="ts">
  import { onMount } from "svelte";
  import Sidebar from "./lib/Sidebar.svelte";
  import TabBar from "./lib/TabBar.svelte";
  import Terminal from "./lib/Terminal.svelte";
  import Dialog from "./lib/Dialog.svelte";
  import type { SessionDef, Workspace } from "../bindings/teer/internal/domain/models";
  import {
    init,
    activeWorkspace,
    activeSessionId,
    opened,
    addSession,
    selectSession,
    closeSession,
    lastError,
  } from "./lib/app";

  onMount(() => {
    init();
  });

  function sessionsOf(ws: Workspace | null): SessionDef[] {
    return (ws?.sessions ?? []).filter(Boolean) as SessionDef[];
  }

  const aw = $derived($activeWorkspace);
  const allSessions = $derived(sessionsOf(aw));
  const openSessions = $derived(allSessions.filter((s) => $opened.has(s.id)));
  const awEnv = $derived((aw?.env ?? {}) as Record<string, string>);
  const awCwd = $derived(aw?.defaultCwd ?? "");

  // Shortcut keyboard dasar (FR-23): Ctrl+T terminal baru, Ctrl+W tutup,
  // Ctrl+Tab pindah tab.
  function onKey(e: KeyboardEvent) {
    if (!aw) return;
    if (e.ctrlKey && (e.key === "t" || e.key === "T")) {
      e.preventDefault();
      addSession(aw.id);
    } else if (e.ctrlKey && (e.key === "w" || e.key === "W")) {
      e.preventDefault();
      const cur = allSessions.find((s) => s.id === $activeSessionId);
      if (cur) closeSession(cur);
    } else if (e.ctrlKey && e.key === "Tab") {
      e.preventDefault();
      if (allSessions.length < 2) return;
      const idx = allSessions.findIndex((s) => s.id === $activeSessionId);
      const next = allSessions[(idx + 1) % allSessions.length];
      if (next) selectSession(next.id);
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<Dialog />

{#if $lastError}
  <div class="banner" role="alert">
    ⚠️ {$lastError}
    <button class="banner-x" onclick={() => lastError.set(null)}>×</button>
  </div>
{/if}

<div class="app">
  <Sidebar />
  <main class="main">
    {#if aw}
      <TabBar workspaceId={aw.id} sessions={allSessions} />
      <div class="stage">
        {#each openSessions as s (s.id)}
          <Terminal
            session={s}
            active={s.id === $activeSessionId}
            wsEnv={awEnv}
            wsCwd={awCwd}
          />
        {/each}
        {#if allSessions.length === 0}
          <div class="placeholder">
            Belum ada terminal. Klik <b>+</b> di atas (atau <kbd>Ctrl</kbd>+<kbd
              >T</kbd
            >).
          </div>
        {/if}
      </div>
    {:else}
      <div class="placeholder big">
        Pilih atau buat workspace di sebelah kiri untuk memulai.
      </div>
    {/if}
  </main>
</div>

<style>
  .app {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
  }
  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    background: #121214;
  }
  .stage {
    position: relative;
    flex: 1;
    min-height: 0;
  }
  .placeholder {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #52525b;
    font-size: 14px;
  }
  .placeholder.big {
    font-size: 15px;
  }
  kbd {
    background: #26262b;
    border: 1px solid #34343a;
    border-radius: 4px;
    padding: 1px 5px;
    font-size: 12px;
    color: #d4d4d8;
  }
  .banner {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 1100;
    background: #7f1d1d;
    color: #fee2e2;
    padding: 9px 16px;
    font-size: 13px;
    display: flex;
    align-items: center;
    gap: 10px;
    line-height: 1.4;
  }
  .banner-x {
    margin-left: auto;
    background: none;
    border: none;
    color: #fecaca;
    font-size: 18px;
    cursor: pointer;
    line-height: 1;
  }
  .banner-x:hover {
    color: #fff;
  }
</style>
