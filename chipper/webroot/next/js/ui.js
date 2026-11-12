// Shared building blocks.
import { html, useEffect, useRef, useState } from "./html.js";
import { Icon } from "./icons.js";
import { toastStore, useStore, dismissToast, confirmStore, robotName } from "./store.js";
import { eyeColor } from "./theme.js";

let uidCounter = 0;
export function useUid(prefix = "f") {
  const ref = useRef(null);
  if (ref.current === null) ref.current = `${prefix}${++uidCounter}`;
  return ref.current;
}

export function Face({ robot, size = 56 }) {
  const off = robot && robot.state === "offline";
  return html`<span class=${"face" + (off ? " off" : "")} style=${`--eye:${eyeColor(robot && robot.eye)};--fs:${size}px`} aria-hidden="true"><i></i><i></i></span>`;
}

export function StatusPill({ robot }) {
  if (robot.state === "online") return html`<span class="pill ok"><span class="dot"></span>Online</span>`;
  if (robot.state === "checking") return html`<span class="pill"><span class="dot mute"></span>Checking</span>`;
  return html`<span class="pill bad"><span class="dot bad"></span>Can’t reach</span>`;
}

export function Battery({ battery }) {
  if (!battery) return html`<div class="bat"><span class="cell none"></span><span>Battery unknown</span></div>`;
  const cls = battery.percent < 20 ? " low" : battery.percent < 50 ? " mid" : "";
  return html`<div class="bat">
    <span class=${"cell" + cls}><i style=${`width:${battery.percent}%`}></i></span>
    <b>${battery.percent}%</b>
    <span>${battery.charging ? "Charging" : battery.on_charger ? "On charger" : "On battery"}</span>
  </div>`;
}

export function Switch({ checked, onChange, label, disabled }) {
  return html`<span class="sw-wrap"><button type="button" class="sw" role="switch" aria-checked=${checked ? "true" : "false"}
    aria-label=${label} disabled=${disabled} onClick=${() => onChange(!checked)}></button></span>`;
}

// options: [[value, label], ...]
export function Segmented({ options, value, onChange, label, disabled }) {
  return html`<div class="sg" role="group" aria-label=${label}>
    ${options.map(([v, text]) => html`<button type="button" key=${v} aria-pressed=${v === value ? "true" : "false"} disabled=${disabled}
      onClick=${() => onChange(v)}>${text}</button>`)}
  </div>`;
}

export function Callout({ icon = "info", kind = "", children }) {
  return html`<div class=${"callout " + kind} role=${kind === "bad" ? "alert" : null}><${Icon} name=${icon} size=${20} /><div class="grow">${children}</div></div>`;
}

export function Loading({ children = "Loading…" }) {
  return html`<div class="loading" role="status"><span class="spinner"></span><span>${children}</span></div>`;
}

// Shown on robot sections that need the robot itself, when it can't be reached.
export function OfflineNotice({ robot, onRetry, what = "this" }) {
  const name = robotName(robot);
  return html`<div class="card offline-note">
    <${Callout} icon="info">
      <b>${name} isn’t answering.</b>
      <span>Check that it’s powered on and on the same Wi-Fi as this computer. Then try again. You can still edit ${name}’s personality and memories while it’s offline, so ${what} will be here when it’s back.</span>
    <//>
    ${onRetry && html`<button type="button" class="btn" onClick=${onRetry}><${Icon} name="refresh" size=${18} />Check again</button>`}
  </div>`;
}

// A failed action: a human sentence plus the raw error tucked away.
export function ErrorNote({ error, onRetry }) {
  if (!error || !error.message) return null;
  return html`<${Callout} icon="alert" kind="warn">
    <span>${error.message}</span>
    ${onRetry && html`<button type="button" class="btn small" onClick=${onRetry}>Try again</button>`}
    ${error.detail && error.detail !== error.message && html`<details><summary>Technical details</summary><pre>${error.detail}</pre></details>`}
  <//>`;
}

const FOCUSABLE = 'a[href],button:not([disabled]),input:not([disabled]):not([type="hidden"]),select:not([disabled]),textarea:not([disabled]),summary,[tabindex]:not([tabindex="-1"])';

// returnFocus: a selector to focus on close when the element that opened the dialog is gone (a menu item, say)
export function Dialog({ title, onClose, children, wide, className = "", returnFocus }) {
  const ref = useRef(null);
  const titleId = useUid("dlg");
  useEffect(() => {
    const d = ref.current;
    const opener = document.activeElement; // what had focus when the dialog opened, to give it back on close
    if (d && !d.open) d.showModal();
    return () => {
      if (d && d.open) d.close();
      const back = opener && opener !== document.body && opener.isConnected ? opener : returnFocus && document.querySelector(returnFocus);
      if (back && typeof back.focus === "function") back.focus();
    };
  }, []);
  // keep Tab inside the dialog
  const trap = (e) => {
    if (e.key !== "Tab") return;
    const items = [...ref.current.querySelectorAll(FOCUSABLE)].filter((el) => el.offsetParent !== null || el === document.activeElement);
    if (!items.length) return;
    const first = items[0], last = items[items.length - 1];
    if (e.shiftKey && (document.activeElement === first || document.activeElement === ref.current)) { e.preventDefault(); last.focus(); }
    else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
  };
  return html`<dialog ref=${ref} role="dialog" aria-modal="true" class=${(wide ? "wide " : "") + className} aria-labelledby=${titleId}
    onClose=${onClose} onKeyDown=${trap} onClick=${(e) => { if (e.target === ref.current) onClose(); }}>
    <div class="dlg"><h2 id=${titleId}>${title}</h2>${children}</div>
  </dialog>`;
}

