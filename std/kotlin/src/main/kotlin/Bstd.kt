package com.banditmoscow1337.bstd

import java.io.IOException
import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.nio.CharBuffer
import java.nio.charset.CodingErrorAction
import java.nio.charset.StandardCharsets
import java.time.Instant

// --- Custom Exceptions ---

/** Base exception for all Bstd library errors. */
open class BstdException(message: String) : IOException(message)

/** Thrown when a varint value exceeds the expected 64-bit range during decoding. */
class BstdOverflowException(message: String = "Varint overflowed a 64-bit integer") : BstdException(message)

/** Thrown when an operation runs out of bytes in the provided buffer. */
class BstdBufferTooSmallException(message: String = "Buffer is too small for the operation") : BstdException(message)

/** Thrown when an unmarshaler of an unexpected functional type is provided. */
class BstdInvalidUnmarshalerException(message: String) : BstdException(message)

/** Thrown when a list or map is missing the expected 4-byte terminator sequence. */
class BstdInvalidTerminatorException(message: String = "Invalid or missing terminator sequence") : BstdException(message)


// --- Functional Type Aliases for Clarity ---

/** A function that calculates the marshalled size of an object of type `T`. */
typealias SizeFunc<T> = (t: T) -> Int

/** A function that marshals an object of type `T` into a [ByteArray]. */
typealias MarshalFunc<T> = (b: ByteArray, n: Int, t: T) -> Int

/** A function that unmarshals an object of type `T` from a [ByteArray]. */
typealias UnmarshalFunc<T> = (b: ByteArray, n: Int) -> Pair<Int, T>

/** A function that skips over a marshalled object in a [ByteArray]. */
typealias SkipFunc = (b: ByteArray, n: Int) -> Int


/**
 * Provides functions for binary serialization (marshalling) and deserialization (unmarshalling).
 * This is a complete Kotlin port of the Go `benc` package, designed for performance and ease of use.
 */
object Bstd {
    private const val MAX_VARINT_LEN_64 = 10 // Max bytes for a 64-bit varint
    private val TERMINATOR_BYTES = byteArrayOf(1, 1, 1, 1)
    private const val MAX_DECODE_BYTES = 64 * 1024 * 1024
    private const val MAX_DECODE_ELEMENTS = 1_048_576

    // --- Private Helpers ---

    /** ZigZag encodes a [Long] to a [ULong] for efficient varint storage of signed numbers. */
    private fun Long.encodeZigZag(): ULong = ((this shl 1) xor (this shr 63)).toULong()

    /** Decodes a ZigZag-encoded [ULong] back to a [Long]. */
    private fun ULong.decodeZigZag(): Long = ((this shr 1).toLong()) xor (-(this and 1u).toLong())

    /** Checks if the buffer `b` has at least `needed` bytes remaining from offset `n`. */
    @JvmStatic
    @Suppress("NOTHING_TO_INLINE")
    private inline fun checkBounds(b: ByteArray, n: Int, needed: Int) {
        if (n < 0 || n > b.size || needed < 0 || b.size - n < needed) {
            throw BstdBufferTooSmallException()
        }
    }

    private fun checkedLength(b: ByteArray, n: Int, length: ULong): Int {
        if (length > MAX_DECODE_BYTES.toULong()) throw BstdOverflowException()
        val size = length.toInt()
        checkBounds(b, n, size)
        return size
    }

    private fun checkedCount(b: ByteArray, n: Int, count: ULong, minimum: Int = 1): Int {
        if (count > MAX_DECODE_ELEMENTS.toULong()) throw BstdOverflowException()
        checkBounds(b, n, 0)
        if (count > ((b.size - n) / minimum).toULong()) throw BstdBufferTooSmallException()
        return count.toInt()
    }

    private fun checkTerminator(b: ByteArray, n: Int): Int {
        checkBounds(b, n, TERMINATOR_BYTES.size)
        if (!b.copyOfRange(n, n + TERMINATOR_BYTES.size).contentEquals(TERMINATOR_BYTES))
            throw BstdInvalidTerminatorException()
        return n + TERMINATOR_BYTES.size
    }

    private fun decodeUtf8(b: ByteArray, n: Int, length: Int): String {
        try {
            return StandardCharsets.UTF_8.newDecoder()
                .onMalformedInput(CodingErrorAction.REPORT)
                .onUnmappableCharacter(CodingErrorAction.REPORT)
                .decode(ByteBuffer.wrap(b, n, length)).toString()
        } catch (ex: java.nio.charset.CharacterCodingException) {
            throw BstdException("invalid UTF-8")
        }
    }

