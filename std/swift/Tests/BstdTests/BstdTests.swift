import XCTest
@testable import Bstd 

final class BstdTests: XCTestCase {

    func testStrictWireLimits() throws {
        var index = 0
        XCTAssertThrowsError(try Bstd.unmarshalString(at: &index,
            from: [0x80, 0x80, 0x80, 0x80, 0x10])) { XCTAssertEqual($0 as? BstdError, .decodeLimit) }
        index = 0
        XCTAssertThrowsError(try Bstd.skipVarint(at: &index,
            in: [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02])) {
            XCTAssertEqual($0 as? BstdError, .overflow)
        }
        index = -1
        XCTAssertThrowsError(try Bstd.unmarshalByte(at: &index, from: [7]))
        index = 0
        XCTAssertThrowsError(try Bstd.skipBool(at: &index, in: [2])) { XCTAssertEqual($0 as? BstdError, .invalidData) }
        XCTAssertEqual(index, 0)
        index = 0
        XCTAssertThrowsError(try Bstd.skipSlice(at: &index, in: [0, 0, 0, 0, 0], skipElement: Bstd.skipByte)) {
            XCTAssertEqual($0 as? BstdError, .missingTerminator)
        }
        XCTAssertEqual(index, 0)
        XCTAssertThrowsError(try Bstd.skipString(at: &index, in: [1, 0xff])) {
            XCTAssertEqual($0 as? BstdError, .invalidUTF8)
        }
        XCTAssertEqual(index, 0)
        let duplicateMap: [UInt8] = [2, 7, 1, 7, 2, 1, 1, 1, 1]
        XCTAssertThrowsError(try Bstd.unmarshalMap(at: &index, from: duplicateMap,
            kUnmarshaler: Bstd.unmarshalByte, vUnmarshaler: Bstd.unmarshalByte)) {
            XCTAssertEqual($0 as? BstdError, .invalidData)
        }
        XCTAssertEqual(index, 0)
        var output = [UInt8](repeating: 0, count: 8)
        index = 0
        try Bstd.marshalTimeNanos(-1, at: &index, into: &output)
        index = 0
        XCTAssertEqual(try Bstd.unmarshalTimeNanos(at: &index, from: output), -1)
    }

    // Test Core Data Types & Roundtrip

