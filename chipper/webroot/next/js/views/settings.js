// Settings: everything that applies to the whole WirePod server, in five sections.
import { html, useState, useEffect, useRef } from "../html.js";
import { Icon } from "../icons.js";
import { href, go } from "../router.js";
import { getJSON, getText, postJSON, describeError } from "../api.js";
import { toast, confirmDialog, useStore, prefsStore, setPref } from "../store.js";
import { Switch, Segmented, Loading, ErrorNote, Dialog, Callout, SecretInput, pageTitle, useUid } from "../ui.js";
import { useServerInfo, loadVersion, loadConfig, versionLabel } from "../server.js";
import { STT_LANGUAGES, KG_PROVIDERS, aiSummary, languageLabel } from "../data.js";
import { ACCENTS } from "../theme.js";
import { TechLog } from "./activity.js";

const UPDATE_GUIDE = "https://github.com/kercre123/wire-pod/wiki/Things-to-Know#updating-wire-pod";
const DEFAULT_PROMPT = "You are a helpful, animated robot called Vector. Keep the response concise yet informative.";

// robot reactions a custom command can send (the same list the classic UI offers)
const INTENTS = ["intent_greeting_hello", "intent_names_ask", "intent_imperative_eyecolor", "intent_character_age", "intent_explore_start", "intent_system_charger", "intent_system_sleep", "intent_greeting_goodmorning", "intent_greeting_goodnight", "intent_greeting_goodbye", "intent_seasonal_happynewyear", "intent_seasonal_happyholidays", "intent_amazon_signin", "intent_imperative_forward", "intent_imperative_turnaround", "intent_imperative_turnleft", "intent_imperative_turnright", "intent_play_rollcube", "intent_play_popawheelie", "intent_play_fistbump", "intent_play_blackjack", "intent_imperative_affirmative", "intent_imperative_negative", "intent_photo_take_extend", "intent_imperative_praise", "intent_imperative_abuse", "intent_weather_extend", "intent_imperative_apologize", "intent_imperative_backup", "intent_imperative_volumedown", "intent_imperative_volumeup", "intent_imperative_lookatme", "intent_imperative_volumelevel_extend", "intent_imperative_shutup", "intent_names_username_extend", "intent_imperative_come", "intent_imperative_love", "intent_knowledge_promptquestion", "intent_clock_checktimer", "intent_global_stop_extend", "intent_clock_settimer_extend", "intent_clock_time", "intent_imperative_quiet", "intent_imperative_dance", "intent_play_pickupcube", "intent_imperative_fetchcube", "intent_imperative_findcube", "intent_play_anytrick", "intent_message_recordmessage_extend", "intent_message_playmessage_extend", "intent_blackjack_hit", "intent_blackjack_stand", "intent_play_keepaway"];

const SECTIONS = [
  ["ai", "AI & knowledge", "spark"],
  ["voice", "Voice & commands", "mic"],
  ["integrations", "Integrations", "plug"],
  ["server", "Server & updates", "server"],
  ["appearance", "Appearance", "palette"],
];

function useLoad(loader, deps) {
  const [st, setSt] = useState({ data: null, error: null });
  const reload = async () => {
    try { setSt({ data: await loader(), error: null }); } catch (e) { setSt({ data: null, error: describeError(e) }); }
  };
  useEffect(() => { reload(); }, deps);
  return { ...st, reload, setData: (d) => setSt({ data: d, error: null }) };
}

const fail = (e) => toast(describeError(e).message, { kind: "bad" });

function SaveBar({ dirty, busy, onSave, onDiscard }) {
  if (!dirty) return null;
  return html`<div class="savebar" role="region" aria-label="Unsaved changes">
    <span><span class="dot warn"></span>Unsaved changes</span>
    <div class="b"><button type="button" class="btn quiet" onClick=${onDiscard} disabled=${busy}>Discard</button>
      <button type="button" class="btn pri" onClick=${onSave} disabled=${busy}>${busy ? "Saving…" : "Save changes"}</button></div>
  </div>`;
}

// a slider that can also be "use the default" (stored as 0)
function OptionalRange({ id, label, value, min, max, step, onChange }) {
  const isDefault = !value;
  return html`<div>
    <label for=${id}>${label}: <span>${isDefault ? "Default" : value}</span></label>
    <input class="range" id=${id} type="range" min=${min} max=${max} step=${step} value=${isDefault ? (min + max) / 2 : value} disabled=${isDefault}
      onInput=${(e) => onChange(parseFloat(e.target.value))} />
    <div class="opt flush"><div><small>Use the server’s default</small></div><${Switch} checked=${isDefault} label=${`Default ${label.toLowerCase()}`} onChange=${(on) => onChange(on ? 0 : (min + max) / 2)} /></div>
  </div>`;
}

