// Deteksi platform sederhana untuk pemilihan modifier keyboard (Cmd vs Ctrl).
// userAgentData lebih akurat bila tersedia; fallback ke platform/userAgent.
const nav = navigator as Navigator & {
  userAgentData?: { platform?: string };
};
const p = (
  nav.userAgentData?.platform ??
  nav.platform ??
  nav.userAgent ??
  ""
).toLowerCase();

export const isMac = p.includes("mac");
