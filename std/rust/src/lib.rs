use std::collections::HashMap;
use std::hash::Hash;
use std::mem::size_of;
use thiserror::Error;
// To use the new time functions, you would need to add `chrono` to your Cargo.toml:
// `chrono = { version = "0.4" }`
use chrono::{DateTime, Utc};

/// The terminator sequence used to mark the end of slices and maps.
/// This specific sequence is chosen as it's unlikely to appear naturally
/// in varint-encoded data.
const TERMINATOR: [u8; 4] = [1, 1, 1, 1];

/// The maximum number of bytes a 64-bit varint can occupy.
const MAX_VARINT_LEN_64: usize = 10;
const MAX_DECODE_BYTES: u64 = 64 * 1024 * 1024;
const MAX_DECODE_ELEMENTS: u64 = 1_048_576;

/// A specialized `Result` type for bstd operations.
pub type Result<T> = std::result::Result<T, Error>;

/// Represents errors that can occur during serialization or deserialization.
#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("buffer is too small to complete the operation")]
    BufferTooSmall,
    #[error("varint is too large and overflows")]
    VarintOverflow,
    #[error("data is not a valid UTF-8 string")]
    InvalidUtf8(#[from] std::str::Utf8Error),
    #[error("expected a terminator sequence, but it was not found")]
    MissingTerminator,
    #[error("value is out of range for the target integer type")]
    OutOfRange,
    #[error("decoded input exceeds the configured limit")]
    DecodeLimit,
    #[error("invalid boolean value")]
    InvalidBool,
    #[error("duplicate map key")]
    DuplicateKey,
}

// ===================================================================================
// Generic Helpers
// ===================================================================================

/// A helper function to advance a slice cursor.
#[inline]
fn advance<'a>(slice: &mut &'a [u8], n: usize) -> Result<&'a [u8]> {
    if slice.len() < n {
        return Err(Error::BufferTooSmall);
    }
    let (head, tail) = slice.split_at(n);
    *slice = tail;
    Ok(head)
}

/// A helper function to write to a slice cursor.
#[inline]
fn write_to_slice<'a>(slice: &mut &'a mut [u8], data: &[u8]) -> Result<()> {
    if slice.len() < data.len() {
        return Err(Error::BufferTooSmall);
    }
    // This cannot be a single call due to lifetime issues with mutable borrows.
    let (head, tail) = std::mem::take(slice).split_at_mut(data.len());
    head.copy_from_slice(data);
    *slice = tail;
    Ok(())
}

fn read_length(reader: &mut &[u8]) -> Result<usize> {
    let raw = unmarshal_uint(reader)?;
    if raw > MAX_DECODE_BYTES { return Err(Error::DecodeLimit); }
    let length = usize::try_from(raw).map_err(|_| Error::OutOfRange)?;
    if length > reader.len() { return Err(Error::BufferTooSmall); }
    Ok(length)
}

fn read_count(reader: &mut &[u8], minimum_bytes: usize) -> Result<usize> {
    let raw = unmarshal_uint(reader)?;
    if raw > MAX_DECODE_ELEMENTS { return Err(Error::DecodeLimit); }
    let count = usize::try_from(raw).map_err(|_| Error::OutOfRange)?;
    if count > reader.len() / minimum_bytes { return Err(Error::BufferTooSmall); }
    Ok(count)
}

// ===================================================================================
// String
// ===================================================================================

/// Returns the number of bytes required to marshal a string.
pub fn size_string(s: &str) -> usize {
    size_uint(s.len() as u64) + s.len()
}

/// Marshals a string into the writer.
/// The format is a varint-encoded length followed by the string's UTF-8 bytes.
///
/// Returns an error if the writer is too small.
pub fn marshal_string(s: &str, writer: &mut &mut [u8]) -> Result<()> {
    if s.len() > writer.len() || size_uint(s.len() as u64) > writer.len() - s.len() {
        return Err(Error::BufferTooSmall);
    }
    marshal_uint(s.len() as u64, writer)?;
    write_to_slice(writer, s.as_bytes())
}

