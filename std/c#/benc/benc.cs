using System;
using System.Buffers.Binary;
using System.Collections.Generic;
using System.Diagnostics.CodeAnalysis;
using System.Runtime.CompilerServices;
using System.Text;

namespace Bstd
{
    // Custom delegates to support `ref` and `in` parameters, which Func/Action do not.
    public delegate T ReadElementDelegate<T>(ref ReadOnlySpan<byte> buffer);
    public delegate void WriteElementDelegate<T>(ref Span<byte> buffer, T item);
    public delegate void SkipElementDelegate(ref ReadOnlySpan<byte> buffer);
    public delegate int GetSizeDelegate<T>(T item);


    /// <summary>
    /// Base exception for bstd serialization errors.
    /// </summary>
    public class BstdException : Exception
    {
        public BstdException(string message) : base(message) { }
    }

    /// <summary>
    /// Thrown when a buffer is too small for a serialization or deserialization operation.
    /// Corresponds to `benc.ErrBufTooSmall` in the Go library.
    /// </summary>
    public class BstdBufferTooSmallException : BstdException
    {
        public BstdBufferTooSmallException(string message = "Buffer is too small for the requested operation.") : base(message) { }
    }

    /// <summary>
    /// Thrown when a varint is malformed, exceeds its maximum size, or overflows the target type.
    /// Corresponds to `benc.ErrOverflow` in the Go library.
    /// </summary>
    public class BstdOverflowException : BstdException
    {
        public BstdOverflowException(string message = "Varint data is corrupted, overflows, or exceeds maximum length.") : base(message) { }
    }

    /// <summary>
    /// Provides high-performance, allocation-minimized methods for serializing and deserializing data types
    /// into a binary format compatible with the Go 'bstd' package.
    /// This implementation uses Span for memory efficiency and safety, and passes spans by `ref` to
    /// advance the buffer position after each read/write operation.
    /// </summary>
    public static class BstdSerializer
    {
        private static ReadOnlySpan<byte> Terminator => new byte[] { 1, 1, 1, 1 };
        
        // MaxVarintLen is the maximum number of bytes a varint can occupy.
        // For 64-bit values, this is 10 bytes.
        private static int MaxVarintLen => IntPtr.Size == 4 ? 5 : 10;
        private static byte MaxLastVarintByte => IntPtr.Size == 4 ? (byte)15 : (byte)1;
        private const int MaxDecodeBytes = 64 * 1024 * 1024;
        private const int MaxDecodeElements = 1_048_576;
        private static readonly UTF8Encoding StrictUtf8 = new UTF8Encoding(false, true);

        private static int CheckedLength(ref ReadOnlySpan<byte> buffer)
        {
            var cursor = buffer;
            nuint raw = ReadVarint(ref cursor);
            if (raw > (nuint)MaxDecodeBytes) throw new BstdOverflowException();
            int length = (int)raw;
            if (length > cursor.Length) throw new BstdBufferTooSmallException();
            buffer = cursor;
            return length;
        }

        private static int CheckedCount(ref ReadOnlySpan<byte> buffer, int minBytes)
        {
            var cursor = buffer;
            nuint raw = ReadVarint(ref cursor);
            if (raw > (nuint)MaxDecodeElements) throw new BstdOverflowException();
            if (raw > (nuint)(cursor.Length / minBytes)) throw new BstdBufferTooSmallException();
            buffer = cursor;
            return (int)raw;
        }

        private static void CheckTerminator(ref ReadOnlySpan<byte> buffer)
        {
            if (buffer.Length < Terminator.Length) throw new BstdBufferTooSmallException();
            if (!buffer.Slice(0, Terminator.Length).SequenceEqual(Terminator)) throw new BstdException("Invalid terminator");
            buffer = buffer.Slice(Terminator.Length);
        }

        #region String

        /// <summary>
        /// Calculates the number of bytes required to serialize a string, including its length prefix.
        /// </summary>
        public static int GetSizeString(string value)
        {
            if (string.IsNullOrEmpty(value)) return GetSizeVarint((nuint)0);
            int byteCount = StrictUtf8.GetByteCount(value);
            return GetSizeVarint((nuint)byteCount) + byteCount;
        }