// ---------- AI & knowledge ----------

const PROVIDER_TILES = [["", "Off", "No AI answers"], ["houndify", "Houndify", "Built-in service"], ["openai", "OpenAI", "ChatGPT"], ["together", "Together", "Hosted models"], ["custom", "Your own server", "OpenAI-compatible"]];

function blankFor(p, keep) {
  return { enable: p !== "", provider: p, key: "", model: "", id: "", endpoint: "", openai_prompt: "", openai_voice: p === "openai" ? "fable" : "", openai_voice_with_english: false, temp: 0, top_p: 0,
    intentgraph: keep.intentgraph, save_chat: keep.save_chat, commands_enable: keep.commands_enable, robotName: keep.robotName || "" };
}

function AIPane() {
  const id = useUid("ai");
  const { data, error, reload, setData } = useLoad(() => getJSON("/api/get_kg_api"), []);
  const [kg, setKg] = useState(null);
  const [busy, setBusy] = useState(false);
  const [test, setTest] = useState(null);
  useEffect(() => { if (data) { setKg({ ...data, provider: data.enable ? data.provider : "" }); setTest(null); } }, [data]);
  if (error) return html`<${ErrorNote} error=${error} onRetry=${reload} />`;
  if (!kg) return html`<${Loading} />`;

  const saved = { ...data, provider: data.enable ? data.provider : "" };
  const dirty = JSON.stringify(kg) !== JSON.stringify(saved);
  const set = (patch) => { setKg({ ...kg, ...patch }); setTest(null); };
  const p = kg.provider;
  const choose = (next) => {
    setTest(null);
    setKg(next === saved.provider ? { ...saved } : blankFor(next, kg));
  };

  const payload = () => {
    const d = { ...blankFor(p, kg), robotName: kg.robotName || "" };
    if (p === "openai") Object.assign(d, { key: kg.key, openai_prompt: kg.openai_prompt, openai_voice: kg.openai_voice, openai_voice_with_english: kg.openai_voice_with_english });
    if (p === "custom") Object.assign(d, { key: kg.key, model: kg.model, openai_prompt: kg.openai_prompt, endpoint: kg.endpoint, temp: kg.temp || 0, top_p: kg.top_p || 0 });
    if (p === "together") Object.assign(d, { key: kg.key, model: kg.model, openai_prompt: kg.openai_prompt });
    if (p === "houndify") Object.assign(d, { key: kg.key, id: kg.id });
    d.intentgraph = p !== "" && kg.intentgraph;
    d.save_chat = (p === "openai" || p === "custom" || p === "together") && kg.save_chat;
    d.commands_enable = (p === "openai" || p === "custom" || p === "together") && kg.commands_enable;
    return d;
  };
  const save = async () => {
    setBusy(true);
    try {
      const body = payload();
      await postJSON("/api/set_kg_api", body);
      setData(body);
      loadConfig();
      toast("AI settings saved");
    } catch (e) { fail(e); }
    setBusy(false);
  };
  const runTest = async () => {
    setTest({ state: "running" });
    try {
      const r = JSON.parse(await postJSON("/api/test_llm", { provider: p, key: kg.key, endpoint: kg.endpoint }, { timeout: 15000 }));
      setTest({ state: r.ok ? "ok" : "bad", message: r.message });
    } catch (e) { setTest({ state: "bad", message: describeError(e).message }); }
  };
  const deleteChats = async () => {
    const ok = await confirmDialog({ title: "Delete saved conversations?", body: "This removes the saved conversations of every robot. It can’t be undone.", confirmLabel: "Delete", danger: true });
    if (!ok) return;
    try { await getText("/api/delete_chats"); toast("Saved conversations deleted"); } catch (e) { fail(e); }
  };

  const llm = p === "openai" || p === "custom" || p === "together";
  const field = (key, label, extra = {}) => html`<div><label for=${`${id}-${key}`}>${label}</label>
    ${extra.secret ? html`<${SecretInput} id=${`${id}-${key}`} value=${kg[key] || ""} onInput=${(v) => set({ [key]: v })} />`
      : html`<input class="inp" id=${`${id}-${key}`} value=${kg[key] || ""} placeholder=${extra.placeholder} autocomplete="off" spellcheck="false" onInput=${(e) => set({ [key]: e.target.value })} />`}
    ${extra.hint && html`<div class="hint"><span>${extra.hint}</span></div>`}</div>`;

  return html`<div class="stack"><div class="card">
    <h2>AI & knowledge</h2><p class="sub">Where robots get answers when you ask a question.</p>
    <span class="lbl" id=${`${id}-prov`}>Provider</span>
    <div class="prov" role="group" aria-labelledby=${`${id}-prov`}>
      ${PROVIDER_TILES.map(([v, n, d]) => html`<button type="button" key=${v} aria-pressed=${p === v ? "true" : "false"} onClick=${() => choose(v)}>${n}<small>${d}</small></button>`)}
    </div>
    ${p === "" && html`<p class="sub" style="margin-top:14px">With no provider, robots only use WirePod’s built-in commands and can’t answer general questions.</p>`}
    <div class="stack" style="margin-top:18px">
      ${p === "houndify" && html`<${Callout}><span>Create a free account at <a href="https://www.houndify.com/signup" target="_blank" rel="noopener">houndify.com</a>, add a domain, and enter its Client ID and Key.</span><//>`}
      ${p === "houndify" && html`<div class="row2">${field("id", "Client ID")}${field("key", "Client key", { secret: true })}</div>`}
      ${p === "openai" && field("key", "OpenAI key", { secret: true })}
      ${p === "together" && html`<${Callout}><span>Create an account at <a href="https://www.together.xyz" target="_blank" rel="noopener">together.xyz</a>, then enter a key. Leave the model blank for the default (meta-llama/Llama-3-70b-chat-hf).</span><//>`}
      ${p === "together" && html`<div class="row2">${field("model", "Model", { placeholder: "meta-llama/Llama-3-70b-chat-hf" })}${field("key", "Together key", { secret: true })}</div>`}
      ${p === "custom" && html`<${Callout}><span>Any host with an OpenAI-compatible API works, such as Ollama or LM Studio. For advanced users.</span><//>`}
      ${p === "custom" && html`<div class="row2">${field("endpoint", "Server address", { placeholder: "http://localhost:11434/v1" })}${field("model", "Model", { placeholder: "llama3" })}</div>`}
      ${p === "custom" && field("key", "API key", { secret: true, hint: "Local servers like Ollama accept any text here, for example “ollama”." })}
      ${llm && html`<div><label for=${`${id}-prompt`}>Instructions for the AI</label>
        <textarea class="inp" id=${`${id}-prompt`} rows="3" placeholder=${DEFAULT_PROMPT} value=${kg.openai_prompt || ""} onInput=${(e) => set({ openai_prompt: e.target.value })}></textarea>
        <div class="hint"><span>Leave blank to use the default. Each robot’s name, personality and memories are added in front of this.</span></div></div>`}
      ${p === "openai" && html`<div><label for=${`${id}-voice`}>Voice for languages other than English</label>
        <select class="inp" id=${`${id}-voice`} onChange=${(e) => set({ openai_voice: e.target.value })}>
          ${["fable", "alloy", "echo", "onyx", "nova", "shimmer"].map((v) => html`<option value=${v} selected=${(kg.openai_voice || "fable") === v}>${v[0].toUpperCase() + v.slice(1)}</option>`)}</select>
        <div class="hint"><span><a href="https://platform.openai.com/docs/guides/text-to-speech/voice-options" target="_blank" rel="noopener">Hear each voice</a></span></div></div>`}
      ${p === "custom" && html`<div class="row2">
        <${OptionalRange} id=${`${id}-temp`} label="Creativity" min=${0.05} max=${2} step=${0.05} value=${kg.temp} onChange=${(v) => set({ temp: v })} />
        <${OptionalRange} id=${`${id}-topp`} label="Variety" min=${0.05} max=${1} step=${0.05} value=${kg.top_p} onChange=${(v) => set({ top_p: v })} /></div>`}
    </div>
    ${p !== "" && html`<div style="margin-top:8px">
      <div class="opt"><div>Use AI to understand what you mean<small>Forwards what was said to the AI when WirePod doesn’t recognize a command</small></div><${Switch} checked=${!!kg.intentgraph} label="Use AI to understand what you mean" onChange=${(v) => set({ intentgraph: v })} /></div>
      ${llm && html`<div class="opt"><div>Let the AI run robot commands<small>For example animations, or “drive forward”, while it answers</small></div><${Switch} checked=${!!kg.commands_enable} label="Let the AI run robot commands" onChange=${(v) => set({ commands_enable: v })} /></div>`}
      ${llm && html`<div class="opt"><div>Save conversations<small>Stored on this computer. Lets robots remember the conversation, and powers “I have a question”.${kg.save_chat && !kg.commands_enable ? " Conversations also need “Let the AI run robot commands” turned on." : ""}</small></div><${Switch} checked=${!!kg.save_chat} label="Save conversations" onChange=${(v) => set({ save_chat: v })} /></div>`}
      ${p === "openai" && html`<div class="opt"><div>Use the OpenAI voice for English too</div><${Switch} checked=${!!kg.openai_voice_with_english} label="Use the OpenAI voice for English" onChange=${(v) => set({ openai_voice_with_english: v })} /></div>`}
    </div>`}
    <div class="dangerzone" style="margin-top:14px">
      <div class="chips">
        ${llm && html`<button type="button" class="btn line" onClick=${runTest} disabled=${test && test.state === "running"}>${test && test.state === "running" ? "Testing…" : "Test connection"}</button>`}
        ${llm && html`<button type="button" class="btn danger" onClick=${deleteChats}>Delete saved conversations</button>`}
      </div>
      ${test && test.message && html`<span class=${"pill " + (test.state === "ok" ? "ok" : "bad")} role="status">${test.message}</span>`}
    </div>
  </div>
  <${SaveBar} dirty=${dirty} busy=${busy} onSave=${save} onDiscard=${() => { setKg({ ...saved }); setTest(null); }} /></div>`;
}

// ---------- Voice & commands ----------

function SpeechCard() {
  const id = useUid("stt");
  const { data, error, reload } = useLoad(() => getJSON("/api/get_stt_info"), []);
  const [lang, setLang] = useState(null);
  const [state, setState] = useState({ phase: "idle", text: "" });
  useEffect(() => { if (data && lang === null) setLang(data.language); }, [data]);
  if (error) return html`<div class="card"><${ErrorNote} error=${error} onRetry=${reload} /></div>`;
  if (!data || lang === null) return html`<div class="card"><${Loading} /></div>`;
  const supported = data.provider === "vosk" || data.provider === "whisper.cpp";
  const langs = STT_LANGUAGES.some(([c]) => c === lang) ? STT_LANGUAGES : [...STT_LANGUAGES, [lang, lang]];

  const poll = async () => {
    for (let i = 0; i < 1200; i++) {
      await new Promise((r) => setTimeout(r, 600));
      let t;
      try { t = await getText("/api/get_download_status"); } catch (e) { setState({ phase: "error", text: describeError(e).message }); return; }
      if (t.includes("success")) { setState({ phase: "done", text: "Language switched." }); reload(); return; }
      if (t.includes("error")) { setState({ phase: "error", text: t }); return; }
      setState({ phase: "busy", text: t.includes("not downloading") ? "Starting the download…" : t });
    }
  };
  const apply = async () => {
    setState({ phase: "busy", text: "Setting language…" });
    try {
      const t = await postJSON("/api/set_stt_info", { language: lang }, { timeout: 60000 });
      if (t.includes("downloading")) { setState({ phase: "busy", text: "Downloading the speech model…" }); await poll(); }
      else { setState({ phase: "done", text: "Language switched." }); reload(); }
    } catch (e) { setState({ phase: "error", text: describeError(e).message }); }
  };
  const busy = state.phase === "busy";
  return html`<div class="card">
    <h2>Speech</h2><p class="sub">The language robots listen for.</p>
    ${!supported && html`<${Callout}><span>The language can only be changed while speech recognition is Vosk or Whisper. It is currently “${data.provider}”.</span><//>`}
    <div class="with-btn" style="margin-top:8px">
      <select class="inp" id=${id} aria-label="Speech language" disabled=${!supported || busy} onChange=${(e) => { setLang(e.target.value); setState({ phase: "idle", text: "" }); }}>
        ${langs.map(([c, l]) => html`<option value=${c} selected=${c === lang}>${l}</option>`)}</select>
      <button type="button" class="btn pri" onClick=${apply} disabled=${!supported || busy || lang === data.language}>Set language</button>
    </div>
    <div class="hint" role="status" aria-live="polite"><span>${busy ? html`<span class="spinner" style="width:12px;height:12px;border-width:2px"></span>` : state.phase === "error" ? html`<span class="err">${state.text}</span>` : html`<span><span class="dot"></span> ${state.phase === "done" ? state.text : `Listening in ${languageLabel(data.language)}`}</span>`} ${busy && state.text}</span></div>
  </div>`;
}

function CommandDialog({ initial, index, onClose, onSaved }) {
  const id = useUid("cmd");
  const edit = index !== null;
  const [f, setF] = useState(initial || { name: "", description: "", utterances: "", intent: INTENTS[0], paramname: "", paramvalue: "", exec: "", execargs: "", luascript: "" });
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);
  const up = (k) => (e) => setF({ ...f, [k]: e.target.value });
  const submit = async (e) => {
    e.preventDefault();
    const utterances = f.utterances.split(",").map((s) => s.trim()).filter(Boolean);
    if (!f.name.trim() || !f.description.trim() || utterances.length === 0) { setErr("A command needs a name, a description and at least one phrase."); return; }
    setBusy(true); setErr("");
    const body = { name: f.name.trim(), description: f.description.trim(), utterances, intent: f.intent, params: { paramname: f.paramname.trim(), paramvalue: f.paramvalue.trim() },
      exec: f.exec.trim(), execargs: f.execargs.split(",").map((s) => s.trim()).filter(Boolean), luascript: f.luascript };
    try {
      await postJSON(edit ? "/api/edit_custom_intent" : "/api/add_custom_intent", edit ? { number: index + 1, ...body } : body);
      toast(edit ? "Command updated" : "Command added");
      onSaved();
    } catch (e2) { setErr(describeError(e2).message); setBusy(false); }
  };
  return html`<${Dialog} title=${edit ? "Edit command" : "Add a command"} wide=${true} onClose=${onClose}>
    <form class="stack" onSubmit=${submit}>
      <div class="row2">
        <div><label for=${`${id}-n`}>Name</label><input class="inp" id=${`${id}-n`} value=${f.name} onInput=${up("name")} /></div>
        <div><label for=${`${id}-d`}>Description</label><input class="inp" id=${`${id}-d`} value=${f.description} onInput=${up("description")} /></div>
      </div>
      <div><label for=${`${id}-u`}>Phrases that trigger it</label><input class="inp" id=${`${id}-u`} value=${f.utterances} onInput=${up("utterances")} placeholder="lights on, turn on the lights" />
        <div class="hint"><span>Separate phrases with commas.</span></div></div>
      <div><label for=${`${id}-i`}>Robot reaction</label>
        <select class="inp" id=${`${id}-i`} onChange=${up("intent")}>${INTENTS.map((i) => html`<option value=${i} selected=${i === f.intent}>${i}</option>`)}</select>
        <div class="hint"><span>The reaction the robot performs after your script runs.</span></div></div>
      <details><summary class="lbl" style="cursor:pointer">Scripts and parameters (advanced)</summary>
        <div class="stack" style="margin-top:12px">
          ${edit && html`<p class="hint"><span>When editing, leave a field empty to keep what it has now.</span></p>`}
          <div class="row2">
            <div><label for=${`${id}-pn`}>Parameter name</label><input class="inp" id=${`${id}-pn`} value=${f.paramname} onInput=${up("paramname")} /></div>
            <div><label for=${`${id}-pv`}>Parameter value</label><input class="inp" id=${`${id}-pv`} value=${f.paramvalue} onInput=${up("paramvalue")} /></div>
          </div>
          <div class="row2">
            <div><label for=${`${id}-e`}>Program to run</label><input class="inp" id=${`${id}-e`} value=${f.exec} onInput=${up("exec")} placeholder="/path/to/script" /></div>
            <div><label for=${`${id}-ea`}>Program arguments</label><input class="inp" id=${`${id}-ea`} value=${f.execargs} onInput=${up("execargs")} placeholder="comma, separated" /></div>
          </div>
          <div><label for=${`${id}-l`}>Lua code to run</label><textarea class="inp" id=${`${id}-l`} rows="6" style="font-family:var(--mono)" value=${f.luascript} onInput=${up("luascript")}></textarea>
            <div class="hint"><span><a href="https://github.com/kercre123/wire-pod/wiki/Custom-Intents" target="_blank" rel="noopener">How custom commands work</a></span></div></div>
        </div></details>
      ${err && html`<p class="err" role="alert">${err}</p>`}
      <div class="foot"><button type="button" class="btn quiet" onClick=${onClose}>Cancel</button><button type="submit" class="btn pri" disabled=${busy}>${busy ? "Saving…" : edit ? "Save command" : "Add command"}</button></div>
    </form>
  <//>`;
}

