"""Independent wire-v1 vectors shared with the other runtimes."""
import json
from pathlib import Path
import unittest

import bstd as b

VECTORS = json.loads((Path(__file__).resolve().parents[2] / 'testdata/conformance/wire-v1.json').read_text())


def codec(name):
    cases = {
        'uint_zero': (0, b.size_uint, b.marshal_uint, b.unmarshal_uint, b.skip_varint),
        'uint_127': (127, b.size_uint, b.marshal_uint, b.unmarshal_uint, b.skip_varint),
        'uint_128': (128, b.size_uint, b.marshal_uint, b.unmarshal_uint, b.skip_varint),
        'uint32_max_varint': (4294967295, b.size_uint, b.marshal_uint, b.unmarshal_uint, b.skip_varint),
        'uint64_max_varint_64bit': ((1 << 64) - 1, b.size_uint, b.marshal_uint, b.unmarshal_uint, b.skip_varint),
        'uint64_max_fixed': ((1 << 64) - 1, b.size_uint64, b.marshal_uint64, b.unmarshal_uint64, b.skip_uint64),
        'int_minus_one': (-1, b.size_int, b.marshal_int, b.unmarshal_int, b.skip_varint),
        'int64_min_varint_64bit': (-(1 << 63), b.size_int, b.marshal_int, b.unmarshal_int, b.skip_varint),
        'int64_max_varint_64bit': ((1 << 63) - 1, b.size_int, b.marshal_int, b.unmarshal_int, b.skip_varint),
        'int64_min_fixed': (-(1 << 63), b.size_int64, b.marshal_int64, b.unmarshal_int64, b.skip_int64),
        'int64_max_fixed': ((1 << 63) - 1, b.size_int64, b.marshal_int64, b.unmarshal_int64, b.skip_int64),
        'string_nul': ('A\0', b.size_string, b.marshal_string, b.unmarshal_string, b.skip_string),
        'string_utf8': ('é', b.size_string, b.marshal_string, b.unmarshal_string, b.skip_string),
        'bytes': (b'\0\xff', b.size_bytes, b.marshal_bytes, b.unmarshal_bytes, b.skip_bytes),
        'array': ([1, -2], lambda a: 4 * len(a), lambda n, out, a: _array_write(n, out, a),
                  lambda n, data: _array_read(n, data), lambda n, data: b.skip_int32(b.skip_int32(n, data), data)),
        'slice': ([1, -2], lambda a: b.size_slice(a, b.size_int32),
                  lambda n, out, a: b.marshal_slice(n, out, a, b.marshal_int32),
                  lambda n, data: b.unmarshal_slice(n, data, b.unmarshal_int32),
                  lambda n, data: b.skip_slice(n, data, b.skip_int32)),
        'empty_slice': ([], lambda a: b.size_slice(a, b.size_int32),
                        lambda n, out, a: b.marshal_slice(n, out, a, b.marshal_int32),
                        lambda n, data: b.unmarshal_slice(n, data, b.unmarshal_int32),
                        lambda n, data: b.skip_slice(n, data, b.skip_int32)),
        'map_single': ({'a': 2}, lambda m: b.size_map(m, b.size_string, b.size_int32),
                       lambda n, out, m: b.marshal_map(n, out, m, b.marshal_string, b.marshal_int32),
                       lambda n, data: b.unmarshal_map(n, data, b.unmarshal_string, b.unmarshal_int32),
                       lambda n, data: b.skip_map(n, data, b.skip_string, b.skip_int32)),
        'nested': ((7, True), lambda v: 5,
                   lambda n, out, v: b.marshal_bool(b.marshal_int32(n, out, v[0]), out, v[1]),
                   lambda n, data: _nested_read(n, data),
                   lambda n, data: b.skip_bool(b.skip_int32(n, data), data)),
        'nil_pointer': (None, lambda v: b.size_optional(v, b.size_int32),
                        lambda n, out, v: b.marshal_optional(n, out, v, b.marshal_int32),
                        lambda n, data: b.unmarshal_optional(n, data, b.unmarshal_int32),
                        lambda n, data: b.skip_optional(n, data, b.skip_int32)),
        'present_pointer': (7, lambda v: b.size_optional(v, b.size_int32),
                            lambda n, out, v: b.marshal_optional(n, out, v, b.marshal_int32),
                            lambda n, data: b.unmarshal_optional(n, data, b.unmarshal_int32),
                            lambda n, data: b.skip_optional(n, data, b.skip_int32)),
        'unix_nanosecond': (1, b.size_time, b.marshal_time_nanos, b.unmarshal_time_nanos, b.skip_time),
        'negative_nanosecond': (-1, b.size_time, b.marshal_time_nanos, b.unmarshal_time_nanos, b.skip_time),
    }
    return cases[name]


def _array_write(n, out, values):
    for value in values: n = b.marshal_int32(n, out, value)
    return n


def _array_read(n, data):
    result = []
    for _ in range(2):
        n, item = b.unmarshal_int32(n, data)
        result.append(item)
    return n, result


def _nested_read(n, data):
    n, value = b.unmarshal_int32(n, data)
    n, flag = b.unmarshal_bool(n, data)
    return n, (value, flag)


class Conformance(unittest.TestCase):
    def test_valid(self):
        for item in VECTORS['valid']:
            with self.subTest(item=item['name']):
                value, size, write, read, skip = codec(item['name'])
                expected = bytes.fromhex(item['hex'])
                out = bytearray(size(value))
                self.assertEqual(write(0, out, value), len(expected))
                self.assertEqual(out, expected)
                self.assertEqual(skip(0, expected), len(expected))
                self.assertEqual(read(0, expected), (len(expected), value))

    def test_invalid(self):
        cases = {
            'missing_slice_terminator': lambda data: b.unmarshal_slice(0, data, b.unmarshal_byte),
            'wrong_slice_terminator': lambda data: b.skip_slice(0, data, b.skip_byte),
            'overflowed_varint': lambda data: b.unmarshal_uint(0, data),
            'truncated_string': lambda data: b.unmarshal_string(0, data),
            'truncated_map': lambda data: b.unmarshal_map(0, data, b.unmarshal_byte, b.unmarshal_byte),
            'invalid_bool': lambda data: b.unmarshal_bool(0, data),
            'invalid_utf8': lambda data: b.skip_string(0, data),
            'wrong_map_terminator': lambda data: b.skip_map(0, data, b.skip_byte, b.skip_byte),
            'duplicate_map_key': lambda data: b.unmarshal_map(0, data, b.unmarshal_byte, b.unmarshal_byte),
            'huge_string': lambda data: b.unmarshal_string(0, data),
            'huge_slice': lambda data: b.unmarshal_slice(0, data, b.unmarshal_byte),
        }
        for item in VECTORS['invalid']:
            with self.subTest(item=item['name']):
                with self.assertRaises(b.BencException) as error:
                    cases[item['name']](bytes.fromhex(item['hex']))
                category = item['error']
                expected = (b.BencError.BUF_TOO_SMALL if category == 'truncated' else
                            b.BencError.OVERFLOW if category in ('overflow', 'limit') else
                            b.BencError.INVALID_DATA)
                self.assertEqual(error.exception.error_type, expected)
