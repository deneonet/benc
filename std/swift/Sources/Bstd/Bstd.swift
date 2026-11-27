import Foundation

// Error Types

public enum BstdError: Error, Equatable {
    /// The provided buffer is too small for the requested operation.
    case bufferTooSmall
    /// A variable-length integer is malformed or exceeds its maximum size.
    case overflow
    /// A byte sequence could not be decoded as a valid UTF-8 string.
    case invalidUTF8
    /// A required terminator sequence for a slice or map was not found.
    case missingTerminator
    case decodeLimit
    case invalidData
}

// Constants

// Made internal (by removing `private`) to be accessible by the test target
struct BstdConstants {
    /// The 4-byte terminator used to mark the end of slices and maps.
    static let terminator: [UInt8] = [1, 1, 1, 1]
    /// Maximum number of bytes for a 64-bit varint.
    static let maxVarintLen64 = 10
    static let maxDecodeBytes = 64 * 1024 * 1024
    static let maxDecodeElements = 1_048_576
}

// Core Bstd Implementation

public struct Bstd {
    private static func hasRoom(_ index: Int, _ size: Int, _ buffer: [UInt8]) -> Bool {
        index >= 0 && index <= buffer.count && size >= 0 && size <= buffer.count - index
    }

    private static func readLength(at index: inout Int, from buffer: [UInt8]) throws -> Int {
        var cursor = index
        let raw = try unmarshalUInt(at: &cursor, from: buffer)
        guard raw <= UInt64(BstdConstants.maxDecodeBytes) else { throw BstdError.decodeLimit }
        let length = Int(raw)
        guard hasRoom(cursor, length, buffer) else { throw BstdError.bufferTooSmall }
        index = cursor
        return length
    }

    private static func readCount(at index: inout Int, from buffer: [UInt8], minimum: Int) throws -> Int {
        var cursor = index
        let raw = try unmarshalUInt(at: &cursor, from: buffer)
        guard raw <= UInt64(BstdConstants.maxDecodeElements) else { throw BstdError.decodeLimit }
        guard hasRoom(cursor, 0, buffer), raw <= UInt64((buffer.count - cursor) / minimum) else { throw BstdError.bufferTooSmall }
        index = cursor
        return Int(raw)
    }

    private static func checkTerminator(at index: inout Int, from buffer: [UInt8]) throws {
        guard hasRoom(index, 4, buffer) else { throw BstdError.bufferTooSmall }
        guard buffer[index..<index+4].elementsEqual(BstdConstants.terminator) else { throw BstdError.missingTerminator }
        index += 4
    }
    
    // Helper Functions
    
    /// Encodes a signed 64-bit integer into an unsigned 64-bit integer using ZigZag encoding.
    private static func encodeZigZag(_ value: Int64) -> UInt64 {
        // (value << 1) duplicates the sign bit, ^ (value >> 63) flips the bits for negative numbers
        return UInt64(bitPattern: (value << 1) ^ (value >> 63))
    }
    
    /// Decodes an unsigned 64-bit integer into a signed 64-bit integer using ZigZag encoding.
    private static func decodeZigZag(_ value: UInt64) -> Int64 {
        // For odd numbers (negative originals), we perform a bitwise NOT on the shifted value.
        // For even numbers (positive originals), we just shift.
        // This correctly reverses the ZigZag encoding: (n >> 1) ^ -(n & 1)
        if (value & 1) == 1 {
            return ~(Int64(bitPattern: value >> 1))
        } else {
            return Int64(bitPattern: value >> 1)
        }
    }
    
    // String Operations
    
    /// Calculates the number of bytes needed to marshal a string.
    public static func sizeString(_ str: String) -> Int {
        let length = str.utf8.count
        return sizeUInt(UInt64(length)) + length
    }
    
