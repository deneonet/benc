/**
 * bstd.test.js - A Jest test suite for the bstd.js library.
 * This file verifies that the `bstd.js` library is a feature-complete and correct
 * port of the original Go implementation.
 *
 * To run this test, you'll need Node.js and Jest.
 * 1. Install Jest: `npm install --save-dev jest`
 * 2. Run tests: `npx jest`
 */

const bstd = require('./std.js');
const {
    // Errors
    BencError, ErrOverflow, ErrBufTooSmall,
    // Varint
    sizeUint, marshalUint, unmarshalUint, skipVarint,
    sizeInt, marshalInt, unmarshalInt,
    // String
    sizeString, marshalString, unmarshalString, skipString,
    marshalUnsafeString, unmarshalUnsafeString,
    // Byte(s)
    skipByte, sizeByte, marshalByte, unmarshalByte,
    skipBytes, sizeBytes, marshalBytes, unmarshalBytesCopied, unmarshalBytesCropped,
    // Slice
    skipSlice, sizeSlice, sizeFixedSlice, marshalSlice, unmarshalSlice,
    // Map
    skipMap, sizeMap, marshalMap, unmarshalMap,
    // Fixed-size
    skipInt8, sizeInt8, marshalInt8, unmarshalInt8,
    skipInt16, sizeInt16, marshalInt16, unmarshalInt16,
    skipInt32, sizeInt32, marshalInt32, unmarshalInt32,
    skipInt64, sizeInt64, marshalInt64, unmarshalInt64,
    skipUint16, sizeUint16, marshalUint16, unmarshalUint16,
    skipUint32, sizeUint32, marshalUint32, unmarshalUint32,
    skipUint64, sizeUint64, marshalUint64, unmarshalUint64,
    skipFloat32, sizeFloat32, marshalFloat32, unmarshalFloat32,
    skipFloat64, sizeFloat64, marshalFloat64, unmarshalFloat64,
    // Bool
    skipBool, sizeBool, marshalBool, unmarshalBool,
    // Time
    skipTime, sizeTime, marshalTime, unmarshalTime,
    // Pointer
    skipPointer, sizePointer, marshalPointer, unmarshalPointer
} = bstd;