function CommandsCard() {
  // the server answers 400 while there are no commands yet, which just means an empty list
  const { data, error, reload } = useLoad(async () => {
    try { return await getJSON("/api/get_custom_intents_json"); }
    catch (e) { if (e.status === 400) return []; throw e; }
  }, []);
  const [dlg, setDlg] = useState(null); // { index, initial }
  const list = data || [];
  const edit = (c, i) => setDlg({ index: i, initial: { name: c.name, description: c.description, utterances: (c.utterances || []).join(", "), intent: c.intent || INTENTS[0], paramname: (c.params && c.params.paramname) || "", paramvalue: (c.params && c.params.paramvalue) || "", exec: c.exec || "", execargs: (c.execargs || []).filter(Boolean).join(", "), luascript: c.luascript || "" } });
  const remove = async (c, i) => {
    const ok = await confirmDialog({ title: `Remove “${c.name}”?`, body: "Robots will stop reacting to its phrases.", confirmLabel: "Remove", danger: true });
    if (!ok) return;
    try { await postJSON("/api/remove_custom_intent", { number: i + 1 }); toast("Command removed"); reload(); } catch (e) { fail(e); }
  };
  return html`<div class="card">
    <div class="top"><div><h2>Custom commands</h2><p class="sub">Phrases that trigger something you define.</p></div>
      <button type="button" class="btn" onClick=${() => setDlg({ index: null, initial: null })}><${Icon} name="plus" size=${18} />Add</button></div>
    ${error && html`<${ErrorNote} error=${error} onRetry=${reload} />`}
    ${data === null && !error && html`<${Loading} />`}
    ${data && list.length === 0 && html`<p class="empty">No custom commands yet.</p>`}
    <div class="list">${list.map((c, i) => html`<div class="item" key=${i}>
      <div class="t"><b>${c.name}</b><small>${(c.utterances || []).map((u) => `“${u}”`).join(", ")}${c.luascript || c.exec ? " · Runs a script" : ""}</small></div>
      <div class="cmd-actions"><button type="button" class="btn quiet" aria-label=${`Edit ${c.name}`} onClick=${() => edit(c, i)}>Edit</button>
        <button type="button" class="btn icon quiet" aria-label=${`Remove ${c.name}`} onClick=${() => remove(c, i)}><${Icon} name="trash" size=${18} /></button></div>
    </div>`)}</div>
    ${dlg && html`<${CommandDialog} initial=${dlg.initial} index=${dlg.index} onClose=${() => setDlg(null)} onSaved=${() => { setDlg(null); reload(); }} />`}
  </div>`;
}