        /// <summary>
        /// Writes a string to the buffer, prefixed with its UTF-8 byte length as a varint.
        /// </summary>
        public static void WriteString(ref Span<byte> buffer, string value)
        {
            int byteCount = string.IsNullOrEmpty(value) ? 0 : StrictUtf8.GetByteCount(value);
            if (GetSizeVarint((nuint)byteCount) > buffer.Length - byteCount)
                throw new BstdBufferTooSmallException();
            WriteVarint(ref buffer, (nuint)byteCount);
            if (byteCount > 0)
            {
                StrictUtf8.GetBytes(value, buffer);
                buffer = buffer.Slice(byteCount);
            }
        }
        
        /// <summary>
        /// Reads a string from the buffer.
        /// </summary>
        public static string ReadString(ref ReadOnlySpan<byte> buffer)
        {
            var cursor = buffer;
            var byteCount = CheckedLength(ref cursor);
            string value;
            try { value = StrictUtf8.GetString(cursor.Slice(0, byteCount)); }
            catch (DecoderFallbackException) { throw new BstdException("Invalid UTF-8"); }
            buffer = cursor.Slice(byteCount);
            return value;
        }

        /// <summary>
        /// Advances the buffer past a serialized string without allocating the string.
        /// </summary>
        public static void SkipString(ref ReadOnlySpan<byte> buffer)
        {
            var cursor = buffer;
            var length = CheckedLength(ref cursor);
            try { StrictUtf8.GetCharCount(cursor.Slice(0, length)); }
            catch (DecoderFallbackException) { throw new BstdException("Invalid UTF-8"); }
            buffer = cursor.Slice(length);
        }

        #endregion
        
        #region Unsafe String (API Parity)

        /// <summary>
        /// Writes a string to the buffer. This method is an alias for WriteString for API compatibility with the Go version's `MarshalUnsafeString`. In C#, the standard `WriteString` is already highly optimized.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public static void WriteStringUnsafe(ref Span<byte> buffer, string value) => WriteString(ref buffer, value);

        /// <summary>
        /// Reads a string from the buffer. This method is an alias for ReadString for API compatibility with the Go version's `UnmarshalUnsafeString`.
        /// </summary>
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public static string ReadStringUnsafe(ref ReadOnlySpan<byte> buffer) => ReadString(ref buffer);
        
        #endregion

        #region Byte Array

        /// <summary>
        /// Calculates the number of bytes required to serialize a byte array.
        /// </summary>
        public static int GetSizeBytes(ReadOnlySpan<byte> value)
        {
            return GetSizeVarint((nuint)value.Length) + value.Length;
        }

        /// <summary>
        /// Writes a byte array to the buffer, prefixed with its length as a varint.
        /// </summary>
        public static void WriteBytes(ref Span<byte> buffer, ReadOnlySpan<byte> value)
        {
            if (GetSizeVarint((nuint)value.Length) > buffer.Length - value.Length) throw new BstdBufferTooSmallException();
            WriteVarint(ref buffer, (nuint)value.Length);
            if (value.Length > 0)
            {
                value.CopyTo(buffer);
                buffer = buffer.Slice(value.Length);
            }
        }

        /// <summary>
        /// Reads a byte array from the buffer, allocating a new array and copying the data.
        /// Slower but safer if the source buffer will be modified.
        /// </summary>
        public static byte[] ReadBytesCopied(ref ReadOnlySpan<byte> buffer)
        {
            var length = CheckedLength(ref buffer);
            if (length > buffer.Length) throw new BstdBufferTooSmallException();
            var value = buffer.Slice(0, length).ToArray();
            buffer = buffer.Slice(length);
            return value;
        }

        /// <summary>
        /// Reads a byte array from the buffer by returning a slice of the original buffer.
        /// Faster (zero-allocation) but unsafe if the source buffer is modified.
        /// </summary>
        public static ReadOnlySpan<byte> ReadBytesCropped(ref ReadOnlySpan<byte> buffer)
        {
            var length = CheckedLength(ref buffer);
            if (length > buffer.Length) throw new BstdBufferTooSmallException();
            var value = buffer.Slice(0, length);
            buffer = buffer.Slice(length);
            return value;
        }

