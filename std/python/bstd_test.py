import unittest
import random
import math
import struct
from datetime import datetime, timezone
from typing import List, Any, Callable, Dict

# Import the bstd module itself
import bstd

# --- Test Helper Functions ---

def deep_equal(a, b):
    """A simple deep equality check for lists and dicts used in tests."""
    if isinstance(a, list) and isinstance(b, list):
        if len(a) != len(b):
            return False
        return all(deep_equal(x, y) for x, y in zip(a, b))
    if isinstance(a, dict) and isinstance(b, dict):
        # In Python, dict order isn't guaranteed for equality, so check items
        if len(a) != len(b):
            return False
        for k, v in a.items():
            if k not in b or not deep_equal(v, b[k]):
                return False
        return True
    if isinstance(a, float) and isinstance(b, float):
        return abs(a - b) < 1e-9 # Handle float precision
    return a == b


class TestBstd(unittest.TestCase):

    def test_data_types(self):
        """Tests marshalling, skipping, and unmarshalling of all basic data types."""
        test_str = "Hello World!"
        test_bs = b'\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a'
        neg_int_val = -12345
        pos_int_val = (2**63) - 1
        uint_max_val = (2**64) - 1
        
        # To accurately test float32, we must account for the precision loss
        # that occurs when a Python float (64-bit) is marshalled to 32 bits.
        # We simulate this conversion on the original value before comparison.
        original_float32 = random.random()
        test_float32 = struct.unpack('<f', struct.pack('<f', original_float32))[0]
        
        values = [
            True, 128, test_float32, random.random(),
            pos_int_val, neg_int_val, -1, -1,
            random.randint(0, 2**31 - 1), random.randint(0, 2**63 - 1),
            uint_max_val, 160, random.randint(0, 2**32 - 1),
            random.randint(0, 2**64 - 1), test_str, test_bs, test_bs,
        ]

        sizers = [
            lambda v=values[0]: bstd.size_bool(v), lambda v=values[1]: bstd.size_byte(v),
            lambda v=values[2]: bstd.size_float32(v), lambda v=values[3]: bstd.size_float64(v),
            lambda v=values[4]: bstd.size_int(v), lambda v=values[5]: bstd.size_int(v),
            lambda v=values[6]: bstd.size_int8(v), lambda v=values[7]: bstd.size_int16(v),
            lambda v=values[8]: bstd.size_int32(v), lambda v=values[9]: bstd.size_int64(v),
            lambda v=values[10]: bstd.size_uint(v), lambda v=values[11]: bstd.size_uint16(v),
            lambda v=values[12]: bstd.size_uint32(v), lambda v=values[13]: bstd.size_uint64(v),
            lambda v=values[14]: bstd.size_string(v), lambda v=values[15]: bstd.size_bytes(v),
            lambda v=values[16]: bstd.size_bytes(v),
        ]
        
        total_size = sum(s() for s in sizers)
        b = bytearray(total_size)
        
        marshallers = [
            bstd.marshal_bool, bstd.marshal_byte, bstd.marshal_float32, bstd.marshal_float64,
            bstd.marshal_int, bstd.marshal_int, bstd.marshal_int8, bstd.marshal_int16,
            bstd.marshal_int32, bstd.marshal_int64, bstd.marshal_uint, bstd.marshal_uint16,
            bstd.marshal_uint32, bstd.marshal_uint64, bstd.marshal_string, bstd.marshal_bytes, bstd.marshal_bytes
        ]

        n = 0
        for i, marshaler in enumerate(marshallers):
            n = marshaler(n, b, values[i])
        self.assertEqual(n, total_size, "Marshal size mismatch")
        
        skippers = [
            bstd.skip_bool, bstd.skip_byte, bstd.skip_float32, bstd.skip_float64, bstd.skip_varint,
            bstd.skip_varint, bstd.skip_int8, bstd.skip_int16, bstd.skip_int32, bstd.skip_int64,
            bstd.skip_varint, bstd.skip_uint16, bstd.skip_uint32, bstd.skip_uint64, bstd.skip_string,
            bstd.skip_bytes, bstd.skip_bytes
        ]
        
        n = 0
        for skipper in skippers:
            n = skipper(n, b)
        self.assertEqual(n, total_size, "Skip size mismatch")
            
        unmarshallers = [
            bstd.unmarshal_bool, bstd.unmarshal_byte, bstd.unmarshal_float32, bstd.unmarshal_float64,
            bstd.unmarshal_int, bstd.unmarshal_int, bstd.unmarshal_int8, bstd.unmarshal_int16,
            bstd.unmarshal_int32, bstd.unmarshal_int64, bstd.unmarshal_uint, bstd.unmarshal_uint16,
            bstd.unmarshal_uint32, bstd.unmarshal_uint64, bstd.unmarshal_string, bstd.unmarshal_bytes, bstd.unmarshal_bytes
        ]
        
        n = 0
        for i, unmarshaller in enumerate(unmarshallers):
            n, v = unmarshaller(n, b)
            self.assertTrue(deep_equal(v, values[i]), f"Unmarshal at index {i} mismatch")
        self.assertEqual(n, total_size, "Unmarshal size mismatch")

    def test_err_buf_too_small(self):
        """Tests general BUF_TOO_SMALL errors."""
        buffers = [
            b'', b'\x01\x02\x03', b'\x01\x02\x03\x04\x05\x06', b'\x80', b'', b'\x01',
            b'\x01\x02\x03', b'\x01\x02\x03\x04\x05\x06', b'\x80', b'\x01',
            b'\x01\x02\x03', b'\x01\x02\x03\x04\x05\x06', b'\x02\x00', b'\x04\x01\x02',
            b'\x08\x01\x02', b'\x0a\x00\x00\x00', b'\x0a\x00\x00\x00'
        ]
        
        unmarshallers = [
            bstd.unmarshal_bool, bstd.unmarshal_float32, bstd.unmarshal_float64, bstd.unmarshal_int,
            bstd.unmarshal_int8, bstd.unmarshal_int16, bstd.unmarshal_int32, bstd.unmarshal_int64,
            bstd.unmarshal_uint, bstd.unmarshal_uint16, bstd.unmarshal_uint32, bstd.unmarshal_uint64,
            bstd.unmarshal_string, bstd.unmarshal_bytes, bstd.unmarshal_bytes,
            lambda n, b: bstd.unmarshal_slice(n, b, bstd.unmarshal_byte),
            lambda n, b: bstd.unmarshal_map(n, b, bstd.unmarshal_byte, bstd.unmarshal_byte)
        ]
        
        for i, unmarshaller in enumerate(unmarshallers):
            with self.assertRaises(bstd.BencException) as cm:
                unmarshaller(0, buffers[i])
            self.assertEqual(cm.exception.error_type, bstd.BencError.BUF_TOO_SMALL, f"Unmarshal #{i}")

    def test_err_buf_too_small_specific(self):
        """Tests specific cases where length is valid but data is truncated."""
        buffers = [b'\x02\x41', b'\x0a\x00\x00\x00\x01']
        
        with self.assertRaises(bstd.BencException) as cm:
            bstd.unmarshal_string(0, buffers[0])
        self.assertEqual(cm.exception.error_type, bstd.BencError.BUF_TOO_SMALL)

        with self.assertRaises(bstd.BencException) as cm:
            bstd.unmarshal_slice(0, buffers[1], bstd.unmarshal_byte)
        self.assertEqual(cm.exception.error_type, bstd.BencError.BUF_TOO_SMALL)

    def test_slices(self):
        slice_data = ["a", "b", "c"]
        s = bstd.size_slice(slice_data, bstd.size_string)
        buf = bytearray(s)
        bstd.marshal_slice(0, buf, slice_data, bstd.marshal_string)
        
        n = bstd.skip_slice(0, buf, bstd.skip_string)
        self.assertEqual(n, s)

        _, ret_slice = bstd.unmarshal_slice(0, buf, bstd.unmarshal_string)
        self.assertEqual(ret_slice, slice_data)

    def test_maps(self):
        m_str = {"k1": "v1", "k2": "v2"}
        s = bstd.size_map(m_str, bstd.size_string, bstd.size_string)
        buf = bytearray(s)
        bstd.marshal_map(0, buf, m_str, bstd.marshal_string, bstd.marshal_string)

        n = bstd.skip_map(0, buf, bstd.skip_string, bstd.skip_string)
        self.assertEqual(n, s)
        
        _, ret_map = bstd.unmarshal_map(0, buf, bstd.unmarshal_string, bstd.unmarshal_string)
        self.assertTrue(deep_equal(ret_map, m_str))
        
    def test_string_edge_cases(self):
        s = bstd.size_string("")
        buf = bytearray(s)
        bstd.marshal_string(0, buf, "")
        n = bstd.skip_string(0, buf)
        self.assertEqual(n, s)
        _, ret_str = bstd.unmarshal_string(0, buf)
        self.assertEqual(ret_str, "")
        
        long_str = "H" * (2**16)
        s = bstd.size_string(long_str)
        buf = bytearray(s)
        bstd.marshal_string(0, buf, long_str)
        n = bstd.skip_string(0, buf)
        self.assertEqual(n, s)
        _, ret_long_str = bstd.unmarshal_string(0, buf)
        self.assertEqual(ret_long_str, long_str)

    def test_varint_overflow(self):
        overflow_buf = b'\x80\x80\x80\x80\x80\x80\x80\x80\x80\x80\x01'
        overflow_edge_case = b'\x80\x80\x80\x80\x80\x80\x80\x80\x80\x02'

        for buf in [overflow_buf, overflow_edge_case]:
            with self.assertRaisesRegex(bstd.BencException, bstd.BencError.OVERFLOW.value):
                bstd.skip_varint(0, buf)
            with self.assertRaisesRegex(bstd.BencException, bstd.BencError.OVERFLOW.value):
                bstd.unmarshal_int(0, buf)
            with self.assertRaisesRegex(bstd.BencException, bstd.BencError.OVERFLOW.value):
                bstd.unmarshal_uint(0, buf)

    def test_detailed_varint_cases(self):
        n, _ = bstd.unmarshal_uint(0, b'\x05')
        self.assertEqual(bstd.skip_varint(0, b'\x05'), n)
        n, _ = bstd.unmarshal_uint(0, b'\x80\x01')
        self.assertEqual(bstd.skip_varint(0, b'\x80\x01'), n)

        _, val = bstd.unmarshal_int(0, b'\x02')
        self.assertEqual(val, 1)
        _, val = bstd.unmarshal_int(0, b'\x03')
        self.assertEqual(val, -2)
        _, val = bstd.unmarshal_int(0, b'\xac\x02')
        self.assertEqual(val, 150)

        _, val = bstd.unmarshal_uint(0, b'\x07')
        self.assertEqual(val, 7)
        _, val = bstd.unmarshal_uint(0, b'\xac\x02')
        self.assertEqual(val, 300)

    def test_time(self):
        t = datetime.fromtimestamp(1663362895.123456789, tz=timezone.utc)
        s = bstd.size_time()
        buf = bytearray(s)
        bstd.marshal_time(0, buf, t)
        
        n = bstd.skip_time(0, buf)
        self.assertEqual(n, s)
        _, ret_time = bstd.unmarshal_time(0, buf)
        self.assertEqual(ret_time, t)

    def test_optional(self):
        val = "hello"
        s = bstd.size_optional(val, bstd.size_string)
        buf = bytearray(s)
        bstd.marshal_optional(0, buf, val, bstd.marshal_string)
        n = bstd.skip_optional(0, buf, bstd.skip_string)
        self.assertEqual(n, s)
        _, ret_ptr = bstd.unmarshal_optional(0, buf, bstd.unmarshal_string)
        self.assertEqual(ret_ptr, val)

        s = bstd.size_optional(None, bstd.size_string)
        buf = bytearray(s)
        bstd.marshal_optional(0, buf, None, bstd.marshal_string)
        n = bstd.skip_optional(0, buf, bstd.skip_string)
        self.assertEqual(n, s)
        _, ret_ptr_nil = bstd.unmarshal_optional(0, buf, bstd.unmarshal_string)
        self.assertIsNone(ret_ptr_nil)

    def test_terminator_errors(self):
        slice_buf = bytearray(bstd.size_slice(["a"], bstd.size_string))
        bstd.marshal_slice(0, slice_buf, ["a"], bstd.marshal_string)
        with self.assertRaisesRegex(bstd.BencException, "buffer was too small"):
            bstd.skip_slice(0, slice_buf[:-1], bstd.skip_string)
            
        map_buf = bytearray(bstd.size_map({"a":"b"}, bstd.size_string, bstd.size_string))
        bstd.marshal_map(0, map_buf, {"a":"b"}, bstd.marshal_string, bstd.marshal_string)
        with self.assertRaisesRegex(bstd.BencException, "buffer was too small"):
            bstd.skip_map(0, map_buf[:-1], bstd.skip_string, bstd.skip_string)

    def test_size_fixed_slice(self):
        slice_data = [1, 2, 3] # int32
        elem_size = 4
        expected_size = bstd.size_uint(len(slice_data)) + len(slice_data) * elem_size + 4
        self.assertEqual(bstd.size_fixed_slice(slice_data, elem_size), expected_size)

    def test_additional_error_paths(self):
        with self.assertRaises(bstd.BencException):
            bstd.unmarshal_time(0, b'\x01\x02\x03')
        with self.assertRaises(ValueError):
             bstd.marshal_time(0, bytearray(8), datetime.now())

        with self.assertRaises(bstd.BencException):
            bstd.skip_optional(0, b'', bstd.skip_byte)

        with self.assertRaises(bstd.BencException):
            bstd.unmarshal_optional(0, b'', bstd.unmarshal_int)
        with self.assertRaises(bstd.BencException):
            bstd.unmarshal_optional(0, b'\x01\x80', bstd.unmarshal_int)

if __name__ == '__main__':
    unittest.main()

