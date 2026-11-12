// Thin client for wire-pod's HTTP API. Everything the UI says to the server goes through here, and
// every failure comes out as an ApiError that describeError() can turn into a human sentence.

const DEFAULT_TIMEOUT = 20000;

export class ApiError extends Error {
  // kind: "network" (could not reach wire-pod), "timeout", "aborted", "http" (wire-pod said no),
  //       "robot" (wire-pod could not get an answer from the robot)
  constructor(message, { status = 0, detail = "", kind = "http" } = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.detail = detail || message;
    this.kind = kind;
  }
}

async function fetchText(url, { method = "GET", body, headers, timeout = DEFAULT_TIMEOUT, signal, keepalive } = {}) {
  const ctl = new AbortController();
  let timedOut = false;
  const timer = setTimeout(() => { timedOut = true; ctl.abort(); }, timeout);
  if (signal) {
    if (signal.aborted) ctl.abort();
    else signal.addEventListener("abort", () => ctl.abort(), { once: true });
  }
  try {
    const res = await fetch(url, { method, body, headers, signal: ctl.signal, cache: "no-store", keepalive });
    const text = await res.text();
    if (!res.ok) throw new ApiError(text.trim() || res.statusText, { status: res.status, detail: text, kind: "http" });
    return text;
  } catch (e) {
    if (e instanceof ApiError) throw e;
    if (e && e.name === "AbortError") throw new ApiError("timed out", { kind: timedOut ? "timeout" : "aborted" });
    throw new ApiError("network", { kind: "network", detail: String(e) });
  } finally {
    clearTimeout(timer);
  }
}

function parseJSON(text, kind) {
  try {
    return JSON.parse(text);
  } catch (e) {
    throw new ApiError(text.trim().slice(0, 200) || "empty response", { kind, detail: text });
  }
}

// ----- wire-pod itself (/api/...) -----

export const getText = (path, opts) => fetchText(path, opts);
export const getJSON = async (path, opts) => parseJSON(await fetchText(path, opts), "http");

export function postJSON(path, data, opts = {}) {
  return fetchText(path, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(data), ...opts });
}

// ----- per-robot data wire-pod keeps itself (/api-sdk/...) -----

function form(params) {
  return new URLSearchParams(params).toString();
}

export function postForm(path, params, opts = {}) {
  return fetchText(path, { method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded" }, body: form(params), ...opts });
}

// ----- the robot itself, through wire-pod (/api-sdk/...) -----
// These handlers answer 200 with "error: ..." (or a raw gRPC error) when the robot cannot be reached.

export async function sdk(name, serial, params = {}, opts = {}) {
  const text = await fetchText(`/api-sdk/${name}?${form({ serial, ...params })}`, { method: "POST", ...opts });
  if (/^\s*error/i.test(text) || /rpc error|^\s*(dial|read|write) tcp/i.test(text)) {
    throw new ApiError(text.trim(), { kind: "robot", detail: text });
  }
  return text;
}

export async function sdkJSON(name, serial, params = {}, opts = {}) {
  return parseJSON(await sdk(name, serial, params, opts), "robot");
}

// ----- turning errors into words -----

const UNREACHABLE = /no route to host|host is down|network is unreachable|connection refused|i\/o timeout|deadline exceeded|unavailable|context deadline|dial tcp|timed out|connection reset|transport is closing|\beof\b/i;
const NOT_PAIRED = /unauthenticated|unauthorized|invalid token|not found in sdk info|robot not found/i;

// Returns { message, detail, offline }. `who` is the robot's name, e.g. "Pip".
export function describeError(err, who = "The robot") {
  const detail = String((err && (err.detail || err.message)) || err || "");
  const clean = detail.replace(/^\s*error:\s*/i, "").trim();
  if (err instanceof ApiError) {
    if (err.kind === "aborted") return { message: "", detail, offline: false };
    if (err.kind === "network") {
      return { message: "Can't reach WirePod. Check that it's still running on this computer, then try again.", detail, offline: false };
    }
    if (err.kind === "timeout") {
      return { message: `${who} took too long to answer. It may be asleep or out of Wi-Fi range.`, detail, offline: true };
    }
    if (NOT_PAIRED.test(detail)) {
      return { message: `WirePod isn't signed in to ${who === "The robot" ? "this robot" : who}. Try pairing it again from Add a robot.`, detail, offline: false };
    }
    if (UNREACHABLE.test(detail) || err.kind === "robot") {
      return {
        message: `${who} isn't answering. Check that it's powered on and on the same Wi-Fi as this computer.`,
        detail,
        offline: UNREACHABLE.test(detail),
      };
    }
    return { message: clean.charAt(0).toUpperCase() + clean.slice(1) || "Something went wrong.", detail, offline: false };
  }
  return { message: "Something went wrong.", detail, offline: false };
}

// ----- the cached status call the home screen polls -----

export const fetchRobotStatuses = async (opts) => (await getJSON("/api-sdk/robots_status", opts)).robots || [];
