<script lang="ts">
  import {
    updateInfo,
    updateProgress,
    updateApplying,
    applyUpdate,
    dismissUpdate,
  } from "@application";
  import { error as setError } from "@application";

  async function onUpdate() {
    const info = $updateInfo;
    if (!info?.downloadUrl) return;
    try {
      await applyUpdate(info.downloadUrl);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Update gagal");
    }
  }
</script>

{#if $updateApplying}
  <div
    class="fixed bottom-4 right-4 z-[2000] w-72 rounded-xl border border-line-3 bg-elevated p-4 shadow-[0_8px_32px_rgba(0,0,0,0.5)]"
  >
    <div class="mb-2 text-sm font-medium text-zinc-100">Memperbarui Teer…</div>
    {#if $updateProgress}
      <div class="h-1.5 w-full overflow-hidden rounded-full bg-zinc-700">
        <div
          class="h-full rounded-full bg-blue-500 transition-all duration-300"
          style="width: {$updateProgress.percent}%"
        ></div>
      </div>
      <div class="mt-1 text-xs text-zinc-400">
        {#if $updateProgress.stage === "downloading"}
          Mengunduh… {$updateProgress.percent}%
        {:else if $updateProgress.stage === "applying"}
          Menerapkan update…
        {:else}
          Selesai, memulai ulang…
        {/if}
      </div>
    {/if}
  </div>
{:else if $updateInfo}
  <div
    class="fixed bottom-4 right-4 z-[2000] w-72 rounded-xl border border-line-3 bg-elevated p-4 shadow-[0_8px_32px_rgba(0,0,0,0.5)]"
  >
    <div class="mb-1 text-sm font-medium text-zinc-100">Update tersedia</div>
    <div class="mb-3 text-xs text-zinc-400">
      {$updateInfo.currentVersion} → <span class="text-green-400">{$updateInfo.latestVersion}</span>
    </div>
    <div class="flex gap-2">
      <button
        class="flex-1 cursor-pointer rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:brightness-110"
        onclick={onUpdate}
      >
        Update Sekarang
      </button>
      <button
        class="cursor-pointer rounded-lg border border-line-2 px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200"
        onclick={dismissUpdate}
      >
        Nanti
      </button>
    </div>
  </div>
{/if}
