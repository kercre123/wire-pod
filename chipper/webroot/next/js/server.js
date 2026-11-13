// Server-wide facts the shell and the home screen show: config and version.
import { useEffect } from "./html.js";
import { createStore, useStore } from "./store.js";
import { getJSON } from "./api.js";

export const serverStore = createStore({ config: null, configError: false, version: null, versionState: "idle" });
let started = false;

let configReads = 0; // only the newest read of the config may apply
export async function loadConfig() {
  const id = ++configReads;
  try {
    const config = await getJSON("/api/get_config", { timeout: 8000 });
    if (id === configReads) serverStore.set({ config, configError: false });
  } catch (e) {
    if (id === configReads) serverStore.set({ configError: true });
  }
}

// After saving the AI settings the UI knows the new value for certain: use it now, and drop any read of the
// old config that is still on its way, rather than waiting for (and trusting) another round trip.
export function setKnowledge(knowledge) {
  configReads++;
  serverStore.set((s) => (s.config ? { config: { ...s.config, knowledge: { ...s.config.knowledge, ...knowledge } }, configError: false } : s));
  loadConfig(); // and confirm with the server in the background
}

// Custom builds are named after the upstream release they are based on, like v1.2.18-custom.5. The server
// compares that string with upstream's v1.2.18 and calls it an update; a build at or ahead of the upstream
// release is not behind it.
const triple = (v) => { const m = /^v?(\d+)\.(\d+)\.(\d+)/.exec(v || ""); return m ? m.slice(1).map(Number) : null; };
export function updateAvailable(v) {
  if (!v || !v.avail) return false;
  if (v.fromsource || !/-[a-z]+\.\d+$/.test(v.installedversion || "")) return true;
  const have = triple(v.installedversion), latest = triple(v.currentversion);
  if (!have || !latest) return true;
  for (let i = 0; i < 3; i++) if (have[i] !== latest[i]) return have[i] < latest[i];
  return false;
}

// get_version_info asks GitHub, so it fails with no internet. That is not an error worth showing loudly.
export async function loadVersion() {
  serverStore.set({ versionState: "loading" });
  try {
    const v = await getJSON("/api/get_version_info", { timeout: 15000 });
    serverStore.set({ version: { ...v, avail: updateAvailable(v) }, versionState: "ok" });
  } catch (e) {
    serverStore.set({ versionState: "unavailable" });
  }
}

export function useServerInfo() {
  const s = useStore(serverStore);
  useEffect(() => {
    if (started) return;
    started = true;
    loadConfig();
    loadVersion();
  }, []);
  return s;
}

// Is AI on? The knowledge settings are the source of truth: switched on and a provider chosen.
// Until the config has loaded nothing is locked, so the screen never flashes a "turn on AI" panel.
export function aiStatus(config) {
  const kg = config && config.knowledge;
  if (!kg) return { known: false, on: true, provider: "", actions: true, canAct: true };
  const on = !!kg.enable && !!kg.provider;
  const provider = on ? kg.provider : ""; // a switched-off AI keeps its provider saved; the UI treats it as none
  return { known: true, on, provider, actions: on && !!kg.commands_enable, canAct: ["openai", "custom", "together"].includes(provider) };
}

export function useAI() {
  return aiStatus(useServerInfo().config);
}

// "v1.0.3" or "commit abc1234" for source builds
export function versionLabel(v) {
  if (!v) return "";
  if (v.fromsource) return v.installedcommit ? `commit ${v.installedcommit}` : "";
  return v.installedversion || "";
}