    /// Marshals a string into the buffer at the given index.
    /// - Throws: `BstdError.bufferTooSmall` if the buffer is not large enough.
    public static func marshalString(_ str: String, at index: inout Int, into buffer: inout [UInt8]) throws {
        let utf8View = str.utf8
        let requiredSizeForLength = sizeUInt(UInt64(utf8View.count))
        guard hasRoom(index, requiredSizeForLength, buffer),
              utf8View.count <= buffer.count - index - requiredSizeForLength else {
            throw BstdError.bufferTooSmall
        }
        
        try marshalUInt(UInt64(utf8View.count), at: &index, into: &buffer)
        
        buffer.replaceSubrange(index..<index + utf8View.count, with: utf8View)
        index += utf8View.count
    }
    
    /// Unmarshals a string from the buffer at the given index.
    /// - Throws: `BstdError` for buffer issues, overflow, or invalid UTF-8 data.
    public static func unmarshalString(at index: inout Int, from buffer: [UInt8]) throws -> String {
        var cursor = index
        let length = try readLength(at: &cursor, from: buffer)
        let stringData = buffer[cursor..<cursor+length]
        guard let string = String(bytes: stringData, encoding: .utf8) else {
            throw BstdError.invalidUTF8
        }
        
        index = cursor + length
        return string
    }
    
    /// Skips over a marshalled string in the buffer.
    /// - Throws: `BstdError` for buffer issues or overflow.
    public static func skipString(at index: inout Int, in buffer: [UInt8]) throws {
        var cursor = index
        let length = try readLength(at: &cursor, from: buffer)
        guard String(bytes: buffer[cursor..<cursor+length], encoding: .utf8) != nil else { throw BstdError.invalidUTF8 }
        index = cursor + length
    }
    
    // "Unsafe" String Operations
    // In Swift, standard library conversions are already highly optimized. These are for API compatibility.

    /// Marshals a string into the buffer at the given index.
    public static func marshalUnsafeString(_ str: String, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalString(str, at: &index, into: &buffer)
    }

    /// Unmarshals a string from the buffer at the given index.
    public static func unmarshalUnsafeString(at index: inout Int, from buffer: [UInt8]) throws -> String {
        return try unmarshalString(at: &index, from: buffer)
    }

    // Slice Operations
    
    /// Calculates the size needed to marshal a slice of elements with dynamic sizes.
    public static func sizeSlice<T>(_ slice: [T], sizer: (T) -> Int) -> Int {
        var size = sizeUInt(UInt64(slice.count))
        size += slice.reduce(0) { $0 + sizer($1) }
        return size + BstdConstants.terminator.count
    }
    
    /// Calculates the size needed to marshal a slice of elements with a fixed size.
    public static func sizeFixedSlice<T>(_ slice: [T], elementSize: Int) -> Int {
        return sizeUInt(UInt64(slice.count)) + (slice.count * elementSize) + BstdConstants.terminator.count
    }
    
    /// Marshals a slice into the buffer.
    public static func marshalSlice<T>(_ slice: [T], at index: inout Int, into buffer: inout [UInt8], marshaler: (T, inout Int, inout [UInt8]) throws -> Void) throws {
        try marshalUInt(UInt64(slice.count), at: &index, into: &buffer)
        
        for item in slice {
            try marshaler(item, &index, &buffer)
        }
        
        guard hasRoom(index, BstdConstants.terminator.count, buffer) else {
            throw BstdError.bufferTooSmall
        }
        
        buffer.replaceSubrange(index..<index+BstdConstants.terminator.count, with: BstdConstants.terminator)
        index += BstdConstants.terminator.count
    }
    
    /// Unmarshals a slice from the buffer.
    public static func unmarshalSlice<T>(at index: inout Int, from buffer: [UInt8], unmarshaler: (inout Int, [UInt8]) throws -> T) throws -> [T] {
        var cursor = index
        let count = try readCount(at: &cursor, from: buffer, minimum: 1)
        var result: [T] = []
        
        for _ in 0..<count {
            result.append(try unmarshaler(&cursor, buffer))
        }
        
        try checkTerminator(at: &cursor, from: buffer)
        index = cursor
        return result
    }
    
    /// Skips over a marshalled slice in the buffer.
    public static func skipSlice(at index: inout Int, in buffer: [UInt8], skipElement: (inout Int, [UInt8]) throws -> Void) throws {
        var cursor = index
        let count = try readCount(at: &cursor, from: buffer, minimum: 1)
        
        for _ in 0..<count {
            try skipElement(&cursor, buffer)
        }
        
        try checkTerminator(at: &cursor, from: buffer)
        index = cursor
    }