const VoicePane = () => html`<div class="stack"><${SpeechCard} /><${CommandsCard} /></div>`;

// ---------- Integrations ----------

function IntegrationsPane() {
  const id = useUid("wx");
  const { data, error, reload, setData } = useLoad(() => getJSON("/api/get_weather_api"), []);
  const [w, setW] = useState(null);
  const [busy, setBusy] = useState(false);
  useEffect(() => { if (data) setW({ provider: data.enable ? data.provider : "", key: data.key || "" }); }, [data]);
  if (error) return html`<${ErrorNote} error=${error} onRetry=${reload} />`;
  if (!w) return html`<${Loading} />`;
  const saved = { provider: data.enable ? data.provider : "", key: data.key || "" };
  const dirty = w.provider !== saved.provider || w.key !== saved.key;
  const save = async () => {
    if (w.provider && !w.key.trim()) { toast("Enter an API key first.", { kind: "bad" }); return; }
    setBusy(true);
    try {
      await postJSON("/api/set_weather_api", { provider: w.provider, key: w.key });
      setData({ ...data, enable: !!w.provider, provider: w.provider, key: w.key.trim() });
      loadConfig();
      toast("Weather saved");
    } catch (e) { fail(e); }
    setBusy(false);
  };
  return html`<div class="stack"><div class="card">
    <h2>Weather</h2><p class="sub">Used when someone asks about the weather.</p>
    <div class="stack">
      <div><label for=${`${id}-p`}>Provider</label>
        <select class="inp" id=${`${id}-p`} onChange=${(e) => setW({ ...w, provider: e.target.value })}>
          <option value="" selected=${w.provider === ""}>None</option>
          <option value="openweathermap.org" selected=${w.provider === "openweathermap.org"}>OpenWeatherMap</option>
          <option value="weatherapi.com" selected=${w.provider === "weatherapi.com"}>WeatherAPI</option></select></div>
      ${w.provider && html`<div><label for=${`${id}-k`}>API key</label><${SecretInput} id=${`${id}-k`} value=${w.key} onInput=${(v) => setW({ ...w, key: v })} /></div>`}
    </div>
    <div class="dangerzone" style="margin-top:14px"><span class=${"pill " + (saved.provider ? "ok" : "")}>${saved.provider ? html`<span class="dot"></span>Set up` : "Not set up"}</span></div>
  </div><${SaveBar} dirty=${dirty} busy=${busy} onSave=${save} onDiscard=${() => setW({ ...saved })} /></div>`;
}