    func testDataTypesRoundtrip() throws {
        let testStr = "Hello World! With Unicode: 🚀"
        let testBs: [UInt8] = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
        let negIntVal: Int64 = -12345
        let posIntVal: Int64 = Int64.max
        let randomFloat = Float.random(in: -1000...1000)
        let randomDouble = Double.random(in: -1000...1000)
        let randomInt32 = Int32.random(in: Int32.min...Int32.max)
        let randomInt64 = Int64.random(in: Int64.min...Int64.max)
        let randomUInt32 = UInt32.random(in: UInt32.min...UInt32.max)
        let randomUInt64 = UInt64.random(in: UInt64.min...UInt64.max)
        
        // Sizers
        let sizers: [() -> Int] = [
            Bstd.sizeBool, Bstd.sizeByte, Bstd.sizeFloat32, Bstd.sizeFloat64,
            { Bstd.sizeInt(posIntVal) }, { Bstd.sizeInt(negIntVal) },
            Bstd.sizeInt8, Bstd.sizeInt16, Bstd.sizeInt32, Bstd.sizeInt64,
            { Bstd.sizeUInt(UInt64.max) }, Bstd.sizeUInt16, Bstd.sizeUInt32, Bstd.sizeUInt64,
            { Bstd.sizeString(testStr) }, { Bstd.sizeString(testStr) },
            { Bstd.sizeBytes(testBs) }, { Bstd.sizeBytes(testBs) }
        ]
        
        let totalSize = sizers.reduce(0) { $0 + $1() }
        var buffer = [UInt8](repeating: 0, count: totalSize)
        var index = 0

        // Marshal
        try Bstd.marshalBool(true, at: &index, into: &buffer)
        try Bstd.marshalByte(128, at: &index, into: &buffer)
        try Bstd.marshalFloat32(randomFloat, at: &index, into: &buffer)
        try Bstd.marshalFloat64(randomDouble, at: &index, into: &buffer)
        try Bstd.marshalInt(posIntVal, at: &index, into: &buffer)
        try Bstd.marshalInt(negIntVal, at: &index, into: &buffer)
        try Bstd.marshalInt8(-1, at: &index, into: &buffer)
        try Bstd.marshalInt16(-1, at: &index, into: &buffer)
        try Bstd.marshalInt32(randomInt32, at: &index, into: &buffer)
        try Bstd.marshalInt64(randomInt64, at: &index, into: &buffer)
        try Bstd.marshalUInt(UInt64.max, at: &index, into: &buffer)
        try Bstd.marshalUInt16(160, at: &index, into: &buffer)
        try Bstd.marshalUInt32(randomUInt32, at: &index, into: &buffer)
        try Bstd.marshalUInt64(randomUInt64, at: &index, into: &buffer)
        try Bstd.marshalString(testStr, at: &index, into: &buffer)
        try Bstd.marshalUnsafeString(testStr, at: &index, into: &buffer)
        try Bstd.marshalBytes(testBs, at: &index, into: &buffer)
        try Bstd.marshalBytes(testBs, at: &index, into: &buffer)

        XCTAssertEqual(index, totalSize, "Marshal failed: final index does not match calculated size.")

        // Skip
        var skipIndex = 0
        try Bstd.skipBool(at: &skipIndex, in: buffer)
        try Bstd.skipByte(at: &skipIndex, in: buffer)
        try Bstd.skipFloat32(at: &skipIndex, in: buffer)
        try Bstd.skipFloat64(at: &skipIndex, in: buffer)
        try Bstd.skipVarint(at: &skipIndex, in: buffer)
        try Bstd.skipVarint(at: &skipIndex, in: buffer)
        try Bstd.skipInt8(at: &skipIndex, in: buffer)
        try Bstd.skipInt16(at: &skipIndex, in: buffer)
        try Bstd.skipInt32(at: &skipIndex, in: buffer)
        try Bstd.skipInt64(at: &skipIndex, in: buffer)
        try Bstd.skipVarint(at: &skipIndex, in: buffer)
        try Bstd.skipUInt16(at: &skipIndex, in: buffer)
        try Bstd.skipUInt32(at: &skipIndex, in: buffer)
        try Bstd.skipUInt64(at: &skipIndex, in: buffer)
        try Bstd.skipString(at: &skipIndex, in: buffer)
        try Bstd.skipString(at: &skipIndex, in: buffer)
        try Bstd.skipBytes(at: &skipIndex, in: buffer)
        try Bstd.skipBytes(at: &skipIndex, in: buffer)
        
        XCTAssertEqual(skipIndex, totalSize, "Skip failed: final index does not match buffer size.")

        // Unmarshal & Verify
        var unmarshalIndex = 0
        XCTAssertEqual(try Bstd.unmarshalBool(at: &unmarshalIndex, from: buffer), true)
        XCTAssertEqual(try Bstd.unmarshalByte(at: &unmarshalIndex, from: buffer), 128)
        XCTAssertEqual(try Bstd.unmarshalFloat32(at: &unmarshalIndex, from: buffer), randomFloat)
        XCTAssertEqual(try Bstd.unmarshalFloat64(at: &unmarshalIndex, from: buffer), randomDouble)
        XCTAssertEqual(try Bstd.unmarshalInt(at: &unmarshalIndex, from: buffer), posIntVal)
        XCTAssertEqual(try Bstd.unmarshalInt(at: &unmarshalIndex, from: buffer), negIntVal)
        XCTAssertEqual(try Bstd.unmarshalInt8(at: &unmarshalIndex, from: buffer), -1)
        XCTAssertEqual(try Bstd.unmarshalInt16(at: &unmarshalIndex, from: buffer), -1)
        XCTAssertEqual(try Bstd.unmarshalInt32(at: &unmarshalIndex, from: buffer), randomInt32)
        XCTAssertEqual(try Bstd.unmarshalInt64(at: &unmarshalIndex, from: buffer), randomInt64)
        XCTAssertEqual(try Bstd.unmarshalUInt(at: &unmarshalIndex, from: buffer), UInt64.max)
        XCTAssertEqual(try Bstd.unmarshalUInt16(at: &unmarshalIndex, from: buffer), 160)
        XCTAssertEqual(try Bstd.unmarshalUInt32(at: &unmarshalIndex, from: buffer), randomUInt32)
        XCTAssertEqual(try Bstd.unmarshalUInt64(at: &unmarshalIndex, from: buffer), randomUInt64)
        XCTAssertEqual(try Bstd.unmarshalString(at: &unmarshalIndex, from: buffer), testStr)
        XCTAssertEqual(try Bstd.unmarshalUnsafeString(at: &unmarshalIndex, from: buffer), testStr)
        XCTAssertEqual(try Bstd.unmarshalBytes(at: &unmarshalIndex, from: buffer), testBs)
        XCTAssertEqual(try Bstd.unmarshalBytes(at: &unmarshalIndex, from: buffer), testBs)
        
        XCTAssertEqual(unmarshalIndex, totalSize, "Unmarshal failed: final index does not match buffer size.")
    }