    // Map Operations
    
    /// Calculates the size needed to marshal a map.
    public static func sizeMap<K: Hashable, V>(_ map: [K: V], kSizer: (K) -> Int, vSizer: (V) -> Int) -> Int {
        var size = sizeUInt(UInt64(map.count))
        size += map.reduce(0) { $0 + kSizer($1.key) + vSizer($1.value) }
        return size + BstdConstants.terminator.count
    }
    
    /// Marshals a map into the buffer.
    public static func marshalMap<K, V>(_ map: [K: V], at index: inout Int, into buffer: inout [UInt8], kMarshaler: (K, inout Int, inout [UInt8]) throws -> Void, vMarshaler: (V, inout Int, inout [UInt8]) throws -> Void) throws {
        try marshalUInt(UInt64(map.count), at: &index, into: &buffer)
        
        for (key, value) in map {
            try kMarshaler(key, &index, &buffer)
            try vMarshaler(value, &index, &buffer)
        }
        
        guard hasRoom(index, BstdConstants.terminator.count, buffer) else {
            throw BstdError.bufferTooSmall
        }
        
        buffer.replaceSubrange(index..<index+BstdConstants.terminator.count, with: BstdConstants.terminator)
        index += BstdConstants.terminator.count
    }
    
    /// Unmarshals a map from the buffer.
    public static func unmarshalMap<K, V>(at index: inout Int, from buffer: [UInt8], kUnmarshaler: (inout Int, [UInt8]) throws -> K, vUnmarshaler: (inout Int, [UInt8]) throws -> V) throws -> [K: V] where K: Hashable {
        var cursor = index
        let count = try readCount(at: &cursor, from: buffer, minimum: 2)
        var result: [K: V] = [:]
        
        for _ in 0..<count {
            let key = try kUnmarshaler(&cursor, buffer)
            let value = try vUnmarshaler(&cursor, buffer)
            if case .some = result.updateValue(value, forKey: key) { throw BstdError.invalidData }
        }
        
        try checkTerminator(at: &cursor, from: buffer)
        index = cursor
        return result
    }
    
    /// Skips over a marshalled map in the buffer.
    public static func skipMap(at index: inout Int, in buffer: [UInt8], skipKey: (inout Int, [UInt8]) throws -> Void, skipValue: (inout Int, [UInt8]) throws -> Void) throws {
        var cursor = index
        let count = try readCount(at: &cursor, from: buffer, minimum: 2)
        
        for _ in 0..<count {
            try skipKey(&cursor, buffer)
            try skipValue(&cursor, buffer)
        }
        
        try checkTerminator(at: &cursor, from: buffer)
        index = cursor
    }

    // Byte & Byte Slice Operations
    
    /// Calculates the size needed to marshal a single byte (always 1).
    public static func sizeByte() -> Int {
        return 1
    }
    
    /// Marshals a single byte into the buffer.
    public static func marshalByte(_ value: UInt8, at index: inout Int, into buffer: inout [UInt8]) throws {
        guard hasRoom(index, 1, buffer) else { throw BstdError.bufferTooSmall }
        buffer[index] = value
        index += 1
    }
    
    /// Unmarshals a single byte from the buffer.
    public static func unmarshalByte(at index: inout Int, from buffer: [UInt8]) throws -> UInt8 {
        guard hasRoom(index, 1, buffer) else { throw BstdError.bufferTooSmall }
        let value = buffer[index]
        index += 1
        return value
    }
    
    /// Skips a single byte in the buffer.
    public static func skipByte(at index: inout Int, in buffer: [UInt8]) throws {
        guard hasRoom(index, 1, buffer) else { throw BstdError.bufferTooSmall }
        index += 1
    }
    
    /// Calculates the size needed to marshal a byte slice.
    public static func sizeBytes(_ data: [UInt8]) -> Int {
        return sizeUInt(UInt64(data.count)) + data.count
    }
    