/// Unmarshals a string slice from the reader without allocating.
/// The returned `&str` is a slice of the input buffer.
pub fn unmarshal_string<'a>(reader: &mut &'a [u8]) -> Result<&'a str> {
    let mut cursor: &'a [u8] = *reader;
    let len = read_length(&mut cursor)?;
    let bytes = advance(&mut cursor, len)?;
    let value = std::str::from_utf8(bytes)?;
    *reader = cursor;
    Ok(value)
}

/// Skips over a marshalled string in the reader.
pub fn skip_string(reader: &mut &[u8]) -> Result<()> {
    let mut cursor = *reader;
    let len = read_length(&mut cursor)?;
    std::str::from_utf8(&cursor[..len])?;
    advance(&mut cursor, len)?;
    *reader = cursor;
    Ok(())
}

// ===================================================================================
// Byte Slice
// ===================================================================================

/// Returns the number of bytes required to marshal a byte slice.
pub fn size_bytes(b: &[u8]) -> usize {
    size_uint(b.len() as u64) + b.len()
}

/// Marshals a byte slice into the writer.
/// The format is a varint-encoded length followed by the bytes.
///
/// Returns an error if the writer is too small.
pub fn marshal_bytes(b: &[u8], writer: &mut &mut [u8]) -> Result<()> {
    if b.len() > writer.len() || size_uint(b.len() as u64) > writer.len() - b.len() {
        return Err(Error::BufferTooSmall);
    }
    marshal_uint(b.len() as u64, writer)?;
    write_to_slice(writer, b)
}

/// Unmarshals a byte slice from the reader without allocating (cropped).
/// The returned `&[u8]` is a slice of the input buffer.
pub fn unmarshal_bytes_cropped<'a>(reader: &mut &'a [u8]) -> Result<&'a [u8]> {
    let mut cursor = *reader;
    let len = read_length(&mut cursor)?;
    let result = advance(&mut cursor, len)?;
    *reader = cursor;
    Ok(result)
}

/// Unmarshals a byte slice from the reader by copying into a new `Vec<u8>`.
pub fn unmarshal_bytes_copied(reader: &mut &[u8]) -> Result<Vec<u8>> {
    Ok(unmarshal_bytes_cropped(reader)?.to_vec())
}

/// Skips over a marshalled byte slice in the reader.
pub fn skip_bytes(reader: &mut &[u8]) -> Result<()> {
    unmarshal_bytes_cropped(reader)?;
    Ok(())
}

// ===================================================================================
// Slice / Vec<T>
// ===================================================================================

/// Returns the number of bytes needed to marshal a slice of elements with dynamic sizes.
pub fn size_slice<T>(slice: &[T], sizer: impl Fn(&T) -> usize) -> usize {
    let len = slice.len();
    size_uint(len as u64)
        + slice.iter().map(sizer).sum::<usize>()
        + TERMINATOR.len()
}

/// Returns the number of bytes needed to marshal a slice of elements with a fixed size.
pub fn size_fixed_slice<T>(slice: &[T], element_size: usize) -> usize {
    let len = slice.len();
    size_uint(len as u64) + len * element_size + TERMINATOR.len()
}

/// Marshals a slice into the writer.
///
/// Returns an error if the writer is too small.
pub fn marshal_slice<T>(
    slice: &[T],
    writer: &mut &mut [u8],
    marshaler: impl Fn(&T, &mut &mut [u8]) -> Result<()>,
) -> Result<()> {
    marshal_uint(slice.len() as u64, writer)?;
    for item in slice {
        marshaler(item, writer)?;
    }
    write_to_slice(writer, &TERMINATOR)
}

