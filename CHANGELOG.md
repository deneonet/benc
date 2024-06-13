## Changelog

#### See you in v1.1.0 :]

- v1.0.9 - refactors
  - removed `btag`
  - removed `bunsafe`, use `bstd.MarshalUnsafeString` and `bstd.UnmarshalUnsafeString` instead
  - removed `bpre`, [how to use buffer reuse in v1.0.9?](https://github.com/deneonet/benc/tree/main?tab=readme-ov-file#buffer-reuse)
  - added maxsize to string, byte slice and map: support now a size over `math.MaxUint16`
  - removed `bstd.MaxSizeUint16`, `bstd.MaxSizeUint32` and `bstd.MaxSizeInt64`, instead use `benc.Bytes2`, `benc.Bytes4`, `benc.Bytes8`
  - removed `bstd.SizeInt`, `bstd.MarshalInt`, `bstd.UnmarshalInt`, use `bstd.SizeInt64`, `bstd.MarshalInt64`, `bstd.UnmarshalInt64` (same for uint), because int and int64 in benc was the same, which is misleading
  - moved `bstd.Marshal`, `bstd.MarshalMF`, `bstd.UnmarshalMF`, `bstd.VerifyMarshal`, `bstd.VerifyUnmarshal` into `benc` package
  - expanded tests to a coverage of ~85%
  - made all unmarshal and skips error-prone (so no panics, view the tests)
  - better error message (from panics and returned errors)
  - [now fully compatible with custom marshal and unmarshal functions, even with data type validation](https://github.com/deneonet/benc/tree/main?tab=readme-ov-file#custom-marshal-and-unmarshal-1)

- v1.0.8 - pull request

- v1.0.7 - enhancement
 - adding the option to set the maximum size of a slice in sizing/encoding/decoding

- v1.0.6 - bug fix
  - bug fix in the byte slice unmarshal

- v1.0.5 - bug fixes + improvements
  - fixed btag message framing not working
  - faster unsafe string conversion
    - replaced deprecated reflect.StringHeader and reflect.SliceHeader
    - fixed unsafe rule violation

- v1.0.4 - bug fixes + new features + improvements
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

- v1.0.3 - new features
  - added custom tags to a marshal, `btag.SMarshal(s)` for string tag, `btag.UMarshal(s)` for uint16 tag (more
    performant),
      append a MF at the end of these functions to get the message framing marshal
    - added pre-allocation for message framing
    - function inline (done)
  - `bstd.MFUnmarshal(s)` to `bstd.UnmarshalMF(s)`, `bstd.MFFinish()` to `bstd.FinishMF(s)`, etc.

- v1.0.2 - new features + improvements
  - added zero memory allocation string to byte slice (and back) conversion, `bunsafe.MarshalString()`
    and `bunsafe.UnmarshalString()`
  - removed that all Size functions require 1 argument: T (expect string), e.g. `bstd.SizeUInt16()`,
    before: `bstd.SizeUInt16(0)`
  - added pre-allocation (message framing not done yet, v1.0.3 fix), `bpre.Marshal(maxSize)`
    - function inline (not done yet), e.g. inlining binary calls

- v1.0.1 - new features + improvements
  - benc -> bstd, e.g. `benc.Marshal(s)` to `bstd.Marshal(s)`
  - all Size function requires 1 argument T (going to be removed again in v1.0.2), e.g. `bstd.SizeUInt16(0)`,
    before: `bstd.SizeUInt16()`
  - added Time, Byte, Faster String encoding, Faster byte slice encoding, Maps and Slices, UInt16, UInt32 and Int16, as well as Float32
  - added [best practices](BESTPRACTICES.md)
