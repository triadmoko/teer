<script lang="ts">
  import { tick } from "svelte";
  import { commandPaletteOpen, closeCommandPalette, buildCommands } from "@application";

  let query = $state("");
  let selectedIdx = $state(0);
  let input = $state<HTMLInputElement | undefined>();

    const allCommands = $derived($commandPaletteOpen ? buildCommands() : []);

  const filtered = $derived(() => {
    const q = query.trim().toLowerCase();
    const list = q
      ? allCommands.filter(
          (c) =>
            c.label.toLowerCase().includes(q) ||
            c.group.toLowerCase().includes(q),
        )
      : allCommands;
    return list;
  });

    $effect(() => {
        query;
    selectedIdx = 0;
  });

    $effect(() => {
    if ($commandPaletteOpen) {
      query = "";
      selectedIdx = 0;
      tick().then(() => input?.focus());
    }
  });

  function run(idx: number) {
    const cmd = filtered()[idx];
    if (!cmd) return;
    closeCommandPalette();
    cmd.run();
  }

  function onKey(e: KeyboardEvent) {
    const list = filtered();
    if (e.key === "Escape") {
      e.preventDefault();
      closeCommandPalette();
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      selectedIdx = Math.min(selectedIdx + 1, list.length - 1);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      selectedIdx = Math.max(selectedIdx - 1, 0);
    } else if (e.key === "Enter") {
      e.preventDefault();
      run(selectedIdx);
    }
  }

  function onBackdrop(e: MouseEvent) {
    if (e.target === e.currentTarget) closeCommandPalette();
  }

    const grouped = $derived(() => {
    const list = filtered();
    const groups: Array<{ group: string; items: Array<{ cmd: typeof list[0]; idx: number }> }> = [];
    const seen = new Map<string, number>();
    for (let i = 0; i < list.length; i++) {
      const c = list[i];
      if (!seen.has(c.group)) {
        seen.set(c.group, groups.length);
        groups.push({ group: c.group, items: [] });
      }
      groups[seen.get(c.group)!].items.push({ cmd: c, idx: i });
    }
    return groups;
  });
</script>

{#if $commandPaletteOpen}
  <div
    class="fixed inset-0 z-[60] flex items-start justify-center pt-[12vh] bg-black/50"
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-label="Command palette"
    onclick={onBackdrop}
    onkeydown={onKey}
  >
    <div
      class="w-[520px] max-w-[90vw] overflow-hidden rounded-xl border border-line bg-surface shadow-2xl"
      role="presentation"
      onclick={(e) => e.stopPropagation()}
    >

      <div class="flex items-center gap-2 border-b border-line px-4 py-3">
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          class="shrink-0 text-zinc-500"
          ><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.35-4.35"></path></svg
        >
        <input
          bind:this={input}
          bind:value={query}
          class="flex-1 bg-transparent text-[14px] text-zinc-100 placeholder:text-zinc-600 outline-none"
          placeholder="Cari perintah…"
          type="text"
          autocomplete="off"
          spellcheck={false}
        />
        <kbd
          class="shrink-0 rounded border border-line px-[5px] py-px text-[11px] text-zinc-600"
          >Esc</kbd
        >
      </div>

      <div class="max-h-[380px] overflow-y-auto py-1">
        {#if filtered().length === 0}
          <div class="px-4 py-6 text-center text-[13px] text-zinc-600">
            Tidak ada perintah yang cocok.
          </div>
        {:else}
          {#each grouped() as grp (grp.group)}
            <div class="px-3 pb-1 pt-2 text-[10px] font-semibold uppercase tracking-wider text-zinc-600">
              {grp.group}
            </div>
            {#each grp.items as { cmd, idx } (cmd.id)}
              <button
                class="flex w-full cursor-pointer items-center gap-2 px-4 py-[7px] text-left text-[13px] {idx ===
                selectedIdx
                  ? 'bg-active text-zinc-50'
                  : 'text-zinc-300 hover:bg-raise'}"
                onclick={() => run(idx)}
                onmouseenter={() => (selectedIdx = idx)}
              >
                <span class="flex-1 truncate">{cmd.label}</span>
                {#if cmd.shortcut}
                  <kbd
                    class="shrink-0 rounded border border-line px-[5px] py-px text-[11px] text-zinc-500"
                    >{cmd.shortcut}</kbd
                  >
                {/if}
              </button>
            {/each}
          {/each}
        {/if}
      </div>
    </div>
  </div>
{/if}
