
import type { Workspace, SessionDef } from "@bindings/teer/internal/domain/models";

export type { Workspace, SessionDef };

export type EnvMap = { [key: string]: string | undefined };

export function sessionsOf(ws: Workspace | null | undefined): SessionDef[] {
  return (ws?.sessions ?? []).filter(Boolean) as SessionDef[];
}

export function mergeEnv(
  wsEnv: EnvMap,
  sessionEnv?: EnvMap,
): Record<string, string> {
  return { ...wsEnv, ...(sessionEnv ?? {}) } as Record<string, string>;
}
