// Controller state aplikasi teer.
//
// Menyatukan store Svelte (workspaces, sesi aktif, status running) dengan aksi
// yang memanggil binding Go (WorkspaceService / SessionService). Komponen UI
// cukup memanggil fungsi-fungsi di sini, bukan binding langsung.

import { writable, derived, get } from "svelte/store";
import { WorkspaceService, SessionService } from "../../bindings/teer/internal/service";
import type { Workspace, SessionDef } from "../../bindings/teer/internal/domain/models";

export const workspaces = writable<Workspace[]>([]);
export const activeWorkspaceId = writable<string | null>(null);
export const activeSessionId = writable<string | null>(null);

/** Set id sesi yang terminalnya sudah di-mount (PTY dibangkitkan). */
export const opened = writable<Set<string>>(new Set());
/** Map id sesi -> sedang berjalan (untuk indikator status di tab, FR-14). */
export const running = writable<Record<string, boolean>>({});

export const activeWorkspace = derived(
  [workspaces, activeWorkspaceId],
  ([$ws, $id]) => $ws.find((w) => w?.id === $id) ?? null,
);

/** Pesan error untuk ditampilkan di banner (mis. backend tidak tersambung). */
export const lastError = writable<string | null>(null);

const BACKEND_HINT =
  "Tidak terhubung ke backend Go. Jalankan dalam mode DESKTOP " +
  "(`task dev` lalu pakai jendela yang muncul, atau `./bin/teer`) — " +
  "bukan membuka URL di browser.";

// Jalankan aksi binding; tangkap error agar tidak gagal diam-diam.
async function guard<T>(fn: () => Promise<T>): Promise<T | undefined> {
  try {
    const r = await fn();
    lastError.set(null);
    return r;
  } catch (e) {
    console.error("teer binding error:", e);
    lastError.set(BACKEND_HINT);
    return undefined;
  }
}

function sessionsOf(ws: Workspace | null | undefined): SessionDef[] {
  return ((ws?.sessions ?? []).filter(Boolean) as SessionDef[]);
}

// ---- Bootstrap & refresh ----

export async function init(): Promise<void> {
  await refresh();
  const list = get(workspaces);
  if (list.length && !get(activeWorkspaceId)) {
    selectWorkspace(list[0].id);
  }
}

/** Muat ulang daftar workspace dari disk & sinkronkan status running. */
export async function refresh(): Promise<void> {
  const result = await guard(async () => {
    const list = ((await WorkspaceService.ListWorkspaces()) ?? []).filter(
      Boolean,
    ) as Workspace[];
    const ids = (await SessionService.ListRunning()) ?? [];
    return { list, ids };
  });
  if (!result) return;

  workspaces.set(result.list);
  const map: Record<string, boolean> = {};
  for (const id of result.ids) map[id] = true;
  running.set(map);
}

// ---- Workspace ----

export function selectWorkspace(id: string): void {
  activeWorkspaceId.set(id);
  const ws = get(workspaces).find((w) => w?.id === id) ?? null;
  const sessions = sessionsOf(ws);

  // Bangkitkan otomatis sesi ber-autoStart (FR-13).
  for (const s of sessions) {
    if (s.autoStart) open(s.id);
  }

  const target = sessions.find((s) => s.autoStart) ?? sessions[0];
  if (target) {
    open(target.id);
    activeSessionId.set(target.id);
  } else {
    activeSessionId.set(null);
  }
}

export async function createWorkspace(
  name: string,
  color: string,
  cwd: string,
): Promise<void> {
  const ws = await guard(() => WorkspaceService.CreateWorkspace(name, color, cwd));
  if (ws === undefined) return; // gagal (backend); banner sudah di-set
  await refresh();
  if (ws) selectWorkspace(ws.id);
}

export async function renameWorkspace(ws: Workspace, name: string): Promise<void> {
  const updated = { ...ws, name };
  await WorkspaceService.UpdateWorkspace(updated as Workspace);
  await refresh();
}

export async function deleteWorkspace(id: string): Promise<void> {
  const ws = get(workspaces).find((w) => w?.id === id) ?? null;
  for (const s of sessionsOf(ws)) {
    await SessionService.CloseSession(s.id);
    closeOpened(s.id);
  }
  await WorkspaceService.DeleteWorkspace(id);

  if (get(activeWorkspaceId) === id) {
    activeWorkspaceId.set(null);
    activeSessionId.set(null);
  }
  await refresh();

  // Pilih workspace lain bila yang aktif barusan dihapus.
  if (!get(activeWorkspaceId)) {
    const list = get(workspaces);
    if (list.length) selectWorkspace(list[0].id);
  }
}

export async function duplicateWorkspace(id: string): Promise<void> {
  await WorkspaceService.DuplicateWorkspace(id);
  await refresh();
}

// ---- Sesi ----

export async function addSession(workspaceId: string): Promise<void> {
  const ws = get(workspaces).find((w) => w?.id === workspaceId) ?? null;
  const name = `terminal ${sessionsOf(ws).length + 1}`;
  const sd = await WorkspaceService.AddSession(
    workspaceId,
    name,
    "",
    ws?.defaultCwd ?? "",
    "",
  );
  await refresh();
  if (sd) selectSession(sd.id);
}

export function selectSession(id: string): void {
  open(id);
  activeSessionId.set(id);
}

export async function renameSession(s: SessionDef, name: string): Promise<void> {
  const updated = { ...s, name };
  await WorkspaceService.UpdateSession(updated as SessionDef);
  await refresh();
}

/** Tutup sesi: kill PTY + hapus definisinya (FR-10). */
export async function closeSession(s: SessionDef): Promise<void> {
  await SessionService.CloseSession(s.id);
  await WorkspaceService.DeleteSession(s.workspaceId, s.id);
  closeOpened(s.id);

  if (get(activeSessionId) === s.id) {
    const ws = get(workspaces).find((w) => w?.id === s.workspaceId) ?? null;
    const remaining = sessionsOf(ws).filter((x) => x.id !== s.id);
    activeSessionId.set(remaining[0]?.id ?? null);
  }
  await refresh();
}

// ---- Status helpers (dipanggil komponen Terminal) ----

export function setRunning(id: string, isRunning: boolean): void {
  running.update((m) => {
    const n = { ...m };
    if (isRunning) n[id] = true;
    else delete n[id];
    return n;
  });
}

// ---- Internal ----

function open(id: string): void {
  opened.update((s) => {
    if (s.has(id)) return s;
    const n = new Set(s);
    n.add(id);
    return n;
  });
}

function closeOpened(id: string): void {
  opened.update((s) => {
    if (!s.has(id)) return s;
    const n = new Set(s);
    n.delete(id);
    return n;
  });
}
