// Adapter Wails untuk persistensi workspace & definisi sesi.
// Membungkus WorkspaceService (binding RPC Go) di balik antarmuka yang bersih.

import { WorkspaceService } from "@bindings/teer/internal/service";
import type { Workspace, SessionDef } from "@domain/models";

export const workspaceRepository = {
  list: () => WorkspaceService.ListWorkspaces(),

  create: (name: string, color: string, defaultCwd: string) =>
    WorkspaceService.CreateWorkspace(name, color, defaultCwd),

  update: (ws: Workspace) => WorkspaceService.UpdateWorkspace(ws),

  remove: (id: string) => WorkspaceService.DeleteWorkspace(id),

  duplicate: (id: string) => WorkspaceService.DuplicateWorkspace(id),

  addSession: (
    workspaceId: string,
    name: string,
    shell: string,
    cwd: string,
    startupCommand: string,
  ) => WorkspaceService.AddSession(workspaceId, name, shell, cwd, startupCommand),

  updateSession: (s: SessionDef) => WorkspaceService.UpdateSession(s),

  deleteSession: (workspaceId: string, sessionId: string) =>
    WorkspaceService.DeleteSession(workspaceId, sessionId),
};
