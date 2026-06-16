<script lang="ts">
  import { IconAlertTriangle, IconX } from "@tabler/icons-svelte";
  import { lastError } from "@application";
</script>

{#if $lastError}
  <div
    class="fixed inset-x-0 top-0 z-[1100] flex items-center gap-[10px] bg-red-900 px-4 py-[9px] text-[13px] leading-[1.4] text-red-100"
    role="alert"
  >
    <IconAlertTriangle size={16} />{$lastError}
    <button
      class="ml-auto flex cursor-pointer items-center border-none bg-transparent leading-none text-red-200 hover:text-white"
      onclick={() => lastError.set(null)}><IconX size={16} /></button
    >
  </div>
{/if}