/// Unmarshals a slice from the reader.
pub fn unmarshal_slice<T>(
    reader: &mut &[u8],
    unmarshaler: impl Fn(&mut &[u8]) -> Result<T>,
) -> Result<Vec<T>> {
    let mut cursor = *reader;
    let len = read_count(&mut cursor, 1)?;
    let mut vec = Vec::new();
    for _ in 0..len {
        vec.push(unmarshaler(&mut cursor)?);
    }
    let terminator = advance(&mut cursor, TERMINATOR.len())?;
    if terminator != TERMINATOR {
        return Err(Error::MissingTerminator);
    }
    *reader = cursor;
    Ok(vec)
}

/// Skips over a marshalled slice in the reader.
pub fn skip_slice(
    reader: &mut &[u8],
    skip_element: impl Fn(&mut &[u8]) -> Result<()>,
) -> Result<()> {
    let mut cursor = *reader;
    let len = read_count(&mut cursor, 1)?;
    for _ in 0..len {
        skip_element(&mut cursor)?;
    }
    let terminator = advance(&mut cursor, TERMINATOR.len())?;
    if terminator != TERMINATOR {
        return Err(Error::MissingTerminator);
    }
    *reader = cursor;
    Ok(())
}

// ===================================================================================
// Map / HashMap<K, V>
// ===================================================================================

/// Returns the bytes needed to marshal a map.
pub fn size_map<K, V>(
    map: &HashMap<K, V>,
    k_sizer: impl Fn(&K) -> usize,
    v_sizer: impl Fn(&V) -> usize,
) -> usize {
    let len = map.len();
    let mut total_size = size_uint(len as u64) + TERMINATOR.len();
    for (k, v) in map.iter() {
        total_size += k_sizer(k);
        total_size += v_sizer(v);
    }
    total_size
}

/// Marshals a map into the writer.
///
/// Returns an error if the writer is too small.
pub fn marshal_map<K, V>(
    map: &HashMap<K, V>,
    writer: &mut &mut [u8],
    k_marshaler: impl Fn(&K, &mut &mut [u8]) -> Result<()>,
    v_marshaler: impl Fn(&V, &mut &mut [u8]) -> Result<()>,
) -> Result<()> {
    marshal_uint(map.len() as u64, writer)?;
    for (k, v) in map.iter() {
        k_marshaler(k, writer)?;
        v_marshaler(v, writer)?;
    }
    write_to_slice(writer, &TERMINATOR)
}

