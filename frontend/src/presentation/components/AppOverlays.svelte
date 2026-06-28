<script lang="ts">
  import Dialog from "./Dialog.svelte";
  import ErrorBanner from "./ErrorBanner.svelte";
  import UpdateNotification from "./UpdateNotification.svelte";
  import SessionFormDialog from "./SessionFormDialog.svelte";
  import WorkspaceSettingsDialog from "./WorkspaceSettingsDialog.svelte";
  import TerminalSettingsDialog from "./TerminalSettingsDialog.svelte";
  import CommandPalette from "./CommandPalette.svelte";
</script>

<Dialog />
<ErrorBanner />
<UpdateNotification />
<SessionFormDialog />
<WorkspaceSettingsDialog />
<TerminalSettingsDialog />
<CommandPalette />
