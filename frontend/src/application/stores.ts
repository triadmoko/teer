
import { writable, derived } from "svelte/store";
import type { Workspace } from "@domain/models";

export const workspaces = writable<Workspace[]>([]);
export const activeWorkspaceId = writable<string | null>(null);
export const activeSessionId = writable<string | null>(null);

export const opened = writable<Set<string>>(new Set());

export const running = writable<Record<string, boolean>>({});

export const broadcastMode = writable(false);

export const activeWorkspace = derived(
  [workspaces, activeWorkspaceId],
  ([$ws, $id]) => $ws.find((w) => w?.id === $id) ?? null,
);
