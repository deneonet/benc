package dev.benc

import java.nio.{ByteBuffer, ByteOrder, CharBuffer}
import java.nio.charset.{CharacterCodingException, CodingErrorAction, StandardCharsets}
import java.time.Instant
import scala.annotation.tailrec
import scala.collection.mutable.ArrayBuffer

// --- Error Handling ---

/** A universal error for all encoding and decoding operations. */
sealed trait BstdError extends Throwable:
  def message: String
  override def getMessage: String = message

object BstdError:
  /** Signals that a variable-length integer was larger than the supported 64 bits. */
  case object VarIntOverflow extends BstdError:
    val message = "VarInt overflowed its maximum bit width."

  /** Signals that the byte array ended unexpectedly while reading data. */
  final case class BufferTooSmall(needed: Int, found: Int) extends BstdError:
    def message = s"Buffer is too small. Required $needed bytes, but only $found are available."

  /** Signals that the data being read doesn't follow the expected format. */
  final case class InvalidData(details: String) extends BstdError:
    def message = s"Invalid data format: $details"

// --- Type Aliases for Readability ---

/** A standard result for a `read` operation. It's either an error or a tuple of (newOffset, decodedValue). */
type ReadResult[A] = Either[BstdError, (Int, A)]

/** A standard result for a `skip` operation. It's either an error or the new offset. */
type SkipResult = Either[BstdError, Int]


/**
 * The core "recipe book" for handling a specific type `A`.
 *
 * This is a type class that defines how any type `A` can be measured, written, read,
 * and skipped over in a binary format.
 */
trait Benc[A]:
  /** Calculates the number of bytes needed to store a value. */
  def size(value: A): Int

  /** Writes a value into a byte array at a given offset, returning the new offset. */
  def write(offset: Int, buffer: Array[Byte], value: A): Int

  /** Reads a value from a byte array at a given offset. */
  def read(offset: Int, buffer: Array[Byte]): ReadResult[A]
  
  /** Skips past a value in a byte array, returning the new offset after it. */
  def skip(offset: Int, buffer: Array[Byte]): SkipResult

object Benc:
  /** A convenient way to get the `Benc` instance for a type `A` (e.g., `Benc[Int]`). */
  def apply[A](using benc: Benc[A]): Benc[A] = benc

/**
 * The main entry point for the Bstd library. Provides simple `encode` and `decode` functions.
 */