    // Generic Error Path Tests

    func testBufferTooSmallErrors() throws {
        // Test unmarshalling from a completely empty buffer
        var index = 0
        XCTAssertThrowsError(try Bstd.unmarshalUInt16(at: &index, from: [])) { XCTAssertEqual($0 as? BstdError, .bufferTooSmall) }
        
        // Test unmarshalling from a partially valid buffer (e.g., length prefix is ok, but data is missing)
        var buffer: [UInt8] = []
        index = 0
        let size = Bstd.sizeUInt(10)
        buffer = [UInt8](repeating: 0, count: size)
        try Bstd.marshalUInt(10, at: &index, into: &buffer) // Marshal length 10
        // NOTE: The buffer only contains the length. Trying to read 10 bytes of data will fail.
        
        index = 0
        XCTAssertThrowsError(try Bstd.unmarshalString(at: &index, from: buffer)) { XCTAssertEqual($0 as? BstdError, .bufferTooSmall) }
    }
    
    func testOverflowErrors() {
        let maxLen = 10
        let overflowBytes = [UInt8](repeating: 0x80, count: maxLen) // 10 bytes with continuation bit
        var index = 0
        XCTAssertThrowsError(try Bstd.unmarshalUInt(at: &index, from: overflowBytes)) { XCTAssertEqual($0 as? BstdError, .overflow) }

        index = 0
        var overflowEdgeCaseBytes = [UInt8](repeating: 0x80, count: maxLen - 1)
        overflowEdgeCaseBytes.append(0x02) // Last byte is > 1
        XCTAssertThrowsError(try Bstd.unmarshalUInt(at: &index, from: overflowEdgeCaseBytes)) { XCTAssertEqual($0 as? BstdError, .overflow) }
    }

    func testTerminatorErrors() throws {
        // Slice: Missing terminator
        let slice: [UInt8] = [1, 2]
        let sliceSize = Bstd.sizeSlice(slice, sizer: { _ in Bstd.sizeByte() })
        var sliceBuf = [UInt8](repeating: 0, count: sliceSize)
        var index = 0
        try Bstd.marshalSlice(slice, at: &index, into: &sliceBuf, marshaler: { val, idx, buf in try Bstd.marshalByte(val, at: &idx, into: &buf) })
        let truncatedSliceBuf = Array(sliceBuf.dropLast()) // Remove terminator
        
        index = 0
        XCTAssertThrowsError(try Bstd.unmarshalSlice(at: &index, from: truncatedSliceBuf, unmarshaler: Bstd.unmarshalByte)) { XCTAssertEqual($0 as? BstdError, .bufferTooSmall) }
        
        // Map: Incorrect terminator
        let map: [UInt8: UInt8] = [1: 2]
        let mapSize = Bstd.sizeMap(map, kSizer: { _ in Bstd.sizeByte() }, vSizer: { _ in Bstd.sizeByte() })
        var mapBuf = [UInt8](repeating: 0, count: mapSize)
        index = 0
        try Bstd.marshalMap(map, at: &index, into: &mapBuf, kMarshaler: { k, idx, buf in try Bstd.marshalByte(k, at: &idx, into: &buf) }, vMarshaler: { v, idx, buf in try Bstd.marshalByte(v, at: &idx, into: &buf) })
        mapBuf[mapBuf.count - 1] = 99 // Corrupt terminator
        
        index = 0
        XCTAssertThrowsError(try Bstd.unmarshalMap(at: &index, from: mapBuf, kUnmarshaler: Bstd.unmarshalByte, vUnmarshaler: Bstd.unmarshalByte)) { XCTAssertEqual($0 as? BstdError, .missingTerminator) }
    }
    
