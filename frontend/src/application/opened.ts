
import { opened } from "./stores";

export function open(id: string): void {
  opened.update((s) => {
    if (s.has(id)) return s;
    const n = new Set(s);
    n.add(id);
    return n;
  });
}

export function closeOpened(id: string): void {
  opened.update((s) => {
    if (!s.has(id)) return s;
    const n = new Set(s);
    n.delete(id);
    return n;
  });
}