    /// Marshals a byte slice into the buffer.
    public static func marshalBytes(_ data: [UInt8], at index: inout Int, into buffer: inout [UInt8]) throws {
        let prefixSize = sizeUInt(UInt64(data.count))
        guard hasRoom(index, prefixSize, buffer), data.count <= buffer.count - index - prefixSize else {
            throw BstdError.bufferTooSmall
        }
        try marshalUInt(UInt64(data.count), at: &index, into: &buffer)
        buffer.replaceSubrange(index..<index+data.count, with: data)
        index += data.count
    }
    
    /// Unmarshals a byte slice from the buffer. In Swift, this always creates a copy.
    public static func unmarshalBytes(at index: inout Int, from buffer: [UInt8]) throws -> [UInt8] {
        var cursor = index
        let length = try readLength(at: &cursor, from: buffer)
        let result = Array(buffer[cursor..<cursor+length])
        index = cursor + length
        return result
    }

    /// Skips over a marshalled byte slice in the buffer.
    public static func skipBytes(at index: inout Int, in buffer: [UInt8]) throws {
        let length = try readLength(at: &index, from: buffer)
        index += length
    }
    
    // Varint Operations
    
    /// Skips over a marshalled varint in the buffer.
    public static func skipVarint(at index: inout Int, in buffer: [UInt8]) throws {
        let originalIndex = index
        guard hasRoom(originalIndex, 0, buffer) else { throw BstdError.bufferTooSmall }
        for i in 0..<BstdConstants.maxVarintLen64 {
            guard i < buffer.count - originalIndex else {
                throw BstdError.bufferTooSmall
            }
            if buffer[originalIndex + i] < 0x80 {
                if i == 9 && buffer[originalIndex + i] > 1 { throw BstdError.overflow }
                index = originalIndex + i + 1
                return
            }
        }
        throw BstdError.overflow
    }

    /// Calculates the size needed for a ZigZag encoded signed 64-bit integer.
    public static func sizeInt(_ value: Int64) -> Int {
        return sizeUInt(encodeZigZag(value))
    }

    /// Marshals a signed 64-bit integer using ZigZag encoding.
    public static func marshalInt(_ value: Int64, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalUInt(encodeZigZag(value), at: &index, into: &buffer)
    }

    /// Unmarshals a signed 64-bit integer using ZigZag encoding.
    public static func unmarshalInt(at index: inout Int, from buffer: [UInt8]) throws -> Int64 {
        return decodeZigZag(try unmarshalUInt(at: &index, from: buffer))
    }

    /// Calculates the size needed for an unsigned 64-bit integer.
    public static func sizeUInt(_ value: UInt64) -> Int {
        if value == 0 { return 1 }
        // This calculates floor(log_128(value)) + 1, which is the number of bytes for a varint.
        let bitWidth = (64 - value.leadingZeroBitCount)
        return (bitWidth + 6) / 7
    }
    
    /// Marshals an unsigned 64-bit integer as a varint.
    public static func marshalUInt(_ value: UInt64, at index: inout Int, into buffer: inout [UInt8]) throws {
        var v = value
        guard hasRoom(index, sizeUInt(value), buffer) else { throw BstdError.bufferTooSmall }
        
        while v >= 0x80 {
            guard index < buffer.count else { throw BstdError.bufferTooSmall }
            buffer[index] = UInt8(v & 0x7F) | 0x80
            v >>= 7
            index += 1
        }
        guard index < buffer.count else { throw BstdError.bufferTooSmall }
        buffer[index] = UInt8(v)
        index += 1
    }
    
    /// Unmarshals an unsigned 64-bit integer from a varint.
    public static func unmarshalUInt(at index: inout Int, from buffer: [UInt8]) throws -> UInt64 {
        var value: UInt64 = 0
        var shift: UInt64 = 0
        let startIndex = index
        guard hasRoom(startIndex, 0, buffer) else { throw BstdError.bufferTooSmall }
        
        for i in 0..<BstdConstants.maxVarintLen64 {
            guard i < buffer.count - startIndex else {
                throw BstdError.bufferTooSmall
            }
            let byte = buffer[startIndex + i]
            if byte < 0x80 {
                // Check for overflow on the last byte. The value must be at most 1.
                if i == BstdConstants.maxVarintLen64 - 1 && byte > 1 {
                    throw BstdError.overflow
                }
                index = startIndex + i + 1
                return value | (UInt64(byte) << shift)
            }
            value |= UInt64(byte & 0x7F) << shift
            shift += 7
        }
        // If the loop completes, it means the continuation bit was set on the 10th byte.
        throw BstdError.overflow
    }
    