// ---------- Server & updates ----------

function VersionCard() {
  const { version: v, versionState } = useServerInfo();
  return html`<div class="card">
    <div class="top"><div>
      <h2>WirePod ${v ? versionLabel(v) : ""}</h2>
      <p class="sub" role="status">${versionState === "loading" ? "Checking for updates…" : versionState === "unavailable" ? "Couldn’t check for updates. That needs an internet connection."
        : v && v.avail ? `An update is available${v.fromsource ? ` (commit ${v.currentcommit})` : ` (${v.currentversion})`}.` : "You’re on the latest version."}</p>
      ${v && !v.fromsource && v.installedcommit && html`<p class="mono">Based on wire-pod commit ${v.installedcommit}</p>`}
    </div>
    <div class="chips"><button type="button" class="btn line" onClick=${loadVersion} disabled=${versionState === "loading"}><${Icon} name="refresh" size=${18} />Check again</button>
      ${v && v.avail && html`<a class="btn pri" href=${UPDATE_GUIDE} target="_blank" rel="noopener"><${Icon} name="external" size=${18} />How to update</a>`}</div></div>
  </div>`;
}

function ConnectionCard() {
  const { config } = useServerInfo();
  const [mode, setMode] = useState(null); // "ip" | "ep"
  const [port, setPort] = useState("443");
  const [busy, setBusy] = useState(false);
  useEffect(() => { if (config && mode === null) { setMode(config.server.epconfig ? "ep" : "ip"); setPort(config.server.port || "443"); } }, [config]);
  if (!config || mode === null) return html`<div class="card"><${Loading} /></div>`;
  const current = config.server.epconfig ? "ep" : "ip";
  const changed = mode !== current || (mode === "ip" && port !== (config.server.port || "443"));
  const run = async (fn, ok) => {
    setBusy(true);
    try {
      const t = await fn();
      if (/^\s*error/i.test(t)) throw new Error(t.replace(/^\s*error:\s*/i, ""));
      toast(ok);
      loadConfig();
    } catch (e) { toast(e.message ? e.message : describeError(e).message, { kind: "bad" }); }
    setBusy(false);
  };
  const apply = async () => {
    if (mode === "ip" && !/^\d+$/.test(port)) { toast("The port must be a number.", { kind: "bad" }); return; }
    const okc = await confirmDialog({ title: "Change how robots connect?", body: "WirePod’s robot connection restarts. Robots reconnect on their own, but one that was paired for the other method may need to be set up again.", confirmLabel: "Change" });
    if (okc) run(() => getText(mode === "ep" ? "/api-chipper/use_ep" : `/api-chipper/use_ip?port=${encodeURIComponent(port)}`, { timeout: 60000 }), "Connection updated");
  };
  const restart = () => run(() => getText("/api-chipper/restart", { timeout: 30000 }), "WirePod’s robot connection restarted");
  const certs = async () => {
    const okc = await confirmDialog({ title: "Regenerate certificates?", body: "Do this if robots can’t connect securely. Robots may need a moment to trust the new certificate.", confirmLabel: "Regenerate" });
    if (okc) run(() => getText("/api/generate_certs", { timeout: 60000 }), "Certificates regenerated");
  };
  return html`<div class="card">
    <h3>Connection</h3><p class="sub">How robots find this computer.</p>
    <${Segmented} label="Connection method" options=${[["ip", "Same Wi-Fi (IP)"], ["ep", "Escape pod"]]} value=${mode} onChange=${setMode} />
    ${mode === "ip" && html`<div style="margin-top:12px;max-width:220px"><label for="srv-port">Port</label><input class="inp" id="srv-port" inputmode="numeric" value=${port} onInput=${(e) => setPort(e.target.value)} /></div>`}
    ${changed && html`<div style="margin-top:12px"><button type="button" class="btn pri" onClick=${apply} disabled=${busy}>Apply connection change</button></div>`}
    <div class="opt" style="margin-top:10px"><div>Restart robot connection<small>Needed after some changes</small></div><button type="button" class="btn" onClick=${restart} disabled=${busy}>Restart</button></div>
    <div class="opt"><div>Security certificates<small>Regenerate if robots can’t connect securely</small></div><button type="button" class="btn" onClick=${certs} disabled=${busy}>Regenerate</button></div>
  </div>`;
}