/// Unmarshals a map from the reader.
pub fn unmarshal_map<'a, K, V>(
    reader: &mut &'a [u8],
    k_unmarshaler: impl Fn(&mut &'a [u8]) -> Result<K>,
    v_unmarshaler: impl Fn(&mut &'a [u8]) -> Result<V>,
) -> Result<HashMap<K, V>>
where
    K: Eq + Hash,
{
    let mut cursor = *reader;
    let len = read_count(&mut cursor, 2)?;
    let mut map = HashMap::new();
    for _ in 0..len {
        let k = k_unmarshaler(&mut cursor)?;
        let v = v_unmarshaler(&mut cursor)?;
        if map.insert(k, v).is_some() { return Err(Error::DuplicateKey); }
    }
    let terminator = advance(&mut cursor, TERMINATOR.len())?;
    if terminator != TERMINATOR {
        return Err(Error::MissingTerminator);
    }
    *reader = cursor;
    Ok(map)
}

/// Skips over a marshalled map by skipping each key and value individually.
pub fn skip_map(
    reader: &mut &[u8],
    skip_key: impl Fn(&mut &[u8]) -> Result<()>,
    skip_value: impl Fn(&mut &[u8]) -> Result<()>,
) -> Result<()> {
    let mut cursor = *reader;
    let len = read_count(&mut cursor, 2)?;
    for _ in 0..len {
        skip_key(&mut cursor)?;
        skip_value(&mut cursor)?;
    }
    let terminator = advance(&mut cursor, TERMINATOR.len())?;
    if terminator != TERMINATOR {
        return Err(Error::MissingTerminator);
    }
    *reader = cursor;
    Ok(())
}

// ===================================================================================
// Varint (u64 / i64)
// ===================================================================================

/// Encodes a signed integer using ZigZag encoding.
///
/// This is a low-level helper function, primarily used internally for
/// `marshal_int`, but exposed for testing purposes.
#[inline]
pub fn encode_zigzag(v: i64) -> u64 {
    ((v << 1) ^ (v >> 63)) as u64
}

/// Decodes a ZigZag-encoded unsigned integer back to a signed one.
///
/// This is a low-level helper function, primarily used internally for
/// `unmarshal_int`, but exposed for testing purposes.
#[inline]
pub fn decode_zigzag(v: u64) -> i64 {
    ((v >> 1) as i64) ^ (-((v & 1) as i64))
}


/// Returns the number of bytes required to marshal a `u64` as a varint.
pub fn size_uint(v: u64) -> usize {
    if v == 0 {
        return 1;
    }
    // Efficiently calculate the number of bytes required.
    // (bits_needed + 6) / 7
    (u64::BITS - v.leading_zeros() + 6) as usize / 7
}

/// Marshals a `u64` as a varint into the writer.
///
/// Returns an error if the writer is too small.
pub fn marshal_uint(mut v: u64, writer: &mut &mut [u8]) -> Result<()> {
    let mut buf = [0u8; MAX_VARINT_LEN_64];
    let mut i = 0;
    while v >= 0x80 {
        buf[i] = (v as u8) | 0x80;
        v >>= 7;
        i += 1;
    }
    buf[i] = v as u8;
    i += 1;
    write_to_slice(writer, &buf[..i])
}

/// Unmarshals a varint-encoded `u64` from the reader.
pub fn unmarshal_uint(reader: &mut &[u8]) -> Result<u64> {
    let mut val: u64 = 0;
    let mut shift: u32 = 0;
    for i in 0..MAX_VARINT_LEN_64 {
        let byte = *reader.get(i).ok_or(Error::BufferTooSmall)?;
        if byte < 0x80 {
            // Last byte
            if i == MAX_VARINT_LEN_64 - 1 && byte > 1 {
                return Err(Error::VarintOverflow);
            }
            advance(reader, i + 1)?;
            return Ok(val | (u64::from(byte) << shift));
        }
        val |= u64::from(byte & 0x7F) << shift;
        shift += 7;
    }
    Err(Error::VarintOverflow)
}

/// Skips over a marshalled varint in the reader.
pub fn skip_uint(reader: &mut &[u8]) -> Result<()> {
    for i in 0..MAX_VARINT_LEN_64 {
        let byte = *reader.get(i).ok_or(Error::BufferTooSmall)?;
        if byte < 0x80 {
            // Add the same overflow check as in unmarshal_uint for the 10-byte case.
            if i == MAX_VARINT_LEN_64 - 1 && byte > 1 {
                return Err(Error::VarintOverflow);
            }
            advance(reader, i + 1)?;
            return Ok(());
        }
    }
    Err(Error::VarintOverflow)
}

/// Returns the number of bytes required to marshal an `i64` as a varint.
pub fn size_int(v: i64) -> usize {
    size_uint(encode_zigzag(v))
}

/// Marshals an `i64` as a ZigZag-encoded varint into the writer.
///
/// Returns an error if the writer is too small.
pub fn marshal_int(v: i64, writer: &mut &mut [u8]) -> Result<()> {
    marshal_uint(encode_zigzag(v), writer)
}

/// Unmarshals a ZigZag-encoded varint `i64` from the reader.
pub fn unmarshal_int(reader: &mut &[u8]) -> Result<i64> {
    unmarshal_uint(reader).map(decode_zigzag)
}

/// Skips over a marshalled zigzag-encoded varint in the reader.
pub fn skip_int(reader: &mut &[u8]) -> Result<()> {
    skip_uint(reader)
}

// ===================================================================================
// Varint (usize / isize) - Platform Dependent
// ===================================================================================

/// Returns the number of bytes required to marshal a `usize` as a varint.
/// Note: The value is always marshalled as a `u64` for platform independence.
pub fn size_usize(v: usize) -> usize {
    size_uint(v as u64)
}

/// Marshals a `usize` into the writer.
/// Note: The value is always marshalled as a `u64` for platform independence.
///
/// Returns an error if the writer is too small.
pub fn marshal_usize(v: usize, writer: &mut &mut [u8]) -> Result<()> {
    marshal_uint(v as u64, writer)
}

/// Unmarshals a `usize` from the reader.
/// Returns an `OutOfRange` error if the unmarshalled `u64` does not fit into a `usize`
/// on the target platform (e.g., a large value on a 32-bit system).
pub fn unmarshal_usize(reader: &mut &[u8]) -> Result<usize> {
    let mut cursor = *reader;
    let val = unmarshal_uint(&mut cursor)?;
    let result = usize::try_from(val).map_err(|_| Error::OutOfRange)?;
    *reader = cursor;
    Ok(result)
}

/// Skips over a marshalled `usize` in the reader.
pub fn skip_usize(reader: &mut &[u8]) -> Result<()> {
    skip_uint(reader)
}

/// Returns the number of bytes required to marshal an `isize` as a varint.
/// Note: The value is always marshalled as an `i64` for platform independence.
pub fn size_isize(v: isize) -> usize {
    size_int(v as i64)
}

/// Marshals an `isize` into the writer.
/// Note: The value is always marshalled as an `i64` for platform independence.
///
/// Returns an error if the writer is too small.
pub fn marshal_isize(v: isize, writer: &mut &mut [u8]) -> Result<()> {
    marshal_int(v as i64, writer)
}

/// Unmarshals an `isize` from the reader.
/// Returns an `OutOfRange` error if the unmarshalled `i64` does not fit into an `isize`
/// on the target platform.
pub fn unmarshal_isize(reader: &mut &[u8]) -> Result<isize> {
    let mut cursor = *reader;
    let val = unmarshal_int(&mut cursor)?;
    let result = isize::try_from(val).map_err(|_| Error::OutOfRange)?;
    *reader = cursor;
    Ok(result)
}

/// Skips over a marshalled `isize` in the reader.
pub fn skip_isize(reader: &mut &[u8]) -> Result<()> {
    skip_int(reader)
}

// ===================================================================================
// Fixed-Size Primitives
// ===================================================================================

// Use a macro to generate functions for fixed-size types to avoid boilerplate.
macro_rules! fixed_size_impl {
    ($type:ty,
     $size_fn:ident,
     $marshal_fn:ident,
     $unmarshal_fn:ident,
     $skip_fn:ident
    ) => {
        /// Returns the number of bytes required to marshal a `$type`.
        pub const fn $size_fn() -> usize {
            size_of::<$type>()
        }

        /// Marshals a `$type` into the writer using little-endian encoding.
        ///
        /// Returns an error if the writer is too small.
        pub fn $marshal_fn(v: $type, writer: &mut &mut [u8]) -> Result<()> {
            let bytes = v.to_le_bytes();
            write_to_slice(writer, &bytes)
        }

        /// Unmarshals a `$type` from the reader using little-endian encoding.
        pub fn $unmarshal_fn(reader: &mut &[u8]) -> Result<$type> {
            let bytes = advance(reader, size_of::<$type>())?;
            Ok(<$type>::from_le_bytes(bytes.try_into().unwrap()))
        }

        /// Skips over a marshalled `$type` in the reader.
        pub fn $skip_fn(reader: &mut &[u8]) -> Result<()> {
            advance(reader, size_of::<$type>())?;
            Ok(())
        }
    };
    // Variant for floats which require `from_bits`
    (float $type:ty,
     $size_fn:ident,
     $marshal_fn:ident,
     $unmarshal_fn:ident,
     $skip_fn:ident,
     $int_type:ty
    ) => {
        /// Returns the number of bytes required to marshal a `$type`.
        pub const fn $size_fn() -> usize {
            size_of::<$type>()
        }

        /// Marshals a `$type` into the writer using little-endian encoding.
        ///
        /// Returns an error if the writer is too small.
        pub fn $marshal_fn(v: $type, writer: &mut &mut [u8]) -> Result<()> {
            let bytes = v.to_bits().to_le_bytes();
            write_to_slice(writer, &bytes)
        }

        /// Unmarshals a `$type` from the reader using little-endian encoding.
        pub fn $unmarshal_fn(reader: &mut &[u8]) -> Result<$type> {
            let bytes = advance(reader, size_of::<$type>())?;
            let int_val = <$int_type>::from_le_bytes(bytes.try_into().unwrap());
            Ok(<$type>::from_bits(int_val))
        }

        /// Skips over a marshalled `$type` in the reader.
        pub fn $skip_fn(reader: &mut &[u8]) -> Result<()> {
            advance(reader, size_of::<$type>())?;
            Ok(())
        }
    };
}

fixed_size_impl!(u16, size_u16, marshal_u16, unmarshal_u16, skip_u16);
fixed_size_impl!(u32, size_u32, marshal_u32, unmarshal_u32, skip_u32);
fixed_size_impl!(u64, size_u64, marshal_u64, unmarshal_u64, skip_u64);
fixed_size_impl!(i16, size_i16, marshal_i16, unmarshal_i16, skip_i16);
fixed_size_impl!(i32, size_i32, marshal_i32, unmarshal_i32, skip_i32);
fixed_size_impl!(i64, size_i64, marshal_i64, unmarshal_i64, skip_i64);
fixed_size_impl!(float f32, size_f32, marshal_f32, unmarshal_f32, skip_f32, u32);
fixed_size_impl!(float f64, size_f64, marshal_f64, unmarshal_f64, skip_f64, u64);

// u8/byte is a special simple case
/// Returns the number of bytes required to marshal a `u8`.
pub const fn size_u8() -> usize { 1 }
/// Marshals a `u8` (byte) into the writer.
/// Returns an error if the writer is too small.
pub fn marshal_u8(v: u8, writer: &mut &mut [u8]) -> Result<()> {
    write_to_slice(writer, &[v])
}
/// Unmarshals a `u8` (byte) from the reader.
pub fn unmarshal_u8(reader: &mut &[u8]) -> Result<u8> {
    Ok(advance(reader, 1)?[0])
}
/// Skips over a marshalled `u8` in the reader.
pub fn skip_u8(reader: &mut &[u8]) -> Result<()> {
    advance(reader, 1)?; Ok(())
}

// i8 is also special, mapping to u8
/// Returns the number of bytes required to marshal an `i8`.
pub const fn size_i8() -> usize { 1 }
/// Marshals an `i8` into the writer.
/// Returns an error if the writer is too small.
pub fn marshal_i8(v: i8, writer: &mut &mut [u8]) -> Result<()> {
    marshal_u8(v as u8, writer)
}
/// Unmarshals an `i8` from the reader.
pub fn unmarshal_i8(reader: &mut &[u8]) -> Result<i8> {
    unmarshal_u8(reader).map(|v| v as i8)
}
/// Skips over a marshalled `i8` in the reader.
pub fn skip_i8(reader: &mut &[u8]) -> Result<()> {
    skip_u8(reader)
}


// bool is also special
/// Returns the number of bytes required to marshal a `bool`.
pub const fn size_bool() -> usize { 1 }
/// Marshals a `bool` into the writer (1 for true, 0 for false).
/// Returns an error if the writer is too small.
pub fn marshal_bool(v: bool, writer: &mut &mut [u8]) -> Result<()> {
    marshal_u8(if v { 1 } else { 0 }, writer)
}
/// Unmarshals a `bool` from the reader.
/// Note: for compatibility with the Go implementation, this is a strict check.
/// Only a value of 1 is considered true.
pub fn unmarshal_bool(reader: &mut &[u8]) -> Result<bool> {
    let val = *reader.first().ok_or(Error::BufferTooSmall)?;
    if val > 1 { return Err(Error::InvalidBool); }
    advance(reader, 1)?;
    Ok(val == 1)
}
/// Skips over a marshalled `bool` in the reader.
pub fn skip_bool(reader: &mut &[u8]) -> Result<()> {
    unmarshal_bool(reader).map(|_| ())
}

// ===================================================================================
// Time (chrono::DateTime<Utc>)
// ===================================================================================

/// Returns the number of bytes required to marshal a `DateTime<Utc>`.
pub const fn size_time() -> usize {
    size_i64()
}

/// Marshals a `DateTime<Utc>` as its nanosecond timestamp (i64).
/// Returns an error if the writer is too small.
pub fn marshal_time(t: DateTime<Utc>, writer: &mut &mut [u8]) -> Result<()> {
    marshal_i64(t.timestamp_nanos_opt().ok_or(Error::OutOfRange)?, writer)
}

/// Unmarshals a `DateTime<Utc>` from its nanosecond timestamp (i64).
pub fn unmarshal_time(reader: &mut &[u8]) -> Result<DateTime<Utc>> {
    let nanos = unmarshal_i64(reader)?;
    // DateTime::from_timestamp_nanos is infallible for any i64.
    Ok(DateTime::from_timestamp_nanos(nanos))
}

/// Skips over a marshalled `DateTime<Utc>` in the reader.
pub fn skip_time(reader: &mut &[u8]) -> Result<()> {
    skip_i64(reader)
}

// ===================================================================================
// Option<T> (for nullable/pointer types)
// ===================================================================================

/// Returns the bytes needed to marshal an `Option<T>`.
/// It adds 1 byte for a boolean flag to indicate if the value is `Some` or `None`.
pub fn size_option<T>(v: &Option<T>, sizer: impl Fn(&T) -> usize) -> usize {
    size_bool() + v.as_ref().map(sizer).unwrap_or(0)
}

/// Marshals an `Option<T>` into the writer.
/// It writes a `bool` (true for `Some`, false for `None`), followed by the marshalled
/// value if it is `Some`.
/// Returns an error if the writer is too small.
pub fn marshal_option<T>(
    v: &Option<T>,
    writer: &mut &mut [u8],
    marshaler: impl Fn(&T, &mut &mut [u8]) -> Result<()>,
) -> Result<()> {
    marshal_bool(v.is_some(), writer)?;
    if let Some(value) = v {
        marshaler(value, writer)?;
    }
    Ok(())
}

/// Unmarshals an `Option<T>` from the reader.
/// It reads a `bool`. If true, it unmarshals the inner value; otherwise, it returns `None`.
pub fn unmarshal_option<T>(
    reader: &mut &[u8],
    unmarshaler: impl Fn(&mut &[u8]) -> Result<T>,
) -> Result<Option<T>> {
    let mut cursor = *reader;
    let result = if unmarshal_bool(&mut cursor)? {
        Some(unmarshaler(&mut cursor)?)
    } else {
        None
    };
    *reader = cursor;
    Ok(result)
}

/// Skips over a marshalled `Option<T>` in the reader.
pub fn skip_option(
    reader: &mut &[u8],
    skip_element: impl Fn(&mut &[u8]) -> Result<()>,
) -> Result<()> {
    let mut cursor = *reader;
    if unmarshal_bool(&mut cursor)? {
        skip_element(&mut cursor)?;
    }
    *reader = cursor;
    Ok(())
}