    // Fixed Size Primitives
    
    public static func sizeInt8() -> Int { return 1 }
    public static func marshalInt8(_ value: Int8, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalByte(UInt8(bitPattern: value), at: &index, into: &buffer)
    }
    public static func unmarshalInt8(at index: inout Int, from buffer: [UInt8]) throws -> Int8 {
        return Int8(bitPattern: try unmarshalByte(at: &index, from: buffer))
    }
    public static func skipInt8(at index: inout Int, in buffer: [UInt8]) throws {
        try skipByte(at: &index, in: buffer)
    }

    public static func sizeUInt16() -> Int { return 2 }
    public static func marshalUInt16(_ value: UInt16, at index: inout Int, into buffer: inout [UInt8]) throws {
        guard hasRoom(index, 2, buffer) else { throw BstdError.bufferTooSmall }
        buffer[index]     = UInt8(truncatingIfNeeded: value)
        buffer[index + 1] = UInt8(truncatingIfNeeded: value >> 8)
        index += 2
    }
    public static func unmarshalUInt16(at index: inout Int, from buffer: [UInt8]) throws -> UInt16 {
        guard hasRoom(index, 2, buffer) else { throw BstdError.bufferTooSmall }
        let value = UInt16(buffer[index]) | (UInt16(buffer[index + 1]) << 8)
        index += 2
        return value
    }
    public static func skipUInt16(at index: inout Int, in buffer: [UInt8]) throws {
        guard hasRoom(index, 2, buffer) else { throw BstdError.bufferTooSmall }
        index += 2
    }
    
    public static func sizeInt16() -> Int { return 2 }
    public static func marshalInt16(_ value: Int16, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalUInt16(UInt16(bitPattern: value), at: &index, into: &buffer)
    }
    public static func unmarshalInt16(at index: inout Int, from buffer: [UInt8]) throws -> Int16 {
        return Int16(bitPattern: try unmarshalUInt16(at: &index, from: buffer))
    }
    public static func skipInt16(at index: inout Int, in buffer: [UInt8]) throws {
        try skipUInt16(at: &index, in: buffer)
    }

    public static func sizeUInt32() -> Int { return 4 }
    public static func marshalUInt32(_ value: UInt32, at index: inout Int, into buffer: inout [UInt8]) throws {
        guard hasRoom(index, 4, buffer) else { throw BstdError.bufferTooSmall }
        buffer[index]     = UInt8(truncatingIfNeeded: value)
        buffer[index + 1] = UInt8(truncatingIfNeeded: value >> 8)
        buffer[index + 2] = UInt8(truncatingIfNeeded: value >> 16)
        buffer[index + 3] = UInt8(truncatingIfNeeded: value >> 24)
        index += 4
    }
    public static func unmarshalUInt32(at index: inout Int, from buffer: [UInt8]) throws -> UInt32 {
        guard hasRoom(index, 4, buffer) else { throw BstdError.bufferTooSmall }
        let value = UInt32(buffer[index]) |
                    UInt32(buffer[index + 1]) << 8 |
                    UInt32(buffer[index + 2]) << 16 |
                    UInt32(buffer[index + 3]) << 24
        index += 4
        return value
    }
    public static func skipUInt32(at index: inout Int, in buffer: [UInt8]) throws {
        guard hasRoom(index, 4, buffer) else { throw BstdError.bufferTooSmall }
        index += 4
    }
    
    public static func sizeInt32() -> Int { return 4 }
    public static func marshalInt32(_ value: Int32, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalUInt32(UInt32(bitPattern: value), at: &index, into: &buffer)
    }
    public static func unmarshalInt32(at index: inout Int, from buffer: [UInt8]) throws -> Int32 {
        return Int32(bitPattern: try unmarshalUInt32(at: &index, from: buffer))
    }
    public static func skipInt32(at index: inout Int, in buffer: [UInt8]) throws {
        try skipUInt32(at: &index, in: buffer)
    }

