<script lang="ts">
  import "../app.css";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Sidebar from "@presentation/components/Sidebar.svelte";
  import AppOverlays from "@presentation/components/AppOverlays.svelte";
  import AppKeyboard from "@presentation/components/AppKeyboard.svelte";
  import {
    init,
    activeWorkspaceId,
    checkUpdate,
    listenUpdateProgress,
  } from "@application";

  let { children } = $props();

  $effect(() => {
    init();
    checkUpdate();
    return listenUpdateProgress();
  });

  // Store → URL sync. Diff-guard cegah loop dengan URL→store di workspace/[id]/+page.svelte.
  $effect(() => {
    const id = $activeWorkspaceId;
    const path = $page.url.pathname;
    if (id && path !== `/workspace/${id}`) {
      goto(`/workspace/${id}`);
    } else if (!id && path.startsWith("/workspace")) {
      goto("/");
    }
  });
</script>

<AppKeyboard />
<AppOverlays />

<div class="flex h-screen w-screen overflow-hidden">
  <Sidebar />
  <main class="relative flex min-w-0 flex-1 flex-col bg-base">
    {@render children()}
  </main>
</div>