    private fun encodeUtf8(value: String): ByteArray {
        try {
            val bytes = StandardCharsets.UTF_8.newEncoder()
                .onMalformedInput(CodingErrorAction.REPORT)
                .onUnmappableCharacter(CodingErrorAction.REPORT)
                .encode(CharBuffer.wrap(value))
            val result = ByteArray(bytes.remaining())
            bytes.get(result)
            return result
        } catch (ex: java.nio.charset.CharacterCodingException) {
            throw BstdException("invalid UTF-16 string")
        }
    }

    /** Creates a little-endian ByteBuffer for a specific portion of the main buffer. */
    @JvmStatic
    private fun bufferFor(b: ByteArray, n: Int, size: Int): ByteBuffer {
        checkBounds(b, n, size)
        return ByteBuffer.wrap(b, n, size).order(ByteOrder.LITTLE_ENDIAN)
    }

    // --- String ---

    /** Returns the number of bytes needed to marshal a string. */
    @JvmStatic
    fun sizeString(str: String): Int {
        val byteLength = encodeUtf8(str).size
        return byteLength + sizeULongAsVarint(byteLength.toULong())
    }

    /** Marshals a string into `b` at offset `n`. */
    @JvmStatic
    fun marshalString(b: ByteArray, n: Int, str: String): Int {
        val strBytes = encodeUtf8(str)
        checkBounds(b, n, strBytes.size)
        checkBounds(b, n, strBytes.size + sizeULongAsVarint(strBytes.size.toULong()))
        val newN = marshalULongAsVarint(b, n, strBytes.size.toULong())
        checkBounds(b, newN, strBytes.size)
        strBytes.copyInto(b, newN)
        return newN + strBytes.size
    }

    /** Unmarshals a string from `b` at offset `n`. Returns the new offset and the string. */
    @JvmStatic
    fun unmarshalString(b: ByteArray, n: Int): Pair<Int, String> {
        val (newN, len) = unmarshalULongAsVarint(b, n)
        val strLen = checkedLength(b, newN, len)
        val str = decodeUtf8(b, newN, strLen)
        return (newN + strLen) to str
    }

    /** Skips a marshalled string in `b` starting at offset `n`. */
    @JvmStatic
    fun skipString(b: ByteArray, n: Int): Int {
        val (newN, len) = unmarshalULongAsVarint(b, n)
        val strLen = checkedLength(b, newN, len)
        decodeUtf8(b, newN, strLen)
        return newN + strLen
    }

    // --- List (Slice) ---

    /** Skips a marshalled list in `b`. Requires a function to skip each element. */
    @JvmStatic
    fun skipSlice(b: ByteArray, n: Int, skipElement: SkipFunc): Int {
        val (countN, elementCount) = unmarshalULongAsVarint(b, n)
        val count = checkedCount(b, countN, elementCount)
        var currentN = countN
        for (i in 0 until count) {
            currentN = skipElement(b, currentN)
        }
        return checkTerminator(b, currentN)
    }

    /** Returns the bytes needed to marshal a list with dynamically-sized elements. */
    @JvmStatic
    fun <T> sizeSlice(slice: List<T>, sizer: SizeFunc<T>): Int {
        val len = slice.size.toULong()
        return TERMINATOR_BYTES.size + sizeULongAsVarint(len) + slice.sumOf(sizer)
    }

    /** Returns the bytes needed to marshal a list with fixed-size elements. */
    @JvmStatic
    fun <T> sizeFixedSlice(slice: List<T>, elementSize: Int): Int {
        val len = slice.size
        return TERMINATOR_BYTES.size + sizeULongAsVarint(len.toULong()) + (len * elementSize)
    }

    /** Marshals a list into `b`. */
    @JvmStatic
    fun <T> marshalSlice(b: ByteArray, n: Int, slice: List<T>, marshaler: MarshalFunc<T>): Int {
        var currentN = marshalULongAsVarint(b, n, slice.size.toULong())
        slice.forEach { element ->
            currentN = marshaler(b, currentN, element)
        }
        checkBounds(b, currentN, TERMINATOR_BYTES.size)
        TERMINATOR_BYTES.copyInto(b, currentN)
        return currentN + TERMINATOR_BYTES.size
    }