    public static func sizeUInt64() -> Int { return 8 }
    public static func marshalUInt64(_ value: UInt64, at index: inout Int, into buffer: inout [UInt8]) throws {
        guard hasRoom(index, 8, buffer) else { throw BstdError.bufferTooSmall }
        buffer[index]     = UInt8(truncatingIfNeeded: value)
        buffer[index + 1] = UInt8(truncatingIfNeeded: value >> 8)
        buffer[index + 2] = UInt8(truncatingIfNeeded: value >> 16)
        buffer[index + 3] = UInt8(truncatingIfNeeded: value >> 24)
        buffer[index + 4] = UInt8(truncatingIfNeeded: value >> 32)
        buffer[index + 5] = UInt8(truncatingIfNeeded: value >> 40)
        buffer[index + 6] = UInt8(truncatingIfNeeded: value >> 48)
        buffer[index + 7] = UInt8(truncatingIfNeeded: value >> 56)
        index += 8
    }
    public static func unmarshalUInt64(at index: inout Int, from buffer: [UInt8]) throws -> UInt64 {
        guard hasRoom(index, 8, buffer) else { throw BstdError.bufferTooSmall }
        let value = UInt64(buffer[index]) |
                    UInt64(buffer[index + 1]) << 8 |
                    UInt64(buffer[index + 2]) << 16 |
                    UInt64(buffer[index + 3]) << 24 |
                    UInt64(buffer[index + 4]) << 32 |
                    UInt64(buffer[index + 5]) << 40 |
                    UInt64(buffer[index + 6]) << 48 |
                    UInt64(buffer[index + 7]) << 56
        index += 8
        return value
    }
    public static func skipUInt64(at index: inout Int, in buffer: [UInt8]) throws {
        guard hasRoom(index, 8, buffer) else { throw BstdError.bufferTooSmall }
        index += 8
    }
    
    public static func sizeInt64() -> Int { return 8 }
    public static func marshalInt64(_ value: Int64, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalUInt64(UInt64(bitPattern: value), at: &index, into: &buffer)
    }
    public static func unmarshalInt64(at index: inout Int, from buffer: [UInt8]) throws -> Int64 {
        return Int64(bitPattern: try unmarshalUInt64(at: &index, from: buffer))
    }
    public static func skipInt64(at index: inout Int, in buffer: [UInt8]) throws {
        try skipUInt64(at: &index, in: buffer)
    }
    
    public static func sizeFloat32() -> Int { return 4 }
    public static func marshalFloat32(_ value: Float, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalUInt32(value.bitPattern, at: &index, into: &buffer)
    }
    public static func unmarshalFloat32(at index: inout Int, from buffer: [UInt8]) throws -> Float {
        return Float(bitPattern: try unmarshalUInt32(at: &index, from: buffer))
    }
    public static func skipFloat32(at index: inout Int, in buffer: [UInt8]) throws {
        try skipUInt32(at: &index, in: buffer)
    }
    
    public static func sizeFloat64() -> Int { return 8 }
    public static func marshalFloat64(_ value: Double, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalUInt64(value.bitPattern, at: &index, into: &buffer)
    }
    public static func unmarshalFloat64(at index: inout Int, from buffer: [UInt8]) throws -> Double {
        return Double(bitPattern: try unmarshalUInt64(at: &index, from: buffer))
    }
    public static func skipFloat64(at index: inout Int, in buffer: [UInt8]) throws {
        try skipUInt64(at: &index, in: buffer)
    }
    
    public static func sizeBool() -> Int { return 1 }
    public static func marshalBool(_ value: Bool, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalByte(value ? 1 : 0, at: &index, into: &buffer)
    }
    public static func unmarshalBool(at index: inout Int, from buffer: [UInt8]) throws -> Bool {
        var cursor = index
        let byte = try unmarshalByte(at: &cursor, from: buffer)
        guard byte <= 1 else { throw BstdError.invalidData }
        index = cursor
        return byte == 1
    }
    public static func skipBool(at index: inout Int, in buffer: [UInt8]) throws {
        _ = try unmarshalBool(at: &index, from: buffer)
    }

