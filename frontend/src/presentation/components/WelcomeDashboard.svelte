<script lang="ts">
  import { IconPlus } from "@tabler/icons-svelte";
  import { workspaces, running, selectWorkspace, createWorkspace, openWorkspaceSettings } from "@application";
  import { sessionsOf, type Workspace } from "@domain/models";

  const palette = ["#60a5fa", "#4ade80", "#facc15", "#f87171", "#c084fc", "#22d3ee"];

  const isMac = $derived(
    typeof navigator !== "undefined" && navigator.platform.toUpperCase().includes("MAC")
  );

  async function onNew() {
    const color = palette[$workspaces.length % palette.length];
    const empty = { id: "", name: "", color, defaultCwd: "", env: {}, sessions: [], createdAt: null, updatedAt: null } as unknown;
    const result = await openWorkspaceSettings(empty as Workspace);
    if (!result || !result.name) return;
    await createWorkspace(result.name, result.color || color, result.defaultCwd);
  }
</script>

<div class="absolute inset-0 flex flex-col items-center justify-center gap-8">
  <div class="flex flex-col items-center gap-3">
    <div class="select-none font-mono text-5xl text-zinc-500">&gt;_</div>
    <div class="text-2xl font-bold tracking-tight text-zinc-200">Teer</div>
    <div class="text-sm text-zinc-500">Terminal Workspace Manager</div>
  </div>

  <button
    onclick={onNew}
    class="flex items-center gap-2 rounded-lg border border-line-2 bg-raise px-4 py-2.5 text-sm text-zinc-300 transition-colors hover:bg-active hover:text-white"
  >
    <IconPlus size={16} />
    Buat Workspace Baru
  </button>

  {#if $workspaces.length > 0}
    <div class="flex w-full max-w-md flex-col items-center gap-3">
      <div class="text-xs uppercase tracking-widest text-zinc-600">Workspace Anda</div>
      <div class="grid w-full grid-cols-2 gap-2 px-4">
        {#each $workspaces as ws (ws.id)}
          {@const sessions = sessionsOf(ws)}
          {@const hasRunning = sessions.some((s) => $running[s.id])}
          <button
            onclick={() => selectWorkspace(ws.id)}
            class="group flex items-center gap-3 rounded-lg border border-line bg-surface px-3 py-2.5 text-left transition-colors hover:bg-raise"
          >
            <span
              class="h-2.5 w-2.5 shrink-0 rounded-full"
              style="background-color: {ws.color}"
            ></span>
            <div class="min-w-0 flex-1">
              <div class="truncate text-sm font-medium text-zinc-300 group-hover:text-zinc-100">
                {ws.name}
              </div>
              <div class="text-xs text-zinc-600">
                {sessions.length} sesi{hasRunning ? " · aktif" : ""}
              </div>
            </div>
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <div class="flex items-center gap-1.5 text-xs text-zinc-700">
    <kbd class="rounded bg-raise px-1.5 py-0.5 font-mono text-[11px] text-zinc-500">
      {isMac ? "⌘" : "Ctrl"}K
    </kbd>
    <span>Command Palette</span>
  </div>
</div>
