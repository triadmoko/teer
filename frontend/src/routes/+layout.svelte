<script lang="ts">
  import "../app.css";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import Sidebar from "@presentation/components/Sidebar.svelte";
  import TabBar from "@presentation/components/TabBar.svelte";
  import Dialog from "@presentation/components/Dialog.svelte";
  import ErrorBanner from "@presentation/components/ErrorBanner.svelte";
  import TerminalStage from "@presentation/components/TerminalStage.svelte";
  import SessionFormDialog from "@presentation/components/SessionFormDialog.svelte";
  import WorkspaceSettingsDialog from "@presentation/components/WorkspaceSettingsDialog.svelte";
  import TerminalSettingsDialog from "@presentation/components/TerminalSettingsDialog.svelte";
  import CommandPalette from "@presentation/components/CommandPalette.svelte";
  import UpdateNotification from "@presentation/components/UpdateNotification.svelte";
  import { sessionsOf, type SessionDef } from "@domain/models";
  import {
    init,
    activeWorkspace,
    activeWorkspaceId,
    activeSessionId,
    workspaces,
    opened,
    running,
    addSession,
    selectSession,
    selectWorkspace,
    closeSession,
    openCommandPalette,
    checkUpdate,
    listenUpdateProgress,
  } from "@application";

  let { children } = $props();

  type OpenEntry = {
    s: SessionDef;
    wsEnv: Record<string, string>;
    wsCwd: string;
    wsStartupCommand: string;
  };

  $effect(() => {
    init();
    checkUpdate();
    return listenUpdateProgress();
  });

  const aw = $derived($activeWorkspace);
  const allSessions = $derived(sessionsOf(aw));

  // Sessions yang dirender sebagai Terminal:
  // - Workspace AKTIF: SEMUA session-nya tampil (cell/tab kosong utk yang belum
  //   jalan), supaya tab mode & grid mode konsisten — keduanya lihat 3 terminal.
  // - Workspace LAIN: hanya yang sudah dibuka/berjalan, supaya Terminal-nya tetap
  //   hidup saat switch workspace (preserves xterm scrollback) tanpa membengkak.
  const allOpenSessions = $derived<OpenEntry[]>(
    $workspaces.flatMap((ws) => {
      const wsEnv = (ws.env ?? {}) as Record<string, string>;
      const wsCwd = ws.defaultCwd ?? "";
      const wsStartupCommand = ws.startupCommand ?? "";
      const isActiveWs = ws.id === $activeWorkspaceId;
      return sessionsOf(ws)
        .filter((s) => isActiveWs || $opened.has(s.id) || $running[s.id])
        .map((s) => ({ s, wsEnv, wsCwd, wsStartupCommand }));
    }),
  );

  // Sinkron store -> URL: saat activeWorkspaceId berubah (klik sidebar,
  // shortcut), mirror ke route. Diff-guard cegah loop dengan +page.svelte.
  $effect(() => {
    const id = $activeWorkspaceId;
    if (id && $page.params.id !== id) {
      goto("/workspace/" + id, {
        replaceState: true,
        keepFocus: true,
        noScroll: true,
      });
    }
  });

  function onKey(e: KeyboardEvent) {
    if (e.ctrlKey && e.shiftKey && (e.key === "p" || e.key === "P")) {
      e.preventDefault();
      openCommandPalette();
      return;
    }
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
<UpdateNotification />
<SessionFormDialog />
<WorkspaceSettingsDialog />
<TerminalSettingsDialog />
<CommandPalette />

<div class="flex h-screen w-screen overflow-hidden">
  <Sidebar />
  <main class="relative flex min-w-0 flex-1 flex-col bg-base">
    {#if aw}
      <TabBar workspaceId={aw.id} sessions={allSessions} />
      <TerminalStage
        {allOpenSessions}
        allSessionsCount={allSessions.length}
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

{@render children()}
