<script lang="ts">
  import { get } from "svelte/store";
  import { goto } from "$app/navigation";
  import { activeWorkspaceId, workspaces } from "@application";

  // Root "/" hanya redirect ke workspace aktif/pertama. TerminalStage & semua
  // state hidup di +layout.svelte, jadi page ini tidak render apa pun.
  $effect(() => {
    const id = get(activeWorkspaceId) ?? get(workspaces)[0]?.id;
    if (id) goto("/workspace/" + id, { replaceState: true });
  });
</script>
