<script lang="ts">
  import { page } from "$app/stores";
</script>

<div
  class="flex h-screen w-screen flex-col items-center justify-center gap-2 bg-base text-zinc-400"
>
  <div class="text-2xl font-semibold text-zinc-200">{$page.status}</div>
  <div class="text-sm">{$page.error?.message ?? "Terjadi kesalahan"}</div>
  <a href="/" class="mt-2 text-sm text-blue-400 hover:underline">Kembali</a>
</div>
