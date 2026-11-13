// Run with: node --test chipper/webroot/next/test/*.mjs
import test from "node:test";
import assert from "node:assert/strict";
import { nextDelay } from "../js/backoff.js";

test("a healthy poll keeps its pace", () => {
  assert.equal(nextDelay(5000, 0), 5000);
});

test("each failure in a row doubles the wait, up to a minute", () => {
  assert.deepEqual([1, 2, 3, 4, 5, 50].map((n) => nextDelay(5000, n)), [10000, 20000, 40000, 60000, 60000, 60000]);
  assert.deepEqual([1, 2, 3, 4, 5, 6, 7, 8].map((n) => nextDelay(500, n)), [1000, 2000, 4000, 8000, 16000, 32000, 60000, 60000]);
});
