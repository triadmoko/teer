<script lang="ts">
  import { page } from "$app/stores";
  import { get } from "svelte/store";
  import TabBar from "@presentation/components/TabBar.svelte";
  import TerminalStage from "@presentation/components/TerminalStage.svelte";
  import {
    activeWorkspace,
    activeWorkspaceId,
    workspaces,
    opened,
    running,
    selectWorkspace,
  } from "@application";
  import { sessionsOf, type SessionDef } from "@domain/models";

  type OpenEntry = {
    s: SessionDef;
    wsEnv: Record<string, string>;
    wsCwd: string;
    wsStartupCommand: string;
  };

  // URL → store. Diff-guard cegah loop dengan sync store→URL di +layout.svelte.
  $effect(() => {
    const id = $page.params.id;
    if (id && get(activeWorkspaceId) !== id) selectWorkspace(id);
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
</script>

{#if aw}
  <TabBar workspaceId={aw.id} sessions={allSessions} />
  <TerminalStage
    {allOpenSessions}
    allSessionsCount={allSessions.length}
  />
{/if}