    func testInvalidUTF8Error() throws {
        // A byte sequence that is not valid UTF-8 (e.g., isolated continuation byte)
        let invalidUTF8Bytes: [UInt8] = [0x80]
        var buffer: [UInt8] = []
        var index = 0
        let size = Bstd.sizeBytes(invalidUTF8Bytes)
        buffer = [UInt8](repeating: 0, count: size)
        try Bstd.marshalBytes(invalidUTF8Bytes, at: &index, into: &buffer)
        
        index = 0
        XCTAssertThrowsError(try Bstd.unmarshalString(at: &index, from: buffer)) { XCTAssertEqual($0 as? BstdError, .invalidUTF8) }
    }


    // Specific Logic & Edge Case Tests

    func testEmptyCollectionsAndStrings() throws {
        // Empty String
        let emptyStr = ""
        let sizeStr = Bstd.sizeString(emptyStr)
        var buffer: [UInt8] = [UInt8](repeating: 0, count: sizeStr)
        var index = 0
        try Bstd.marshalString(emptyStr, at: &index, into: &buffer)
        var unmarshalIndex = 0
        XCTAssertEqual(try Bstd.unmarshalString(at: &unmarshalIndex, from: buffer), emptyStr)
        XCTAssertEqual(buffer.count, unmarshalIndex)

        // Empty Slice
        let emptySlice: [Int64] = []
        let sizeSlice = Bstd.sizeSlice(emptySlice, sizer: Bstd.sizeInt)
        buffer = [UInt8](repeating: 0, count: sizeSlice)
        index = 0
        try Bstd.marshalSlice(emptySlice, at: &index, into: &buffer, marshaler: { _,_,_ in })
        unmarshalIndex = 0
        let resSlice: [Int64] = try Bstd.unmarshalSlice(at: &unmarshalIndex, from: buffer, unmarshaler: { _,_ in XCTFail("Unmarshaler should not be called for empty slice"); return 0 })
        XCTAssertTrue(resSlice.isEmpty)
        XCTAssertEqual(buffer.count, unmarshalIndex)

        // Empty Map
        let emptyMap: [String: String] = [:]
        let sizeMap = Bstd.sizeMap(emptyMap, kSizer: Bstd.sizeString, vSizer: Bstd.sizeString)
        buffer = [UInt8](repeating: 0, count: sizeMap)
        index = 0
        try Bstd.marshalMap(emptyMap, at: &index, into: &buffer, kMarshaler: { _,_,_ in }, vMarshaler: { _,_,_ in })
        unmarshalIndex = 0
        let resMap: [String: String] = try Bstd.unmarshalMap(at: &unmarshalIndex, from: buffer, kUnmarshaler: { _,_ in "" }, vUnmarshaler: { _,_ in "" })
        XCTAssertTrue(resMap.isEmpty)
        XCTAssertEqual(buffer.count, unmarshalIndex)
    }

    func testTime() throws {
        // Test with a specific date to ensure nanosecond precision
        let date = Date(timeIntervalSince1970: 1663362895.123456789)
        let size = Bstd.sizeTime()
        var buffer = [UInt8](repeating: 0, count: size)
        var index = 0
        try Bstd.marshalTime(date, at: &index, into: &buffer)

        var skipIndex = 0
        try Bstd.skipTime(at: &skipIndex, in: buffer)
        XCTAssertEqual(skipIndex, buffer.count)

        var unmarshalIndex = 0
        let retTime = try Bstd.unmarshalTime(at: &unmarshalIndex, from: buffer)
        // Adjust accuracy to account for Double's precision limitations with nanoseconds.
        // 1e-7 corresponds to a 100-nanosecond tolerance, which is reasonable.
        XCTAssertEqual(date.timeIntervalSince1970, retTime.timeIntervalSince1970, accuracy: 1e-7)
    }
    
