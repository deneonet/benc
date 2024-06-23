# benc dev

I have to sometimes change PCs, so I created this repo, and made it public, if someone is interested in the development of benc.

## Disclaimers:
- .md files (like README) may contain false information, unfinished statements or invalid URLs.
- Code is unfinished and could be broken, some parts may be just for testing. Code will change.

## Changes since v1.0.9:
- Made a code generator (see cmd/bencgen)
- Implementation for bencgen (see impl/gen), for `forward and backward compatibility`
- Implemented varint marshalling (much faster and better than the Max size system Benc currently uses)
- Mod Path changed (github.com/deneonet/benc) to go.kine.bz/benc (cleaner imports, I own the domain, so its unique)
- Strucure changes, probably going to remove Message Framing too (already done)

## Todos:
- Fix bcd
- Implement maps to bencgen
- Finish byte slice
- mdv rework
- Add bencgen tests
- Add more tests (testing)
- Write README for root, bstd and mdv