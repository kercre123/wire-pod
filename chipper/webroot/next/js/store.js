// Small shared state: robot status polling, toasts, confirm dialogs and appearance prefs.
import { useEffect, useReducer } from "./html.js";
import { fetchRobotStatuses, describeError } from "./api.js";

export function createStore(initial) {
  let state = initial;
  const subs = new Set();
  return {
    get: () => state,
    set(patch) {
      state = { ...state, ...(typeof patch === "function" ? patch(state) : patch) };
      subs.forEach((f) => f());
    },
    subscribe(f) {
      subs.add(f);
      return () => subs.delete(f);
    },
  };
}

export function useStore(store) {
  const [, bump] = useReducer((n) => n + 1, 0);
  useEffect(() => store.subscribe(bump), [store]);
  return store.get();
}

// ---------- robots: one cached status call, polled ----------

export const robotsStore = createStore({ loaded: false, robots: [], error: null });
let pollers = 0;
let timer = null;
let inflight = false;

async function tick() {
  if (inflight) return;
  inflight = true;
  clearTimeout(timer);
  try {
    const robots = await fetchRobotStatuses({ timeout: 10000 });
    robotsStore.set({ loaded: true, robots, error: null });
  } catch (e) {
    robotsStore.set({ loaded: true, error: describeError(e).message });
  } finally {
    inflight = false;
  }
  schedule();
}

function schedule() {
  clearTimeout(timer);
  if (pollers === 0 || document.hidden) return;
  const checking = robotsStore.get().robots.some((r) => r.state === "checking");
  timer = setTimeout(tick, checking ? 1200 : 5000);
}

document.addEventListener("visibilitychange", () => {
  if (!document.hidden && pollers > 0) tick();
});

export function refreshRobots() {
  return tick();
}

// Components that show robot status call this; polling runs while at least one is mounted.
export function useRobots() {
  const s = useStore(robotsStore);
  useEffect(() => {
    pollers++;
    if (pollers === 1 || !robotsStore.get().loaded) tick();
    else schedule();
    return () => {
      pollers--;
      if (pollers === 0) clearTimeout(timer);
    };
  }, []);
  return s;
}

export function robotName(r) {
  return r ? r.name || `Vector ${r.esn}` : "Robot";
}

// For titles that show the serial number next to the name: a robot with no custom name is just "Vector".
export function robotTitle(r) {
  return r ? r.name || "Vector" : "Robot";
}

// ---------- toasts ----------

export const toastStore = createStore({ items: [] });
let toastId = 0;

// toast("Saved") or toast("Memory deleted", { undo: fn }) or toast("Couldn't save", { kind: "bad" })
export function toast(message, { undo, kind = "ok", ms, persist = false } = {}) {
  const id = ++toastId;
  const item = { id, message, undo, kind, persist };
  toastStore.set((s) => ({ items: [...s.items.slice(-2), item] }));
  setTimeout(() => dismissToast(id), ms || (undo ? 7000 : kind === "bad" ? 7000 : 3500));
  return id;
}

// Moving to another screen clears what the last one said, except messages meant for the next one.
export function clearToasts() {
  toastStore.set((s) => (s.items.some((t) => !t.persist) ? { items: s.items.filter((t) => t.persist) } : s));
}

export function dismissToast(id) {
  toastStore.set((s) => ({ items: s.items.filter((t) => t.id !== id) }));
}

// ---------- confirm dialog ----------

export const confirmStore = createStore({ current: null });

// await confirmDialog({ title, body, confirmLabel, danger }) -> true | false
export function confirmDialog(opts) {
  return new Promise((resolve) => {
    confirmStore.set({ current: { ...opts, resolve } });
  });
}

// ---------- appearance ----------

const PREFS_KEY = "wirepod-next-prefs";
export const DEFAULT_PREFS = { theme: "system", accent: "teal", accentFollowsRobot: false, size: "default", font: "default" };

function loadPrefs() {
  try {
    return { ...DEFAULT_PREFS, ...JSON.parse(localStorage.getItem(PREFS_KEY) || "{}") };
  } catch (e) {
    return { ...DEFAULT_PREFS };
  }
}

export const prefsStore = createStore({ prefs: loadPrefs() });

export function setPref(patch) {
  const prefs = { ...prefsStore.get().prefs, ...patch };
  try {
    localStorage.setItem(PREFS_KEY, JSON.stringify(prefs));
  } catch (e) {
    // private mode or blocked storage: the choice still applies until the page closes
  }
  prefsStore.set({ prefs });
}
