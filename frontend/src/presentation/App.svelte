<script lang="ts">
  import Sidebar from "./components/Sidebar.svelte";
  import TabBar from "./components/TabBar.svelte";
  import Dialog from "./components/Dialog.svelte";
  import ErrorBanner from "./components/ErrorBanner.svelte";
  import TerminalStage from "./components/TerminalStage.svelte";
  import SessionFormDialog from "./components/SessionFormDialog.svelte";
  import WorkspaceSettingsDialog from "./components/WorkspaceSettingsDialog.svelte";
  import TerminalSettingsDialog from "./components/TerminalSettingsDialog.svelte";
  import CommandPalette from "./components/CommandPalette.svelte";
  import { sessionsOf } from "@domain/models";
  import {
    init,
    activeWorkspace,
    activeSessionId,
    workspaces,
    opened,
    addSession,
    selectSession,
    selectWorkspace,
    closeSession,
    openCommandPalette,
  } from "@application";

  $effect(() => {
    init();
  });

  const aw = $derived($activeWorkspace);
  const allSessions = $derived(sessionsOf(aw));
  const openSessions = $derived(allSessions.filter((s) => $opened.has(s.id)));
  const awEnv = $derived((aw?.env ?? {}) as Record<string, string>);
  const awCwd = $derived(aw?.defaultCwd ?? "");

  // Shortcut keyboard (FR-23):
  // Terminal: Ctrl+T baru, Ctrl+W tutup, Ctrl+Tab pindah tab.
  // Workspace: Ctrl+Shift+1..9 aktifkan workspace ke-N.
  function onKey(e: KeyboardEvent) {
    // Command palette (FR-24): Ctrl+Shift+P.
    if (e.ctrlKey && e.shiftKey && (e.key === "p" || e.key === "P")) {
      e.preventDefault();
      openCommandPalette();
      return;
    }
    // Pindah workspace ke-N (Ctrl+Shift+1..9).
    if (e.ctrlKey && e.shiftKey && e.key >= "1" && e.key <= "9") {
      e.preventDefault();
      const idx = parseInt(e.key) - 1;
      const ws = $workspaces[idx];
      if (ws) selectWorkspace(ws.id);
      return;
    }
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
<ErrorBanner />
<SessionFormDialog />
<WorkspaceSettingsDialog />
<TerminalSettingsDialog />
<CommandPalette />

<div class="flex h-screen w-screen overflow-hidden">
  <Sidebar />
  <main class="flex min-w-0 flex-1 flex-col bg-base">
    {#if aw}
      <TabBar workspaceId={aw.id} sessions={allSessions} />
      <TerminalStage
        {openSessions}
        allSessionsCount={allSessions.length}
        {awEnv}
        {awCwd}
      />
    {:else}
      <div
        class="absolute inset-0 flex items-center justify-center text-[15px] text-zinc-600"
      >
        Pilih atau buat workspace di sebelah kiri untuk memulai.
      </div>
    {/if}
  </main>
</div>