        /// <summary>
        /// Advances the buffer past a serialized byte array.
        /// </summary>
        public static void SkipBytes(ref ReadOnlySpan<byte> buffer)
        {
            var cursor = buffer;
            var length = CheckedLength(ref cursor);
            buffer = cursor.Slice(length);
        }

        #endregion

        #region Varint (Signed and Unsigned)

        // Native-sized (nint/nuint), equivalent to Go's int/uint
        public static int GetSizeVarint(nint value) => GetSizeVarint(EncodeZigZag(value));
        public static void WriteVarint(ref Span<byte> buffer, nint value) => WriteVarint(ref buffer, EncodeZigZag(value));
        public static nint ReadSignedVarint(ref ReadOnlySpan<byte> buffer) => DecodeZigZag(ReadVarint(ref buffer));
        
        /// <summary>
        /// Calculates the number of bytes required to serialize an unsigned native integer as a varint.
        /// </summary>
        public static int GetSizeVarint(nuint value)
        {
            int i = 1;
            while (value >= 0x80)
            {
                value >>= 7;
                i++;
            }
            return i;
        }

        /// <summary>
        /// Writes an unsigned native integer to the buffer as a varint.
        /// </summary>
        public static void WriteVarint(ref Span<byte> buffer, nuint value)
        {
            if (buffer.Length < GetSizeVarint(value)) throw new BstdBufferTooSmallException();
            while (value >= 0x80)
            {
                buffer[0] = (byte)(value | 0x80);
                value >>= 7;
                buffer = buffer.Slice(1);
            }
            buffer[0] = (byte)value;
            buffer = buffer.Slice(1);
        }

        /// <summary>
        /// Reads an unsigned native integer varint from the buffer.
        /// </summary>
        public static nuint ReadVarint(ref ReadOnlySpan<byte> buffer)
        {
            var cursor = buffer;
            nuint value = 0;
            int shift = 0;
            for (int i = 0; i < MaxVarintLen; i++)
            {
                if (cursor.IsEmpty) throw new BstdBufferTooSmallException();

                byte b = cursor[0];
                cursor = cursor.Slice(1);
                
                // Overflow check from Go lib: `if i == maxVarintLen-1 && b > 1`
                if (i == MaxVarintLen - 1 && b > MaxLastVarintByte) throw new BstdOverflowException();

                value |= (nuint)(b & 0x7F) << shift;
                if (b < 0x80) { buffer = cursor; return value; }
                
                shift += 7;
            }
            throw new BstdOverflowException();
        }
        
        /// <summary>
        /// Advances the buffer past a varint without decoding it.
        /// </summary>
        public static void SkipVarint(ref ReadOnlySpan<byte> buffer)
        {
             var cursor = buffer;
             for (int i = 0; i < MaxVarintLen; ++i)
             {
                if (cursor.IsEmpty) throw new BstdBufferTooSmallException();
                byte b = cursor[0];
                cursor = cursor.Slice(1);
                if (b < 0x80)
                {
                    if (i == MaxVarintLen - 1 && b > MaxLastVarintByte) throw new BstdOverflowException();
                    buffer = cursor;
                    return;
                }
            }
            throw new BstdOverflowException();
        }

        #endregion

        #region Collections (Array/Slice)

        /// <summary>
        /// Calculates the size of a collection where each element has a variable size.
        /// </summary>
        public static int GetSizeArray<T>(ICollection<T> items, GetSizeDelegate<T> sizeElement)
        {
            int size = GetSizeVarint((nuint)items.Count);
            foreach (var item in items) size += sizeElement(item);
            return size + Terminator.Length;
        }
        
        /// <summary>
        /// Calculates the size of a collection where each element has a fixed size.
        /// </summary>
        public static int GetSizeFixedArray<T>(ICollection<T> items, int elementSize)
        {
            return GetSizeVarint((nuint)items.Count) + (items.Count * elementSize) + Terminator.Length;
        }

        /// <summary>
        /// Writes a collection to the buffer.
        /// </summary>
        public static void WriteArray<T>(ref Span<byte> buffer, ICollection<T> items, WriteElementDelegate<T> writeElement)
        {
            WriteVarint(ref buffer, (nuint)items.Count);
            foreach (var item in items) writeElement(ref buffer, item);
            Terminator.CopyTo(buffer);
            buffer = buffer.Slice(Terminator.Length);
        }

