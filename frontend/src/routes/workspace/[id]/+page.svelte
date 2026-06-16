<script lang="ts">
  import { page } from "$app/stores";
  import { get } from "svelte/store";
  import { selectWorkspace, activeWorkspaceId } from "@application";

  // Route -> store. selectWorkspace() set activeWorkspaceId, buka autoStart
  // session (opened -> TerminalStage mount di layout), set activeSessionId.
  // Diff-guard cegah loop dengan sync store->URL di +layout.svelte.
  $effect(() => {
    const id = $page.params.id;
    if (id && get(activeWorkspaceId) !== id) selectWorkspace(id);
  });
</script>