const ServerPane = () => html`<div class="stack"><${VersionCard} /><${ConnectionCard} />
  <div class="card"><h3>Live log</h3><p class="sub">What WirePod is doing right now. It updates by itself.</p><${TechLog} label="" /></div></div>`;

// ---------- Appearance ----------

function AppearancePane() {
  const { prefs } = useStore(prefsStore);
  return html`<div class="card"><h2>Appearance</h2><p class="sub">How this app looks. Saves instantly, on this device.</p>
    <div class="opt"><div>Theme</div><${Segmented} label="Theme" options=${[["system", "System"], ["light", "Light"], ["dark", "Dark"]]} value=${prefs.theme} onChange=${(v) => setPref({ theme: v })} /></div>
    <div class="opt"><div id="ap-acc">Accent color${prefs.accentFollowsRobot && html`<small>Following the robot you’re viewing</small>`}</div>
      <div class="swatches" role="group" aria-labelledby="ap-acc">${ACCENTS.map((a) => html`<button type="button" key=${a.id} style=${`--c:${a.swatch}`} aria-label=${a.name} aria-pressed=${prefs.accent === a.id ? "true" : "false"}
        disabled=${prefs.accentFollowsRobot} onClick=${() => setPref({ accent: a.id })}></button>`)}</div></div>
    <div class="opt"><div>Accent follows the robot’s eye color<small>Highlights change to match the robot you’re viewing</small></div><${Switch} checked=${prefs.accentFollowsRobot} label="Accent follows the robot’s eye color" onChange=${(v) => setPref({ accentFollowsRobot: v })} /></div>
    <div class="opt"><div>Text size</div><${Segmented} label="Text size" options=${[["default", "Default"], ["large", "Large"]]} value=${prefs.size} onChange=${(v) => setPref({ size: v })} /></div>
    <div class="opt"><div>Font<small>Retro uses the classic UI’s pixel font</small></div><${Segmented} label="Font" options=${[["default", "Default"], ["retro", "Retro"]]} value=${prefs.font} onChange=${(v) => setPref({ font: v })} /></div>
  </div>`;
}

