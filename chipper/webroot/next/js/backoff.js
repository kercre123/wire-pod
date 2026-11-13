// How long to wait before the next poll after `fails` in a row failed: doubles each time, up to `max` (one minute).
export const nextDelay = (every, fails, max = 60000) => Math.min(max, every * 2 ** Math.min(fails, 10));
