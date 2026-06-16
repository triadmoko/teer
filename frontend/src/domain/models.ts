// Lapisan domain: entitas inti aplikasi teer.
//
// Tipe Workspace & SessionDef berasal dari binding Go (generated), namun seluruh
// kode aplikasi mengimpornya dari sini agar lapisan dalam (application,
// presentation) tidak bergantung langsung pada path binding/infrastruktur.

import type { Workspace, SessionDef } from "@bindings/teer/internal/domain/models";

export type { Workspace, SessionDef };

/** Peta variabel lingkungan; nilai bisa undefined sesuai bentuk binding. */
export type EnvMap = { [key: string]: string | undefined };

/** Ambil daftar sesi valid (buang null) dari sebuah workspace. */
export function sessionsOf(ws: Workspace | null | undefined): SessionDef[] {
  return (ws?.sessions ?? []).filter(Boolean) as SessionDef[];
}

/** Gabungkan env workspace dengan env sesi (env sesi menimpa workspace). */
export function mergeEnv(
  wsEnv: EnvMap,
  sessionEnv?: EnvMap,
): Record<string, string> {
  return { ...wsEnv, ...(sessionEnv ?? {}) } as Record<string, string>;
}
