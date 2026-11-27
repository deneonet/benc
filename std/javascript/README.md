# Benc for JavaScript and TypeScript

This is the single JS/TS runtime package. Its `std.js` implementation and
`std.d.ts` declarations share one release. Install from this repository:

```sh
npm install ./std/javascript
```

Use `require('@banditmoscow1337/benc')` or a TypeScript import from the same
package. Generated `*_benc.js` files expect `./std.js` next to them; copy
`node_modules/@banditmoscow1337/benc/std.js` into the generated output
directory. Generated TypeScript declarations use the adjacent `std.d.ts`.
The code generator parses source declarations into Go AST and emits the
requested target; this package is only the wire runtime.
