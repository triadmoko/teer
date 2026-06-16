
import { writable, get } from "svelte/store";
import { workspaces, activeWorkspaceId, activeSessionId, opened } from "./stores";
import { sessionsOf } from "@domain/models";
import type { Workspace } from "@domain/models";
import { selectWorkspace, createWorkspace, editWorkspace, deleteWorkspace } from "./workspace";
import {
  addSession,
  closeSession,
  renameSession,
  selectSession,
  restartSession,
} from "./session";
import { layoutMode, gridCols } from "./layout";
import { broadcastMode } from "./stores";
import { openTerminalSettings } from "./terminalSettings";
import { openWorkspaceSettings } from "./workspaceSettingsDialog";

export interface PaletteCommand {
  id: string;
  label: string;
  group: string;
  shortcut?: string;
}

export const commandPaletteOpen = writable(false);

export function openCommandPalette(): void {
  commandPaletteOpen.set(true);
}

export function closeCommandPalette(): void {
  commandPaletteOpen.set(false);
}

const PALETTE_COLORS = ["#60a5fa", "#4ade80", "#facc15", "#f87171", "#c084fc", "#22d3ee"];

export function buildCommands(): Array<PaletteCommand & { run: () => void | Promise<void> }> {
  const wsList = get(workspaces);
  const activeWsId = get(activeWorkspaceId);
  const activeWs = wsList.find((w) => w?.id === activeWsId) ?? null;
  const activeWsSessions = sessionsOf(activeWs);
  const activeSessId = get(activeSessionId);
  const activeSess = activeWsSessions.find((s) => s.id === activeSessId) ?? null;
  const openedSet = get(opened);

  const cmds: Array<PaletteCommand & { run: () => void | Promise<void> }> = [];

    cmds.push({
    id: "ws:new",
    label: "Workspace Baru",
    group: "Workspace",
    run: async () => {
      const color = PALETTE_COLORS[wsList.length % PALETTE_COLORS.length];
      const empty = { id: "", name: "", color, defaultCwd: "", env: {}, sessions: [], createdAt: null, updatedAt: null } as unknown as Workspace;
      const result = await openWorkspaceSettings(empty);
      if (!result || !result.name) return;
      await createWorkspace(result.name, result.color || color, result.defaultCwd);
    },
  });

  if (activeWs) {
    cmds.push({
      id: "ws:edit",
      label: `Edit Workspace: ${activeWs.name}`,
      group: "Workspace",
      run: () => editWorkspace(activeWs),
    });
    cmds.push({
      id: "ws:delete",
      label: `Hapus Workspace: ${activeWs.name}`,
      group: "Workspace",
      run: () => deleteWorkspace(activeWs.id),
    });
  }

  for (const ws of wsList) {
    if (!ws || ws.id === activeWsId) continue;
    cmds.push({
      id: `ws:switch:${ws.id}`,
      label: `Pindah ke Workspace: ${ws.name}`,
      group: "Workspace",
      run: () => selectWorkspace(ws.id),
    });
  }

    if (activeWs) {
    cmds.push({
      id: "sess:new",
      label: "Terminal Baru",
      group: "Terminal",
      shortcut: "Ctrl+T",
      run: () => addSession(activeWs.id),
    });
  }

  if (activeSess) {
    cmds.push({
      id: "sess:close",
      label: `Tutup Terminal: ${activeSess.name}`,
      group: "Terminal",
      shortcut: "Ctrl+W",
      run: () => closeSession(activeSess),
    });
    cmds.push({
      id: "sess:restart",
      label: `Restart Terminal: ${activeSess.name}`,
      group: "Terminal",
      run: () => restartSession(activeSess.id),
    });
  }

  for (const s of activeWsSessions) {
    if (s.id === activeSessId) continue;
    cmds.push({
      id: `sess:switch:${s.id}`,
      label: `Pindah ke Terminal: ${s.name}`,
      group: "Terminal",
      run: () => selectSession(s.id),
    });
  }

    cmds.push({
    id: "layout:tabs",
    label: "Tampilan: Mode Tab",
    group: "Tampilan",
    run: () => layoutMode.set("tabs"),
  });
  cmds.push({
    id: "layout:grid",
    label: "Tampilan: Mode Grid",
    group: "Tampilan",
    run: () => layoutMode.set("grid"),
  });
  for (const c of [1, 2, 3, 4] as const) {
    cmds.push({
      id: `layout:cols:${c}`,
      label: `Tampilan: Grid ${c} Kolom`,
      group: "Tampilan",
      run: () => { layoutMode.set("grid"); gridCols.set(c); },
    });
  }

    const bcast = get(broadcastMode);
  cmds.push({
    id: "broadcast:toggle",
    label: bcast ? "Nonaktifkan Broadcast Input" : "Aktifkan Broadcast Input",
    group: "Alat",
    run: () => broadcastMode.update((v) => !v),
  });

    cmds.push({
    id: "settings:terminal",
    label: "Pengaturan Terminal (Font, Tema)",
    group: "Pengaturan",
    run: () => openTerminalSettings(),
  });

  return cmds;
}
