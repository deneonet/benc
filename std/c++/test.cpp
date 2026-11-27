// To compile and run this test (requires C++20):
// g++ -std=c++20 -o std_test std.hpp.test.cpp && ./std_test

#include <iostream>
#include <vector>
#include <string>
#include <map>
#include <cmath>
#include <limits>
#include <random>

#include "std.hpp"

void test_strict_wire_values() {
    using namespace bstd;
    const std::byte bad_bool[] = {std::byte{2}};
    const std::byte bad_string[] = {std::byte{1}, std::byte{0xff}};
    const std::byte bad_collection[] = {std::byte{0}, std::byte{0}, std::byte{0}, std::byte{0}, std::byte{0}};
    const std::byte huge_count[] = {std::byte{0x80}, std::byte{0x80}, std::byte{0x80}, std::byte{0x80}, std::byte{0x10}};
    if (std::get<Error>(unmarshal_bool(bad_bool, 0)) != Error::InvalidData) std::abort();
    if (std::get<Error>(unmarshal_string(bad_string, 0)) != Error::InvalidData) std::abort();
    if (std::get<Error>(skip_string(bad_string, 0)) != Error::InvalidData) std::abort();
    auto skip_byte = [](std::span<const std::byte> b, std::size_t n) { return bstd::skip_byte(b, n); };
    if (std::get<Error>(skip_slice(bad_collection, 0, skip_byte)) != Error::InvalidData) std::abort();
    if (std::get<Error>(skip_slice(huge_count, 0, skip_byte)) != Error::DecodeLimit) std::abort();
    const std::byte duplicate_map[] = {std::byte{2}, std::byte{1}, std::byte{2}, std::byte{1}, std::byte{3}, std::byte{1}, std::byte{1}, std::byte{1}, std::byte{1}};
    auto result = unmarshal_map<uint8_t, uint8_t>(duplicate_map, 0, unmarshal_uint8, unmarshal_uint8);
    if (std::get<Error>(result) != Error::InvalidData) std::abort();
}

//-///////////////////////////////////////////////////////////////////////////
// Minimal Testing Framework
//-///////////////////////////////////////////////////////////////////////////

int tests_run = 0;
int tests_failed = 0;

