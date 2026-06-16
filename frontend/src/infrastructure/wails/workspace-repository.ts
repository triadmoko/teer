
import { WorkspaceService } from "@bindings/teer/internal/service";
import { Call } from "@wailsio/runtime";
import type { Workspace, SessionDef } from "@domain/models";

export const workspaceRepository = {
  list: () => WorkspaceService.ListWorkspaces(),

  create: (name: string, color: string, defaultCwd: string) =>
    WorkspaceService.CreateWorkspace(name, color, defaultCwd),

  update: (ws: Workspace) => WorkspaceService.UpdateWorkspace(ws),

  remove: (id: string) => WorkspaceService.DeleteWorkspace(id),

  duplicate: (id: string) => WorkspaceService.DuplicateWorkspace(id),

  addSession: (workspaceId: string, name: string, shell: string, cwd: string) =>
    WorkspaceService.AddSession(workspaceId, name, shell, cwd),

  updateSession: (s: SessionDef) => WorkspaceService.UpdateSession(s),

  deleteSession: (workspaceId: string, sessionId: string) =>
    WorkspaceService.DeleteSession(workspaceId, sessionId),

      reorder: (ids: string[]): Promise<void> =>
    Call.ByName("teer/internal/service.WorkspaceService.ReorderWorkspaces", ids),
};
