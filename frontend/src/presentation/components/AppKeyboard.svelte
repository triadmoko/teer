<script lang="ts">
  import {
    activeWorkspace,
    activeSessionId,
    workspaces,
    addSession,
    selectSession,
    closeSession,
    selectWorkspace,
    openCommandPalette,
  } from "@application";
  import { sessionsOf } from "@domain/models";

  const aw = $derived($activeWorkspace);
  const allSessions = $derived(sessionsOf(aw));

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
