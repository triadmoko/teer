<script lang="ts">
  import { tick } from "svelte";
  import { IconPlus, IconTrash } from "@tabler/icons-svelte";
  import { workspaceSettingsDialog } from "@application/workspaceSettingsDialog";

  const palette = [
    "#60a5fa",
    "#4ade80",
    "#facc15",
    "#f87171",
    "#c084fc",
    "#22d3ee",
    "#fb923c",
    "#a78bfa",
    "#34d399",
    "#f472b6",
  ];

  let nameInput = $state<HTMLInputElement | undefined>();
  let name = $state("");
  let color = $state(palette[0]);
  let defaultCwd = $state("");
  // env vars direpresentasikan sebagai array [key, value] agar bisa di-edit per-baris
  let envRows = $state<[string, string][]>([]);

  $effect(() => {
    const d = $workspaceSettingsDialog;
    if (!d) return;
    const ws = d.workspace;
    name = ws.name;
    color = ws.color || palette[0];
    defaultCwd = ws.defaultCwd;
    const entries = Object.entries(ws.env ?? {}).filter(
      ([k]) => k !== undefined,
    ) as [string, string][];
    envRows = entries.length ? entries : [];
    tick().then(() => {
      nameInput?.focus();
      nameInput?.select();
    });
  });

  function addEnvRow() {
    envRows = [...envRows, ["", ""]];
  }

  function removeEnvRow(i: number) {
    envRows = envRows.filter((_, idx) => idx !== i);
  }

  function setEnvKey(i: number, val: string) {
    envRows = envRows.map((row, idx) => (idx === i ? [val, row[1]] : row));
  }

  function setEnvVal(i: number, val: string) {
    envRows = envRows.map((row, idx) => (idx === i ? [row[0], val] : row));
  }

  function buildEnv(): Record<string, string> {
    const out: Record<string, string> = {};
    for (const [k, v] of envRows) {
      if (k.trim()) out[k.trim()] = v;
    }
    return out;
  }

  function close(ok: boolean) {
    const d = $workspaceSettingsDialog;
    workspaceSettingsDialog.set(null);
    if (!d) return;
    if (!ok) {
      d.resolve(null);
      return;
    }
    d.resolve({
      name: name.trim() || d.workspace.name,
      color,
      defaultCwd: defaultCwd.trim(),
      env: buildEnv(),
    });
  }

  function onKey(e: KeyboardEvent) {
    if (!$workspaceSettingsDialog) return;
    if (e.key === "Escape") {
      e.preventDefault();
      close(false);
    } else if (e.key === "Enter" && e.ctrlKey) {
      e.preventDefault();
      close(true);
    }
  }

  function onOverlayClick(e: MouseEvent) {
    if (e.target !== e.currentTarget) return;
    close(false);
  }
</script>

<svelte:window onkeydown={onKey} />

{#if $workspaceSettingsDialog}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-[1000] flex items-center justify-center bg-black/50"
    onclick={onOverlayClick}
    onkeydown={(e) =>
      (e.key === "Enter" || e.key === " ") &&
      onOverlayClick(e as unknown as MouseEvent)}
  >
    <div
      class="w-[500px] max-w-[calc(100vw-40px)] rounded-xl border border-line-3 bg-elevated p-[18px] shadow-[0_20px_50px_rgba(0,0,0,0.5)]"
      role="dialog"
      aria-modal="true"
    >
      <div class="mb-4 text-sm font-semibold text-zinc-50">
        Pengaturan workspace
      </div>

      <div class="flex flex-col gap-3">
        <!-- Nama -->
        <label class="flex flex-col gap-1">
          <span class="text-[11px] text-zinc-400">Nama</span>
          <input
            bind:this={nameInput}
            bind:value={name}
            class="w-full rounded-lg border border-zinc-700 bg-base px-[11px] py-[8px] text-sm text-zinc-50 outline-none focus:border-blue-400"
            type="text"
            autocomplete="off"
          />
        </label>

        <!-- Warna -->
        <div class="flex flex-col gap-1">
          <span class="text-[11px] text-zinc-400">Warna</span>
          <div class="flex gap-2">
            {#each palette as c (c)}
              <button
                class="h-5 w-5 cursor-pointer rounded-full border-2 transition-transform {color ===
                c
                  ? 'border-white scale-110'
                  : 'border-transparent hover:scale-110'}"
                style="background:{c}"
                title={c}
                onclick={() => (color = c)}
              ></button>
            {/each}
          </div>
        </div>

        <!-- Default CWD -->
        <label class="flex flex-col gap-1">
          <span class="text-[11px] text-zinc-400">Default working directory</span>
          <input
            bind:value={defaultCwd}
            class="w-full rounded-lg border border-zinc-700 bg-base px-[11px] py-[8px] text-sm text-zinc-50 outline-none focus:border-blue-400"
            placeholder="~"
            type="text"
            autocomplete="off"
          />
        </label>

        <!-- Env vars -->
        <div class="flex flex-col gap-1">
          <div class="flex items-center justify-between">
            <span class="text-[11px] text-zinc-400">Environment variables</span>
            <button
              class="flex cursor-pointer items-center gap-1 rounded border-none bg-transparent text-[11px] text-zinc-500 hover:text-zinc-200"
              onclick={addEnvRow}><IconPlus size={12} /> Tambah</button
            >
          </div>
          {#if envRows.length === 0}
            <div class="text-[11px] text-zinc-600">Belum ada env var.</div>
          {:else}
            <div class="flex max-h-[160px] flex-col gap-1 overflow-y-auto">
              {#each envRows as row, i (i)}
                <div class="flex gap-1">
                  <input
                    class="w-[40%] rounded border border-zinc-700 bg-base px-2 py-1 text-[12px] text-zinc-200 outline-none focus:border-blue-400"
                    placeholder="KEY"
                    type="text"
                    value={row[0]}
                    onchange={(e) =>
                      setEnvKey(i, (e.currentTarget as HTMLInputElement).value)}
                  />
                  <input
                    class="min-w-0 flex-1 rounded border border-zinc-700 bg-base px-2 py-1 text-[12px] text-zinc-200 outline-none focus:border-blue-400"
                    placeholder="value"
                    type="text"
                    value={row[1]}
                    onchange={(e) =>
                      setEnvVal(i, (e.currentTarget as HTMLInputElement).value)}
                  />
                  <button
                    class="flex cursor-pointer items-center rounded border-none bg-transparent px-1 text-zinc-600 hover:text-red-400"
                    onclick={() => removeEnvRow(i)}
                    title="Hapus"
                  >
                    <IconTrash size={13} />
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <p class="mt-3 text-[11px] text-zinc-600">
        Ctrl+Enter untuk simpan &bull; Escape untuk batal
      </p>

      <div class="mt-4 flex justify-end gap-2">
        <button
          class="cursor-pointer rounded-lg border border-line-2 bg-active px-4 py-2 text-[13px] text-zinc-300 hover:brightness-110"
          onclick={() => close(false)}>Batal</button
        >
        <button
          class="cursor-pointer rounded-lg border border-line-2 bg-blue-600 px-4 py-2 text-[13px] text-white hover:brightness-110"
          onclick={() => close(true)}>Simpan</button
        >
      </div>
    </div>
  </div>
{/if}