    /**
     * Unmarshals a list from `b`.
     */
    @JvmStatic
    fun <T> unmarshalSlice(b: ByteArray, n: Int, unmarshaler: UnmarshalFunc<T>): Pair<Int, List<T>> {
        val (countN, elementCount) = unmarshalULongAsVarint(b, n)
        val s = checkedCount(b, countN, elementCount)
        var currentN = countN
        val resultList = ArrayList<T>(s)
        for (i in 0 until s) {
            val (nextN, element) = unmarshaler(b, currentN)
            resultList.add(element)
            currentN = nextN
        }
        return checkTerminator(b, currentN) to resultList
    }

    // --- Map ---

    /** Skips a marshalled map in `b`. Requires functions to skip keys and values. */
    @JvmStatic
    fun skipMap(b: ByteArray, n: Int, skipKey: SkipFunc, skipValue: SkipFunc): Int {
        val (countN, pairCount) = unmarshalULongAsVarint(b, n)
        val count = checkedCount(b, countN, pairCount, 2)
        var currentN = countN
        for (i in 0 until count) {
            currentN = skipKey(b, currentN)
            currentN = skipValue(b, currentN)
        }
        return checkTerminator(b, currentN)
    }

    /** Returns the bytes needed to marshal a map. */
    @JvmStatic
    fun <K, V> sizeMap(map: Map<K, V>, keySizer: SizeFunc<K>, valueSizer: SizeFunc<V>): Int {
        var totalSize = TERMINATOR_BYTES.size + sizeULongAsVarint(map.size.toULong())
        for ((key, value) in map) {
            totalSize += keySizer(key)
            totalSize += valueSizer(value)
        }
        return totalSize
    }

    /** Marshals a map into `b`. */
    @JvmStatic
    fun <K, V> marshalMap(b: ByteArray, n: Int, map: Map<K, V>, keyMarshaler: MarshalFunc<K>, valueMarshaler: MarshalFunc<V>): Int {
        var currentN = marshalULongAsVarint(b, n, map.size.toULong())
        for ((key, value) in map) {
            currentN = keyMarshaler(b, currentN, key)
            currentN = valueMarshaler(b, currentN, value)
        }
        checkBounds(b, currentN, TERMINATOR_BYTES.size)
        TERMINATOR_BYTES.copyInto(b, currentN)
        return currentN + TERMINATOR_BYTES.size
    }

    /**
     * Unmarshals a map from `b`.
     *
     * **Note**: This function returns a [LinkedHashMap] to provide predictable iteration order,
     * a feature not guaranteed by standard Maps or by the original Go implementation.
     *
     */
    @JvmStatic
    fun <K, V> unmarshalMap(b: ByteArray, n: Int, keyUnmarshaler: UnmarshalFunc<K>, valueUnmarshaler: UnmarshalFunc<V>): Pair<Int, Map<K, V>> {
        val (countN, pairCount) = unmarshalULongAsVarint(b, n)
        val s = checkedCount(b, countN, pairCount, 2)
        var currentN = countN
        val resultMap = LinkedHashMap<K, V>(s)
        for (i in 0 until s) {
            val (keyN, key) = keyUnmarshaler(b, currentN)
            val (valueN, value) = valueUnmarshaler(b, keyN)
            if (resultMap.containsKey(key)) throw BstdException("duplicate map key")
            resultMap[key] = value
            currentN = valueN
        }
        return checkTerminator(b, currentN) to resultMap
    }

    // --- Byte & ByteArray ---

