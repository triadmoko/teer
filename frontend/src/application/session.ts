// Lapisan aplikasi: use case sesi terminal (tambah, pilih, rename, tutup).

import { get } from "svelte/store";
import { sessionsOf, type SessionDef } from "@domain/models";
import { workspaceRepository, sessionGateway } from "@infrastructure/wails";
import { workspaces, activeSessionId, running } from "./stores";
import { open, closeOpened } from "./opened";
import { refresh } from "./workspace";

export async function addSession(workspaceId: string): Promise<void> {
  const ws = get(workspaces).find((w) => w?.id === workspaceId) ?? null;
  const name = `terminal ${sessionsOf(ws).length + 1}`;
  const sd = await workspaceRepository.addSession(
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
  await workspaceRepository.updateSession(updated as SessionDef);
  await refresh();
}

/** Tutup sesi: kill PTY + hapus definisinya (FR-10). */
export async function closeSession(s: SessionDef): Promise<void> {
  await sessionGateway.close(s.id);
  await workspaceRepository.deleteSession(s.workspaceId, s.id);
  closeOpened(s.id);

  if (get(activeSessionId) === s.id) {
    const ws = get(workspaces).find((w) => w?.id === s.workspaceId) ?? null;
    const remaining = sessionsOf(ws).filter((x) => x.id !== s.id);
    activeSessionId.set(remaining[0]?.id ?? null);
  }
  await refresh();
}

/** Setel status running sebuah sesi (dipanggil oleh komponen Terminal). */
export function setRunning(id: string, isRunning: boolean): void {
  running.update((m) => {
    const n = { ...m };
    if (isRunning) n[id] = true;
    else delete n[id];
    return n;
  });
}
