// A robot's page: header with status, then Personality / Memories / Settings / Controls / Activity.
import { html, useState, useEffect, useRef } from "../html.js";
import { Icon } from "../icons.js";
import { href, go, ROBOT_TABS } from "../router.js";
import { useRobots, robotName, refreshRobots, toast } from "../store.js";
import { sdk, describeError } from "../api.js";
import { Face, StatusPill, Battery, Tabs, Loading, Callout, Dialog, pageTitle } from "../ui.js";
import { Personality } from "./personality.js";
import { Memories } from "./memories.js";

const TABS = [
  ["personality", "Personality", "identity"],
  ["memories", "Memories", "memory"],
  ["settings", "Settings", "toggle"],
  ["controls", "Controls", "pad"],
  ["activity", "Activity", "pulse"],
];

function MoreMenu({ robot, onClose, onInfo }) {
  const ref = useRef(null);
  const name = robotName(robot);
  useEffect(() => {
    const away = (e) => { if (ref.current && !ref.current.contains(e.target) && !e.target.closest("[data-more]")) onClose(); };
    const esc = (e) => { if (e.key === "Escape") { onClose(); const b = document.querySelector("[data-more]"); if (b) b.focus(); } };
    document.addEventListener("pointerdown", away);
    document.addEventListener("keydown", esc);
    const first = ref.current && ref.current.querySelector("button");
    if (first) first.focus();
    return () => { document.removeEventListener("pointerdown", away); document.removeEventListener("keydown", esc); };
  }, []);
  const online = robot.state === "online";
  const disconnect = async () => {
    onClose();
    try {
      await sdk("disconnect", robot.esn, {}, { timeout: 15000 });
      toast(`Disconnected from ${name}. WirePod reconnects when you use it again.`);
    } catch (e) {
      toast(describeError(e, name).message, { kind: "bad" });
    }
  };
  return html`<div class="menu" role="menu" ref=${ref} aria-label=${`More for ${name}`}>
    <button type="button" role="menuitem" onClick=${() => { onClose(); onInfo(); }} disabled=${!online}><${Icon} name="info" size=${18} />Robot debug info</button>
    <button type="button" role="menuitem" onClick=${disconnect} disabled=${!online}><${Icon} name="plug" size=${18} />Disconnect from WirePod</button>
  </div>`;
}

function DebugInfo({ robot, onClose }) {
  const [text, setText] = useState(null);
  useEffect(() => {
    sdk("print_robot_info", robot.esn).then(setText, (e) => setText(describeError(e, robotName(robot)).message));
  }, []);
  return html`<${Dialog} title=${`${robotName(robot)} debug info`} onClose=${onClose} wide=${true}>
    ${text === null ? html`<${Loading} />` : html`<pre class="log" tabindex="0">${text}</pre>`}
    <div class="foot"><button type="button" class="btn pri" onClick=${onClose}>Close</button></div>
  <//>`;
}

export function RobotPage({ esn, tab }) {
  const { loaded, robots } = useRobots();
  const robot = robots.find((r) => r.esn === esn.toLowerCase());
  const [menu, setMenu] = useState(false);
  const [info, setInfo] = useState(false);
  const [sleeping, setSleeping] = useState(false);
  useEffect(() => pageTitle(robot ? robotName(robot) : "Robot"), [robot && robot.name, robot && robot.esn]);

  if (!robot) {
    if (!loaded) return html`<div class="page"><${Loading}>Opening robot…<//></div>`;
    return html`<div class="page"><h1>Robot not found</h1><p class="sub">WirePod doesn’t know a robot with the serial ${esn}. <a href=${href.robots}>Back to your robots</a>.</p></div>`;
  }
  const name = robotName(robot);
  const online = robot.state === "online";
  const sleep = async () => {
    setSleeping(true);
    try {
      await sdk("cloud_intent", robot.esn, { intent: "intent_system_sleep" });
      toast(`${name} is going to sleep`);
    } catch (e) {
      toast(describeError(e, name).message, { kind: "bad" });
    } finally {
      setSleeping(false);
    }
  };
  const sayHref = () => {
    go(href.robot(robot.esn, "controls"));
    setTimeout(() => { const i = document.getElementById("say-text"); if (i) i.focus(); }, 80);
  };

  let panel;
  switch (tab) {
    case "memories": panel = html`<${Memories} robot=${robot} key=${robot.esn} />`; break;
    case "personality": panel = html`<${Personality} robot=${robot} key=${robot.esn} />`; break;
    default: panel = html`<div class="card"><p class="sub">This section is on its way. The <a href="/sdkapp/settings.html?serial=${robot.esn}">classic Bot Settings</a> page has it in the meantime.</p></div>`;
  }

  return html`<div class="page" style="gap:18px">
    <div class="crumb"><a href=${href.robots}><${Icon} name="left" size=${16} />Robots</a><span aria-hidden="true">/</span><span aria-current="page">${name}</span></div>
    <div class="rhead">
      <${Face} robot=${robot} size=${72} />
      <div class="who">
        <h1>${name}</h1>
        <div class="meta"><span class="mono">${robot.esn}</span><${StatusPill} robot=${robot} />${robot.state !== "checking" && html`<${Battery} battery=${robot.battery} />`}</div>
      </div>
      <div class="acts">
        <button type="button" class="btn" onClick=${sayHref} disabled=${!online}><${Icon} name="send" size=${18} />Say something</button>
        <button type="button" class="btn" onClick=${sleep} disabled=${!online || sleeping}><${Icon} name="moon" size=${18} />Sleep</button>
        <button type="button" class="btn icon" data-more aria-label=${`More for ${name}`} aria-haspopup="menu" aria-expanded=${menu ? "true" : "false"} onClick=${() => setMenu(!menu)}><${Icon} name="more" size=${20} /></button>
        ${menu && html`<${MoreMenu} robot=${robot} onClose=${() => setMenu(false)} onInfo=${() => setInfo(true)} />`}
      </div>
    </div>
    ${robot.state === "offline" && html`<${Callout} icon="info">
      <span><b>${name} isn’t answering.</b> Check that it’s powered on and on the same Wi-Fi as this computer. You can still edit its personality and memories.</span>
      <button type="button" class="btn small" onClick=${refreshRobots}><${Icon} name="refresh" size=${16} />Check again</button>
    <//>`}
    <${Tabs} tabs=${TABS} value=${tab} onChange=${(t) => go(href.robot(robot.esn, t))} label="Robot sections" idPrefix="rt" />
    <div role="tabpanel" id="rt-panel" aria-labelledby=${`rt-tab-${tab}`} tabindex="-1">${panel}</div>
    ${info && html`<${DebugInfo} robot=${robot} onClose=${() => setInfo(false)} />`}
  </div>`;
}
