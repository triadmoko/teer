// Lapisan aplikasi: state inti (sumber kebenaran tunggal untuk UI).

import { writable, derived } from "svelte/store";
import type { Workspace } from "@domain/models";

export const workspaces = writable<Workspace[]>([]);
export const activeWorkspaceId = writable<string | null>(null);
export const activeSessionId = writable<string | null>(null);

/** Set id sesi yang terminalnya sudah di-mount (PTY dibangkitkan). */
export const opened = writable<Set<string>>(new Set());
/** Map id sesi -> sedang berjalan (untuk indikator status di tab, FR-14). */
export const running = writable<Record<string, boolean>>({});

/** Bila true, input keyboard dikirim ke semua sesi yang terbuka (FR-16). */
export const broadcastMode = writable(false);

export const activeWorkspace = derived(
  [workspaces, activeWorkspaceId],
  ([$ws, $id]) => $ws.find((w) => w?.id === $id) ?? null,
);