#define FAIL(msg, ...) \
    do { \
        fprintf(stderr, "FAIL (%s:%d): " msg "\n", __func__, __LINE__, ##__VA_ARGS__); \
        tests_failed++; \
    } while (0)

#define ASSERT(condition) \
    do { \
        if (!(condition)) { \
            FAIL("Assertion failed: %s", #condition); \
        } \
    } while (0)

#define RUN_TEST(test) \
    do { \
        fprintf(stdout, "--- Running %s ---\n", #test); \
        test(); \
        tests_run++; \
    } while (0)


//-///////////////////////////////////////////////////////////////////////////
// Test Cases
//-///////////////////////////////////////////////////////////////////////////

void test_data_types() {
    // --- Setup Test Values ---
    const bool val_bool = true;
    const std::byte val_byte{128};
    const float val_f32 = 123.456f;
    const double val_f64 = 789.0123;
    const intptr_t val_neg_int = -12345;
    const intptr_t val_pos_int = std::numeric_limits<intptr_t>::max();
    const int8_t val_i8 = -1;
    const int16_t val_i16 = -1234;
    const int32_t val_i32 = 123456;
    const int64_t val_i64 = -123456789012LL;
    const uintptr_t val_uint = std::numeric_limits<uintptr_t>::max();
    const uint16_t val_u16 = 65000;
    const uint32_t val_u32 = 4000000000U;
    const uint64_t val_u64 = 18000000000000000000ULL;
    const std::string val_str = "Hello World!";
    const std::vector<std::byte> val_bs = {
        std::byte{0}, std::byte{1}, std::byte{2}, std::byte{3}, std::byte{4},
        std::byte{5}, std::byte{6}, std::byte{7}, std::byte{8}, std::byte{9}
    };

    // --- Sizing ---
    size_t total_size = 0;
    total_size += bstd::size_bool();
    total_size += bstd::size_byte();
    total_size += bstd::size_float32();
    total_size += bstd::size_float64();
    total_size += bstd::size_intptr(val_pos_int);
    total_size += bstd::size_intptr(val_neg_int);
    total_size += bstd::size_int8();
    total_size += bstd::size_int16();
    total_size += bstd::size_int32();
    total_size += bstd::size_int64();
    total_size += bstd::size_uintptr(val_uint);
    total_size += bstd::size_uint16();
    total_size += bstd::size_uint32();
    total_size += bstd::size_uint64();
    total_size += bstd::size_string(val_str);
    total_size += bstd::size_bytes(val_bs);

    // --- Marshaling ---
    std::vector<std::byte> buffer(total_size);
    size_t offset = 0;
    offset = bstd::marshal_bool(buffer, offset, val_bool);
    offset = bstd::marshal_byte(buffer, offset, val_byte);
    offset = bstd::marshal_float32(buffer, offset, val_f32);
    offset = bstd::marshal_float64(buffer, offset, val_f64);
    offset = bstd::marshal_intptr(buffer, offset, val_pos_int);
    offset = bstd::marshal_intptr(buffer, offset, val_neg_int);
    offset = bstd::marshal_int8(buffer, offset, val_i8);
    offset = bstd::marshal_int16(buffer, offset, val_i16);
    offset = bstd::marshal_int32(buffer, offset, val_i32);
    offset = bstd::marshal_int64(buffer, offset, val_i64);
    offset = bstd::marshal_uintptr(buffer, offset, val_uint);
    offset = bstd::marshal_uint16(buffer, offset, val_u16);
    offset = bstd::marshal_uint32(buffer, offset, val_u32);
    offset = bstd::marshal_uint64(buffer, offset, val_u64);
    offset = bstd::marshal_string(buffer, offset, val_str);
    offset = bstd::marshal_bytes(buffer, offset, val_bs);
    ASSERT(offset == total_size);

    // --- Skipping ---
    offset = 0;
    bstd::SkipResult skip_res;
    skip_res = bstd::skip_bool(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_byte(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_float32(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_float64(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_varint(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_varint(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_int8(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_int16(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_int32(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_int64(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_varint(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_uint16(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_uint32(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_uint64(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_string(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    skip_res = bstd::skip_bytes(buffer, offset); ASSERT(std::holds_alternative<size_t>(skip_res)); offset = std::get<size_t>(skip_res);
    ASSERT(offset == total_size);

    // --- Unmarshaling ---
    offset = 0;
    auto res_bool = bstd::unmarshal_bool(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<bool>>(res_bool).value == val_bool); offset = std::get<bstd::UnmarshalResult<bool>>(res_bool).new_offset;
    auto res_byte = bstd::unmarshal_byte(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<std::byte>>(res_byte).value == val_byte); offset = std::get<bstd::UnmarshalResult<std::byte>>(res_byte).new_offset;
    auto res_f32 = bstd::unmarshal_float32(buffer, offset); ASSERT(std::abs(std::get<bstd::UnmarshalResult<float>>(res_f32).value - val_f32) < 1e-6); offset = std::get<bstd::UnmarshalResult<float>>(res_f32).new_offset;
    auto res_f64 = bstd::unmarshal_float64(buffer, offset); ASSERT(std::abs(std::get<bstd::UnmarshalResult<double>>(res_f64).value - val_f64) < 1e-9); offset = std::get<bstd::UnmarshalResult<double>>(res_f64).new_offset;
    auto res_pos_int = bstd::unmarshal_intptr(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<intptr_t>>(res_pos_int).value == val_pos_int); offset = std::get<bstd::UnmarshalResult<intptr_t>>(res_pos_int).new_offset;
    auto res_neg_int = bstd::unmarshal_intptr(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<intptr_t>>(res_neg_int).value == val_neg_int); offset = std::get<bstd::UnmarshalResult<intptr_t>>(res_neg_int).new_offset;
    auto res_i8 = bstd::unmarshal_int8(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<int8_t>>(res_i8).value == val_i8); offset = std::get<bstd::UnmarshalResult<int8_t>>(res_i8).new_offset;
    auto res_i16 = bstd::unmarshal_int16(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<int16_t>>(res_i16).value == val_i16); offset = std::get<bstd::UnmarshalResult<int16_t>>(res_i16).new_offset;
    auto res_i32 = bstd::unmarshal_int32(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<int32_t>>(res_i32).value == val_i32); offset = std::get<bstd::UnmarshalResult<int32_t>>(res_i32).new_offset;
    auto res_i64 = bstd::unmarshal_int64(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<int64_t>>(res_i64).value == val_i64); offset = std::get<bstd::UnmarshalResult<int64_t>>(res_i64).new_offset;
    auto res_uint = bstd::unmarshal_uintptr(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<uintptr_t>>(res_uint).value == val_uint); offset = std::get<bstd::UnmarshalResult<uintptr_t>>(res_uint).new_offset;
    auto res_u16 = bstd::unmarshal_uint16(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<uint16_t>>(res_u16).value == val_u16); offset = std::get<bstd::UnmarshalResult<uint16_t>>(res_u16).new_offset;
    auto res_u32 = bstd::unmarshal_uint32(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<uint32_t>>(res_u32).value == val_u32); offset = std::get<bstd::UnmarshalResult<uint32_t>>(res_u32).new_offset;
    auto res_u64 = bstd::unmarshal_uint64(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<uint64_t>>(res_u64).value == val_u64); offset = std::get<bstd::UnmarshalResult<uint64_t>>(res_u64).new_offset;
    auto res_str = bstd::unmarshal_string(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<std::string>>(res_str).value == val_str); offset = std::get<bstd::UnmarshalResult<std::string>>(res_str).new_offset;
    auto res_bs = bstd::unmarshal_bytes_copied(buffer, offset); ASSERT(std::get<bstd::UnmarshalResult<std::vector<std::byte>>>(res_bs).value == val_bs); offset = std::get<bstd::UnmarshalResult<std::vector<std::byte>>>(res_bs).new_offset;
    ASSERT(offset == total_size);
}

void test_err_buf_too_small() {
    std::vector<std::byte> empty_buf = {};
    std::vector<std::byte> short_buf_1 = {std::byte{0x80}};
    std::vector<std::byte> short_buf_3 = {std::byte{1}, std::byte{2}, std::byte{3}};

    auto res_bool = bstd::unmarshal_bool(empty_buf, 0);
    ASSERT(std::get<bstd::Error>(res_bool) == bstd::Error::BufferTooSmall);

    auto res_i32 = bstd::unmarshal_int32(short_buf_3, 0);
    ASSERT(std::get<bstd::Error>(res_i32) == bstd::Error::BufferTooSmall);

    auto res_varint = bstd::unmarshal_uintptr(short_buf_1, 0);
    ASSERT(std::get<bstd::Error>(res_varint) == bstd::Error::BufferTooSmall);

    auto skip_res_varint = bstd::skip_varint(short_buf_1, 0);
    ASSERT(std::get<bstd::Error>(skip_res_varint) == bstd::Error::BufferTooSmall);
}

void test_slices() {
    std::vector<std::string> slice = {"slice_element_1", "slice_element_2", "slice_element_3"};

    auto sizer = [](const std::string& s) { return bstd::size_string(s); };
    auto marshaler = [](std::span<std::byte> b, size_t n, const std::string& s) { return bstd::marshal_string(b, n, s); };
    auto unmarshaler = [](std::span<const std::byte> b, size_t n) { return bstd::unmarshal_string(b, n); };
    auto skipper = [](std::span<const std::byte> b, size_t n) { return bstd::skip_string(b, n); };

    size_t total_size = bstd::size_slice(slice, sizer);
    std::vector<std::byte> buffer(total_size);

    size_t offset = bstd::marshal_slice(buffer, 0, slice, marshaler);
    ASSERT(offset == total_size);

    auto skip_res = bstd::skip_slice(buffer, 0, skipper);
    ASSERT(std::get<size_t>(skip_res) == total_size);

    auto unmarshal_res = bstd::unmarshal_slice<std::string>(buffer, 0, unmarshaler);
    ASSERT(std::get<bstd::UnmarshalResult<std::vector<std::string>>>(unmarshal_res).value == slice);
}

void test_maps() {
    std::map<std::string, int32_t> test_map = {
        {"key1", 100},
        {"key2", -200},
        {"key3", 300}
    };

    auto key_sizer = [](const std::string& s) { return bstd::size_string(s); };
    auto val_sizer = []() { return bstd::size_int32(); };
    auto key_marshaler = [](std::span<std::byte> b, size_t n, const std::string& s) { return bstd::marshal_string(b, n, s); };
    auto val_marshaler = [](std::span<std::byte> b, size_t n, int32_t v) { return bstd::marshal_int32(b, n, v); };
    auto key_unmarshaler = [](std::span<const std::byte> b, size_t n) { return bstd::unmarshal_string(b, n); };
    auto val_unmarshaler = [](std::span<const std::byte> b, size_t n) { return bstd::unmarshal_int32(b, n); };
    auto key_skipper = [](std::span<const std::byte> b, size_t n) { return bstd::skip_string(b, n); };
    auto val_skipper = [](std::span<const std::byte> b, size_t n) { return bstd::skip_int32(b, n); };

    size_t total_size = bstd::size_map(test_map, key_sizer, val_sizer);
    std::vector<std::byte> buffer(total_size);

    size_t offset = bstd::marshal_map(buffer, 0, test_map, key_marshaler, val_marshaler);
    ASSERT(offset == total_size);

    auto skip_res = bstd::skip_map(buffer, 0, key_skipper, val_skipper);
    ASSERT(std::get<size_t>(skip_res) == total_size);

    auto unmarshal_res = bstd::unmarshal_map<std::string, int32_t>(buffer, 0, key_unmarshaler, val_unmarshaler);
    ASSERT((std::get<bstd::UnmarshalResult<std::map<std::string, int32_t>>>(unmarshal_res).value == test_map));
}

void test_time() {
    bstd::time_point now(std::chrono::nanoseconds(1663362895123456789LL));
    
    size_t size = bstd::size_time();
    std::vector<std::byte> buffer(size);

    size_t offset = bstd::marshal_time(buffer, 0, now);
    ASSERT(offset == size);

    auto skip_res = bstd::skip_time(buffer, 0);
    ASSERT(std::get<size_t>(skip_res) == size);

    auto unmarshal_res = bstd::unmarshal_time(buffer, 0);
    ASSERT(std::get<bstd::UnmarshalResult<bstd::time_point>>(unmarshal_res).value == now);
}

void test_pointer() {
    // Test non-nil pointer
    std::optional<std::string> opt_str = "hello optional";
    auto sizer = [](const std::string& s) { return bstd::size_string(s); };
    auto marshaler = [](std::span<std::byte> b, size_t n, const std::string& s) { return bstd::marshal_string(b, n, s); };
    auto unmarshaler = [](std::span<const std::byte> b, size_t n) { return bstd::unmarshal_string(b, n); };
    auto skipper = [](std::span<const std::byte> b, size_t n) { return bstd::skip_string(b, n); };

    size_t size = bstd::size_pointer(opt_str, sizer);
    std::vector<std::byte> buffer(size);
    
    size_t offset = bstd::marshal_pointer(buffer, 0, opt_str, marshaler);
    ASSERT(offset == size);
    
    auto skip_res = bstd::skip_pointer(buffer, 0, skipper);
    ASSERT(std::get<size_t>(skip_res) == size);
    
    auto unmarshal_res = bstd::unmarshal_pointer<std::string>(buffer, 0, unmarshaler);
    ASSERT(std::get<bstd::UnmarshalResult<std::optional<std::string>>>(unmarshal_res).value == opt_str);

    // Test nil pointer
    std::optional<std::string> nil_opt = std::nullopt;
    size = bstd::size_pointer(nil_opt, sizer);
    buffer.resize(size);

    offset = bstd::marshal_pointer(buffer, 0, nil_opt, marshaler);
    ASSERT(offset == size);

    skip_res = bstd::skip_pointer(buffer, 0, skipper);
    ASSERT(std::get<size_t>(skip_res) == size);

    unmarshal_res = bstd::unmarshal_pointer<std::string>(buffer, 0, unmarshaler);
    ASSERT(std::get<bstd::UnmarshalResult<std::optional<std::string>>>(unmarshal_res).value == nil_opt);
}

void test_varint_overflow() {
    std::vector<std::byte> overflow_buf(bstd::detail::MAX_VARINT_LEN);
    for(size_t i=0; i < overflow_buf.size() - 1; ++i) {
        overflow_buf[i] = std::byte{0x80};
    }
    overflow_buf.back() = std::byte{0x02};

    auto res = bstd::unmarshal_uintptr(overflow_buf, 0);
    ASSERT(std::get<bstd::Error>(res) == bstd::Error::Overflow);
    
    auto skip_res = bstd::skip_varint(overflow_buf, 0);
    ASSERT(std::get<bstd::Error>(skip_res) == bstd::Error::Overflow);
}


//-///////////////////////////////////////////////////////////////////////////
// Main Test Runner
//-///////////////////////////////////////////////////////////////////////////

int main() {
    RUN_TEST(test_data_types);
    RUN_TEST(test_err_buf_too_small);
    RUN_TEST(test_slices);
    RUN_TEST(test_maps);
    RUN_TEST(test_time);
    RUN_TEST(test_pointer);
    RUN_TEST(test_varint_overflow);
    RUN_TEST(test_strict_wire_values);

    if (tests_failed == 0) {
        printf("\nSUCCESS: All %d tests passed.\n", tests_run);
        return 0;
    } else {
        printf("\nFAILURE: %d out of %d tests failed.\n", tests_failed, tests_run);
        return 1;
    }
}

