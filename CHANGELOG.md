## Changelog

#### See you in v1.0.5 :]

- v1.0.3 to v1.0.4 - mid update
  - fixed message framing bugs
  - added to be able to skip a data type in the unmarshal process, e.g. `bstd.SkipString(...)`
    - added out-of-order deserialization
  - added `bstd.VerifyUnmarshalMF(...)` to verify a message framing unmarshal
    - removed `bstd.FinishMF(...)`
  - data type validation support, e.g. prefixes a encoded string with e.g. `1` to indicate that is a string and has to
    be decoded with `bstd.UnmarshalString(...)`
    - accessible using `bmd`, e.g. `bmd.MarshalString(...)`
    - to add the data type metadata to `bunsafe` strings or `btag` just append `MD` at the end of the function,
      e.g `btag.SMarshalMD(...)`
  - fixed byte slice bugs
  - replaced `bstd.UnmarshalStringTag(...)` with `btag.SUnmarshal(...)`
  - replaced `bstd.UnmarshalUIntTag(...)` with `btag.UUnmarshal(...)`
- v1.0.2 to v1.0.3 - small update
  - added custom tags to a marshal, `btag.SMarshal(s)` for string tag, `btag.UMarshal(s)` for uint16 tag (more
    performant),
      append a MF at the end of these functions to get the message framing marshal
    - added pre-allocation for message framing
    - function inline (done)
  - `bstd.MFUnmarshal(s)` to `bstd.UnmarshalMF(s)`, `bstd.MFFinish()` to `bstd.FinishMF(s)`, etc.

- v1.0.1 to v1.0.2 - small update
  - added zero memory allocation string to byte slice (and back) conversion, `bunsafe.MarshalString()`
    and `bunsafe.UnmarshalString()`
  - removed that all Size functions require 1 argument: T (expect string), e.g. `bstd.SizeUInt16()`,
    before: `bstd.SizeUInt16(0)`
  - added pre-allocation (message framing not done yet, v1.0.3 fix), `bpre.Marshal(maxSize)`
    - function inline (not done yet), e.g. inlining binary calls

- v1 to v1.0.1 - mid update
  - benc -> bstd, e.g. `benc.Marshal(s)` to `bstd.Marshal(s)`
  - all Size function requires 1 argument T (going to be removed again in v1.0.2), e.g. `bstd.SizeUInt16(0)`,
    before: `bstd.SizeUInt16()`
  - added Time, Byte, Faster String encoding, Faster byte slice encoding, Maps and Slices, UInt16, UInt32 and Int16, as well as Float32
  - added [best practices](BESTPRACTICES.md)