    func testPointer() throws {
        // Non-nil
        let val = "hello world"
        let ptr: String? = val
        let size = Bstd.sizePointer(ptr, sizer: Bstd.sizeString)
        var buffer = [UInt8](repeating: 0, count: size)
        var index = 0
        try Bstd.marshalPointer(ptr, at: &index, into: &buffer, marshaler: Bstd.marshalString)

        var skipIndex = 0
        try Bstd.skipPointer(at: &skipIndex, in: buffer, skipElement: Bstd.skipString)
        XCTAssertEqual(skipIndex, buffer.count)

        var unmarshalIndex = 0
        let retPtr = try Bstd.unmarshalPointer(at: &unmarshalIndex, from: buffer, unmarshaler: Bstd.unmarshalString)
        XCTAssertEqual(retPtr, val)

        // Nil
        let nilPtr: String? = nil
        let sizeNil = Bstd.sizePointer(nilPtr, sizer: Bstd.sizeString)
        buffer = [UInt8](repeating: 0, count: sizeNil)
        index = 0
        try Bstd.marshalPointer(nilPtr, at: &index, into: &buffer, marshaler: Bstd.marshalString)

        skipIndex = 0
        try Bstd.skipPointer(at: &skipIndex, in: buffer, skipElement: Bstd.skipString)
        XCTAssertEqual(skipIndex, buffer.count)

        unmarshalIndex = 0
        let retNilPtr = try Bstd.unmarshalPointer(at: &unmarshalIndex, from: buffer, unmarshaler: Bstd.unmarshalString)
        XCTAssertNil(retNilPtr)
    }
    
    func testFixedSliceSize() {
        let slice: [Int32] = [1, 2, 3]
        let elemSize = Bstd.sizeInt32()
        let expected = Bstd.sizeUInt(UInt64(slice.count)) + (slice.count * elemSize) + BstdConstants.terminator.count
        XCTAssertEqual(Bstd.sizeFixedSlice(slice, elementSize: elemSize), expected)
    }
    
    // Complex Nested Structures Test
    
    func testComplexStructure() throws {
        // A map where keys are strings and values are slices of optional integers
        let complexData: [String: [Int64?]] = [
            "primes": [2, 3, 5, 7, nil, 11],
            "fibonacci": [1, 1, 2, 3, 5, 8, 13],
            "empty": [],
            "nils": [nil, nil]
        ]
        
        // Define sizer for the nested structure
        let valueSizer: ([Int64?]) -> Int = { slice in
            return Bstd.sizeSlice(slice, sizer: { Bstd.sizePointer($0, sizer: Bstd.sizeInt) })
        }
        
        // Calculate size
        let totalSize = Bstd.sizeMap(complexData, kSizer: Bstd.sizeString, vSizer: valueSizer)
        var buffer = [UInt8](repeating: 0, count: totalSize)
        var index = 0
        
        // Define marshaler for the nested structure
        let valueMarshaler: ([Int64?], inout Int, inout [UInt8]) throws -> Void = { slice, index, buffer in
            try Bstd.marshalSlice(slice, at: &index, into: &buffer) { optInt, idx, buf in
                try Bstd.marshalPointer(optInt, at: &idx, into: &buf, marshaler: Bstd.marshalInt)
            }
        }
        
        // Marshal the data
        try Bstd.marshalMap(complexData, at: &index, into: &buffer, kMarshaler: Bstd.marshalString, vMarshaler: valueMarshaler)
        XCTAssertEqual(index, totalSize)
        
        // Define skipper for the nested structure
        let valueSkipper: (inout Int, [UInt8]) throws -> Void = { index, buffer in
            try Bstd.skipSlice(at: &index, in: buffer) { idx, buf in
                try Bstd.skipPointer(at: &idx, in: buf, skipElement: Bstd.skipVarint)
            }
        }
        
        // Skip the data
        var skipIndex = 0
        try Bstd.skipMap(at: &skipIndex, in: buffer, skipKey: Bstd.skipString, skipValue: valueSkipper)
        XCTAssertEqual(skipIndex, totalSize)
        
        // Define unmarshaler for the nested structure
        let valueUnmarshaler: (inout Int, [UInt8]) throws -> [Int64?] = { index, buffer in
            return try Bstd.unmarshalSlice(at: &index, from: buffer) { idx, buf in
                return try Bstd.unmarshalPointer(at: &idx, from: buf, unmarshaler: Bstd.unmarshalInt)
            }
        }
        
        // Unmarshal the data
        var unmarshalIndex = 0
        let result = try Bstd.unmarshalMap(at: &unmarshalIndex, from: buffer, kUnmarshaler: Bstd.unmarshalString, vUnmarshaler: valueUnmarshaler)
        XCTAssertEqual(unmarshalIndex, totalSize)
        
        // Verify the result
        XCTAssertEqual(result.count, complexData.count)
        for (key, value) in complexData {
            XCTAssertNotNil(result[key])
            XCTAssertEqual(result[key]!, value)
        }
    }
}
