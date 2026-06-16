
import { writable } from "svelte/store";

export interface SessionFormResult {
  name: string;
  shell: string;
  cwd: string;
  startupCommand: string;
  autoStart: boolean;
}

export interface SessionFormPrefill {
  name?: string;
  shell?: string;
  cwd?: string;
  startupCommand?: string;
  autoStart?: boolean;
}

interface SessionFormDialogState {
  workspaceDefaultCwd: string;
  prefill: SessionFormPrefill | null;
  resolve: (result: SessionFormResult | null) => void;
}

export const sessionFormDialog = writable<SessionFormDialogState | null>(null);

export function openSessionForm(
  workspaceDefaultCwd: string,
  prefill: SessionFormPrefill | null = null,
): Promise<SessionFormResult | null> {
  return new Promise((resolve) => {
    sessionFormDialog.set({ workspaceDefaultCwd, prefill, resolve });
  });
}
