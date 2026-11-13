// Run with: node --test chipper/webroot/next/test
import test from "node:test";
import assert from "node:assert/strict";
import { QUICK, runQuick } from "../js/quick.js";

const q = (label) => QUICK.find((x) => x.label === label);
const setup = (controlled, call) => {
  const log = [];
  return { log, deps: { esn: "e1", name: "Lulu", controlled, release: async () => log.push("release"), toast: (m, o) => log.push(["toast", m, o && o.kind]), call: async (...a) => { log.push(["call", ...a]); return call ? call(...a) : "done"; }, pause: async () => log.push("pause") } };
};

test("controls off: intent is sent, nothing is released", async () => {
  const { log, deps } = setup(false);
  await runQuick(q("Dance"), deps);
  assert.deepEqual(log, [["call", "cloud_intent", "e1", { intent: "intent_imperative_dance" }], ["toast", "Dance: sent to Lulu", undefined]]);
});

test("controls on: release first, tell the user, then send the intent", async () => {
  const { log, deps } = setup(true);
  await runQuick(q("Fetch cube"), deps);
  assert.deepEqual(log, ["release", ["toast", "Released controls so Lulu can fetch the cube", undefined], "pause", ["call", "cloud_intent", "e1", { intent: "intent_imperative_fetchcube" }], ["toast", "Fetch cube: sent to Lulu", undefined]]);
});

test("go to charger with controls on uses drive_on_charger and keeps control", async () => {
  const { log, deps } = setup(true, () => '{"result":"BEHAVIOR_COMPLETE_STATE"}');
  await runQuick(q("Go to charger"), deps);
  assert.deepEqual(log, [["call", "drive_on_charger", "e1"], ["toast", "Go to charger: sent to Lulu", undefined]]);
});

test("go to charger with controls off keeps the intent", async () => {
  const { log, deps } = setup(false);
  await runQuick(q("Go to charger"), deps);
  assert.equal(log[0][1], "cloud_intent");
});

test("charger that won't start says so", async () => {
  const { log, deps } = setup(true, () => '{"result":"BEHAVIOR_WONT_ACTIVATE_STATE"}');
  await runQuick(q("Go to charger"), deps);
  assert.equal(log[1][2], "bad");
});

test("a server error becomes a friendly toast, not silence", async () => {
  const { log, deps } = setup(false, () => { throw Object.assign(new Error("error: rpc error: busy"), { name: "ApiError", kind: "http", detail: "error: rpc error: busy" }); });
  await runQuick(q("Sleep"), deps);
  const t = log.at(-1);
  assert.equal(t[2], "bad");
  assert.match(t[1], /^Lulu didn’t take “Sleep”\./);
});