        /// <summary>
        /// Reads a collection from the buffer.
        /// </summary>
        public static T[] ReadArray<T>(ref ReadOnlySpan<byte> buffer, ReadElementDelegate<T> readElement)
        {
            var cursor = buffer;
            var count = CheckedCount(ref cursor, 1);
            var items = new T[count];
            for (int i = 0; i < count; i++) items[i] = readElement(ref cursor);
            CheckTerminator(ref cursor);
            buffer = cursor;
            return items;
        }

        /// <summary>
        /// Advances the buffer past a collection.
        /// </summary>
        public static void SkipArray(ref ReadOnlySpan<byte> buffer, SkipElementDelegate skipElement)
        {
            var cursor = buffer;
            var count = CheckedCount(ref cursor, 1);
            for (int i = 0; i < count; i++) skipElement(ref cursor);
            CheckTerminator(ref cursor);
            buffer = cursor;
        }

        #endregion

        #region Collections (Dictionary/Map)

        /// <summary>
        /// Calculates the size of a dictionary.
        /// </summary>
        public static int GetSizeDictionary<TKey, TValue>(IDictionary<TKey, TValue> map, GetSizeDelegate<TKey> sizeKey, GetSizeDelegate<TValue> sizeValue) where TKey : notnull
        {
            int size = GetSizeVarint((nuint)map.Count);
            foreach (var kvp in map)
            {
                size += sizeKey(kvp.Key);
                size += sizeValue(kvp.Value);
            }
            return size + Terminator.Length;
        }

        /// <summary>
        /// Writes a dictionary to the buffer.
        /// </summary>
        public static void WriteDictionary<TKey, TValue>(ref Span<byte> buffer, IDictionary<TKey, TValue> map, WriteElementDelegate<TKey> writeKey, WriteElementDelegate<TValue> writeValue) where TKey : notnull
        {
            WriteVarint(ref buffer, (nuint)map.Count);
            foreach (var kvp in map)
            {
                writeKey(ref buffer, kvp.Key);
                writeValue(ref buffer, kvp.Value);
            }
            Terminator.CopyTo(buffer);
            buffer = buffer.Slice(Terminator.Length);
        }

        /// <summary>
        /// Reads a dictionary from the buffer.
        /// </summary>
        public static Dictionary<TKey, TValue> ReadDictionary<TKey, TValue>(ref ReadOnlySpan<byte> buffer, ReadElementDelegate<TKey> readKey, ReadElementDelegate<TValue> readValue) where TKey : notnull
        {
            var cursor = buffer;
            var count = CheckedCount(ref cursor, 2);
            var map = new Dictionary<TKey, TValue>(count);
            for (int i = 0; i < count; i++)
            {
                var key = readKey(ref cursor);
                var value = readValue(ref cursor);
                if (map.ContainsKey(key)) throw new BstdException("Duplicate map key");
                map.Add(key, value);
            }
            CheckTerminator(ref cursor);
            buffer = cursor;
            return map;
        }
        
        /// <summary>
        /// Advances the buffer past a dictionary.
        /// </summary>
        public static void SkipDictionary(ref ReadOnlySpan<byte> buffer, SkipElementDelegate skipKey, SkipElementDelegate skipValue)
        {
            var cursor = buffer;
            var count = CheckedCount(ref cursor, 2);
            for (int i = 0; i < count; i++)
            {
                skipKey(ref cursor);
                skipValue(ref cursor);
            }
            CheckTerminator(ref cursor);
            buffer = cursor;
        }

        #endregion

        #region Pointer / Nullable types

        /// <summary>
        /// Calculates the size of a nullable reference type.
        /// </summary>
        public static int GetSizePointerClass<T>(T? value, GetSizeDelegate<T> sizeFunc) where T : class
        {
            return sizeof(bool) + (value != null ? sizeFunc(value) : 0);
        }

        /// <summary>
        /// Calculates the size of a nullable value type.
        /// </summary>
        public static int GetSizePointerStruct<T>(T? value, GetSizeDelegate<T> sizeFunc) where T : struct
        {
            return sizeof(bool) + (value.HasValue ? sizeFunc(value.Value) : 0);
        }

