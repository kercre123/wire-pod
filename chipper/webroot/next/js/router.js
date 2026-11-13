// Hash router. The server has no fallback route, so the app lives at /next/#/...
import { useReducer, useEffect, useRef } from "./html.js";

export const ROBOT_TABS = ["personality", "memories", "settings", "controls", "activity"];
export const SETTINGS_SECTIONS = ["ai", "voice", "integrations", "server", "appearance"];

export function parseHash(hash = location.hash) {
  const q = hash.indexOf("?");
  const query = new URLSearchParams(q >= 0 ? hash.slice(q + 1) : "");
  if (q >= 0) hash = hash.slice(0, q);
  const parts = hash.replace(/^#\/?/, "").split("/").filter(Boolean).map(decodeURIComponent);
  const [first, a, b] = parts;
  switch (first) {
    case undefined:
    case "robots":
      return { name: "robots" };
    case "robot":
      return a ? { name: "robot", esn: a, tab: ROBOT_TABS.includes(b) ? b : "personality" } : { name: "robots" };
    case "settings":
      // ?return=#/robot/... says where to go back to after the section's work is done (only robot screens)
      return { name: "settings", section: SETTINGS_SECTIONS.includes(a) ? a : null, back: /^#\/robot\//.test(query.get("return") || "") ? query.get("return") : null };
    case "add":
      return { name: "add" };
    case "welcome":
      return { name: "welcome" };
    default:
      return { name: "notfound" };
  }
}

export const href = {
  robots: "#/",
  robot: (esn, tab) => `#/robot/${encodeURIComponent(esn)}${tab && tab !== "personality" ? "/" + tab : ""}`,
  settings: (section, back) => `#/settings${section ? "/" + section : ""}${back ? "?return=" + encodeURIComponent(back) : ""}`,
  add: "#/add",
  welcome: "#/welcome",
};

// The current route lives in one place, so a navigation produces exactly one new route object and one render,
// and re-reading an unchanged hash produces none.
let current = parseHash();
const listeners = new Set();
function sync() {
  const next = parseHash();
  if (JSON.stringify(next) === JSON.stringify(current)) return;
  current = next;
  listeners.forEach((f) => f());
}
window.addEventListener("hashchange", sync);

export function go(hash, { replace = false } = {}) {
  if (replace) {
    history.replaceState(null, "", hash); // no hashchange event, so nothing else reacts to a redirect
    sync();
  } else location.hash = hash;
}

export function useRoute() {
  const [, bump] = useReducer((n) => n + 1, 0);
  const rendered = useRef(current);
  rendered.current = current;
  useEffect(() => {
    listeners.add(bump);
    sync(); // the hash may have been replaced before this subscribed
    if (rendered.current !== current) bump(); // or changed after this rendered but before we listened
    return () => listeners.delete(bump);
  }, []);
  return current;
}
