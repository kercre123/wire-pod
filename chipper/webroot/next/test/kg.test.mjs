// Run with: node --test chipper/webroot/next/test
import test from "node:test";
import assert from "node:assert/strict";
import { chooseKg, kgPayload } from "../js/kg.js";

const custom = { enable: true, provider: "custom", key: "lm-studio", id: "", model: "qwen/qwen3.6-35b-a3b", intentgraph: true, robotName: "Vector", openai_prompt: "Be brief.", openai_voice: "", openai_voice_with_english: false, save_chat: true, commands_enable: true, endpoint: "http://localhost:1234/v1", top_p: 0.9, temp: 0.7 };
const openai = { ...custom, provider: "openai", key: "sk-x", model: "gpt-4.1", endpoint: "", openai_voice: "fable", temp: 0, top_p: 0 };

// the screen: pick a tile, then save; the server stores the body as sent
const save = (kg, saved) => kgPayload(kg, saved);

for (const saved of [custom, openai]) {
  test(`${saved.provider}: Off then save only sets enable:false`, () => {
    const body = save(chooseKg("", saved, saved), saved);
    assert.deepEqual(body, { ...saved, enable: false });
  });

  test(`${saved.provider}: Off, save, then back on restores every setting`, () => {
    const off = save(chooseKg("", saved, saved), saved); // what the server now holds
    const kg = chooseKg(saved.provider, off, off);
    assert.deepEqual(kg, saved); // pre-filled
    assert.deepEqual(save(kg, off), saved); // and saved unchanged
  });
}

test("Off after switching tiles in the same session still keeps what was saved", () => {
  const kg = chooseKg("openai", custom, custom); // blank OpenAI form, never saved
  assert.equal(kg.key, "");
  assert.deepEqual(save(chooseKg("", kg, custom), custom), { ...custom, enable: false });
});

test("switching to a different provider starts blank but keeps the shared flags", () => {
  const body = save(chooseKg("openai", custom, custom), custom);
  assert.equal(body.provider, "openai");
  assert.equal(body.key, "");
  assert.equal(body.intentgraph, true);
  assert.equal(body.commands_enable, true);
});