        /// <summary>
        /// Writes a nullable reference type to the buffer.
        /// </summary>
        public static void WritePointerClass<T>(ref Span<byte> buffer, T? value, WriteElementDelegate<T> writeFunc) where T : class
        {
            WriteBool(ref buffer, value != null);
            if (value != null) writeFunc(ref buffer, value);
        }
        
        /// <summary>
        /// Writes a nullable value type to the buffer.
        /// </summary>
        public static void WritePointerStruct<T>(ref Span<byte> buffer, T? value, WriteElementDelegate<T> writeFunc) where T : struct
        {
            WriteBool(ref buffer, value.HasValue);
            if (value.HasValue) writeFunc(ref buffer, value.Value);
        }

        /// <summary>
        /// Reads a nullable reference type from the buffer.
        /// </summary>
        [return: MaybeNull]
        public static T ReadPointerClass<T>(ref ReadOnlySpan<byte> buffer, ReadElementDelegate<T> readFunc) where T : class
        {
            if (!ReadBool(ref buffer)) return null;
            return readFunc(ref buffer);
        }

        /// <summary>
        /// Reads a nullable value type from the buffer.
        /// </summary>
        public static T? ReadPointerStruct<T>(ref ReadOnlySpan<byte> buffer, ReadElementDelegate<T> readFunc) where T : struct
        {
            return ReadBool(ref buffer) ? readFunc(ref buffer) : (T?)null;
        }

        /// <summary>
        /// Advances the buffer past a nullable type.
        /// </summary>
        public static void SkipPointer(ref ReadOnlySpan<byte> buffer, SkipElementDelegate skipElement)
        {
            if (ReadBool(ref buffer)) skipElement(ref buffer);
        }

        #endregion

        #region Fixed-Size Primitives & Time

        public static int GetSizeBool() => 1;
        public static void WriteBool(ref Span<byte> buffer, bool value) { buffer[0] = (byte)(value ? 1 : 0); buffer = buffer.Slice(1); }
        public static bool ReadBool(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < 1) throw new BstdBufferTooSmallException(); if (buffer[0] > 1) throw new BstdException("Invalid boolean"); var val = buffer[0] == 1; buffer = buffer.Slice(1); return val; }
        public static void SkipBool(ref ReadOnlySpan<byte> buffer) { ReadBool(ref buffer); }

