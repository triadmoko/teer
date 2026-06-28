<script lang="ts">
  import "../app.css";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import Sidebar from "@presentation/components/Sidebar.svelte";
  import TabBar from "@presentation/components/TabBar.svelte";
  import TerminalStage from "@presentation/components/TerminalStage.svelte";
  import AppOverlays from "@presentation/components/AppOverlays.svelte";
  import AppKeyboard from "@presentation/components/AppKeyboard.svelte";
  import { sessionsOf, type SessionDef } from "@domain/models";
  import {
    init,
    activeWorkspace,
    activeWorkspaceId,
    workspaces,
    opened,
    running,
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
</script>

<AppKeyboard />
<AppOverlays />

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
