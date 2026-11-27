#pragma once

// --- Compiler/Language Standard Check ---
#if __cplusplus < 202002L
    #error "bstd.hpp requires a C++20 compliant compiler. Please use the -std=c++20 flag."
#endif

#include <cassert>
#include <cstddef>
#include <cstdint>
#include <cstring> // For std::memcpy
#include <map>
#include <span>      
#include <string>
#include <string_view>
#include <system_error>
#include <type_traits>
#include <variant>
#include <vector>
#include <algorithm>
#include <chrono>
#include <optional>
#include <stdexcept>

// --- Deeper C++20 Feature Check ---
#if defined(__cpp_lib_span) && __cpp_lib_span >= 202002L
    // OK
#else
    #error "bstd.hpp requires full C++20 std::span support. Your compiler's standard library is too old. Please use GCC 10+, Clang 12+, or MSVC v14.28+."
#endif

#if defined(__GNUC__) && !defined(__clang__) && (__GNUC__ < 10)
    #warning "GCC versions prior to 10 may have incomplete support for C++20 features like std::span."
#endif


namespace bstd {

// --- Public Error and Result Types ---

/**
 * @brief Error codes for bstd operations.
 */
enum class Error {
    Ok = 0,             ///< No error
    Overflow,           ///< A variable-length integer overflowed its boundaries
    BufferTooSmall,     ///< The provided buffer is too small for the operation
    InvalidData,        ///< Malformed terminator, boolean, UTF-8 or duplicate key
    DecodeLimit,        ///< Decoding exceeds the resource budget
};

/**
 * @brief Represents the successful result of an unmarshal operation.
 * @tparam T The type of the unmarshalled value.
 */
template<typename T>
struct UnmarshalResult {
    T value;
    std::size_t new_offset;
};

/**
 * @brief A variant representing either a successful result or an error.
 * @tparam T The type of the value in the successful result.
 */
template<typename T>
using Result = std::variant<UnmarshalResult<T>, Error>;

/**
 * @brief A variant representing either a new offset or an error, used by Skip functions.
 */
using SkipResult = std::variant<std::size_t, Error>;


// --- Implementation Details ---
namespace detail {
constexpr std::size_t MAX_DECODE_ELEMENTS = 1'048'576;
constexpr std::size_t MAX_DECODE_BYTES = 64 * 1024 * 1024;
inline bool has_bytes(std::span<const std::byte> b, std::size_t n, std::size_t amount) noexcept {
    return n <= b.size() && amount <= b.size() - n;
}
inline Error check_terminator(std::span<const std::byte> b, std::size_t n) noexcept {
    if (!has_bytes(b, n, 4)) return Error::BufferTooSmall;
    for (std::size_t i = 0; i < 4; ++i)
        if (b[n + i] != std::byte{1}) return Error::InvalidData;
    return Error::Ok;
}
inline bool valid_utf8(std::span<const std::byte> input) noexcept {
    std::size_t i = 0;
    while (i < input.size()) {
        auto c = std::to_integer<uint8_t>(input[i++]);
        if (c < 0x80) continue;
        int trail;
        uint32_t cp;
        if (c >= 0xC2 && c <= 0xDF) { trail = 1; cp = c & 0x1F; }
        else if (c >= 0xE0 && c <= 0xEF) { trail = 2; cp = c & 0x0F; }
        else if (c >= 0xF0 && c <= 0xF4) { trail = 3; cp = c & 0x07; }
        else return false;
        if (static_cast<std::size_t>(trail) > input.size() - i) return false;
        for (int j = 0; j < trail; ++j) {
            auto next = std::to_integer<uint8_t>(input[i++]);
            if ((next & 0xC0) != 0x80) return false;
            cp = (cp << 6) | (next & 0x3F);
        }
        if ((trail == 2 && cp < 0x800) || (trail == 3 && cp < 0x10000) ||
            (cp >= 0xD800 && cp <= 0xDFFF) || cp > 0x10FFFF) return false;
    }
    return true;
}

#if INTPTR_MAX == INT32_MAX
    constexpr int MAX_VARINT_LEN = 5;
#elif INTPTR_MAX == INT64_MAX
    constexpr int MAX_VARINT_LEN = 10;
#else
    #error "bstd.hpp: Unsupported pointer size."
#endif

constexpr std::byte SLICE_MAP_TERMINATOR[] = {std::byte{1}, std::byte{1}, std::byte{1}, std::byte{1}};

template<typename T, std::enable_if_t<std::is_signed_v<T>, int> = 0>
constexpr std::make_unsigned_t<T> encode_zigzag(T t) noexcept {
    return (static_cast<std::make_unsigned_t<T>>(t) << 1) ^ (t >> (sizeof(T) * 8 - 1));
}

template<typename T, std::enable_if_t<std::is_unsigned_v<T>, int> = 0>
constexpr std::make_signed_t<T> decode_zigzag(T t) noexcept {
    using Signed = std::make_signed_t<T>;
    return static_cast<Signed>((t >> 1) ^ (-(t & 1)));
}

// --- Varint Functions ---

inline std::size_t size_uint_var(uintptr_t v) noexcept {
    std::size_t i = 1;
    while (v >= 0x80) {
        v >>= 7;
        i++;
    }
    return i;
}

inline std::size_t marshal_uint_var(std::span<std::byte> b, std::size_t n, uintptr_t v) noexcept {
    assert(b.size() >= n + size_uint_var(v) && "marshal_uint_var: buffer too small");
    std::size_t i = n;
    while (v >= 0x80) {
        b[i++] = static_cast<std::byte>(v) | std::byte{0x80};
        v >>= 7;
    }
    b[i++] = static_cast<std::byte>(v);
    return i;
}

inline Result<uintptr_t> unmarshal_uint_var(std::span<const std::byte> b, std::size_t n) noexcept {
    uintptr_t x = 0;
    uintptr_t s = 0;
    for (std::size_t i = 0; i < MAX_VARINT_LEN; ++i) {
        if (n + i >= b.size()) {
            return Error::BufferTooSmall;
        }
        const std::byte B = b[n + i];
        if (std::to_integer<uint8_t>(B) < 0x80) {
            if (i == MAX_VARINT_LEN - 1 && std::to_integer<uint8_t>(B) > ((1u << (sizeof(uintptr_t) * 8 - 7 * i)) - 1)) {
                return Error::Overflow;
            }
            x |= static_cast<uintptr_t>(std::to_integer<uint8_t>(B)) << s;
            return UnmarshalResult<uintptr_t>{x, n + i + 1};
        }
        x |= static_cast<uintptr_t>(std::to_integer<uint8_t>(B) & 0x7f) << s;
        s += 7;
    }
    return Error::Overflow;
}

inline SkipResult skip_varint(std::span<const std::byte> b, std::size_t n) noexcept {
    for (std::size_t i = 0; i < MAX_VARINT_LEN; ++i) {
        if (n + i >= b.size()) return Error::BufferTooSmall;
        if (std::to_integer<uint8_t>(b[n + i]) < 0x80) {
            if (i == MAX_VARINT_LEN - 1 && std::to_integer<uint8_t>(b[n + i]) > ((1u << (sizeof(uintptr_t) * 8 - 7 * i)) - 1)) {
                return Error::Overflow;
            }
            return n + i + 1;
        }
    }
    return Error::Overflow;
}
// --- Fixed-size Primitives (Little-Endian) ---

template<typename T>
inline void encode_little_endian(std::span<std::byte> b, std::size_t n, T val) noexcept {
    assert(b.size() >= n + sizeof(T) && "encode_little_endian: buffer too small");
    uint64_t bits;
    if constexpr (std::is_floating_point_v<T>) {
        static_assert(sizeof(bits) >= sizeof(T));
        bits = 0;
        std::memcpy(&bits, &val, sizeof(T));
    } else {
        bits = static_cast<uint64_t>(static_cast<std::make_unsigned_t<T>>(val));
    }
    for (std::size_t i = 0; i < sizeof(T); ++i) {
        b[n + i] = static_cast<std::byte>(bits >> (i * 8));
    }
}

template<>
inline void encode_little_endian<bool>(std::span<std::byte> b, std::size_t n, bool val) noexcept {
    assert(b.size() >= n + 1 && "encode_little_endian<bool>: buffer too small");
    b[n] = static_cast<std::byte>(val ? 1 : 0);
}

template<typename T>
inline T decode_little_endian(std::span<const std::byte> b, std::size_t n) noexcept {
    assert(b.size() >= n + sizeof(T) && "decode_little_endian: buffer too small");
    uint64_t bits = 0;
    for (std::size_t i = 0; i < sizeof(T); ++i) {
        bits |= static_cast<uint64_t>(std::to_integer<uint8_t>(b[n + i])) << (i * 8);
    }
    if constexpr (std::is_floating_point_v<T>) {
         T val;
         std::memcpy(&val, &bits, sizeof(T));
         return val;
    }
    return static_cast<T>(bits);
}

template <typename F, typename... Args>
struct is_callable_with_args {
    template <typename U, typename = decltype(std::declval<U>()(std::declval<Args>()...))>
    static std::true_type test(int);
    template <typename>
    static std::false_type test(...);
    static constexpr bool value = decltype(test<F>(0))::value;
};

} // namespace detail


// --- Public API ---

// --- Fixed-size Primitives ---

#define BSTD_DEFINE_FIXED_TYPE_FUNCS(TypeName, CppType) \
    inline SkipResult skip_##TypeName(std::span<const std::byte> b, std::size_t n) noexcept { \
        if (b.size() < n + sizeof(CppType)) return Error::BufferTooSmall; \
        return n + sizeof(CppType); \
    } \
    constexpr std::size_t size_##TypeName() noexcept { return sizeof(CppType); } \
    inline std::size_t marshal_##TypeName(std::span<std::byte> b, std::size_t n, CppType v) noexcept { \
        detail::encode_little_endian(b, n, v); \
        return n + sizeof(CppType); \
    } \
    inline Result<CppType> unmarshal_##TypeName(std::span<const std::byte> b, std::size_t n) noexcept { \
        if (b.size() < n + sizeof(CppType)) return Result<CppType>{Error::BufferTooSmall}; \
        return UnmarshalResult<CppType>{detail::decode_little_endian<CppType>(b, n), n + sizeof(CppType)}; \
    }

