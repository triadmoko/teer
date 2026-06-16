
import { writable } from "svelte/store";

export interface SessionFormResult {
  name: string;
  shell: string;
  cwd: string;
  autoStart: boolean;
}

interface SessionFormDialogState {
  workspaceDefaultCwd: string;
  resolve: (result: SessionFormResult | null) => void;
}

export const sessionFormDialog = writable<SessionFormDialogState | null>(null);

export function openSessionForm(
  workspaceDefaultCwd: string,
): Promise<SessionFormResult | null> {
  return new Promise((resolve) => {
    sessionFormDialog.set({ workspaceDefaultCwd, resolve });
  });
}
