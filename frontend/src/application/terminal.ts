
import { get } from "svelte/store";
import { mergeEnv, type SessionDef } from "@domain/models";
import { sessionGateway, terminalTransport } from "@infrastructure/wails";
import { opened } from "./stores";
export { broadcastMode } from "./stores";

export function startSession(
  session: SessionDef,
  wsEnv: Record<string, string>,
  wsCwd: string,
  wsStartupCommand: string,
  size: { cols: number; rows: number },
): Promise<void> {
  return sessionGateway.start({
    id: session.id,
    shell: session.shell ?? "",
    cwd: session.cwd || wsCwd,
    env: mergeEnv(wsEnv, session.env),
    startupCommand: session.autoStart ? (session.startupCommand || wsStartupCommand) : "",
    cols: size.cols,
    rows: size.rows,
  });
}

export const writeSession = (id: string, data: string): Promise<void> =>
  sessionGateway.write(id, data);

export const resizeSession = (
  id: string,
  cols: number,
  rows: number,
): Promise<void> => sessionGateway.resize(id, cols, rows);

export const onSessionOutput = (
  id: string,
  cb: (bytes: Uint8Array) => void,
): (() => void) => terminalTransport.onOutput(id, cb);

export function broadcastWrite(data: string): void {
  for (const id of get(opened)) {
    sessionGateway.write(id, data);
  }
}

export const onSessionExit = (
  id: string,
  cb: (code: number) => void,
): (() => void) => terminalTransport.onExit(id, cb);

// persistScrollback menyimpan snapshot scrollback (string hasil serialize xterm)
// ke disk lewat backend. Best-effort.
export const persistScrollback = (id: string, data: string): Promise<void> =>
  sessionGateway.saveScrollback(id, data);

// restoreScrollback mengambil snapshot scrollback terakhir; "" bila tidak ada.
export const restoreScrollback = (id: string): Promise<string> =>
  sessionGateway.loadScrollback(id);