    /** Skips a single byte. */
    @JvmStatic
    fun skipByte(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 1); return n + 1 }

    /** Returns the size of a marshalled byte (always 1). */
    @JvmStatic
    fun sizeByte(): Int = 1

    /** Marshals a single byte. */
    @JvmStatic
    fun marshalByte(b: ByteArray, n: Int, value: Byte): Int {
        checkBounds(b, n, 1)
        b[n] = value
        return n + 1
    }

    /** Unmarshals a single byte. */
    @JvmStatic
    fun unmarshalByte(b: ByteArray, n: Int): Pair<Int, Byte> {
        checkBounds(b, n, 1)
        return (n + 1) to b[n]
    }

    /** Skips a marshalled byte array. */
    @JvmStatic
    fun skipBytes(b: ByteArray, n: Int): Int {
        val (offset, len) = unmarshalULongAsVarint(b, n)
        return offset + checkedLength(b, offset, len)
    }

    /** Returns the bytes needed to marshal a byte array. */
    @JvmStatic
    fun sizeBytes(bs: ByteArray): Int = bs.size + sizeULongAsVarint(bs.size.toULong())

    /** Marshals a byte array into `b`. */
    @JvmStatic
    fun marshalBytes(b: ByteArray, n: Int, bs: ByteArray): Int {
        checkBounds(b, n, bs.size)
        checkBounds(b, n, bs.size + sizeULongAsVarint(bs.size.toULong()))
        val newN = marshalULongAsVarint(b, n, bs.size.toULong())
        checkBounds(b, newN, bs.size)
        bs.copyInto(b, newN)
        return newN + bs.size
    }

    /**
     * Unmarshals a byte array from `b`.
     * This method always returns a new, copied byte array to ensure data integrity.
     */
    @JvmStatic
    fun unmarshalBytes(b: ByteArray, n: Int): Pair<Int, ByteArray> {
        val (newN, len) = unmarshalULongAsVarint(b, n)
        val bytesLen = checkedLength(b, newN, len)
        return (newN + bytesLen) to b.copyOfRange(newN, newN + bytesLen)
    }

    // --- Varint Signed (Long) ---

    /** Skips a varint in `b`. */
    @JvmStatic
    fun skipVarint(b: ByteArray, n: Int): Int {
        for (i in 0 until MAX_VARINT_LEN_64) {
            val index = n + i
            checkBounds(b, index, 1)
            if (b[index] >= 0) { // Check if MSB is 0
                if (i == MAX_VARINT_LEN_64 - 1 && b[index] > 1) throw BstdOverflowException()
                return n + i + 1
            }
        }
        throw BstdOverflowException()
    }

    /** Returns the bytes needed to marshal a [Long] as a ZigZag-encoded varint. */
    @JvmStatic
    fun sizeLongAsVarint(v: Long): Int = sizeULongAsVarint(v.encodeZigZag())

    /** Marshals a [Long] as a ZigZag-encoded varint. */
    @JvmStatic
    fun marshalLongAsVarint(b: ByteArray, n: Int, v: Long): Int = marshalULongAsVarint(b, n, v.encodeZigZag())

    /** Unmarshals a ZigZag-encoded varint into a [Long]. */
    @JvmStatic
    fun unmarshalLongAsVarint(b: ByteArray, n: Int): Pair<Int, Long> {
        val (newN, ulongVal) = unmarshalULongAsVarint(b, n)
        return newN to ulongVal.decodeZigZag()
    }

    // --- Varint Unsigned (ULong) ---

    /**
     * Returns the bytes needed to marshal a [ULong] as a varint.
     *
     * **Note**: [ULong] is used here to maintain direct binary compatibility with Go's `uint` type,
     * which is 64-bit on target platforms and does not use ZigZag encoding.
     */
    @JvmStatic
    fun sizeULongAsVarint(v: ULong): Int {
        return when {
            v < 1uL shl 7 -> 1
            v < 1uL shl 14 -> 2
            v < 1uL shl 21 -> 3
            v < 1uL shl 28 -> 4
            v < 1uL shl 35 -> 5
            v < 1uL shl 42 -> 6
            v < 1uL shl 49 -> 7
            v < 1uL shl 56 -> 8
            v < 1uL shl 63 -> 9
            else -> 10
        }
    }

    /** Marshals a [ULong] as a varint. */
    @JvmStatic
    fun marshalULongAsVarint(b: ByteArray, n: Int, v: ULong): Int {
        checkBounds(b, n, sizeULongAsVarint(v))
        var value = v
        var i = n
        while (value >= 0x80u) {
            checkBounds(b, i, 1)
            b[i] = (value.toByte().toInt() or 0x80).toByte()
            value = value shr 7
            i++
        }
        checkBounds(b, i, 1)
        b[i] = value.toByte()
        return i + 1
    }

    /** Unmarshals a varint into a [ULong]. */
    @JvmStatic
    fun unmarshalULongAsVarint(b: ByteArray, n: Int): Pair<Int, ULong> {
        var result = 0uL
        var shift = 0
        for (i in 0 until MAX_VARINT_LEN_64) {
            val index = n + i
            checkBounds(b, index, 1)
            val byte = b[index]
            result = result or ((byte.toLong() and 0x7F).toULong() shl shift)
            if (byte >= 0) { // MSB is 0, this is the last byte
                if (i == MAX_VARINT_LEN_64 - 1 && byte > 1) throw BstdOverflowException()
                return (n + i + 1) to result
            }
            shift += 7
        }
        throw BstdOverflowException()
    }

    // --- Fixed-size Primitives ---

    // Int64 / Long
    @JvmStatic fun skipInt64(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 8); return n + 8 }
    @JvmStatic fun sizeInt64(): Int = 8
    @JvmStatic fun marshalInt64(b: ByteArray, n: Int, v: Long): Int { bufferFor(b, n, 8).putLong(v); return n + 8 }
    @JvmStatic fun unmarshalInt64(b: ByteArray, n: Int): Pair<Int, Long> = (n + 8) to bufferFor(b, n, 8).long

    // UInt64 / ULong
    @JvmStatic fun skipUInt64(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 8); return n + 8 }
    @JvmStatic fun sizeUInt64(): Int = 8
    @JvmStatic fun marshalUInt64(b: ByteArray, n: Int, v: ULong): Int { bufferFor(b, n, 8).putLong(v.toLong()); return n + 8 }
    @JvmStatic fun unmarshalUInt64(b: ByteArray, n: Int): Pair<Int, ULong> = (n + 8) to bufferFor(b, n, 8).long.toULong()

    // Int32 / Int
    @JvmStatic fun skipInt32(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 4); return n + 4 }
    @JvmStatic fun sizeInt32(): Int = 4
    @JvmStatic fun marshalInt32(b: ByteArray, n: Int, v: Int): Int { bufferFor(b, n, 4).putInt(v); return n + 4 }
    @JvmStatic fun unmarshalInt32(b: ByteArray, n: Int): Pair<Int, Int> = (n + 4) to bufferFor(b, n, 4).int

    // UInt32 / UInt
    @JvmStatic fun skipUInt32(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 4); return n + 4 }
    @JvmStatic fun sizeUInt32(): Int = 4
    @JvmStatic fun marshalUInt32(b: ByteArray, n: Int, v: UInt): Int { bufferFor(b, n, 4).putInt(v.toInt()); return n + 4 }
    @JvmStatic fun unmarshalUInt32(b: ByteArray, n: Int): Pair<Int, UInt> = (n + 4) to bufferFor(b, n, 4).int.toUInt()

    // Int16 / Short
    @JvmStatic fun skipInt16(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 2); return n + 2 }
    @JvmStatic fun sizeInt16(): Int = 2
    @JvmStatic fun marshalInt16(b: ByteArray, n: Int, v: Short): Int { bufferFor(b, n, 2).putShort(v); return n + 2 }
    @JvmStatic fun unmarshalInt16(b: ByteArray, n: Int): Pair<Int, Short> = (n + 2) to bufferFor(b, n, 2).short

    // UInt16 / UShort
    @JvmStatic fun skipUInt16(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 2); return n + 2 }
    @JvmStatic fun sizeUInt16(): Int = 2
    @JvmStatic fun marshalUInt16(b: ByteArray, n: Int, v: UShort): Int { bufferFor(b, n, 2).putShort(v.toShort()); return n + 2 }
    @JvmStatic fun unmarshalUInt16(b: ByteArray, n: Int): Pair<Int, UShort> = (n + 2) to bufferFor(b, n, 2).short.toUShort()

    // Int8 / Byte
    @JvmStatic fun skipInt8(b: ByteArray, n: Int): Int = skipByte(b, n)
    @JvmStatic fun sizeInt8(): Int = sizeByte()
    @JvmStatic fun marshalInt8(b: ByteArray, n: Int, v: Byte): Int = marshalByte(b, n, v)
    @JvmStatic fun unmarshalInt8(b: ByteArray, n: Int): Pair<Int, Byte> = unmarshalByte(b, n)

    // Float64 / Double
    @JvmStatic fun skipFloat64(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 8); return n + 8 }
    @JvmStatic fun sizeFloat64(): Int = 8
    @JvmStatic fun marshalFloat64(b: ByteArray, n: Int, v: Double): Int { bufferFor(b, n, 8).putDouble(v); return n + 8 }
    @JvmStatic fun unmarshalFloat64(b: ByteArray, n: Int): Pair<Int, Double> = (n + 8) to bufferFor(b, n, 8).double

    // Float32 / Float
    @JvmStatic fun skipFloat32(_b: ByteArray, n: Int): Int { checkBounds(_b, n, 4); return n + 4 }
    @JvmStatic fun sizeFloat32(): Int = 4
    @JvmStatic fun marshalFloat32(b: ByteArray, n: Int, v: Float): Int { bufferFor(b, n, 4).putFloat(v); return n + 4 }
    @JvmStatic fun unmarshalFloat32(b: ByteArray, n: Int): Pair<Int, Float> = (n + 4) to bufferFor(b, n, 4).float

    // --- Boolean ---

    /** Skips a boolean value (1 byte). */
    @JvmStatic
    fun skipBoolean(_b: ByteArray, n: Int): Int = unmarshalBoolean(_b, n).first

    /** Returns the size of a marshalled boolean (always 1). */
    @JvmStatic
    fun sizeBoolean(): Int = 1

    /** Marshals a boolean value. */
    @JvmStatic
    fun marshalBoolean(b: ByteArray, n: Int, v: Boolean): Int {
        checkBounds(b, n, 1)
        b[n] = if (v) 1 else 0
        return n + 1
    }

    /** Unmarshals a boolean value. */
    @JvmStatic
    fun unmarshalBoolean(b: ByteArray, n: Int): Pair<Int, Boolean> {
        checkBounds(b, n, 1)
        if (b[n] != 0.toByte() && b[n] != 1.toByte()) throw BstdException("invalid boolean")
        return (n + 1) to (b[n].toInt() == 1)
    }

    // --- Time (Instant) ---
    // Go version uses UnixNano as int64. We use java.time.Instant.

    /** Skips a marshalled time value (8 bytes). */
    @JvmStatic
    fun skipTime(b: ByteArray, n: Int): Int = skipInt64(b, n)

    /** Returns the size of a marshalled time value (always 8). */
    @JvmStatic
    fun sizeTime(): Int = 8

    /** Marshals an [Instant] as its nanosecond epoch value. */
    @JvmStatic
    fun marshalTime(b: ByteArray, n: Int, t: Instant): Int {
        val nano = try {
            Math.addExact(Math.multiplyExact(t.epochSecond, 1_000_000_000L), t.nano.toLong())
        } catch (ex: ArithmeticException) { throw BstdOverflowException("time exceeds signed 64-bit nanoseconds") }
        return marshalInt64(b, n, nano)
    }

    /** Unmarshals an [Instant] from its nanosecond epoch value. */
    @JvmStatic
    fun unmarshalTime(b: ByteArray, n: Int): Pair<Int, Instant> {
        val (newN, nano) = unmarshalInt64(b, n)
        val sec = nano / 1_000_000_000
        val nanoOfSec = (nano % 1_000_000_000).toInt()
        return newN to Instant.ofEpochSecond(sec, nanoOfSec.toLong())
    }

    // --- Pointer (Nullable) ---
    // In Kotlin, Go pointers are represented as nullable types (`T?`).

    /** Skips a marshalled nullable field. */
    @JvmStatic
    fun skipPointer(b: ByteArray, n: Int, skipElement: SkipFunc): Int {
        val (newN, hasValue) = unmarshalBoolean(b, n)
        return if (hasValue) skipElement(b, newN) else newN
    }

    /** Returns the bytes needed to marshal a nullable value. */
    @JvmStatic
    fun <T : Any> sizePointer(v: T?, sizeFn: SizeFunc<T>): Int =
        sizeBoolean() + (if (v != null) sizeFn(v) else 0)


    /** Marshals a nullable value, prefixed with a boolean flag. */
    @JvmStatic
    fun <T : Any> marshalPointer(b: ByteArray, n: Int, v: T?, marshalFn: MarshalFunc<T>): Int {
        val newN = marshalBoolean(b, n, v != null)
        return if (v != null) marshalFn(b, newN, v) else newN
    }

    /** Unmarshals a nullable value. Returns null if the value was marshalled as nil. */
    @JvmStatic
    fun <T : Any> unmarshalPointer(b: ByteArray, n: Int, unmarshaler: UnmarshalFunc<T>): Pair<Int, T?> {
        val (newN, hasValue) = unmarshalBoolean(b, n)
        return if (hasValue) unmarshaler(b, newN) else newN to null
    }
}
