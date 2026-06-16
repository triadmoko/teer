
import { Events } from "@wailsio/runtime";
import { b64ToBytes } from "../encoding";

export const terminalTransport = {

  onOutput(sessionId: string, cb: (bytes: Uint8Array) => void): () => void {
    return Events.On(`session:${sessionId}:out`, (ev) => {
      cb(b64ToBytes(ev.data as string));
    });
  },

  onExit(sessionId: string, cb: (code: number) => void): () => void {
    return Events.On(`session:${sessionId}:exit`, (ev) => {
      const code = (ev.data as { code?: number })?.code ?? 0;
      cb(code);
    });
  },
};
