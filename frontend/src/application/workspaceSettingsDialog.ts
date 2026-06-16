
import { writable } from "svelte/store";
import type { Workspace } from "@domain/models";

export interface WorkspaceFormResult {
  name: string;
  color: string;
  defaultCwd: string;
  startupCommand: string;
  env: Record<string, string>;
}

interface WorkspaceSettingsDialogState {
  workspace: Workspace;
  resolve: (result: WorkspaceFormResult | null) => void;
}

export const workspaceSettingsDialog =
  writable<WorkspaceSettingsDialogState | null>(null);

export function openWorkspaceSettings(
  workspace: Workspace,
): Promise<WorkspaceFormResult | null> {
  return new Promise((resolve) => {
    workspaceSettingsDialog.set({ workspace, resolve });
  });
}