BSTD_DEFINE_FIXED_TYPE_FUNCS(uint64, uint64_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(uint32, uint32_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(uint16, uint16_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(uint8,  uint8_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(byte,   std::byte)
BSTD_DEFINE_FIXED_TYPE_FUNCS(int64,  int64_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(int32,  int32_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(int16,  int16_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(int8,   int8_t)
BSTD_DEFINE_FIXED_TYPE_FUNCS(float64,double)
BSTD_DEFINE_FIXED_TYPE_FUNCS(float32,float)

#undef BSTD_DEFINE_FIXED_TYPE_FUNCS

inline SkipResult skip_bool(std::span<const std::byte> b, std::size_t n) noexcept {
    if (!detail::has_bytes(b, n, 1)) return Error::BufferTooSmall;
    if (std::to_integer<uint8_t>(b[n]) > 1) return Error::InvalidData;
    return n + 1;
}
constexpr std::size_t size_bool() noexcept { return 1; }
inline std::size_t marshal_bool(std::span<std::byte> b, std::size_t n, bool v) noexcept {
    detail::encode_little_endian<bool>(b, n, v);
    return n + 1;
}
inline Result<bool> unmarshal_bool(std::span<const std::byte> b, std::size_t n) noexcept {
    auto offset = skip_bool(b, n);
    if (auto* err = std::get_if<Error>(&offset)) return *err;
    return UnmarshalResult<bool>{b[n] == std::byte{1}, n + 1};
}

// --- Varint Primitives (platform-dependent size) ---

inline SkipResult skip_varint(std::span<const std::byte> b, std::size_t n) noexcept { return detail::skip_varint(b, n); }
inline std::size_t size_uintptr(uintptr_t v) noexcept { return detail::size_uint_var(v); }
inline std::size_t marshal_uintptr(std::span<std::byte> b, std::size_t n, uintptr_t v) noexcept { return detail::marshal_uint_var(b, n, v); }
inline Result<uintptr_t> unmarshal_uintptr(std::span<const std::byte> b, std::size_t n) noexcept { return detail::unmarshal_uint_var(b, n); }

inline std::size_t size_intptr(intptr_t v) noexcept { return detail::size_uint_var(detail::encode_zigzag(v)); }
inline std::size_t marshal_intptr(std::span<std::byte> b, std::size_t n, intptr_t v) noexcept { return detail::marshal_uint_var(b, n, detail::encode_zigzag(v)); }
inline Result<intptr_t> unmarshal_intptr(std::span<const std::byte> b, std::size_t n) noexcept {
    auto res = detail::unmarshal_uint_var(b, n);
    if (auto* val = std::get_if<UnmarshalResult<uintptr_t>>(&res)) {
        return UnmarshalResult<intptr_t>{detail::decode_zigzag(val->value), val->new_offset};
    }
    return std::get<Error>(res);
}


// --- String ---

inline std::size_t size_string(std::string_view str) noexcept {
    return detail::size_uint_var(str.length()) + str.length();
}

inline std::size_t marshal_string(std::span<std::byte> b, std::size_t n, std::string_view str) {
    if (!detail::valid_utf8({reinterpret_cast<const std::byte*>(str.data()), str.size()}))
        throw std::invalid_argument("benc: string is not valid UTF-8");
    assert(b.size() >= n + size_string(str) && "marshal_string: buffer too small");
    n = detail::marshal_uint_var(b, n, str.length());
    std::memcpy(b.data() + n, str.data(), str.length());
    return n + str.length();
}

inline Result<std::string> unmarshal_string(std::span<const std::byte> b, std::size_t n) noexcept {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    if (len > detail::MAX_DECODE_BYTES) return Error::DecodeLimit;
    if (!detail::has_bytes(b, offset, len)) return Error::BufferTooSmall;
    if (!detail::valid_utf8(b.subspan(offset, len))) return Error::InvalidData;
    std::string str(reinterpret_cast<const char*>(b.data() + offset), static_cast<std::size_t>(len));
    return UnmarshalResult<std::string>{std::move(str), offset + static_cast<std::size_t>(len)};
}

inline Result<std::string_view> unmarshal_string_view(std::span<const std::byte> b, std::size_t n) noexcept {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    if (len > detail::MAX_DECODE_BYTES) return Error::DecodeLimit;
    if (!detail::has_bytes(b, offset, len)) return Error::BufferTooSmall;
    if (!detail::valid_utf8(b.subspan(offset, len))) return Error::InvalidData;
    std::string_view sv(reinterpret_cast<const char*>(b.data() + offset), static_cast<std::size_t>(len));
    return UnmarshalResult<std::string_view>{sv, offset + static_cast<std::size_t>(len)};
}

inline SkipResult skip_string(std::span<const std::byte> b, std::size_t n) noexcept {
    auto res = unmarshal_string_view(b, n);
    if (auto* err = std::get_if<Error>(&res)) return *err;
    return std::get<UnmarshalResult<std::string_view>>(res).new_offset;
}

// --- Bytes ---

inline std::size_t size_bytes(std::span<const std::byte> bs) noexcept {
    return detail::size_uint_var(bs.size()) + bs.size();
}

inline std::size_t marshal_bytes(std::span<std::byte> b, std::size_t n, std::span<const std::byte> bs) noexcept {
    assert(b.size() >= n + size_bytes(bs) && "marshal_bytes: buffer too small");
    n = detail::marshal_uint_var(b, n, bs.size());
    std::memcpy(b.data() + n, bs.data(), bs.size());
    return n + bs.size();
}

inline Result<std::vector<std::byte>> unmarshal_bytes_copied(std::span<const std::byte> b, std::size_t n) noexcept {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    if (len > detail::MAX_DECODE_BYTES) return Error::DecodeLimit;
    if (!detail::has_bytes(b, offset, len)) return Error::BufferTooSmall;
    return UnmarshalResult<std::vector<std::byte>>{{b.begin() + offset, b.begin() + offset + len}, offset + len};
}

inline Result<std::span<const std::byte>> unmarshal_bytes_cropped(std::span<const std::byte> b, std::size_t n) noexcept {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    if (len > detail::MAX_DECODE_BYTES) return Error::DecodeLimit;
    if (!detail::has_bytes(b, offset, len)) return Error::BufferTooSmall;
    return UnmarshalResult<std::span<const std::byte>>{b.subspan(offset, len), offset + len};
}

inline SkipResult skip_bytes(std::span<const std::byte> b, std::size_t n) noexcept {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    if (len > detail::MAX_DECODE_BYTES) return Error::DecodeLimit;
    if (!detail::has_bytes(b, offset, len)) return Error::BufferTooSmall;
    return offset + static_cast<std::size_t>(len);
}

// --- Slice ---
template<typename Skipper>
SkipResult skip_slice(std::span<const std::byte> b, std::size_t n, Skipper element_skipper) {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    n = offset;
    if (len > detail::MAX_DECODE_ELEMENTS) return Error::DecodeLimit;
    if (!detail::has_bytes(b, n, 4) || len > b.size() - n - 4) return Error::BufferTooSmall;

    for (uintptr_t i = 0; i < len; ++i) {
        auto skip_res = element_skipper(b, n);
        if (auto* err = std::get_if<Error>(&skip_res)) return *err;
        n = std::get<std::size_t>(skip_res);
    }
    
    auto err = detail::check_terminator(b, n);
    if (err != Error::Ok) return err;
    return n + 4;
}

template<typename T, typename Sizer>
std::size_t size_slice(const std::vector<T>& slice, Sizer sizer) {
    std::size_t s = detail::size_uint_var(slice.size()) + 4; // Length + terminator
    for (const auto& t : slice) {
        if constexpr (detail::is_callable_with_args<Sizer, const T&>::value) {
            s += sizer(t);
        } else {
            s += sizer();
        }
    }
    return s;
}

template<typename T>
std::size_t size_fixed_slice(const std::vector<T>& slice, std::size_t elem_size) noexcept {
    return detail::size_uint_var(slice.size()) + 4 + (slice.size() * elem_size);
}

template<typename T, typename Marshaler>
std::size_t marshal_slice(std::span<std::byte> b, std::size_t n, const std::vector<T>& slice, Marshaler marshaler) {
    n = detail::marshal_uint_var(b, n, slice.size());
    for (const auto& t : slice) {
        n = marshaler(b, n, t);
    }
    assert(b.size() >= n + 4 && "marshal_slice: buffer too small for terminator");
    std::memcpy(b.data() + n, detail::SLICE_MAP_TERMINATOR, 4);
    return n + 4;
}

template<typename T, typename Unmarshaler>
Result<std::vector<T>> unmarshal_slice(std::span<const std::byte> b, std::size_t n, Unmarshaler unmarshaler) {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    n = offset;
    if (len > detail::MAX_DECODE_ELEMENTS) return Error::DecodeLimit;
    if (!detail::has_bytes(b, n, 4) || len > b.size() - n - 4) return Error::BufferTooSmall;

    std::vector<T> ts;

    for (uintptr_t i = 0; i < len; ++i) {
        auto item_res = unmarshaler(b, n);
        if (auto* err = std::get_if<Error>(&item_res)) return *err;
        auto& [item, item_offset] = std::get<UnmarshalResult<T>>(item_res);
        ts.push_back(std::move(item));
        n = item_offset;
    }

    auto err = detail::check_terminator(b, n);
    if (err != Error::Ok) return err;
    return UnmarshalResult<std::vector<T>>{std::move(ts), n + 4};
}


// --- Map ---
template<typename KeySkipper, typename ValueSkipper>
SkipResult skip_map(std::span<const std::byte> b, std::size_t n, KeySkipper key_skipper, ValueSkipper value_skipper) {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    n = offset;
    if (len > detail::MAX_DECODE_ELEMENTS) return Error::DecodeLimit;
    if (!detail::has_bytes(b, n, 4) || len > (b.size() - n - 4) / 2) return Error::BufferTooSmall;

    for (uintptr_t i = 0; i < len; ++i) {
        auto key_skip_res = key_skipper(b, n);
        if (auto* err = std::get_if<Error>(&key_skip_res)) return *err;
        n = std::get<std::size_t>(key_skip_res);

        auto val_skip_res = value_skipper(b, n);
        if (auto* err = std::get_if<Error>(&val_skip_res)) return *err;
        n = std::get<std::size_t>(val_skip_res);
    }
    
    auto err = detail::check_terminator(b, n);
    if (err != Error::Ok) return err;
    return n + 4;
}

template<typename K, typename V, typename KSizer, typename VSizer>
std::size_t size_map(const std::map<K, V>& m, KSizer k_sizer, VSizer v_sizer) {
    std::size_t s = detail::size_uint_var(m.size()) + 4; // Length + terminator
    for (const auto& [key, val] : m) {
        if constexpr (detail::is_callable_with_args<KSizer, const K&>::value) {
            s += k_sizer(key);
        } else {
            s += k_sizer();
        }
        if constexpr (detail::is_callable_with_args<VSizer, const V&>::value) {
            s += v_sizer(val);
        } else {
            s += v_sizer();
        }
    }
    return s;
}

template<typename K, typename V, typename KMarshaler, typename VMarshaler>
std::size_t marshal_map(std::span<std::byte> b, std::size_t n, const std::map<K, V>& m, KMarshaler k_marshaler, VMarshaler v_marshaler) {
    n = detail::marshal_uint_var(b, n, m.size());
    for (const auto& [key, val] : m) {
        n = k_marshaler(b, n, key);
        n = v_marshaler(b, n, val);
    }
    assert(b.size() >= n + 4 && "marshal_map: buffer too small for terminator");
    std::memcpy(b.data() + n, detail::SLICE_MAP_TERMINATOR, 4);
    return n + 4;
}

template<typename K, typename V, typename KUnmarshaler, typename VUnmarshaler>
Result<std::map<K, V>> unmarshal_map(std::span<const std::byte> b, std::size_t n, KUnmarshaler k_unmarshaler, VUnmarshaler v_unmarshaler) {
    auto len_res = detail::unmarshal_uint_var(b, n);
    if (auto* err = std::get_if<Error>(&len_res)) return *err;
    auto& [len, offset] = std::get<UnmarshalResult<uintptr_t>>(len_res);
    n = offset;
    if (len > detail::MAX_DECODE_ELEMENTS) return Error::DecodeLimit;
    if (!detail::has_bytes(b, n, 4) || len > (b.size() - n - 4) / 2) return Error::BufferTooSmall;

    std::map<K, V> m;
    for (uintptr_t i = 0; i < len; ++i) {
        auto key_res = k_unmarshaler(b, n);
        if (auto* err = std::get_if<Error>(&key_res)) return *err;
        auto& [key, key_offset] = std::get<UnmarshalResult<K>>(key_res);
        n = key_offset;

        auto val_res = v_unmarshaler(b, n);
        if (auto* err = std::get_if<Error>(&val_res)) return *err;
        auto& [val, val_offset] = std::get<UnmarshalResult<V>>(val_res);
        n = val_offset;
        
        if (!m.emplace(std::move(key), std::move(val)).second) return Error::InvalidData;
    }
    
    auto err = detail::check_terminator(b, n);
    if (err != Error::Ok) return err;
    return UnmarshalResult<std::map<K, V>>{std::move(m), n + 4};
}

// --- Time ---

using time_point = std::chrono::time_point<std::chrono::system_clock, std::chrono::nanoseconds>;

constexpr std::size_t size_time() noexcept { return size_int64(); }

inline std::size_t marshal_time(std::span<std::byte> b, std::size_t n, time_point t) noexcept {
    return marshal_int64(b, n, t.time_since_epoch().count());
}

inline Result<time_point> unmarshal_time(std::span<const std::byte> b, std::size_t n) noexcept {
    auto res = unmarshal_int64(b, n);
    if (auto* val = std::get_if<UnmarshalResult<int64_t>>(&res)) {
        time_point t{std::chrono::nanoseconds(val->value)};
        return UnmarshalResult<time_point>{t, val->new_offset};
    }
    return std::get<Error>(res);
}

inline SkipResult skip_time(std::span<const std::byte> b, std::size_t n) noexcept {
    return skip_int64(b, n);
}


// --- Pointer (Nullable) ---

template<typename T, typename Sizer>
std::size_t size_pointer(const std::optional<T>& opt, Sizer sizer) {
    std::size_t s = size_bool();
    if (opt.has_value()) {
         if constexpr (detail::is_callable_with_args<Sizer, const T&>::value) {
            s += sizer(*opt);
        } else {
            s += sizer();
        }
    }
    return s;
}

template<typename T, typename Marshaler>
std::size_t marshal_pointer(std::span<std::byte> b, std::size_t n, const std::optional<T>& opt, Marshaler marshaler) {
    n = marshal_bool(b, n, opt.has_value());
    if (opt.has_value()) {
        n = marshaler(b, n, *opt);
    }
    return n;
}

template<typename T, typename Unmarshaler>
Result<std::optional<T>> unmarshal_pointer(std::span<const std::byte> b, std::size_t n, Unmarshaler unmarshaler) {
    auto has_value_res = unmarshal_bool(b, n);
    if (auto* err = std::get_if<Error>(&has_value_res)) return *err;
    
    auto& [has_value, offset] = std::get<UnmarshalResult<bool>>(has_value_res);
    n = offset;

    if (has_value) {
        auto item_res = unmarshaler(b, n);
        if (auto* err = std::get_if<Error>(&item_res)) return *err;
        auto& [item, item_offset] = std::get<UnmarshalResult<T>>(item_res);
        return UnmarshalResult<std::optional<T>>{std::move(item), item_offset};
    } else {
        return UnmarshalResult<std::optional<T>>{std::nullopt, n};
    }
}

template<typename Skipper>
SkipResult skip_pointer(std::span<const std::byte> b, std::size_t n, Skipper skipper) {
    auto has_value_res = unmarshal_bool(b, n);
    if (auto* err = std::get_if<Error>(&has_value_res)) return *err;

    auto& [has_value, offset] = std::get<UnmarshalResult<bool>>(has_value_res);
    n = offset;

    if (has_value) {
        return skipper(b, n);
    }
    return n;
}


} // namespace bstd