object Bstd:
  private val CollectionTerminator: Array[Byte] = Array(1, 1, 1, 1)
  private val MaxDecodeBytes = 64 * 1024 * 1024
  private val MaxDecodeElements = 1_048_576

  private def validRange(offset: Int, buffer: Array[Byte], size: Int): Boolean =
    offset >= 0 && offset <= buffer.length && size >= 0 && size <= buffer.length - offset

  private def checkedLength(n: Long, offset: Int, buffer: Array[Byte]): Either[BstdError, Int] =
    if java.lang.Long.compareUnsigned(n, MaxDecodeBytes.toLong) > 0 then Left(BstdError.VarIntOverflow)
    else if !validRange(offset, buffer, n.toInt) then Left(BstdError.BufferTooSmall(n.toInt, buffer.length - offset))
    else Right(n.toInt)

  private def checkedCount(n: Long, offset: Int, buffer: Array[Byte], minimum: Int = 1): Either[BstdError, Int] =
    if java.lang.Long.compareUnsigned(n, MaxDecodeElements.toLong) > 0 then Left(BstdError.VarIntOverflow)
    else if !validRange(offset, buffer, 0) || n > (buffer.length - offset) / minimum then Left(BstdError.BufferTooSmall(n.toInt, buffer.length - offset))
    else Right(n.toInt)

  private def decodeUtf8(buffer: Array[Byte], offset: Int, length: Int): Either[BstdError, String] =
    try Right(StandardCharsets.UTF_8.newDecoder()
      .onMalformedInput(CodingErrorAction.REPORT)
      .onUnmappableCharacter(CodingErrorAction.REPORT)
      .decode(ByteBuffer.wrap(buffer, offset, length)).toString)
    catch case _: CharacterCodingException => Left(BstdError.InvalidData("invalid UTF-8"))

  private def encodeUtf8(value: String): Array[Byte] =
    try
      val encoded = StandardCharsets.UTF_8.newEncoder()
        .onMalformedInput(CodingErrorAction.REPORT)
        .onUnmappableCharacter(CodingErrorAction.REPORT)
        .encode(CharBuffer.wrap(value))
      val result = new Array[Byte](encoded.remaining())
      encoded.get(result)
      result
    catch case _: CharacterCodingException => throw new IllegalArgumentException("invalid UTF-16 string")

  /** Encodes any value of type `A` into a compact byte array. */
  def encode[A](value: A)(using benc: Benc[A]): Array[Byte] =
    val buffer = new Array[Byte](benc.size(value))
    benc.write(0, buffer, value)
    buffer

  /** Decodes a byte array back into a value of type `A`. */
  def decode[A](buffer: Array[Byte])(using benc: Benc[A]): Either[BstdError, A] =
    benc.read(0, buffer).map(_._2) // We only care about the value, not the final offset.

  // --- Internal "Magic": How VarInts Work ---

  /**
   * Why ZigZag?
   * Regular signed integers waste space in VarInt encoding. For example, -1 would be a huge
   * number if stored directly. ZigZag encoding cleverly maps small negative numbers to
   * small positive numbers, making them highly compressible.
   * -1 -> 1, -2 -> 3, -3 -> 5, etc.
   */
  private def encodeZigZag(n: Long): Long = (n << 1) ^ (n >> 63)
  private def decodeZigZag(n: Long): Long = (n >>> 1) ^ -(n & 1)

  /** Calculates the size of a VarInt. */
  private def sizeVarUInt(v: Long): Int =
    var value = v
    var count = 1
    while java.lang.Long.compareUnsigned(value, 0x80L) >= 0 do
      value = value >>> 7
      count += 1
    count
  
  /**
   * The VarInt Format:
   * This is a space-saving way to store integers. Each byte uses 7 bits for data. The 8th bit
   * (the most significant bit, 0x80) is a "continuation flag". If it's set, it means more
   * bytes will follow. If it's not set, this is the last byte of the integer.
   */
  private def writeVarUInt(offset: Int, buffer: Array[Byte], v: Long): Int =
    var value = v
    var currentOffset = offset
    while (java.lang.Long.compareUnsigned(value, 0x80L) >= 0) {
      buffer(currentOffset) = (value | 0x80).toByte
      value >>>= 7
      currentOffset += 1
    }
    buffer(currentOffset) = value.toByte
    currentOffset + 1

  /** Reads a VarInt from the buffer. */
  private def readVarUInt(offset: Int, buffer: Array[Byte]): ReadResult[Long] =
    if !validRange(offset, buffer, 0) then return Left(BstdError.BufferTooSmall(1, buffer.length))
    var value: Long = 0
    var shift: Int = 0
    var currentOffset = offset
    while (currentOffset < buffer.length) {
      val byte = buffer(currentOffset)
      if (shift >= 64 || (shift == 63 && (byte & 0xff) > 1)) return Left(BstdError.VarIntOverflow)

      value |= (byte & 0x7f).toLong << shift
      if ((byte & 0x80) == 0) return Right((currentOffset + 1, value))
      
      shift += 7
      currentOffset += 1
    }
    Left(BstdError.BufferTooSmall(currentOffset + 1, buffer.length))

  /** Skips a VarInt in the buffer without decoding it. */
  private def skipVarUInt(offset: Int, buffer: Array[Byte]): SkipResult =
    if !validRange(offset, buffer, 0) then return Left(BstdError.BufferTooSmall(1, buffer.length))
    var currentOffset = offset
    var count = 0
    while (currentOffset < buffer.length) {
      count += 1
      if (count > 10) return Left(BstdError.VarIntOverflow) // A 64-bit VarInt is at most 10 bytes
      if (count == 10 && (buffer(currentOffset) & 0xff) > 1) return Left(BstdError.VarIntOverflow)
      if ((buffer(currentOffset) & 0x80) == 0) return Right(currentOffset + 1)
      currentOffset += 1
    }
    Left(BstdError.BufferTooSmall(currentOffset + 1, buffer.length))

  /**
   * This object contains all the built-in "recipes" (Benc instances) for standard types.
   */
  object instances:
    /**
     * Note: For performance, whole numbers like `Int` and `Long` are stored using a fixed
     * number of bytes (4 and 8, respectively) in little-endian order.
     * For variable-length encoding, which is better for small numbers, import a recipe
     * explicitly from `Bstd.instances.VarInt`.
     */
    
    // --- Boolean ---
    given Benc[Boolean] with
      def size(value: Boolean): Int = 1
      def write(offset: Int, buffer: Array[Byte], value: Boolean): Int =
        buffer(offset) = if (value) 1.toByte else 0.toByte
        offset + 1
      def read(offset: Int, buffer: Array[Byte]): ReadResult[Boolean] =
        if (validRange(offset, buffer, 1) && buffer(offset) >= 0 && buffer(offset) <= 1) Right((offset + 1, buffer(offset) == 1.toByte))
        else if validRange(offset, buffer, 1) then Left(BstdError.InvalidData("invalid boolean"))
        else Left(BstdError.BufferTooSmall(offset + 1, buffer.length))
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        read(offset, buffer).map(_._1)

    // --- Byte ---
    given Benc[Byte] with
      def size(value: Byte): Int = 1
      def write(offset: Int, buffer: Array[Byte], value: Byte): Int =
        buffer(offset) = value
        offset + 1
      def read(offset: Int, buffer: Array[Byte]): ReadResult[Byte] =
        if (validRange(offset, buffer, 1)) Right((offset + 1, buffer(offset)))
        else Left(BstdError.BufferTooSmall(offset + 1, buffer.length))
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        if (validRange(offset, buffer, 1)) Right(offset + 1)
        else Left(BstdError.BufferTooSmall(offset + 1, buffer.length))

    // --- Fixed-size Primitives (e.g., Int, Long, Double) ---
    private def fixedSizeBenc[A](sizeBytes: Int)(
      put: (ByteBuffer, A) => Unit, get: ByteBuffer => A
    ): Benc[A] = new Benc[A]:
      def size(value: A): Int = sizeBytes
      def write(offset: Int, buffer: Array[Byte], value: A): Int =
        val byteBuffer = ByteBuffer.wrap(buffer).order(ByteOrder.LITTLE_ENDIAN).position(offset)
        put(byteBuffer, value)
        offset + sizeBytes
      def read(offset: Int, buffer: Array[Byte]): ReadResult[A] =
        if (validRange(offset, buffer, sizeBytes))
          val byteBuffer = ByteBuffer.wrap(buffer).order(ByteOrder.LITTLE_ENDIAN).position(offset)
          Right((offset + sizeBytes, get(byteBuffer)))
        else Left(BstdError.BufferTooSmall(offset + sizeBytes, buffer.length))
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        if (validRange(offset, buffer, sizeBytes)) Right(offset + sizeBytes)
        else Left(BstdError.BufferTooSmall(offset + sizeBytes, buffer.length))

    given Benc[Short] = fixedSizeBenc[Short](2)(_.putShort(_), _.getShort())
    given Benc[Int]   = fixedSizeBenc[Int](4)(_.putInt(_), _.getInt())
    given Benc[Long]  = fixedSizeBenc[Long](8)(_.putLong(_), _.getLong())
    given Benc[Float] = fixedSizeBenc[Float](4)(_.putFloat(_), _.getFloat())
    given Benc[Double]= fixedSizeBenc[Double](8)(_.putDouble(_), _.getDouble())

    // --- Recipes for Variable-Length Integers ---
    object VarInt:
      given bencLong: Benc[Long] with
        def size(value: Long): Int = sizeVarUInt(encodeZigZag(value))
        def write(offset: Int, buffer: Array[Byte], value: Long): Int = writeVarUInt(offset, buffer, encodeZigZag(value))
        def read(offset: Int, buffer: Array[Byte]): ReadResult[Long] =
          readVarUInt(offset, buffer).map { case (newOffset, value) => (newOffset, decodeZigZag(value)) }
        def skip(offset: Int, buffer: Array[Byte]): SkipResult = skipVarUInt(offset, buffer)
      
      given bencInt: Benc[Int] with
        def size(value: Int): Int = bencLong.size(value.toLong)
        def write(offset: Int, buffer: Array[Byte], value: Int): Int = bencLong.write(offset, buffer, value.toLong)
        def read(offset: Int, buffer: Array[Byte]): ReadResult[Int] = 
          bencLong.read(offset, buffer).flatMap { case (newOffset, longVal) =>
            if longVal < Int.MinValue || longVal > Int.MaxValue then Left(BstdError.VarIntOverflow)
            else Right((newOffset, longVal.toInt))
          }
        def skip(offset: Int, buffer: Array[Byte]): SkipResult = bencLong.skip(offset, buffer)

    object VarUInt:
      given bencLong: Benc[Long] with
        def size(value: Long): Int = sizeVarUInt(value)
        def write(offset: Int, buffer: Array[Byte], value: Long): Int = writeVarUInt(offset, buffer, value)
        def read(offset: Int, buffer: Array[Byte]): ReadResult[Long] = readVarUInt(offset, buffer)
        def skip(offset: Int, buffer: Array[Byte]): SkipResult = skipVarUInt(offset, buffer)

      given bencInt: Benc[Int] with
        def size(value: Int): Int =
          require(value >= 0, "unsigned varint cannot be negative")
          bencLong.size(value.toLong)
        def write(offset: Int, buffer: Array[Byte], value: Int): Int =
          require(value >= 0, "unsigned varint cannot be negative")
          bencLong.write(offset, buffer, value.toLong)
        def read(offset: Int, buffer: Array[Byte]): ReadResult[Int] = 
          bencLong.read(offset, buffer).flatMap { case (newOffset, longVal) =>
            if longVal < 0 || longVal > Int.MaxValue then Left(BstdError.VarIntOverflow)
            else Right((newOffset, longVal.toInt))
          }
        def skip(offset: Int, buffer: Array[Byte]): SkipResult = bencLong.skip(offset, buffer)
          
    // --- String and Byte Array (Length-prefixed format) ---
    private val bencLength = VarUInt.bencInt
    given Benc[String] with
      def size(value: String): Int =
        val byteLength = encodeUtf8(value).length
        bencLength.size(byteLength) + byteLength
      def write(offset: Int, buffer: Array[Byte], value: String): Int =
        val bytes = encodeUtf8(value)
        val offsetAfterLength = bencLength.write(offset, buffer, bytes.length)
        System.arraycopy(bytes, 0, buffer, offsetAfterLength, bytes.length)
        offsetAfterLength + bytes.length
      def read(offset: Int, buffer: Array[Byte]): ReadResult[String] =
        for {
          (offsetAfterLength, rawLength) <- readVarUInt(offset, buffer)
          byteLength <- checkedLength(rawLength, offsetAfterLength, buffer)
          endOffset = offsetAfterLength + byteLength
          str <- decodeUtf8(buffer, offsetAfterLength, byteLength)
        } yield (endOffset, str)
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        for {
          (offsetAfterLength, rawLength) <- readVarUInt(offset, buffer)
          byteLength <- checkedLength(rawLength, offsetAfterLength, buffer)
          endOffset = offsetAfterLength + byteLength
          _ <- decodeUtf8(buffer, offsetAfterLength, byteLength)
        } yield endOffset

    given Benc[Array[Byte]] with
      def size(value: Array[Byte]): Int = bencLength.size(value.length) + value.length
      def write(offset: Int, buffer: Array[Byte], value: Array[Byte]): Int =
        val offsetAfterLength = bencLength.write(offset, buffer, value.length)
        System.arraycopy(value, 0, buffer, offsetAfterLength, value.length)
        offsetAfterLength + value.length
      def read(offset: Int, buffer: Array[Byte]): ReadResult[Array[Byte]] =
        for {
          (offsetAfterLength, rawLength) <- readVarUInt(offset, buffer)
          byteLength <- checkedLength(rawLength, offsetAfterLength, buffer)
          endOffset = offsetAfterLength + byteLength
        } yield (endOffset, buffer.slice(offsetAfterLength, endOffset))
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        for {
          (offsetAfterLength, rawLength) <- readVarUInt(offset, buffer)
          byteLength <- checkedLength(rawLength, offsetAfterLength, buffer)
        } yield offsetAfterLength + byteLength
        
    // --- Time (stored as a fixed 64-bit integer of nanoseconds since Unix epoch) ---
    given Benc[Instant] with
      private val bencNano = Benc[Long]
      def size(value: Instant): Int = 8
      def write(offset: Int, buffer: Array[Byte], value: Instant): Int =
        val totalNanos = Math.addExact(Math.multiplyExact(value.getEpochSecond, 1_000_000_000L), value.getNano.toLong)
        bencNano.write(offset, buffer, totalNanos)
      def read(offset: Int, buffer: Array[Byte]): ReadResult[Instant] =
        bencNano.read(offset, buffer).map { case (newOffset, totalNanos) =>
          val instant = Instant.ofEpochSecond(totalNanos / 1_000_000_000, totalNanos % 1_000_000_000)
          (newOffset, instant)
        }
      def skip(offset: Int, buffer: Array[Byte]): SkipResult = bencNano.skip(offset, buffer)

    // --- Collections (Format: Count -> Elements -> Terminator) ---
    private def checkTerminator(offset: Int, buffer: Array[Byte]): SkipResult =
      val finalOffset = offset + CollectionTerminator.length
      if (!validRange(offset, buffer, CollectionTerminator.length)) Left(BstdError.BufferTooSmall(finalOffset, buffer.length))
      else if (buffer.view.slice(offset, finalOffset).sameElements(CollectionTerminator)) Right(finalOffset)
      else Left(BstdError.InvalidData("Collection terminator not found or malformed."))

    given[A](using bencA: Benc[A]): Benc[Seq[A]] with
      def size(values: Seq[A]): Int =
        bencLength.size(values.length) + values.map(bencA.size).sum + CollectionTerminator.length
      def write(offset: Int, buffer: Array[Byte], values: Seq[A]): Int =
        var currentOffset = bencLength.write(offset, buffer, values.length)
        values.foreach(v => currentOffset = bencA.write(currentOffset, buffer, v))
        System.arraycopy(CollectionTerminator, 0, buffer, currentOffset, CollectionTerminator.length)
        currentOffset + CollectionTerminator.length
      def read(offset: Int, buffer: Array[Byte]): ReadResult[Seq[A]] =
        readVarUInt(offset, buffer).flatMap { case (offsetAfterCount, rawCount) =>
          checkedCount(rawCount, offsetAfterCount, buffer).flatMap(elementCount =>
            readElements(offsetAfterCount, elementCount, buffer, new ArrayBuffer[A]()))
        }
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        readVarUInt(offset, buffer).flatMap { case (offsetAfterCount, rawCount) =>
          checkedCount(rawCount, offsetAfterCount, buffer).flatMap(elementCount =>
            skipElements(offsetAfterCount, elementCount, buffer))
        }
      
      @tailrec private def readElements(offset: Int, remaining: Int, buffer: Array[Byte], acc: ArrayBuffer[A]): ReadResult[Seq[A]] =
        if remaining == 0 then checkTerminator(offset, buffer).map(finalOffset => (finalOffset, acc.toSeq))
        else bencA.read(offset, buffer) match
          case Left(err) => Left(err)
          case Right((newOffset, value)) => readElements(newOffset, remaining - 1, buffer, acc.addOne(value))
      
      @tailrec private def skipElements(offset: Int, remaining: Int, buffer: Array[Byte]): SkipResult =
        if remaining == 0 then checkTerminator(offset, buffer)
        else bencA.skip(offset, buffer) match
          case Left(err) => Left(err)
          case Right(newOffset) => skipElements(newOffset, remaining - 1, buffer)
        
    given[K, V](using bencK: Benc[K], bencV: Benc[V]): Benc[Map[K, V]] with
      def size(value: Map[K, V]): Int =
        bencLength.size(value.size) + value.map { case (k, v) => bencK.size(k) + bencV.size(v) }.sum + CollectionTerminator.length
      def write(offset: Int, buffer: Array[Byte], value: Map[K, V]): Int =
        var currentOffset = bencLength.write(offset, buffer, value.size)
        value.foreach { case (k, v) =>
          currentOffset = bencK.write(currentOffset, buffer, k)
          currentOffset = bencV.write(currentOffset, buffer, v)
        }
        System.arraycopy(CollectionTerminator, 0, buffer, currentOffset, CollectionTerminator.length)
        currentOffset + CollectionTerminator.length
      def read(offset: Int, buffer: Array[Byte]): ReadResult[Map[K, V]] =
        readVarUInt(offset, buffer).flatMap { case (offsetAfterCount, rawCount) =>
          checkedCount(rawCount, offsetAfterCount, buffer, 2).flatMap(pairCount =>
            readPairs(offsetAfterCount, pairCount, buffer, Map.empty))
        }
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        readVarUInt(offset, buffer).flatMap { case (offsetAfterCount, rawCount) =>
          checkedCount(rawCount, offsetAfterCount, buffer, 2).flatMap(pairCount =>
            skipPairs(offsetAfterCount, pairCount, buffer))
        }
      
      @tailrec private def readPairs(offset: Int, remaining: Int, buffer: Array[Byte], acc: Map[K, V]): ReadResult[Map[K, V]] =
        if remaining == 0 then checkTerminator(offset, buffer).map(finalOffset => (finalOffset, acc))
        else (for {
            (offsetAfterKey, key)     <- bencK.read(offset, buffer)
            (offsetAfterValue, value) <- bencV.read(offsetAfterKey, buffer)
          } yield (offsetAfterValue, key -> value)) match
          case Left(err) => Left(err)
          case Right((newOffset, pair)) =>
            if acc.contains(pair._1) then Left(BstdError.InvalidData("duplicate map key"))
            else readPairs(newOffset, remaining - 1, buffer, acc + pair)
          
      @tailrec private def skipPairs(offset: Int, remaining: Int, buffer: Array[Byte]): SkipResult =
        if remaining == 0 then checkTerminator(offset, buffer)
        else (for {
            offsetAfterKey   <- bencK.skip(offset, buffer)
            offsetAfterValue <- bencV.skip(offsetAfterKey, buffer)
          } yield offsetAfterValue) match
          case Left(err) => Left(err)
          case Right(newOffset) => skipPairs(newOffset, remaining - 1, buffer)

    // --- Option (for pointers/nullable types, stored as a boolean flag + optional data) ---
    given[A](using bencA: Benc[A]): Benc[Option[A]] with
      private val bencBool = Benc[Boolean]
      def size(value: Option[A]): Int = 1 + value.map(bencA.size).getOrElse(0)
      def write(offset: Int, buffer: Array[Byte], value: Option[A]): Int = value match
        case Some(v) => bencA.write(bencBool.write(offset, buffer, true), buffer, v)
        case None    => bencBool.write(offset, buffer, false)
      def read(offset: Int, buffer: Array[Byte]): ReadResult[Option[A]] =
        bencBool.read(offset, buffer).flatMap {
          case (offsetAfterFlag, true)  => bencA.read(offsetAfterFlag, buffer).map { case (finalOffset, v) => (finalOffset, Some(v)) }
          case (offsetAfterFlag, false) => Right((offsetAfterFlag, None))
        }
      def skip(offset: Int, buffer: Array[Byte]): SkipResult =
        bencBool.read(offset, buffer).flatMap {
          case (offsetAfterFlag, true)  => bencA.skip(offsetAfterFlag, buffer)
          case (offsetAfterFlag, false) => Right(offsetAfterFlag)
        }
