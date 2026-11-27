using Microsoft.VisualStudio.TestTools.UnitTesting;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace Bstd.Tests
{
    [TestClass]
    public class BstdSerializerTests
    {
        [TestMethod]
        public void StrictWireBounds()
        {
            var huge = new ReadOnlySpan<byte>(new byte[] { 0x80, 0x80, 0x80, 0x80, 0x10 });
            try { BstdSerializer.ReadString(ref huge); Assert.Fail("expected overflow"); }
            catch (BstdOverflowException) { }
            var wrong = new ReadOnlySpan<byte>(new byte[] { 0, 0, 0, 0, 0 });
            try { BstdSerializer.SkipArray(ref wrong, BstdSerializer.SkipByte); Assert.Fail("expected terminator error"); }
            catch (BstdException) { }
            Assert.AreEqual(5, wrong.Length);
            var invalid = new ReadOnlySpan<byte>(new byte[] { 1, 0xff });
            try { BstdSerializer.SkipString(ref invalid); Assert.Fail("expected UTF-8 error"); }
            catch (BstdException) { }
            Assert.AreEqual(2, invalid.Length);
            var boolBytes = new ReadOnlySpan<byte>(new byte[] { 2 });
            try { BstdSerializer.ReadBool(ref boolBytes); Assert.Fail("expected invalid boolean"); }
            catch (BstdException) { }
            var time = new Span<byte>(new byte[8]);
            BstdSerializer.WriteTimeNanos(ref time, 1);
            var readTime = new ReadOnlySpan<byte>(new byte[] { 1, 0, 0, 0, 0, 0, 0, 0 });
            Assert.AreEqual(1L, BstdSerializer.ReadTimeNanos(ref readTime));
        }

        // Helper to perform a full serialization-deserialization-skip cycle and assert equality.
        private void AssertRoundtrip<T>(
            T value,
            GetSizeDelegate<T> getSize,
            WriteElementDelegate<T> write,
            ReadElementDelegate<T> read,
            SkipElementDelegate skip,
            Action<T, T> assertEqual)
        {
            // 1. Get Size
            int size = getSize(value);
            // Handle zero-size case for things like empty strings or collections
            if (size == 0)
            {
                var zBuffer = new Span<byte>(new byte[0]);
                var zWriteBuffer = zBuffer;
                write(ref zWriteBuffer, value);
                Assert.AreEqual(0, zWriteBuffer.Length);

                var zReadBuffer = (ReadOnlySpan<byte>)zBuffer;
                T zReadValue = read(ref zReadBuffer);
                assertEqual(value, zReadValue);
                return;
            }

            var buffer = new Span<byte>(new byte[size]);

            // 2. Write
            var writeBuffer = buffer;
            write(ref writeBuffer, value);
            Assert.AreEqual(0, writeBuffer.Length, "Buffer not fully used on write.");

            // 3. Skip
            var skipBuffer = (ReadOnlySpan<byte>)buffer;
            skip(ref skipBuffer);
            Assert.AreEqual(0, skipBuffer.Length, "Buffer not fully skipped.");

            // 4. Read
            var readBuffer = (ReadOnlySpan<byte>)buffer;
            T readValue = read(ref readBuffer);
            Assert.AreEqual(0, readBuffer.Length, "Buffer not fully read.");

            // 5. Assert Equality
            assertEqual(value, readValue);
        }

        [TestMethod]
        public void TestMainDataTypes()
        {
            var rand = new Random();
            var testStr = "Hello World!";
            var testBs = new byte[] { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 };
            nint negIntVal = -12345;
            nint posIntVal = nint.MaxValue;
            nuint uintVal = nuint.MaxValue;

            var values = new object[]
            {
                true,
                (byte)128,
                rand.NextSingle(),
                rand.NextDouble(),
                posIntVal,
                negIntVal,
                (sbyte)-1,
                (short)-1,
                (int)rand.Next(),
                (long)rand.NextInt64(),
                uintVal,
                (ushort)160,
                (uint)rand.Next(),
                (ulong)rand.NextInt64(), // Positive long to fit in ulong
                testStr,
                testStr, // For unsafe version
                testBs,  // For cropped
                testBs   // For copied
            };

            // 1. Calculate total size
            int totalSize = BstdSerializer.GetSizeBool() +
                            BstdSerializer.GetSizeByte() +
                            BstdSerializer.GetSizeSingle() +
                            BstdSerializer.GetSizeDouble() +
                            BstdSerializer.GetSizeVarint(posIntVal) +
                            BstdSerializer.GetSizeVarint(negIntVal) +
                            BstdSerializer.GetSizeSByte() +
                            BstdSerializer.GetSizeInt16() +
                            BstdSerializer.GetSizeInt32() +
                            BstdSerializer.GetSizeInt64() +
                            BstdSerializer.GetSizeVarint(uintVal) +
                            BstdSerializer.GetSizeUInt16() +
                            BstdSerializer.GetSizeUInt32() +
                            BstdSerializer.GetSizeUInt64() +
                            BstdSerializer.GetSizeString(testStr) +
                            BstdSerializer.GetSizeString(testStr) +
                            BstdSerializer.GetSizeBytes(testBs) +
                            BstdSerializer.GetSizeBytes(testBs);

            var buffer = new Span<byte>(new byte[totalSize]);
            var writeBuffer = buffer;

            // 2. Write all values
            BstdSerializer.WriteBool(ref writeBuffer, (bool)values[0]);
            BstdSerializer.WriteByte(ref writeBuffer, (byte)values[1]);
            BstdSerializer.WriteSingle(ref writeBuffer, (float)values[2]);
            BstdSerializer.WriteDouble(ref writeBuffer, (double)values[3]);
            BstdSerializer.WriteVarint(ref writeBuffer, (nint)values[4]);
            BstdSerializer.WriteVarint(ref writeBuffer, (nint)values[5]);
            BstdSerializer.WriteSByte(ref writeBuffer, (sbyte)values[6]);
            BstdSerializer.WriteInt16(ref writeBuffer, (short)values[7]);
            BstdSerializer.WriteInt32(ref writeBuffer, (int)values[8]);
            BstdSerializer.WriteInt64(ref writeBuffer, (long)values[9]);
            BstdSerializer.WriteVarint(ref writeBuffer, (nuint)values[10]);
            BstdSerializer.WriteUInt16(ref writeBuffer, (ushort)values[11]);
            BstdSerializer.WriteUInt32(ref writeBuffer, (uint)values[12]);
            BstdSerializer.WriteUInt64(ref writeBuffer, (ulong)values[13]);
            BstdSerializer.WriteString(ref writeBuffer, (string)values[14]);
            BstdSerializer.WriteStringUnsafe(ref writeBuffer, (string)values[15]);
            BstdSerializer.WriteBytes(ref writeBuffer, (byte[])values[16]);
            BstdSerializer.WriteBytes(ref writeBuffer, (byte[])values[17]);
            Assert.AreEqual(0, writeBuffer.Length);

            // 3. Skip all values
            var skipBuffer = (ReadOnlySpan<byte>)buffer;
            BstdSerializer.SkipBool(ref skipBuffer);
            BstdSerializer.SkipByte(ref skipBuffer);
            BstdSerializer.SkipSingle(ref skipBuffer);
            BstdSerializer.SkipDouble(ref skipBuffer);
            BstdSerializer.SkipVarint(ref skipBuffer);
            BstdSerializer.SkipVarint(ref skipBuffer);
            BstdSerializer.SkipSByte(ref skipBuffer);
            BstdSerializer.SkipInt16(ref skipBuffer);
            BstdSerializer.SkipInt32(ref skipBuffer);
            BstdSerializer.SkipInt64(ref skipBuffer);
            BstdSerializer.SkipVarint(ref skipBuffer);
            BstdSerializer.SkipUInt16(ref skipBuffer);
            BstdSerializer.SkipUInt32(ref skipBuffer);
            BstdSerializer.SkipUInt64(ref skipBuffer);
            BstdSerializer.SkipString(ref skipBuffer);
            BstdSerializer.SkipString(ref skipBuffer);
            BstdSerializer.SkipBytes(ref skipBuffer);
            BstdSerializer.SkipBytes(ref skipBuffer);
            Assert.AreEqual(0, skipBuffer.Length);

            // 4. Read and verify all values
            var readBuffer = (ReadOnlySpan<byte>)buffer;
            Assert.AreEqual((bool)values[0], BstdSerializer.ReadBool(ref readBuffer));
            Assert.AreEqual((byte)values[1], BstdSerializer.ReadByte(ref readBuffer));
            Assert.AreEqual((float)values[2], BstdSerializer.ReadSingle(ref readBuffer));
            Assert.AreEqual((double)values[3], BstdSerializer.ReadDouble(ref readBuffer));
            Assert.AreEqual((nint)values[4], BstdSerializer.ReadSignedVarint(ref readBuffer));
            Assert.AreEqual((nint)values[5], BstdSerializer.ReadSignedVarint(ref readBuffer));
            Assert.AreEqual((sbyte)values[6], BstdSerializer.ReadSByte(ref readBuffer));
            Assert.AreEqual((short)values[7], BstdSerializer.ReadInt16(ref readBuffer));
            Assert.AreEqual((int)values[8], BstdSerializer.ReadInt32(ref readBuffer));
            Assert.AreEqual((long)values[9], BstdSerializer.ReadInt64(ref readBuffer));
            Assert.AreEqual((nuint)values[10], BstdSerializer.ReadVarint(ref readBuffer));
            Assert.AreEqual((ushort)values[11], BstdSerializer.ReadUInt16(ref readBuffer));
            Assert.AreEqual((uint)values[12], BstdSerializer.ReadUInt32(ref readBuffer));
            Assert.AreEqual((ulong)values[13], BstdSerializer.ReadUInt64(ref readBuffer));
            Assert.AreEqual((string)values[14], BstdSerializer.ReadString(ref readBuffer));
            Assert.AreEqual((string)values[15], BstdSerializer.ReadStringUnsafe(ref readBuffer));
            CollectionAssert.AreEqual((byte[])values[16], BstdSerializer.ReadBytesCropped(ref readBuffer).ToArray());
            CollectionAssert.AreEqual((byte[])values[17], BstdSerializer.ReadBytesCopied(ref readBuffer));
            Assert.AreEqual(0, readBuffer.Length);
        }

        [TestMethod]
        public void TestErrBufTooSmall()
        {
            Assert.ThrowsExactly<BstdBufferTooSmallException>(() => { var b = new ReadOnlySpan<byte>(new byte[0]); BstdSerializer.ReadBool(ref b); });
            Assert.ThrowsExactly<BstdBufferTooSmallException>(() => { var b = new ReadOnlySpan<byte>(new byte[3]); BstdSerializer.ReadInt32(ref b); });
            Assert.ThrowsExactly<BstdBufferTooSmallException>(() => { var b = new ReadOnlySpan<byte>(new byte[7]); BstdSerializer.ReadInt64(ref b); });
            Assert.ThrowsExactly<BstdBufferTooSmallException>(() => { var b = new ReadOnlySpan<byte>(new byte[] { 2, 0 }); BstdSerializer.ReadString(ref b); });
            Assert.ThrowsExactly<BstdBufferTooSmallException>(() => { var b = new ReadOnlySpan<byte>(new byte[] { 0x80 }); BstdSerializer.ReadVarint(ref b); });
        }
        
        [TestMethod]
        public void TestVarintOverflowAndZigZag()
        {
            var overflowBytes = new byte[] { 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01 };
            var edgeCaseBytes = new byte[] { 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x02 };

            Assert.ThrowsExactly<BstdOverflowException>(() => { var b = new ReadOnlySpan<byte>(overflowBytes); BstdSerializer.ReadVarint(ref b); });
            Assert.ThrowsExactly<BstdOverflowException>(() => { var b = new ReadOnlySpan<byte>(edgeCaseBytes); BstdSerializer.ReadVarint(ref b); });
            Assert.ThrowsExactly<BstdOverflowException>(() => { var b = new ReadOnlySpan<byte>(overflowBytes); BstdSerializer.SkipVarint(ref b); });
            Assert.ThrowsExactly<BstdOverflowException>(() => { var b = new ReadOnlySpan<byte>(edgeCaseBytes); BstdSerializer.SkipVarint(ref b); });

            // Specific ZigZag values from Go test
            AssertRoundtrip((nint)1, BstdSerializer.GetSizeVarint, BstdSerializer.WriteVarint, BstdSerializer.ReadSignedVarint, BstdSerializer.SkipVarint, Assert.AreEqual);
            AssertRoundtrip((nint)(-2), BstdSerializer.GetSizeVarint, BstdSerializer.WriteVarint, BstdSerializer.ReadSignedVarint, BstdSerializer.SkipVarint, Assert.AreEqual);
            AssertRoundtrip((nint)150, BstdSerializer.GetSizeVarint, BstdSerializer.WriteVarint, BstdSerializer.ReadSignedVarint, BstdSerializer.SkipVarint, Assert.AreEqual);
        }

        [TestMethod]
        public void TestSlices()
        {
            var slice = new[] { "sliceelement1", "sliceelement2", "sliceelement3" };
            AssertRoundtrip(
                slice,
                items => BstdSerializer.GetSizeArray(items, BstdSerializer.GetSizeString),
                (ref Span<byte> b, string[] items) => BstdSerializer.WriteArray(ref b, items, BstdSerializer.WriteString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.ReadArray(ref b, BstdSerializer.ReadString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.SkipArray(ref b, BstdSerializer.SkipString),
                CollectionAssert.AreEqual
            );
        }

        [TestMethod]
        public void TestMaps()
        {
            var map = new Dictionary<string, string>
            {
                ["mapkey1"] = "mapvalue1",
                ["mapkey2"] = "mapvalue2",
                ["mapkey3"] = "mapvalue3"
            };

            AssertRoundtrip(
                map,
                m => BstdSerializer.GetSizeDictionary(m, BstdSerializer.GetSizeString, BstdSerializer.GetSizeString),
                (ref Span<byte> b, Dictionary<string,string> m) => BstdSerializer.WriteDictionary(ref b, m, BstdSerializer.WriteString, BstdSerializer.WriteString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.ReadDictionary(ref b, BstdSerializer.ReadString, BstdSerializer.ReadString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.SkipDictionary(ref b, BstdSerializer.SkipString, BstdSerializer.SkipString),
                (expected, actual) => {
                    Assert.AreEqual(expected.Count, actual.Count);
                    foreach(var kvp in expected)
                    {
                        Assert.IsTrue(actual.ContainsKey(kvp.Key));
                        Assert.AreEqual(kvp.Value, actual[kvp.Key]);
                    }
                }
            );
        }
        
        [TestMethod]
        public void TestEmptyAndLongStrings()
        {
            // Empty String
            AssertRoundtrip("", BstdSerializer.GetSizeString, BstdSerializer.WriteString, BstdSerializer.ReadString, BstdSerializer.SkipString, Assert.AreEqual);
            
            // Long String
            var longStr = new string('H', ushort.MaxValue + 1);
            AssertRoundtrip(longStr, BstdSerializer.GetSizeString, BstdSerializer.WriteString, BstdSerializer.ReadString, BstdSerializer.SkipString, Assert.AreEqual);
        }
        
        [TestMethod]
        public void TestTime()
        {
            var now = new DateTime(2022, 9, 17, 10, 14, 55, DateTimeKind.Utc).AddTicks(1234567);
             AssertRoundtrip(
                now,
                _ => BstdSerializer.GetSizeTime(),
                (ref Span<byte> b, DateTime val) => BstdSerializer.WriteTime(ref b, val),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.ReadTime(ref b),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.SkipTime(ref b),
                (expected, actual) => Assert.AreEqual(expected.ToUniversalTime(), actual.ToUniversalTime())
            );
        }
        
        [TestMethod]
        public void TestPointer()
        {
            // Non-Nil Class
            string nonNilValue = "hello world";
            AssertRoundtrip(
                nonNilValue,
                v => BstdSerializer.GetSizePointerClass(v, BstdSerializer.GetSizeString),
                (ref Span<byte> b, string v) => BstdSerializer.WritePointerClass(ref b, v, BstdSerializer.WriteString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.ReadPointerClass(ref b, BstdSerializer.ReadString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.SkipPointer(ref b, BstdSerializer.SkipString),
                Assert.AreEqual
            );
            
            // Nil Class
            string? nilValue = null;
            AssertRoundtrip<string?>(
                nilValue,
                v => BstdSerializer.GetSizePointerClass(v, BstdSerializer.GetSizeString),
                (ref Span<byte> b, string? v) => BstdSerializer.WritePointerClass(ref b, v, BstdSerializer.WriteString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.ReadPointerClass(ref b, BstdSerializer.ReadString),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.SkipPointer(ref b, BstdSerializer.SkipString),
                Assert.AreEqual
            );

            // Non-Nil Struct
            int? nonNilStruct = 12345;
             AssertRoundtrip(
                nonNilStruct,
                v => BstdSerializer.GetSizePointerStruct(v, _ => BstdSerializer.GetSizeInt32()),
                (ref Span<byte> b, int? v) => BstdSerializer.WritePointerStruct(ref b, v, (ref Span<byte> buf, int item) => BstdSerializer.WriteInt32(ref buf, item)),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.ReadPointerStruct(ref b, (ref ReadOnlySpan<byte> buf) => BstdSerializer.ReadInt32(ref buf)),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.SkipPointer(ref b, (ref ReadOnlySpan<byte> buf) => BstdSerializer.SkipInt32(ref buf)),
                Assert.AreEqual
            );
            
            // Nil struct
            int? nilStruct = null;
            AssertRoundtrip(
                nilStruct,
                v => BstdSerializer.GetSizePointerStruct(v, _ => BstdSerializer.GetSizeInt32()),
                (ref Span<byte> b, int? v) => BstdSerializer.WritePointerStruct(ref b, v, (ref Span<byte> buf, int item) => BstdSerializer.WriteInt32(ref buf, item)),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.ReadPointerStruct(ref b, (ref ReadOnlySpan<byte> buf) => BstdSerializer.ReadInt32(ref buf)),
                (ref ReadOnlySpan<byte> b) => BstdSerializer.SkipPointer(ref b, (ref ReadOnlySpan<byte> buf) => BstdSerializer.SkipInt32(ref buf)),
                Assert.AreEqual
            );
        }

        [TestMethod]
        public void TestTerminatorErrors()
        {
            // Slice terminator error
            var slice = new[] { "a" };
            var sliceSize = BstdSerializer.GetSizeArray(slice, BstdSerializer.GetSizeString);
            var sliceBuffer = new Span<byte>(new byte[sliceSize]);
            var writeBuf = sliceBuffer;
            BstdSerializer.WriteArray(ref writeBuf, slice, BstdSerializer.WriteString);
            var truncatedSliceBuffer = new ReadOnlySpan<byte>(sliceBuffer.ToArray(), 0, sliceSize - 1);
            try
            {
                BstdSerializer.SkipArray(ref truncatedSliceBuffer, BstdSerializer.SkipString);
                Assert.Fail("Expected BstdBufferTooSmallException was not thrown for slice terminator.");
            }
            catch (BstdBufferTooSmallException) { /* Expected */ }

            // Map terminator error
            var map = new Dictionary<string, string> { { "a", "b" } };
            var mapSize = BstdSerializer.GetSizeDictionary(map, BstdSerializer.GetSizeString, BstdSerializer.GetSizeString);
            var mapBuffer = new Span<byte>(new byte[mapSize]);
            writeBuf = mapBuffer;
            BstdSerializer.WriteDictionary(ref writeBuf, map, BstdSerializer.WriteString, BstdSerializer.WriteString);
            var truncatedMapBuffer = new ReadOnlySpan<byte>(mapBuffer.ToArray(), 0, mapSize - 1);
            try
            {
                BstdSerializer.SkipDictionary(ref truncatedMapBuffer, BstdSerializer.SkipString, BstdSerializer.SkipString);
                Assert.Fail("Expected BstdBufferTooSmallException was not thrown for map terminator.");
            }
            catch (BstdBufferTooSmallException) { /* Expected */ }
        }

        [TestMethod]
        public void TestFinalCoverage()
        {
            // GetSizeFixedArray
            var fixedSlice = new int[] { 1, 2, 3 };
            var expectedSize = BstdSerializer.GetSizeVarint((nuint)fixedSlice.Length) + (fixedSlice.Length * sizeof(int)) + 4;
            Assert.AreEqual(expectedSize, BstdSerializer.GetSizeFixedArray(fixedSlice, sizeof(int)));

            // UnmarshalTime Error
            try
            {
                var badTimeBuf = new ReadOnlySpan<byte>(new byte[] { 1, 2, 3 });
                BstdSerializer.ReadTime(ref badTimeBuf);
                Assert.Fail("Expected BstdBufferTooSmallException for ReadTime.");
            }
            catch (BstdBufferTooSmallException) { /* Expected */ }
            
            // SkipPointer Error
            try
            {
                var emptyBuf = ReadOnlySpan<byte>.Empty;
                BstdSerializer.SkipPointer(ref emptyBuf, BstdSerializer.SkipByte);
                Assert.Fail("Expected BstdBufferTooSmallException for SkipPointer.");
            }
            catch (BstdBufferTooSmallException) { /* Expected */ }
            
            // UnmarshalPointer Errors
            try
            {
                var b = ReadOnlySpan<byte>.Empty;
                BstdSerializer.ReadPointerClass<string>(ref b, BstdSerializer.ReadString);
                Assert.Fail("Expected BstdBufferTooSmallException for ReadPointerClass (empty).");
            }
            catch (BstdBufferTooSmallException) { /* Expected */ }

            try
            {
                var b = new ReadOnlySpan<byte>(new byte[] { 1, 0x80 });
                BstdSerializer.ReadPointerClass<string>(ref b, BstdSerializer.ReadString);
                Assert.Fail("Expected BstdBufferTooSmallException for ReadPointerClass (truncated).");
            }
            catch (BstdBufferTooSmallException) { /* Expected */ }
        }
    }
}

