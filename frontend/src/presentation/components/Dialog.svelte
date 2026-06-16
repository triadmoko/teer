<script lang="ts">
  import { tick } from "svelte";
  import { dialog } from "@application";

  let value = $state("");
  let input = $state<HTMLInputElement | undefined>();

    $effect(() => {
    const d = $dialog;
    if (!d) return;
    value = d.defaultValue;
    tick().then(() => input?.focus());
    tick().then(() => input?.select());
  });

  function close(result: string | boolean | null) {
    const d = $dialog;
    dialog.set(null);
    d?.resolve(result);
  }

  function onConfirm() {
    if (!$dialog) return;
    if ($dialog.kind === "prompt") {
      const v = value.trim();
      close(v === "" ? null : v);
    } else {
      close(true);
    }
  }

  function onKey(e: KeyboardEvent) {
    if (!$dialog) return;
    if (e.key === "Escape") {
      e.preventDefault();
      close($dialog.kind === "confirm" ? false : null);
    } else if (e.key === "Enter") {
      e.preventDefault();
      onConfirm();
    }
  }

  function onOverlayClick(e: MouseEvent) {
    if (e.target !== e.currentTarget) return;
    close($dialog && $dialog.kind === "confirm" ? false : null);
  }
</script>

<svelte:window onkeydown={onKey} />

{#if $dialog}
  <div
    class="fixed inset-0 z-[1000] flex items-center justify-center bg-black/50"
    role="button"
    tabindex="-1"
    onclick={onOverlayClick}
    onkeydown={(e) =>
      (e.key === "Enter" || e.key === " ") &&
      onOverlayClick(e as unknown as MouseEvent)}
  >
    <div
      class="w-[360px] max-w-[calc(100vw-40px)] rounded-xl border border-line-3 bg-elevated p-[18px] shadow-[0_20px_50px_rgba(0,0,0,0.5)]"
      role="dialog"
      aria-modal="true"
    >
      <div class="mb-3 text-sm text-zinc-50">{$dialog.title}</div>

      {#if $dialog.kind === "prompt"}
        <input
          bind:this={input}
          bind:value
          class="w-full select-text rounded-lg border border-zinc-700 bg-base px-[11px] py-[9px] text-sm text-zinc-50 outline-none focus:border-blue-400"
          placeholder={$dialog.placeholder}
          type="text"
          autocomplete="off"
        />
      {/if}

      <div class="mt-4 flex justify-end gap-2">
        <button
          class="cursor-pointer rounded-lg border border-line-2 bg-active px-4 py-2 text-[13px] text-zinc-300 hover:brightness-110"
          onclick={() => close($dialog && $dialog.kind === "confirm" ? false : null)}
          >Batal</button
        >
        <button
          class="cursor-pointer rounded-lg border border-line-2 px-4 py-2 text-[13px] text-white hover:brightness-110 {$dialog.danger
            ? 'bg-red-600'
            : 'bg-blue-600'}"
          onclick={onConfirm}
        >
          {$dialog.confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}