export function Toasts() {
  const { items } = useStore(toastStore);
  return html`<div class="toasts" role="status" aria-live="polite" aria-atomic="false">
    ${items.map((t) => html`<div class=${"toast " + (t.kind === "bad" ? "bad" : "")} key=${t.id}>
      <span>${t.message}</span>
      ${t.undo && html`<button type="button" onClick=${() => { dismissToast(t.id); t.undo(); }}>Undo</button>`}
      ${!t.undo && t.kind === "bad" && html`<button type="button" aria-label="Dismiss" onClick=${() => dismissToast(t.id)}>OK</button>`}
    </div>`)}
  </div>`;
}

export function ConfirmHost() {
  const { current } = useStore(confirmStore);
  if (!current) return null;
  const done = (v) => {
    confirmStore.set({ current: null });
    current.resolve(v);
  };
  return html`<${Dialog} title=${current.title} onClose=${() => done(false)}>
    ${current.body && html`<p class="sub">${current.body}</p>`}
    <div class="foot">
      <button type="button" class="btn quiet" onClick=${() => done(false)}>Cancel</button>
      <button type="button" class=${"btn " + (current.danger ? "danger fill" : "pri")} onClick=${() => done(true)}>${current.confirmLabel || "OK"}</button>
    </div>
  <//>`;
}

// A text input with a visible show/hide button, for API keys and passwords.
export function SecretInput({ id, value, onInput, placeholder }) {
  const [show, setShow] = useState(false);
  return html`<div class="with-btn">
    <input class="inp" id=${id} type=${show ? "text" : "password"} value=${value} placeholder=${placeholder} autocomplete="off" spellcheck="false"
      onInput=${(e) => onInput(e.target.value)} />
    <button type="button" class="btn line" aria-pressed=${show ? "true" : "false"} onClick=${() => setShow(!show)}>${show ? "Hide" : "Show"}</button>
  </div>`;
}

// Hold-to-act button: pointer down starts, up/leave/cancel stops. Space and Enter work too.
export function HoldButton({ onStart, onStop, label, children, className = "btn", disabled }) {
  const active = useRef(false);
  const start = (e) => {
    if (active.current || disabled) return;
    active.current = true;
    onStart(e);
  };
  const stop = (e) => {
    if (!active.current) return;
    active.current = false;
    onStop(e);
  };
  useEffect(() => () => { if (active.current) onStop(); }, []);
  return html`<button type="button" class=${className} aria-label=${label} disabled=${disabled}
    onPointerDown=${(e) => { e.preventDefault(); e.currentTarget.setPointerCapture && e.currentTarget.setPointerCapture(e.pointerId); start(e); }}
    onPointerUp=${stop} onPointerCancel=${stop} onLostPointerCapture=${stop} onBlur=${stop}
    onKeyDown=${(e) => { if ((e.key === " " || e.key === "Enter") && !e.repeat) { e.preventDefault(); start(e); } }}
    onKeyUp=${(e) => { if (e.key === " " || e.key === "Enter") { e.preventDefault(); stop(e); } }}
    onContextMenu=${(e) => e.preventDefault()}>${children}</button>`;
}

export function pageTitle(title) {
  document.title = title ? `${title} · WirePod` : "WirePod";
}

// Tabs with the keyboard pattern screen readers expect: arrows move, Home and End jump.
// tabs: [[id, label, icon], ...]
export function Tabs({ tabs, value, onChange, label, idPrefix }) {
  const move = (e, i) => {
    let n = null;
    if (e.key === "ArrowRight") n = (i + 1) % tabs.length;
    else if (e.key === "ArrowLeft") n = (i - 1 + tabs.length) % tabs.length;
    else if (e.key === "Home") n = 0;
    else if (e.key === "End") n = tabs.length - 1;
    if (n === null) return;
    e.preventDefault();
    onChange(tabs[n][0]);
    requestAnimationFrame(() => {
      const el = document.getElementById(`${idPrefix}-tab-${tabs[n][0]}`);
      if (el) el.focus();
    });
  };
  // on a narrow screen the strip scrolls: keep the active tab in view and fade the edge that has more
  const strip = useRef(null);
  const [edge, setEdge] = useState({ l: false, r: false });
  const measure = () => {
    const el = strip.current;
    if (el) setEdge({ l: el.scrollLeft > 2, r: el.scrollLeft + el.clientWidth < el.scrollWidth - 2 });
  };
  useEffect(() => {
    const el = strip.current;
    const on = el && el.querySelector('[aria-selected="true"]');
    if (el && on) el.scrollLeft = Math.max(0, on.offsetLeft - (el.clientWidth - on.offsetWidth) / 2);
    measure();
  }, [value]);
  useEffect(() => {
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, []);
  return html`<div class=${"tabs-wrap" + (edge.l ? " fade-l" : "") + (edge.r ? " fade-r" : "")}><div class="tabs" role="tablist" aria-label=${label} ref=${strip} onScroll=${measure}>
    ${tabs.map(([id, text, icon], i) => html`<button type="button" role="tab" key=${id} class="tab" id=${`${idPrefix}-tab-${id}`}
      aria-selected=${value === id ? "true" : "false"} aria-controls=${`${idPrefix}-panel`} tabindex=${value === id ? 0 : -1}
      onClick=${() => onChange(id)} onKeyDown=${(e) => move(e, i)}><${Icon} name=${icon} size=${18} />${text}</button>`)}
  </div></div>`;
}
