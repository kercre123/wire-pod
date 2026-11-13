# Vendored libraries

Plain ES modules, committed as-is. The page maps the bare names `preact`, `preact/hooks` and
`htm` to these files with an import map (see `../index.html`). There is no build step and nothing
needs Node at serve time.

| File | Package | Version | License |
| --- | --- | --- | --- |
| `preact.module.js` | `preact` `dist/preact.module.js` | 10.29.8 | MIT (`LICENSE-preact.txt`) |
| `hooks.module.js` | `preact` `hooks/dist/hooks.module.js` | 10.29.8 | MIT (`LICENSE-preact.txt`) |
| `htm.module.js` | `htm` `dist/htm.module.js` | 3.1.1 | Apache-2.0 (`LICENSE-htm.txt`) |

To update: `npm pack preact@<version> htm@<version>`, copy the three files above, and change the
versions here. The only edit made to the upstream files is removing the trailing
`sourceMappingURL` comment, because the maps are not shipped.

Fonts in `../fonts` (Bricolage Grotesque, Figtree, JetBrains Mono; latin subset, variable) are
licensed under the SIL Open Font License 1.1; the license texts sit next to them.
