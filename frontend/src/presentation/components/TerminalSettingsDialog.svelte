<script lang="ts">
  import { IconX } from "@tabler/icons-svelte";
  import {
    terminalSettingsDialog,
    terminalFontSize,
    terminalFontFamily,
    terminalThemeName,
  } from "@application";
  import {
    THEMES,
    FONT_FAMILIES,
    FONT_FAMILY_LABELS,
    FONT_SIZES,
  } from "@domain/terminalSettings";

  function close() {
    const state = $terminalSettingsDialog;
    if (!state) return;
    terminalSettingsDialog.set(null);
    state.resolve();
  }

  function onBackdrop(e: MouseEvent) {
    if (e.target === e.currentTarget) close();
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") close();
  }
</script>

{#if $terminalSettingsDialog}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-label="Pengaturan Terminal"
    onclick={onBackdrop}
    onkeydown={onKeydown}
  >
    <div
      class="w-[420px] rounded-xl border border-line bg-surface shadow-2xl"
      role="presentation"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div
        class="flex items-center justify-between border-b border-line px-5 py-4"
      >
        <span class="text-[14px] font-semibold text-zinc-100"
          >Pengaturan Terminal</span
        >
        <button
          class="flex cursor-pointer items-center rounded border-none bg-transparent p-1 text-zinc-500 hover:bg-line-3 hover:text-zinc-50"
          onclick={close}><IconX size={15} /></button
        >
      </div>

      <!-- Body -->
      <div class="space-y-5 px-5 py-5">
        <!-- Font Size -->
        <div class="space-y-2">
          <label class="block text-[12px] font-medium text-zinc-400" for="font-size">
            Ukuran Font
          </label>
          <div class="flex flex-wrap gap-[6px]">
            {#each FONT_SIZES as size (size)}
              <button
                id={size === FONT_SIZES[0] ? "font-size" : undefined}
                class="min-w-[36px] cursor-pointer rounded-[6px] border px-2 py-[5px] text-[13px] leading-none {$terminalFontSize ===
                size
                  ? 'border-blue-500 bg-blue-500/20 text-blue-300'
                  : 'border-line bg-raise text-zinc-400 hover:bg-line-3 hover:text-zinc-50'}"
                onclick={() => terminalFontSize.set(size)}>{size}</button
              >
            {/each}
          </div>
        </div>

        <!-- Font Family -->
        <div class="space-y-2">
          <label class="block text-[12px] font-medium text-zinc-400" for="font-family">
            Font Family
          </label>
          <select
            id="font-family"
            class="w-full cursor-pointer rounded-[7px] border border-line bg-raise px-3 py-[7px] text-[13px] text-zinc-200 outline-none focus:border-blue-500"
            bind:value={$terminalFontFamily}
          >
            {#each FONT_FAMILIES as ff, i (ff)}
              <option value={ff}>{FONT_FAMILY_LABELS[i]}</option>
            {/each}
          </select>
        </div>

        <!-- Theme -->
        <div class="space-y-2">
          <span class="block text-[12px] font-medium text-zinc-400">Tema Warna</span>
          <div class="grid grid-cols-2 gap-[6px]">
            {#each THEMES as theme (theme.name)}
              <button
                class="flex cursor-pointer items-center gap-2 rounded-[7px] border px-3 py-2 text-left text-[13px] {$terminalThemeName ===
                theme.name
                  ? 'border-blue-500 bg-blue-500/10 text-zinc-100'
                  : 'border-line bg-raise text-zinc-400 hover:bg-line-3 hover:text-zinc-200'}"
                onclick={() => terminalThemeName.set(theme.name)}
              >
                <!-- swatch warna bg/fg/aksen -->
                <span class="flex shrink-0 gap-[3px]">
                  <span
                    class="h-3 w-3 rounded-sm"
                    style="background:{theme.background}"
                  ></span>
                  <span
                    class="h-3 w-3 rounded-sm"
                    style="background:{theme.green}"
                  ></span>
                  <span
                    class="h-3 w-3 rounded-sm"
                    style="background:{theme.blue}"
                  ></span>
                </span>
                {theme.name}
              </button>
            {/each}
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="border-t border-line px-5 py-4">
        <button
          class="w-full cursor-pointer rounded-[7px] border border-line-2 bg-raise py-[7px] text-[13px] text-zinc-300 hover:bg-active hover:text-white"
          onclick={close}>Tutup</button
        >
      </div>
    </div>
  </div>
{/if}