describe('DataTypes', () => {
    test('should correctly marshal, skip, and unmarshal various data types', () => {
        const testStr = "Hello World!";
        const testBs = new Uint8Array([0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
        const negIntVal = -12345n;
        const posIntVal = 9223372036854775807n; // Max signed 64-bit int
        const maxUintVal = 18446744073709551615n; // Max unsigned 64-bit int

        const values = [
            true,
            128,
            Math.random(), // float64
            Math.fround(Math.random()), // float32
            posIntVal,
            negIntVal,
            -1,
            -1,
            2147483647,
            9223372036854775807n,
            maxUintVal,
            160,
            4294967295,
            18446744073709551615n,
            testStr,
            testStr,
            testBs,
            testBs,
        ];

        const sizers = [
            () => sizeBool(),
            () => sizeByte(),
            () => sizeFloat64(),
            () => sizeFloat32(),
            () => sizeInt(posIntVal),
            () => sizeInt(negIntVal),
            () => sizeInt8(),
            () => sizeInt16(),
            () => sizeInt32(),
            () => sizeInt64(),
            () => sizeUint(maxUintVal),
            () => sizeUint16(),
            () => sizeUint32(),
            () => sizeUint64(),
            () => sizeString(testStr),
            () => sizeString(testStr),
            () => sizeBytes(testBs),
            () => sizeBytes(testBs),
        ];
        
        const marshalers = [
            (n, b, v) => marshalBool(n, b, v),
            (n, b, v) => marshalByte(n, b, v),
            (n, b, v) => marshalFloat64(n, b, v),
            (n, b, v) => marshalFloat32(n, b, v),
            (n, b, v) => marshalInt(n, b, v),
            (n, b, v) => marshalInt(n, b, v),
            (n, b, v) => marshalInt8(n, b, v),
            (n, b, v) => marshalInt16(n, b, v),
            (n, b, v) => marshalInt32(n, b, v),
            (n, b, v) => marshalInt64(n, b, v),
            (n, b, v) => marshalUint(n, b, v),
            (n, b, v) => marshalUint16(n, b, v),
            (n, b, v) => marshalUint32(n, b, v),
            (n, b, v) => marshalUint64(n, b, v),
            (n, b, v) => marshalString(n, b, v),
            (n, b, v) => marshalUnsafeString(n, b, v),
            (n, b, v) => marshalBytes(n, b, v),
            (n, b, v) => marshalBytes(n, b, v),
        ];

        const skippers = [
            skipBool, skipByte, skipFloat64, skipFloat32, skipVarint, skipVarint,
            skipInt8, skipInt16, skipInt32, skipInt64, skipVarint, skipUint16,
            skipUint32, skipUint64, skipString, skipString, skipBytes, skipBytes
        ];
        
        const unmarshalers = [
            (n, b) => unmarshalBool(n, b),
            (n, b) => unmarshalByte(n, b),
            (n, b) => unmarshalFloat64(n, b),
            (n, b) => unmarshalFloat32(n, b),
            (n, b) => unmarshalInt(n, b),
            (n, b) => unmarshalInt(n, b),
            (n, b) => unmarshalInt8(n, b),
            (n, b) => unmarshalInt16(n, b),
            (n, b) => unmarshalInt32(n, b),
            (n, b) => unmarshalInt64(n, b),
            (n, b) => unmarshalUint(n, b),
            (n, b) => unmarshalUint16(n, b),
            (n, b) => unmarshalUint32(n, b),
            (n, b) => unmarshalUint64(n, b),
            (n, b) => unmarshalString(n, b),
            (n, b) => unmarshalUnsafeString(n, b),
            (n, b) => unmarshalBytesCropped(n, b),
            (n, b) => unmarshalBytesCopied(n, b),
        ];

        const s = sizers.reduce((acc, sizer) => acc + sizer(), 0);
        const b = new Uint8Array(s);
        let n = 0;
        for (let i = 0; i < marshalers.length; i++) {
            n = marshalers[i](n, b, values[i]);
        }
        expect(n).toBe(s);

        // Test Skip
        let skipN = 0;
        for (const skipper of skippers) {
            skipN = skipper(skipN, b);
        }
        expect(skipN).toBe(b.length);

        // Test Unmarshal
        let unmarshalN = 0;
        for (let i = 0; i < unmarshalers.length; i++) {
            const [newN, val] = unmarshalers[i](unmarshalN, b);
            
            if (typeof values[i] === 'number' && !Number.isInteger(values[i])) {
                 expect(val).toBeCloseTo(values[i]);
            } else {
                 expect(val).toEqual(values[i]);
            }
            unmarshalN = newN;
        }
        expect(unmarshalN).toBe(b.length);
    });
});

describe('Error Handling', () => {
    test('should throw ErrBufTooSmall on insufficient buffer', () => {
        const skipSliceOfBytes = (n, b) => skipSlice(n, b, skipByte);
        const skipMapOfBytes = (n, b) => skipMap(n, b, skipByte, skipByte);
        const unmarshalSliceOfBytes = (n, b) => unmarshalSlice(n, b, unmarshalByte);
        const unmarshalMapOfBytes = (n, b) => unmarshalMap(n, b, unmarshalByte, unmarshalByte);

        const funcs = [
            // Unmarshal funcs
            unmarshalBool, unmarshalByte, unmarshalFloat32, unmarshalFloat64, unmarshalInt,
            unmarshalInt8, unmarshalInt16, unmarshalInt32, unmarshalInt64, unmarshalUint,
            unmarshalUint16, unmarshalUint32, unmarshalUint64, unmarshalString, unmarshalUnsafeString,
            unmarshalBytesCropped, unmarshalBytesCopied, unmarshalSliceOfBytes, unmarshalMapOfBytes,
            // Skip funcs
            skipBool, skipByte, skipFloat32, skipFloat64, skipVarint, skipInt8, skipInt16,
            skipInt32, skipInt64, skipUint16, skipUint32, skipUint64, skipString, skipBytes,
            skipSliceOfBytes, skipMapOfBytes
        ];
        
        for (const func of funcs) {
            expect(() => func(0, new Uint8Array([]))).toThrow(ErrBufTooSmall);
        }
        
        expect(() => unmarshalString(0, new Uint8Array([2, 0]))).toThrow(ErrBufTooSmall);
        expect(() => unmarshalBytesCopied(0, new Uint8Array([10, 0]))).toThrow(ErrBufTooSmall);
        expect(() => unmarshalSlice(0, new Uint8Array([10, 0, 0, 0, 1]), unmarshalByte)).toThrow(ErrBufTooSmall);
    });
});


describe('Complex Types', () => {
    test('should handle slices correctly', () => {
        const slice = ["sliceelement1", "sliceelement2", "sliceelement3"];
        const s = sizeSlice(slice, sizeString);
        const buf = new Uint8Array(s);
        marshalSlice(0, buf, slice, marshalString);

        const finalSkipN = skipSlice(0, buf, skipString);
        expect(finalSkipN).toBe(s);

        const [finalN, retSlice] = unmarshalSlice(0, buf, unmarshalString);
        expect(finalN).toBe(s);
        expect(retSlice).toEqual(slice);
    });

    test('should handle maps correctly', () => {
        const m = new Map([
            ["mapkey1", "mapvalue1"],
            ["mapkey2", "mapvalue2"],
        ]);
        const s = sizeMap(m, sizeString, sizeString);
        const buf = new Uint8Array(s);
        marshalMap(0, buf, m, marshalString, marshalString);
        
        const finalSkipN = skipMap(0, buf, skipString, skipString);
        expect(finalSkipN).toBe(s);
        
        const [finalN, retMap] = unmarshalMap(0, buf, unmarshalString, unmarshalString);
        expect(finalN).toBe(s);
        expect(retMap).toEqual(m);
    });

    test('should handle maps with non-string keys', () => {
        const m = new Map([
            [1, "mapvalue1"],
            [2, "mapvalue2"],
        ]);
        const s = sizeMap(m, sizeInt32, sizeString);
        const buf = new Uint8Array(s);
        marshalMap(0, buf, m, marshalInt32, marshalString);

        const finalSkipN = skipMap(0, buf, skipInt32, skipString);
        expect(finalSkipN).toBe(s);

        const [finalN, retMap] = unmarshalMap(0, buf, unmarshalInt32, unmarshalString);
        expect(finalN).toBe(s);
        expect(retMap).toEqual(m);
    });
});


describe('Edge Cases', () => {
    test('should handle empty strings', () => {
        const str = "";
        const s = sizeString(str);
        const buf = new Uint8Array(s);
        marshalString(0, buf, str);
        
        const finalSkipN = skipString(0, buf);
        expect(finalSkipN).toBe(s);

        const [finalN, retStr] = unmarshalString(0, buf);
        expect(finalN).toBe(s);
        expect(retStr).toBe(str);
    });

    test('should handle long strings', () => {
        const str = "H".repeat(65536); // MaxUint16 + 1
        const s = sizeString(str);
        const buf = new Uint8Array(s);
        marshalString(0, buf, str);
        
        const finalSkipN = skipString(0, buf);
        expect(finalSkipN).toBe(s);

        const [finalN, retStr] = unmarshalString(0, buf);
        expect(finalN).toBe(s);
        expect(retStr).toBe(str);
    });
});

describe('Varint Specific Tests', () => {
    test('should correctly skip varints', () => {
        expect(skipVarint(0, new Uint8Array([0x05]))).toBe(1);
        expect(skipVarint(0, new Uint8Array([0x80, 0x01]))).toBe(2);
        expect(() => skipVarint(0, new Uint8Array([0x80]))).toThrow(ErrBufTooSmall);
        expect(() => skipVarint(0, new Uint8Array(Array(11).fill(0x80)))).toThrow(ErrOverflow);
    });

    test('should correctly unmarshal signed ints (varint)', () => {
        let [n, val] = unmarshalInt(0, new Uint8Array([0x02]));
        expect(n).toBe(1);
        expect(val).toBe(1n);
        
        [n, val] = unmarshalInt(0, new Uint8Array([0x03]));
        expect(n).toBe(1);
        expect(val).toBe(-2n);

        [n, val] = unmarshalInt(0, new Uint8Array([0xAC, 0x02]));
        expect(n).toBe(2);
        expect(val).toBe(150n);
        
        expect(() => unmarshalInt(0, new Uint8Array([0x80]))).toThrow(ErrBufTooSmall);
        expect(() => unmarshalInt(0, new Uint8Array(Array(11).fill(0x80)))).toThrow(ErrOverflow);
    });
    
    test('should correctly unmarshal unsigned ints (varint)', () => {
        let [n, val] = unmarshalUint(0, new Uint8Array([0x07]));
        expect(n).toBe(1);
        expect(val).toBe(7n);

        [n, val] = unmarshalUint(0, new Uint8Array([0xAC, 0x02]));
        expect(n).toBe(2);
        expect(val).toBe(300n);
        
        expect(() => unmarshalUint(0, new Uint8Array([0x80]))).toThrow(ErrBufTooSmall);
        expect(() => unmarshalUint(0, new Uint8Array(Array(11).fill(0x80)))).toThrow(ErrOverflow);
    });
});

describe('Time and Pointer Types', () => {
    test('should handle Time (Date) objects', () => {
        const now = new Date(1663362895123); // Milliseconds
        const s = sizeTime();
        const buf = new Uint8Array(s);
        marshalTime(0, buf, now);
        
        const finalSkipN = skipTime(0, buf);
        expect(finalSkipN).toBe(s);
        
        const [finalN, retTime] = unmarshalTime(0, buf);
        expect(finalN).toBe(s);
        expect(retTime.toDate()).toEqual(now);
        expect(retTime.unixNano).toBe(BigInt(now.getTime()) * 1000000n);
    });

    test('should handle non-null pointers', () => {
        const val = "hello world";
        const s = sizePointer(val, sizeString);
        const buf = new Uint8Array(s);
        marshalPointer(0, buf, val, marshalString);
        
        const finalSkipN = skipPointer(0, buf, skipString);
        expect(finalSkipN).toBe(s);
        
        const [finalN, retPtr] = unmarshalPointer(0, buf, unmarshalString);
        expect(finalN).toBe(s);
        expect(retPtr).not.toBeNull();
        expect(retPtr).toBe(val);
    });
    
    test('should handle null pointers', () => {
        const ptr = null;
        const s = sizePointer(ptr, sizeString);
        const buf = new Uint8Array(s);
        marshalPointer(0, buf, ptr, marshalString);
        
        const finalSkipN = skipPointer(0, buf, skipString);
        expect(finalSkipN).toBe(s);
        
        const [finalN, retPtr] = unmarshalPointer(0, buf, unmarshalString);
        expect(finalN).toBe(s);
        expect(retPtr).toBeNull();
    });
});

