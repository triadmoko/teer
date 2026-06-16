
import { get } from "svelte/store";
import { sessionsOf, mergeEnv, type Workspace } from "@domain/models";
import { workspaceRepository, sessionGateway } from "@infrastructure/wails";
import { guard } from "./error";
import {
  workspaces,
  activeWorkspaceId,
  activeSessionId,
  running,
} from "./stores";
import { open, closeOpened } from "./opened";
import { openWorkspaceSettings } from "./workspaceSettingsDialog";

export { workspaceSettingsDialog } from "./workspaceSettingsDialog";

export async function init(): Promise<void> {
  await refresh();
  const list = get(workspaces);
  if (list.length && !get(activeWorkspaceId)) {
    selectWorkspace(list[0].id);
  }
}

export async function refresh(): Promise<void> {
  const result = await guard(async () => {
    const list = ((await workspaceRepository.list()) ?? []).filter(
      Boolean,
    ) as Workspace[];
    const ids = (await sessionGateway.listRunning()) ?? [];
    return { list, ids };
  });
  if (!result) return;

  workspaces.set(result.list);
  const map: Record<string, boolean> = {};
  for (const id of result.ids) map[id] = true;
  running.set(map);
}

export function selectWorkspace(id: string): void {
  activeWorkspaceId.set(id);
  const ws = get(workspaces).find((w) => w?.id === id) ?? null;
  const sessions = sessionsOf(ws);
  const wsEnv = (ws?.env ?? {}) as Record<string, string>;
  const wsCwd = ws?.defaultCwd ?? "";
  const wsStartupCommand = ws?.startupCommand ?? "";

  for (const s of sessions) {
    if (s.autoStart) {
      open(s.id);
      sessionGateway.start({
        id: s.id,
        shell: s.shell ?? "",
        cwd: s.cwd || wsCwd,
        env: mergeEnv(wsEnv, s.env),
        startupCommand: wsStartupCommand,
        cols: 80,
        rows: 24,
      }).catch(() => {});
    }
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
  const ws = await guard(() => workspaceRepository.create(name, color, cwd));
  if (ws === undefined) return;
  await refresh();
  if (ws) selectWorkspace(ws.id);
}

export async function renameWorkspace(
  ws: Workspace,
  name: string,
): Promise<void> {
  const updated = { ...ws, name };
  await workspaceRepository.update(updated as Workspace);
  await refresh();
}

export async function deleteWorkspace(id: string): Promise<void> {
  const ws = get(workspaces).find((w) => w?.id === id) ?? null;
  for (const s of sessionsOf(ws)) {
    await sessionGateway.close(s.id);
    closeOpened(s.id);
  }
  await workspaceRepository.remove(id);

  if (get(activeWorkspaceId) === id) {
    activeWorkspaceId.set(null);
    activeSessionId.set(null);
  }
  await refresh();

    if (!get(activeWorkspaceId)) {
    const list = get(workspaces);
    if (list.length) selectWorkspace(list[0].id);
  }
}

export async function duplicateWorkspace(id: string): Promise<void> {
  await workspaceRepository.duplicate(id);
  await refresh();
}

export async function reorderWorkspaces(ids: string[]): Promise<void> {
  await workspaceRepository.reorder(ids);
  await refresh();
}

export async function editWorkspace(ws: Workspace): Promise<void> {
  const result = await openWorkspaceSettings(ws);
  if (!result) return;
  await workspaceRepository.update({
    ...ws,
    name: result.name,
    color: result.color,
    defaultCwd: result.defaultCwd,
    startupCommand: result.startupCommand,
    env: result.env,
  } as Workspace);
  await refresh();
}