    // Time / Date Operations
    
    /// Calculates the size needed to marshal a Date (as a 64-bit integer).
    public static func sizeTime() -> Int {
        return sizeInt64()
    }
    
    /// Marshals a Date as its nanoseconds since the Unix epoch.
    public static func marshalTime(_ date: Date, at index: inout Int, into buffer: inout [UInt8]) throws {
        let timeInterval = date.timeIntervalSince1970
        guard timeInterval.isFinite, timeInterval > -9_223_372_037, timeInterval < 9_223_372_037 else {
            throw BstdError.overflow
        }
        let whole = Int64(timeInterval)
        let (seconds, overflow) = whole.multipliedReportingOverflow(by: 1_000_000_000)
        guard !overflow else { throw BstdError.overflow }
        let fractional = Int64((timeInterval - Double(whole)) * 1_000_000_000)
        let (nanos, sumOverflow) = seconds.addingReportingOverflow(fractional)
        guard !sumOverflow else { throw BstdError.overflow }
        try marshalInt64(nanos, at: &index, into: &buffer)
    }

    /// Exact wire representation. Use this for nanosecond-preserving round trips.
    public static func marshalTimeNanos(_ nanos: Int64, at index: inout Int, into buffer: inout [UInt8]) throws {
        try marshalInt64(nanos, at: &index, into: &buffer)
    }

    public static func unmarshalTimeNanos(at index: inout Int, from buffer: [UInt8]) throws -> Int64 {
        try unmarshalInt64(at: &index, from: buffer)
    }
    
    /// Unmarshals a Date from its nanoseconds since the Unix epoch.
    public static func unmarshalTime(at index: inout Int, from buffer: [UInt8]) throws -> Date {
        let totalNanoseconds = try unmarshalInt64(at: &index, from: buffer)
        let seconds = totalNanoseconds / 1_000_000_000
        let nanoseconds = totalNanoseconds % 1_000_000_000
        let timeInterval = TimeInterval(seconds) + TimeInterval(nanoseconds) / 1_000_000_000
        return Date(timeIntervalSince1970: timeInterval)
    }
    
    /// Skips over a marshalled Date.
    public static func skipTime(at index: inout Int, in buffer: [UInt8]) throws {
        try skipInt64(at: &index, in: buffer)
    }
    
    // Pointer / Optional Operations

    /// Calculates the size needed to marshal an optional value.
    public static func sizePointer<T>(_ value: T?, sizer: (T) -> Int) -> Int {
        if let unwrapped = value {
            return sizeBool() + sizer(unwrapped)
        }
        return sizeBool()
    }

    /// Marshals an optional value, prefixed with a boolean to indicate its presence.
    public static func marshalPointer<T>(_ value: T?, at index: inout Int, into buffer: inout [UInt8], marshaler: (T, inout Int, inout [UInt8]) throws -> Void) throws {
        try marshalBool(value != nil, at: &index, into: &buffer)
        if let unwrapped = value {
            try marshaler(unwrapped, &index, &buffer)
        }
    }

    /// Unmarshals an optional value. Returns `nil` if the boolean prefix is false.
    public static func unmarshalPointer<T>(at index: inout Int, from buffer: [UInt8], unmarshaler: (inout Int, [UInt8]) throws -> T) throws -> T? {
        var cursor = index
        let hasValue = try unmarshalBool(at: &cursor, from: buffer)
        if hasValue {
            let value = try unmarshaler(&cursor, buffer)
            index = cursor
            return value
        }
        index = cursor
        return nil
    }

    /// Skips over a marshalled optional value.
    public static func skipPointer(at index: inout Int, in buffer: [UInt8], skipElement: (inout Int, [UInt8]) throws -> Void) throws {
        var cursor = index
        let hasValue = try unmarshalBool(at: &cursor, from: buffer)
        if hasValue {
            try skipElement(&cursor, buffer)
        }
        index = cursor
    }
}
