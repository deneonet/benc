const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const vectors = JSON.parse(fs.readFileSync(path.join(__dirname, '../../testdata/conformance/wire-v1.json')));
const runtimes = {
  'javascript/typescript': require('./std.js'),
};

function cases(b) {
  const scalar = (value, size, marshal, unmarshal, skip) => ({ value, size: () => size(value), marshal: (n, buf) => marshal(n, buf, value), unmarshal, skip });
  const uint = value => scalar(BigInt(value), b.sizeUint, b.marshalUint, b.unmarshalUint, b.skipVarint);
  const int = value => scalar(BigInt(value), b.sizeInt, b.marshalInt, b.unmarshalInt, b.skipVarint);
  const fixed = (value, size, marshal, unmarshal, skip) => scalar(value, size, marshal, unmarshal, skip);
  const arr = [1, -2], map = new Map([['a', 2]]);
  const slice = value => scalar(value, v => b.sizeSlice(v, b.sizeInt32),
    (n, buf, v) => b.marshalSlice(n, buf, v, b.marshalInt32),
    (n, buf) => b.unmarshalSlice(n, buf, b.unmarshalInt32),
    (n, buf) => b.skipSlice(n, buf, b.skipInt32));
  const optional = value => scalar(value, v => b.sizePointer(v, b.sizeInt32),
    (n, buf, v) => b.marshalPointer(n, buf, v, b.marshalInt32),
    (n, buf) => b.unmarshalPointer(n, buf, b.unmarshalInt32),
    (n, buf) => b.skipPointer(n, buf, b.skipInt32));
  return {
    uint_zero: uint(0), uint_127: uint(127), uint_128: uint(128),
    uint32_max_varint: uint(4294967295), uint64_max_varint_64bit: uint((1n << 64n) - 1n),
    uint64_max_fixed: fixed((1n << 64n) - 1n, b.sizeUint64, b.marshalUint64, b.unmarshalUint64, b.skipUint64),
    int_minus_one: int(-1), int64_min_varint_64bit: int(-(1n << 63n)),
    int64_max_varint_64bit: int((1n << 63n) - 1n),
    int64_min_fixed: fixed(-(1n << 63n), b.sizeInt64, b.marshalInt64, b.unmarshalInt64, b.skipInt64),
    int64_max_fixed: fixed((1n << 63n) - 1n, b.sizeInt64, b.marshalInt64, b.unmarshalInt64, b.skipInt64),
    string_nul: scalar('A\0', b.sizeString, b.marshalString, b.unmarshalString, b.skipString),
    string_utf8: scalar('é', b.sizeString, b.marshalString, b.unmarshalString, b.skipString),
    bytes: scalar(new Uint8Array([0, 255]), b.sizeBytes, b.marshalBytes, b.unmarshalBytesCopied, b.skipBytes),
    array: { value: arr, size: () => 8,
      marshal: (n, buf) => arr.reduce((at, v) => b.marshalInt32(at, buf, v), n),
      unmarshal: (n, buf) => { const [at, first] = b.unmarshalInt32(n, buf); const [end, second] = b.unmarshalInt32(at, buf); return [end, [first, second]]; },
      skip: (n, buf) => b.skipInt32(b.skipInt32(n, buf), buf) },
    slice: slice(arr), empty_slice: slice([]),
    map_single: scalar(map, v => b.sizeMap(v, b.sizeString, b.sizeInt32),
      (n, buf, v) => b.marshalMap(n, buf, v, b.marshalString, b.marshalInt32),
      (n, buf) => b.unmarshalMap(n, buf, b.unmarshalString, b.unmarshalInt32),
      (n, buf) => b.skipMap(n, buf, b.skipString, b.skipInt32)),
    nested: { value: [7, true], size: () => 5,
      marshal: (n, buf) => b.marshalBool(b.marshalInt32(n, buf, 7), buf, true),
      unmarshal: (n, buf) => { const [at, x] = b.unmarshalInt32(n, buf); const [end, flag] = b.unmarshalBool(at, buf); return [end, [x, flag]]; },
      skip: (n, buf) => b.skipBool(b.skipInt32(n, buf), buf) },
    nil_pointer: optional(null), present_pointer: optional(7),
    unix_nanosecond: scalar(new b.Time(1n), b.sizeTime, b.marshalTime, b.unmarshalTime, b.skipTime),
    negative_nanosecond: scalar(new b.Time(-1n), b.sizeTime, b.marshalTime, b.unmarshalTime, b.skipTime),
  };
}

for (const [language, b] of Object.entries(runtimes)) {
  const valid = cases(b);
  for (const vector of vectors.valid) {
    test(`${language}: ${vector.name}`, () => {
      const c = valid[vector.name];
      assert.ok(c, 'missing codec');
      const expected = Buffer.from(vector.hex, 'hex');
      const output = new Uint8Array(c.size());
      assert.equal(c.marshal(0, output), expected.length);
      assert.deepEqual(Buffer.from(output), expected);
      assert.equal(c.skip(0, expected), expected.length);
      const [end, value] = c.unmarshal(0, expected);
      assert.equal(end, expected.length);
      if (vector.type === 'time.Time') assert.equal(value.unixNano, c.value.unixNano);
      else assert.deepEqual(value, c.value);
    });
  }
  const invalid = {
    missing_slice_terminator: buf => b.unmarshalSlice(0, buf, b.unmarshalByte),
    wrong_slice_terminator: buf => b.skipSlice(0, buf, b.skipByte),
    overflowed_varint: buf => b.unmarshalUint(0, buf),
    truncated_string: buf => b.unmarshalString(0, buf),
    truncated_map: buf => b.unmarshalMap(0, buf, b.unmarshalByte, b.unmarshalByte),
    invalid_bool: buf => b.unmarshalBool(0, buf),
    invalid_utf8: buf => b.skipString(0, buf),
    wrong_map_terminator: buf => b.skipMap(0, buf, b.skipByte, b.skipByte),
    duplicate_map_key: buf => b.unmarshalMap(0, buf, b.unmarshalByte, b.unmarshalByte),
    huge_string: buf => b.unmarshalString(0, buf),
    huge_slice: buf => b.unmarshalSlice(0, buf, b.unmarshalByte),
  };
  for (const vector of vectors.invalid) {
    test(`${language}: ${vector.name}`, () => {
      assert.ok(invalid[vector.name], 'missing negative codec');
      const expected = vector.error === 'truncated' ? b.ErrBufTooSmall :
        ['overflow', 'limit'].includes(vector.error) ? b.ErrOverflow : b.ErrInvalidData;
      assert.throws(() => invalid[vector.name](Buffer.from(vector.hex, 'hex')),
        error => error === expected);
    });
  }
  test(`${language}: numeric bounds and buffer preflight`, () => {
    assert.throws(() => b.sizeUint(-1), b.ErrOverflow);
    assert.throws(() => b.sizeUint(Number.MAX_SAFE_INTEGER + 1), b.ErrOverflow);
    assert.throws(() => b.sizeInt(1n << 63n), b.ErrOverflow);
    assert.throws(() => b.marshalUint(0, new Uint8Array(0), 128n), b.ErrBufTooSmall);
    assert.throws(() => b.unmarshalByte(-1, Uint8Array.of(7)), b.ErrBufTooSmall);
  });
}
