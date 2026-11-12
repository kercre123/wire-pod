// Hash router. The server has no fallback route, so the app lives at /next/#/...
import { useState, useEffect } from "./html.js";

export const ROBOT_TABS = ["personality", "memories", "settings", "controls", "activity"];
export const SETTINGS_SECTIONS = ["ai", "voice", "integrations", "server", "appearance"];

export function parseHash(hash = location.hash) {
  const parts = hash.replace(/^#\/?/, "").split("/").filter(Boolean).map(decodeURIComponent);
  const [first, a, b] = parts;
  switch (first) {
    case undefined:
    case "robots":
      return { name: "robots" };
    case "robot":
      return a ? { name: "robot", esn: a, tab: ROBOT_TABS.includes(b) ? b : "personality" } : { name: "robots" };
    case "settings":
      return { name: "settings", section: SETTINGS_SECTIONS.includes(a) ? a : null };
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
  settings: (section) => `#/settings${section ? "/" + section : ""}`,
  add: "#/add",
  welcome: "#/welcome",
};

export function go(hash, { replace = false } = {}) {
  if (replace) history.replaceState(null, "", hash);
  else location.hash = hash;
  if (replace) window.dispatchEvent(new HashChangeEvent("hashchange"));
}

export function useRoute() {
  const [route, setRoute] = useState(parseHash());
  useEffect(() => {
    const f = () => setRoute(parseHash());
    window.addEventListener("hashchange", f);
    return () => window.removeEventListener("hashchange", f);
  }, []);
  return route;
}