        public static int GetSizeByte() => 1;
        public static void WriteByte(ref Span<byte> buffer, byte value) { buffer[0] = value; buffer = buffer.Slice(1); }
        public static byte ReadByte(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < 1) throw new BstdBufferTooSmallException(); var val = buffer[0]; buffer = buffer.Slice(1); return val; }
        public static void SkipByte(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < 1) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(1); }
        
        public static int GetSizeSByte() => 1;
        public static void WriteSByte(ref Span<byte> buffer, sbyte value) { buffer[0] = (byte)value; buffer = buffer.Slice(1); }
        public static sbyte ReadSByte(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < 1) throw new BstdBufferTooSmallException(); var val = (sbyte)buffer[0]; buffer = buffer.Slice(1); return val; }
        public static void SkipSByte(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < 1) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(1); }
        
        public static int GetSizeInt16() => sizeof(short);
        public static void WriteInt16(ref Span<byte> buffer, short value) { BinaryPrimitives.WriteInt16LittleEndian(buffer, value); buffer = buffer.Slice(sizeof(short)); }
        public static short ReadInt16(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(short)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadInt16LittleEndian(buffer); buffer = buffer.Slice(sizeof(short)); return val; }
        public static void SkipInt16(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(short)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(short)); }

        public static int GetSizeUInt16() => sizeof(ushort);
        public static void WriteUInt16(ref Span<byte> buffer, ushort value) { BinaryPrimitives.WriteUInt16LittleEndian(buffer, value); buffer = buffer.Slice(sizeof(ushort)); }
        public static ushort ReadUInt16(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(ushort)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadUInt16LittleEndian(buffer); buffer = buffer.Slice(sizeof(ushort)); return val; }
        public static void SkipUInt16(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(ushort)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(ushort)); }

        public static int GetSizeInt32() => sizeof(int);
        public static void WriteInt32(ref Span<byte> buffer, int value) { BinaryPrimitives.WriteInt32LittleEndian(buffer, value); buffer = buffer.Slice(sizeof(int)); }
        public static int ReadInt32(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(int)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadInt32LittleEndian(buffer); buffer = buffer.Slice(sizeof(int)); return val; }
        public static void SkipInt32(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(int)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(int)); }

        public static int GetSizeUInt32() => sizeof(uint);
        public static void WriteUInt32(ref Span<byte> buffer, uint value) { BinaryPrimitives.WriteUInt32LittleEndian(buffer, value); buffer = buffer.Slice(sizeof(uint)); }
        public static uint ReadUInt32(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(uint)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadUInt32LittleEndian(buffer); buffer = buffer.Slice(sizeof(uint)); return val; }
        public static void SkipUInt32(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(uint)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(uint)); }

        public static int GetSizeInt64() => sizeof(long);
        public static void WriteInt64(ref Span<byte> buffer, long value) { BinaryPrimitives.WriteInt64LittleEndian(buffer, value); buffer = buffer.Slice(sizeof(long)); }
        public static long ReadInt64(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(long)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadInt64LittleEndian(buffer); buffer = buffer.Slice(sizeof(long)); return val; }
        public static void SkipInt64(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(long)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(long)); }

        public static int GetSizeUInt64() => sizeof(ulong);
        public static void WriteUInt64(ref Span<byte> buffer, ulong value) { BinaryPrimitives.WriteUInt64LittleEndian(buffer, value); buffer = buffer.Slice(sizeof(ulong)); }
        public static ulong ReadUInt64(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(ulong)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadUInt64LittleEndian(buffer); buffer = buffer.Slice(sizeof(ulong)); return val; }
        public static void SkipUInt64(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(ulong)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(ulong)); }

        public static int GetSizeSingle() => sizeof(float);
        public static void WriteSingle(ref Span<byte> buffer, float value) { BinaryPrimitives.WriteSingleLittleEndian(buffer, value); buffer = buffer.Slice(sizeof(float)); }
        public static float ReadSingle(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(float)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadSingleLittleEndian(buffer); buffer = buffer.Slice(sizeof(float)); return val; }
        public static void SkipSingle(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(float)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(float)); }

        public static int GetSizeDouble() => sizeof(double);
        public static void WriteDouble(ref Span<byte> buffer, double value) { BinaryPrimitives.WriteDoubleLittleEndian(buffer, value); buffer = buffer.Slice(sizeof(double)); }
        public static double ReadDouble(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(double)) throw new BstdBufferTooSmallException(); var val = BinaryPrimitives.ReadDoubleLittleEndian(buffer); buffer = buffer.Slice(sizeof(double)); return val; }
        public static void SkipDouble(ref ReadOnlySpan<byte> buffer) { if(buffer.Length < sizeof(double)) throw new BstdBufferTooSmallException(); buffer = buffer.Slice(sizeof(double)); }
        
        public static int GetSizeTime() => sizeof(long);
        public static void WriteTime(ref Span<byte> buffer, DateTime value)
        {
            long unixNano = checked((value.ToUniversalTime().Ticks - DateTime.UnixEpoch.Ticks) * 100);
            WriteInt64(ref buffer, unixNano);
        }
        public static void WriteTimeNanos(ref Span<byte> buffer, long unixNano) => WriteInt64(ref buffer, unixNano);
        public static long ReadTimeNanos(ref ReadOnlySpan<byte> buffer) => ReadInt64(ref buffer);
        public static DateTime ReadTime(ref ReadOnlySpan<byte> buffer)
        {
            var cursor = buffer;
            long unixNano = ReadTimeNanos(ref cursor);
            if (unixNano % 100 != 0) throw new BstdException("DateTime cannot represent nanoseconds; use ReadTimeNanos");
            var value = DateTime.UnixEpoch.AddTicks(unixNano / 100);
            buffer = cursor;
            return value;
        }
        public static void SkipTime(ref ReadOnlySpan<byte> buffer) => SkipInt64(ref buffer);

        #endregion

        #region Helpers

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static nuint EncodeZigZag(nint value) => (nuint)((value << 1) ^ (value >> (IntPtr.Size * 8 - 1)));
        
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        private static nint DecodeZigZag(nuint value) => (nint)(value >> 1) ^ (-(nint)(value & 1));

        #endregion
    }
}
