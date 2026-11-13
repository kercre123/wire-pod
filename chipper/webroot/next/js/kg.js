// What the AI pane does to the saved knowledge settings, kept apart from the screen so it can be tested.
// "Off" only flips `enable`: the server keeps the provider and every other setting, so turning AI back on restores them.
const LLM = ["openai", "custom", "together"];

export function blankFor(p, keep) {
  return { enable: p !== "", provider: p, key: "", model: "", id: "", endpoint: "", openai_prompt: "", openai_voice: p === "openai" ? "fable" : "", openai_voice_with_english: false, temp: 0, top_p: 0,
    intentgraph: keep.intentgraph, save_chat: keep.save_chat, commands_enable: keep.commands_enable, robotName: keep.robotName || "" };
}

// The screen's state after a provider tile is picked ("" is Off). `saved` is the config on the server.
export function chooseKg(next, kg, saved) {
  if (next === "") return { ...saved, enable: false };
  return next === saved.provider ? { ...saved, enable: true } : blankFor(next, kg);
}

// The body to save. The saved provider keeps every field it had, so re-enabling it never blanks anything.
export function kgPayload(kg, saved) {
  const p = kg.enable ? kg.provider : "";
  if (p === "") return { ...saved, enable: false };
  const d = { ...(p === saved.provider ? saved : blankFor(p, kg)), enable: true, provider: p, robotName: kg.robotName || "" };
  if (p === "openai") Object.assign(d, { key: kg.key, model: kg.model, openai_prompt: kg.openai_prompt, openai_voice: kg.openai_voice, openai_voice_with_english: kg.openai_voice_with_english });
  if (p === "custom") Object.assign(d, { key: kg.key, model: kg.model, openai_prompt: kg.openai_prompt, endpoint: kg.endpoint, temp: kg.temp || 0, top_p: kg.top_p || 0 });
  if (p === "together") Object.assign(d, { key: kg.key, model: kg.model, openai_prompt: kg.openai_prompt });
  if (p === "houndify") Object.assign(d, { key: kg.key, id: kg.id });
  d.intentgraph = kg.intentgraph;
  d.save_chat = LLM.includes(p) && kg.save_chat;
  d.commands_enable = LLM.includes(p) && kg.commands_enable;
  return d;
}