// ---------- the page ----------

function useWide() {
  const q = window.matchMedia("(min-width: 900px)");
  const [w, setW] = useState(q.matches);
  useEffect(() => { const f = () => setW(q.matches); q.addEventListener("change", f); return () => q.removeEventListener("change", f); }, []);
  return w;
}

export function Settings({ section }) {
  const wide = useWide();
  const server = useServerInfo();
  const { prefs } = useStore(prefsStore);
  useEffect(() => { if (wide && !section) go(href.settings("ai"), { replace: true }); }, [wide, section]);
  const meta = SECTIONS.find((s) => s[0] === section);
  useEffect(() => pageTitle(meta ? meta[1] : "Settings"), [section]);
  const summaries = {
    ai: server.config ? aiSummary(server.config.knowledge) : "",
    voice: server.config ? languageLabel(server.config.STT.language) : "",
    integrations: server.config ? (server.config.weather.enable ? "Weather set up" : "Weather not set up") : "",
    server: server.version ? versionLabel(server.version) + (server.version.avail ? " · update ready" : "") : "",
    appearance: { system: "System theme", light: "Light theme", dark: "Dark theme" }[prefs.theme],
  };
  let pane = null;
  switch (section) {
    case "ai": pane = html`<${AIPane} />`; break;
    case "voice": pane = html`<${VoicePane} />`; break;
    case "integrations": pane = html`<${IntegrationsPane} />`; break;
    case "server": pane = html`<${ServerPane} />`; break;
    case "appearance": pane = html`<${AppearancePane} />`; break;
  }
  return html`<div class="page">
    ${section && html`<div class="crumb only-mobile"><a href=${href.settings()}><${Icon} name="left" size=${16} />Settings</a></div>`}
    <div class="top"><div><h1>${meta && !wide ? meta[1] : "Settings"}</h1>
      <p class="sub">${section && !wide ? "" : "These apply to the whole server and every robot. Robot-specific settings live on each robot’s page."}</p></div></div>
    <div class="sgrid">
      <nav class="subnav" aria-label="Settings sections">${SECTIONS.map(([k, n, ic]) => html`<a key=${k} href=${href.settings(k)} aria-current=${section === k ? "page" : null}><${Icon} name=${ic} size=${20} />${n}</a>`)}</nav>
      <div>
        ${!section && html`<div class="card slist slist-wrap">${SECTIONS.map(([k, n, ic]) => html`<a class="srow" key=${k} href=${href.settings(k)}><span class="ico"><${Icon} name=${ic} size=${20} /></span><span class="t"><b>${n}</b><small>${summaries[k]}</small></span><${Icon} name="right" size=${18} /></a>`)}</div>`}
        ${pane}
      </div>
    </div>
  </div>`;
}
