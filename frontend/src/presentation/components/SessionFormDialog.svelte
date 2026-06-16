<script lang="ts">
  import { tick } from "svelte";
  import { sessionFormDialog } from "@application/sessionFormDialog";
  import { IconFolderOpen } from "@tabler/icons-svelte";
  import { PickDirectory } from "@bindings/teer/internal/service/dialogservice";

  let nameInput = $state<HTMLInputElement | undefined>();
  let name = $state("");
  let shell = $state("");
  let cwd = $state("");
  let startupCommand = $state("");
  let autoStart = $state(false);
  let isEdit = $state(false);

  $effect(() => {
    const d = $sessionFormDialog;
    if (!d) return;
    const p = d.prefill;
    isEdit = p !== null;
    name = p?.name ?? "";
    shell = p?.shell ?? "";
    cwd = p?.cwd ?? d.workspaceDefaultCwd;
    startupCommand = p?.startupCommand ?? "";
    autoStart = p?.autoStart ?? false;
    tick().then(() => {
      nameInput?.focus();
      nameInput?.select();
    });
  });

  function close(ok: boolean) {
    const d = $sessionFormDialog;
    sessionFormDialog.set(null);
    if (!d) return;
    if (!ok) {
      d.resolve(null);
      return;
    }
    d.resolve({
      name: name.trim() || "terminal",
      shell: shell.trim(),
      cwd: cwd.trim(),
      startupCommand: startupCommand.trim(),
      autoStart,
    });
  }

  function onKey(e: KeyboardEvent) {
    if (!$sessionFormDialog) return;
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

  async function browseCwd() {
    const picked = await PickDirectory();
    if (picked) cwd = picked;
  }
</script>

<svelte:window onkeydown={onKey} />

{#if $sessionFormDialog}

  <div
    class="fixed inset-0 z-[1000] flex items-center justify-center bg-black/50"
    role="presentation"
    onclick={onOverlayClick}
    onkeydown={(e) =>
      (e.key === "Enter" || e.key === " ") &&
      onOverlayClick(e as unknown as MouseEvent)}
  >
    <div
      class="w-[440px] max-w-[calc(100vw-40px)] rounded-xl border border-line-3 bg-elevated p-[18px] shadow-[0_20px_50px_rgba(0,0,0,0.5)]"
      role="dialog"
      aria-modal="true"
    >
      <div class="mb-4 text-sm font-semibold text-zinc-50">
        {isEdit ? "Edit terminal" : "Terminal baru"}
      </div>

      <div class="flex flex-col gap-3">

        <label class="flex flex-col gap-1">
          <span class="text-[11px] text-zinc-400">Nama</span>
          <input
            bind:this={nameInput}
            bind:value={name}
            class="w-full rounded-lg border border-zinc-700 bg-base px-[11px] py-[8px] text-sm text-zinc-50 outline-none focus:border-blue-400"
            placeholder="mis. server, worker, db"
            type="text"
            autocomplete="off"
          />
        </label>

        <label class="flex flex-col gap-1">
          <span class="text-[11px] text-zinc-400"
            >Shell <span class="text-zinc-600">(kosong = default)</span></span
          >
          <input
            bind:value={shell}
            class="w-full rounded-lg border border-zinc-700 bg-base px-[11px] py-[8px] text-sm text-zinc-50 outline-none focus:border-blue-400"
            placeholder="/bin/bash"
            type="text"
            autocomplete="off"
          />
        </label>

        <div class="flex flex-col gap-1">
          <span class="text-[11px] text-zinc-400"
            >Working directory <span class="text-zinc-600"
              >(kosong = ikut workspace)</span
            ></span
          >
          <div class="flex gap-1">
            <input
              bind:value={cwd}
              class="min-w-0 flex-1 rounded-lg border border-zinc-700 bg-base px-[11px] py-[8px] text-sm text-zinc-50 outline-none focus:border-blue-400"
              placeholder="~"
              type="text"
              autocomplete="off"
            />
            <button
              class="flex cursor-pointer items-center rounded-lg border border-zinc-700 bg-base px-2 text-zinc-400 hover:border-blue-400 hover:text-blue-400"
              onclick={browseCwd}
              title="Pilih direktori"
              type="button"
            >
              <IconFolderOpen size={15} />
            </button>
          </div>
        </div>

        <label class="flex flex-col gap-1">
          <span class="text-[11px] text-zinc-400"
            >Startup command <span class="text-zinc-600"
              >(kosong = ikut workspace)</span
            ></span
          >
          <textarea
            bind:value={startupCommand}
            class="w-full resize-y rounded-lg border border-zinc-700 bg-base px-[11px] py-[8px] text-sm text-zinc-50 outline-none focus:border-blue-400"
            placeholder="mis. npm run dev"
            rows="2"
            autocomplete="off"
          ></textarea>
        </label>

        <label class="flex cursor-pointer items-center gap-2">
          <input
            bind:checked={autoStart}
            class="h-4 w-4 accent-blue-500"
            type="checkbox"
          />
          <span class="text-[13px] text-zinc-300"
            >Auto-start saat workspace dibuka</span
          >
        </label>
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
          onclick={() => close(true)}>{isEdit ? "Simpan" : "Buat terminal"}</button
        >
      </div>
    </div>
  </div>
{/if}
