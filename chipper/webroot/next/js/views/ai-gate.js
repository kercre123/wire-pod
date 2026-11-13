// What the Personality and Memories screens show while AI is off: an invitation, not an error.
// Their data is kept; turning AI on brings it all back.
import { html, useState } from "../html.js";
import { Icon } from "../icons.js";
import { href } from "../router.js";
import { getJSON, postJSON, describeError } from "../api.js";
import { robotName, toast } from "../store.js";
import { loadConfig } from "../server.js";
import { Callout } from "../ui.js";

export function AiUnlock({ robot, tab }) {
  const name = robotName(robot);
  const back = href.robot(robot.esn, tab);
  return html`<div class="unlock" role="region" aria-labelledby="unlock-h">
    <span class="badge"><${Icon} name="spark" size=${26} /></span>
    <div><h2 id="unlock-h">Unlock ${name}’s personality</h2>
      <p class="sub" style="margin-top:6px;max-width:56ch">Give ${name} its own personality, let it remember things about you, and answer any question.</p></div>
    <ul>
      <li><${Icon} name="identity" size=${20} /><span><b>A personality of its own.</b> Choose how ${name} talks and what it cares about.</span></li>
      <li><${Icon} name="memory" size=${20} /><span><b>Remembers you.</b> Names, birthdays, the dog’s favorite toy.</span></li>
      <li><${Icon} name="mic" size=${20} /><span><b>Answers anything.</b> Not just the commands it already knows.</span></li>
    </ul>
    <div class="chips"><a class="btn pri" href=${href.settings("ai", back)}><${Icon} name="spark" size=${18} />Turn on AI</a></div>
    <details><summary>Learn more</summary>
      <p style="margin:8px 0 0">Run a model on this computer with LM Studio or Ollama and it’s free and private: nothing you say leaves your home. Or use a cloud service such as OpenAI, which starts faster but sends what you say to them and may cost money.</p></details>
    <p class="hint" style="margin:0"><span>Anything ${name} already knows is kept and comes back when AI is on.</span></p>
  </div>`;
}

// AI is on but the AI may not run robot commands: the robot can't add memories by itself. One click fixes it.
export function MemoriesNote({ robot, canAct }) {
  const name = robotName(robot);
  const [busy, setBusy] = useState(false);
  const turnOn = async () => {
    setBusy(true);
    try {
      const kg = await getJSON("/api/get_kg_api");
      await postJSON("/api/set_kg_api", { ...kg, commands_enable: true });
      await loadConfig(); // the whole UI sees the new setting at once
      toast("The AI can now run robot commands");
    } catch (e) {
      toast(describeError(e).message, { kind: "bad" });
    }
    setBusy(false);
  };
  return html`<${Callout}>
    <span>${name} won’t save new memories by itself until “Let the AI run robot commands” is on. You can still add and delete them here.</span>
    ${canAct ? html`<button type="button" class="btn small" onClick=${turnOn} disabled=${busy}>${busy ? "Turning on…" : "Turn it on"}</button>`
      : html`<a class="btn small" href=${href.settings("ai")}>Open AI settings</a>`}
  <//>`;
}
