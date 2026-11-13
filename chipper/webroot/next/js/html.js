// htm bound to Preact, plus the hooks every view uses.
import { h, render, Fragment } from "preact";
import htm from "htm";
export const html = htm.bind(h);
export { h, render, Fragment };
export { useState, useEffect, useRef, useCallback, useMemo, useLayoutEffect, useReducer } from "preact/hooks";
