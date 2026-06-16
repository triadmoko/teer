<script lang="ts">
  import { tick } from "svelte";
  import { dialog } from "./dialog";

  let value = $state("");
  let input = $state<HTMLInputElement | undefined>();

  // Saat dialog baru muncul, isi nilai awal & fokus ke input.
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
    class="overlay"
    role="button"
    tabindex="-1"
    onclick={onOverlayClick}
    onkeydown={(e) =>
      (e.key === "Enter" || e.key === " ") &&
      onOverlayClick(e as unknown as MouseEvent)}
  >
    <div class="modal" role="dialog" aria-modal="true">
      <div class="title">{$dialog.title}</div>

      {#if $dialog.kind === "prompt"}
        <input
          bind:this={input}
          bind:value
          class="field"
          placeholder={$dialog.placeholder}
          type="text"
          autocomplete="off"
        />
      {/if}

      <div class="actions">
        <button
          class="btn ghost"
          onclick={() => close($dialog && $dialog.kind === "confirm" ? false : null)}
          >Batal</button
        >
        <button class="btn" class:danger={$dialog.danger} onclick={onConfirm}>
          {$dialog.confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  .modal {
    width: 360px;
    max-width: calc(100vw - 40px);
    background: #1c1c20;
    border: 1px solid #34343a;
    border-radius: 12px;
    padding: 18px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
  }
  .title {
    color: #fafafa;
    font-size: 14px;
    margin-bottom: 12px;
  }
  .field {
    width: 100%;
    background: #121214;
    border: 1px solid #3f3f46;
    border-radius: 8px;
    color: #fafafa;
    padding: 9px 11px;
    font-size: 14px;
    outline: none;
    user-select: text;
  }
  .field:focus {
    border-color: #60a5fa;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 16px;
  }
  .btn {
    padding: 8px 16px;
    border-radius: 8px;
    border: 1px solid #2c2c31;
    background: #2563eb;
    color: #fff;
    cursor: pointer;
    font-size: 13px;
  }
  .btn:hover {
    filter: brightness(1.1);
  }
  .btn.ghost {
    background: #26262b;
    color: #d4d4d8;
  }
  .btn.danger {
    background: #dc2626;
  }
</style>
