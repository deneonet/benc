/**
 * @file test.c
 * @brief Merged main test runner for the benc serialization library.
 * This file combines tests from two separate suites into a single file.
 */

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <stdbool.h>
#include <stdint.h>
#include <math.h>
#include <time.h>

// Define the implementation in this file
#define BSTD_IMPLEMENTATION
#include "gen.h" // Assumed to contain generators and TestStruct definition

//-///////////////////////////////////////////////////////////////////////////
// Unified Testing Framework
//-///////////////////////////////////////////////////////////////////////////

static int tests_passed = 0;
static int tests_failed = 0;

#define TEST_ASSERT(condition, msg, ...) \
    do { \
        if (!(condition)) { \
            printf("\n    [FAIL] %s:%d: Assertion failed: " msg, __FILE__, __LINE__, ##__VA_ARGS__); \
            return false; \
        } \
    } while (0)

#define RUN_TEST(test_func) \
    do { \
        printf("- Running %s...", #test_func); \
        fflush(stdout); \
        if (test_func()) { \
            printf("\r- Running %s... [PASS]\n", #test_func); \
            tests_passed++; \
        } else { \
            /* The failing test should have already printed its error message */ \
            printf("\r- Running %s... [FAIL]\n", #test_func); \
            tests_failed++; \
        } \
    } while (0)


//-///////////////////////////////////////////////////////////////////////////
// Tests adapted from the first file (test.c)
//-///////////////////////////////////////////////////////////////////////////

// --- Helper functions for string slices/maps ---
static bstd_status skip_string_element(const uint8_t* buf, size_t buf_len, size_t* offset) {
    return bstd_skip_string(buf, buf_len, offset);
}

static size_t size_string_element(const void* element) {
    const char* str = *(const char**)element;
    return bstd_size_string(str, strlen(str));
}

static bstd_status marshal_string_element(uint8_t* buf, size_t buf_len, size_t* offset, const void* element) {
    const char* str = *(const char**)element;
    return bstd_marshal_string(buf, buf_len, offset, str, strlen(str));
}

static bstd_status unmarshal_string_element(const uint8_t* buf, size_t buf_len, size_t* offset, void* out_element) {
    return bstd_unmarshal_string_alloc(buf, buf_len, offset, (char**)out_element);
}

static void free_string_element(void* element) {
    free(*(char**)element);
}

bool test_data_types(void) {
    const char* test_str = "Hello World!";
    const uint8_t test_bs[] = {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10};
    const intptr_t neg_int_val = -12345;
    const intptr_t pos_int_val = INTPTR_MAX;

    size_t total_size = 0;
    total_size += bstd_size_bool();
    total_size += bstd_size_byte();
    total_size += bstd_size_float32();
    total_size += bstd_size_float64();
    total_size += bstd_size_int(pos_int_val);
    total_size += bstd_size_int(neg_int_val);
    total_size += bstd_size_int8();
    total_size += bstd_size_int16();
    total_size += bstd_size_int32();
    total_size += bstd_size_int64();
    total_size += bstd_size_uint(UINTPTR_MAX);
    total_size += bstd_size_uint16();
    total_size += bstd_size_uint32();
    total_size += bstd_size_uint64();
    total_size += bstd_size_string(test_str, strlen(test_str));
    total_size += bstd_size_bytes(test_bs, sizeof(test_bs));

    uint8_t* buf = (uint8_t*)malloc(total_size);
    TEST_ASSERT(buf != NULL, "malloc failed");

    size_t offset = 0;
    bstd_marshal_bool(buf, total_size, &offset, true);
    bstd_marshal_byte(buf, total_size, &offset, 128);
    bstd_marshal_float32(buf, total_size, &offset, 123.456f);
    bstd_marshal_float64(buf, total_size, &offset, 789.0123);
    bstd_marshal_int(buf, total_size, &offset, pos_int_val);
    bstd_marshal_int(buf, total_size, &offset, neg_int_val);
    bstd_marshal_int8(buf, total_size, &offset, -1);
    bstd_marshal_int16(buf, total_size, &offset, -1234);
    bstd_marshal_int32(buf, total_size, &offset, 123456);
    bstd_marshal_int64(buf, total_size, &offset, -123456789012LL);
    bstd_marshal_uint(buf, total_size, &offset, UINTPTR_MAX);
    bstd_marshal_uint16(buf, total_size, &offset, 65000);
    bstd_marshal_uint32(buf, total_size, &offset, 4000000000U);
    bstd_marshal_uint64(buf, total_size, &offset, 18000000000000000000ULL);
    bstd_marshal_string(buf, total_size, &offset, test_str, strlen(test_str));
    bstd_marshal_bytes(buf, total_size, &offset, test_bs, sizeof(test_bs));
    TEST_ASSERT(offset == total_size, "offset mismatch after marshaling");

    // Test Skip functions
    offset = 0;
    TEST_ASSERT(bstd_skip_bool(buf, total_size, &offset) == BSTD_OK, "skip bool");
    TEST_ASSERT(bstd_skip_byte(buf, total_size, &offset) == BSTD_OK, "skip byte");
    TEST_ASSERT(bstd_skip_float32(buf, total_size, &offset) == BSTD_OK, "skip float32");
    TEST_ASSERT(bstd_skip_float64(buf, total_size, &offset) == BSTD_OK, "skip float64");
    TEST_ASSERT(bstd_skip_varint(buf, total_size, &offset) == BSTD_OK, "skip varint pos");
    TEST_ASSERT(bstd_skip_varint(buf, total_size, &offset) == BSTD_OK, "skip varint neg");
    TEST_ASSERT(bstd_skip_int8(buf, total_size, &offset) == BSTD_OK, "skip int8");
    TEST_ASSERT(bstd_skip_int16(buf, total_size, &offset) == BSTD_OK, "skip int16");
    TEST_ASSERT(bstd_skip_int32(buf, total_size, &offset) == BSTD_OK, "skip int32");
    TEST_ASSERT(bstd_skip_int64(buf, total_size, &offset) == BSTD_OK, "skip int64");
    TEST_ASSERT(bstd_skip_varint(buf, total_size, &offset) == BSTD_OK, "skip varint uint");
    TEST_ASSERT(bstd_skip_uint16(buf, total_size, &offset) == BSTD_OK, "skip uint16");
    TEST_ASSERT(bstd_skip_uint32(buf, total_size, &offset) == BSTD_OK, "skip uint32");
    TEST_ASSERT(bstd_skip_uint64(buf, total_size, &offset) == BSTD_OK, "skip uint64");
    TEST_ASSERT(bstd_skip_string(buf, total_size, &offset) == BSTD_OK, "skip string");
    TEST_ASSERT(bstd_skip_bytes(buf, total_size, &offset) == BSTD_OK, "skip bytes");
    TEST_ASSERT(offset == total_size, "offset mismatch after skipping");

    // Test Unmarshal functions
    offset = 0;
    bool v_bool;
    uint8_t v_byte;
    float v_f32;
    double v_f64;
    intptr_t v_pos_int, v_neg_int;
    int8_t v_i8;
    int16_t v_i16;
    int32_t v_i32;
    int64_t v_i64;
    uintptr_t v_uint;
    uint16_t v_u16;
    uint32_t v_u32;
    uint64_t v_u64;
    char* v_str;
    uint8_t* v_bs;
    size_t v_bs_len;

    TEST_ASSERT(bstd_unmarshal_bool(buf, total_size, &offset, &v_bool) == BSTD_OK && v_bool == true, "unmarshal bool");
    TEST_ASSERT(bstd_unmarshal_byte(buf, total_size, &offset, &v_byte) == BSTD_OK && v_byte == 128, "unmarshal byte");
    TEST_ASSERT(bstd_unmarshal_float32(buf, total_size, &offset, &v_f32) == BSTD_OK && fabsf(v_f32 - 123.456f) < 1e-6, "unmarshal float32");
    TEST_ASSERT(bstd_unmarshal_float64(buf, total_size, &offset, &v_f64) == BSTD_OK && fabs(v_f64 - 789.0123) < 1e-9, "unmarshal float64");
    TEST_ASSERT(bstd_unmarshal_int(buf, total_size, &offset, &v_pos_int) == BSTD_OK && v_pos_int == pos_int_val, "unmarshal int pos");
    TEST_ASSERT(bstd_unmarshal_int(buf, total_size, &offset, &v_neg_int) == BSTD_OK && v_neg_int == neg_int_val, "unmarshal int neg");
    TEST_ASSERT(bstd_unmarshal_int8(buf, total_size, &offset, &v_i8) == BSTD_OK && v_i8 == -1, "unmarshal int8");
    TEST_ASSERT(bstd_unmarshal_int16(buf, total_size, &offset, &v_i16) == BSTD_OK && v_i16 == -1234, "unmarshal int16");
    TEST_ASSERT(bstd_unmarshal_int32(buf, total_size, &offset, &v_i32) == BSTD_OK && v_i32 == 123456, "unmarshal int32");
    TEST_ASSERT(bstd_unmarshal_int64(buf, total_size, &offset, &v_i64) == BSTD_OK && v_i64 == -123456789012LL, "unmarshal int64");
    TEST_ASSERT(bstd_unmarshal_uint(buf, total_size, &offset, &v_uint) == BSTD_OK && v_uint == UINTPTR_MAX, "unmarshal uint");
    TEST_ASSERT(bstd_unmarshal_uint16(buf, total_size, &offset, &v_u16) == BSTD_OK && v_u16 == 65000, "unmarshal uint16");
    TEST_ASSERT(bstd_unmarshal_uint32(buf, total_size, &offset, &v_u32) == BSTD_OK && v_u32 == 4000000000U, "unmarshal uint32");
    TEST_ASSERT(bstd_unmarshal_uint64(buf, total_size, &offset, &v_u64) == BSTD_OK && v_u64 == 18000000000000000000ULL, "unmarshal uint64");
    TEST_ASSERT(bstd_unmarshal_string_alloc(buf, total_size, &offset, &v_str) == BSTD_OK && strcmp(v_str, test_str) == 0, "unmarshal string");
    TEST_ASSERT(bstd_unmarshal_bytes_alloc(buf, total_size, &offset, &v_bs, &v_bs_len) == BSTD_OK && v_bs_len == sizeof(test_bs) && memcmp(v_bs, test_bs, v_bs_len) == 0, "unmarshal bytes");
    TEST_ASSERT(offset == total_size, "offset mismatch after unmarshaling");

    free(v_str);
    free(v_bs);
    free(buf);
    return true;
}

bool test_err_buf_too_small(void) {
    uint8_t empty_buf[] = {};
    uint8_t short_buf_1[] = {0x80};
    uint8_t short_buf_3[] = {1, 2, 3};
    size_t offset;
    bool v_bool;
    int32_t v_i32;
    intptr_t v_int;
    char* v_str;

    offset = 0;
    TEST_ASSERT(bstd_unmarshal_bool(empty_buf, sizeof(empty_buf), &offset, &v_bool) == BSTD_ERR_BUF_TOO_SMALL, "unmarshal bool from empty");
    offset = 0;
    TEST_ASSERT(bstd_unmarshal_int32(short_buf_3, sizeof(short_buf_3), &offset, &v_i32) == BSTD_ERR_BUF_TOO_SMALL, "unmarshal int32 from short");
    offset = 0;
    TEST_ASSERT(bstd_unmarshal_int(short_buf_1, sizeof(short_buf_1), &offset, &v_int) == BSTD_ERR_BUF_TOO_SMALL, "unmarshal varint from short");
    offset = 0;
    TEST_ASSERT(bstd_unmarshal_string_alloc(short_buf_1, sizeof(short_buf_1), &offset, &v_str) == BSTD_ERR_BUF_TOO_SMALL, "unmarshal string from short");

    offset = 0;
    TEST_ASSERT(bstd_skip_bool(empty_buf, sizeof(empty_buf), &offset) == BSTD_ERR_BUF_TOO_SMALL, "skip bool from empty");
    offset = 0;
    TEST_ASSERT(bstd_skip_int32(short_buf_3, sizeof(short_buf_3), &offset) == BSTD_ERR_BUF_TOO_SMALL, "skip int32 from short");
    offset = 0;
    TEST_ASSERT(bstd_skip_varint(short_buf_1, sizeof(short_buf_1), &offset) == BSTD_ERR_BUF_TOO_SMALL, "skip varint from short");
    offset = 0;
    TEST_ASSERT(bstd_skip_string(short_buf_1, sizeof(short_buf_1), &offset) == BSTD_ERR_BUF_TOO_SMALL, "skip string from short");
    
    return true;
}

bool test_slices(void) {
    const char* slice[] = {"sliceelement1", "sliceelement2", "sliceelement3"};
    size_t count = sizeof(slice) / sizeof(slice[0]);

    size_t total_size = bstd_size_slice(slice, count, sizeof(char*), size_string_element);
    uint8_t* buf = (uint8_t*)malloc(total_size);
    TEST_ASSERT(buf != NULL, "malloc failed");

    size_t offset = 0;
    TEST_ASSERT(bstd_marshal_slice(buf, total_size, &offset, slice, count, sizeof(char*), marshal_string_element) == BSTD_OK, "marshal slice");
    TEST_ASSERT(offset == total_size, "offset mismatch after marshal");

    offset = 0;
    TEST_ASSERT(bstd_skip_slice(buf, total_size, &offset, skip_string_element) == BSTD_OK, "skip slice");
    TEST_ASSERT(offset == total_size, "offset mismatch after skip");

    char** ret_slice = NULL;
    size_t ret_count = 0;
    offset = 0;
    TEST_ASSERT(bstd_unmarshal_slice_alloc_ex(buf, total_size, &offset, (void**)&ret_slice, &ret_count, sizeof(char*), unmarshal_string_element, free_string_element) == BSTD_OK, "unmarshal slice");
    TEST_ASSERT(offset == total_size, "offset mismatch after unmarshal");
    TEST_ASSERT(ret_count == count, "unmarshaled slice count mismatch");
    for (size_t i = 0; i < count; ++i) {
        TEST_ASSERT(strcmp(slice[i], ret_slice[i]) == 0, "slice element mismatch at index %zu", i);
    }
    
    bstd_free_slice(ret_slice, ret_count, sizeof(char*), free_string_element);
    free(buf);
    return true;
}

bool test_maps(void) {
    const char* keys[] = {"mapkey1", "mapkey2", "mapkey3"};
    const char* values[] = {"mapvalue1", "mapvalue2", "mapvalue3"};
    size_t count = sizeof(keys) / sizeof(keys[0]);

    size_t total_size = bstd_size_map(keys, values, count, sizeof(char*), sizeof(char*), size_string_element, size_string_element);
    uint8_t* buf = (uint8_t*)malloc(total_size);
    TEST_ASSERT(buf != NULL, "malloc failed");

    size_t offset = 0;
    TEST_ASSERT(bstd_marshal_map(buf, total_size, &offset, keys, values, count, sizeof(char*), sizeof(char*), marshal_string_element, marshal_string_element) == BSTD_OK, "marshal map");
    TEST_ASSERT(offset == total_size, "offset mismatch after marshal");

    offset = 0;
    TEST_ASSERT(bstd_skip_map(buf, total_size, &offset, skip_string_element, skip_string_element) == BSTD_OK, "skip map");
    TEST_ASSERT(offset == total_size, "offset mismatch after skip");

    char** ret_keys = NULL;
    char** ret_values = NULL;
    size_t ret_count = 0;
    offset = 0;
    TEST_ASSERT(bstd_unmarshal_map_alloc_ex(buf, total_size, &offset, (void**)&ret_keys, (void**)&ret_values, &ret_count, sizeof(char*), sizeof(char*), unmarshal_string_element, unmarshal_string_element, free_string_element, free_string_element) == BSTD_OK, "unmarshal map");
    TEST_ASSERT(offset == total_size, "offset mismatch after unmarshal");
    TEST_ASSERT(ret_count == count, "unmarshaled map count mismatch");
    
    int found_mask = 0;
    for(size_t i=0; i < ret_count; ++i) {
        for(size_t j=0; j < count; ++j) {
            if (strcmp(ret_keys[i], keys[j]) == 0 && strcmp(ret_values[i], values[j]) == 0) {
                found_mask |= (1 << j);
            }
        }
    }
    TEST_ASSERT(found_mask == (1 << count) - 1, "not all map pairs were found after unmarshaling");

    bstd_free_map(ret_keys, ret_values, ret_count, sizeof(char*), sizeof(char*), free_string_element, free_string_element);
    free(buf);
    return true;
}

bool test_time(void) {
    int64_t now = 1663362895123456789LL;
    size_t size = bstd_size_time();
    uint8_t buf[8];
    TEST_ASSERT(size == 8, "time size should be 8 bytes");

    size_t offset = 0;
    TEST_ASSERT(bstd_marshal_time(buf, size, &offset, now) == BSTD_OK, "marshal time");
    TEST_ASSERT(offset == size, "offset mismatch after marshal time");

    offset = 0;
    TEST_ASSERT(bstd_skip_time(buf, size, &offset) == BSTD_OK, "skip time");
    TEST_ASSERT(offset == size, "offset mismatch after skip time");

    int64_t ret_time;
    offset = 0;
    TEST_ASSERT(bstd_unmarshal_time(buf, size, &offset, &ret_time) == BSTD_OK, "unmarshal time");
    TEST_ASSERT(ret_time == now, "time value mismatch");
    return true;
}

bool test_pointer(void) {
    // Test non-nil pointer
    const char* val = "hello world";
    const char** ptr = &val;

    size_t size = bstd_size_pointer(ptr, size_string_element);
    uint8_t* buf = (uint8_t*)malloc(size);
    TEST_ASSERT(buf != NULL, "malloc failed for non-nil test");

    size_t offset = 0;
    TEST_ASSERT(bstd_marshal_pointer(buf, size, &offset, ptr, marshal_string_element) == BSTD_OK, "marshal non-nil pointer");
    TEST_ASSERT(offset == size, "offset mismatch for non-nil marshal");
    
    offset = 0;
    TEST_ASSERT(bstd_skip_pointer(buf, size, &offset, skip_string_element) == BSTD_OK, "skip non-nil pointer");
    TEST_ASSERT(offset == size, "offset mismatch for non-nil skip");

    char** ret_ptr = NULL;
    offset = 0;
    TEST_ASSERT(bstd_unmarshal_pointer_alloc_ex(buf, size, &offset, (void**)&ret_ptr, sizeof(char*), unmarshal_string_element, free_string_element) == BSTD_OK, "unmarshal non-nil pointer");
    TEST_ASSERT(ret_ptr != NULL, "unmarshaled pointer should not be nil");
    TEST_ASSERT(strcmp(*ret_ptr, val) == 0, "dereferenced pointer value mismatch");
    
    bstd_free_pointer(ret_ptr, free_string_element);
    free(buf);

    // Test nil pointer
    char** nil_ptr = NULL;
    size = bstd_size_pointer(nil_ptr, size_string_element);
    buf = (uint8_t*)malloc(size);
    TEST_ASSERT(buf != NULL, "malloc failed for nil test");
    
    offset = 0;
    TEST_ASSERT(bstd_marshal_pointer(buf, size, &offset, nil_ptr, marshal_string_element) == BSTD_OK, "marshal nil pointer");
    TEST_ASSERT(offset == size, "offset mismatch for nil marshal");
    
    offset = 0;
    TEST_ASSERT(bstd_skip_pointer(buf, size, &offset, skip_string_element) == BSTD_OK, "skip nil pointer");
    TEST_ASSERT(offset == size, "offset mismatch for nil skip");

    offset = 0;
    TEST_ASSERT(bstd_unmarshal_pointer_alloc_ex(buf, size, &offset, (void**)&ret_ptr, sizeof(char*), unmarshal_string_element, free_string_element) == BSTD_OK, "unmarshal nil pointer");
    TEST_ASSERT(ret_ptr == NULL, "unmarshaled pointer should be nil");

    free(buf);
    return true;
}

bool test_varint_overflow(void) {
    uint8_t overflow_buf[MAX_VARINT_LEN + 1];
    memset(overflow_buf, 0x80, sizeof(overflow_buf));
    overflow_buf[MAX_VARINT_LEN - 1] = 0x82; // Malformed last byte
    
    size_t offset = 0;
    uintptr_t val;
    TEST_ASSERT(bstd_unmarshal_uint(overflow_buf, MAX_VARINT_LEN, &offset, &val) == BSTD_ERR_OVERFLOW, "unmarshal should return overflow");

    offset = 0;
    TEST_ASSERT(bstd_skip_varint(overflow_buf, MAX_VARINT_LEN, &offset) == BSTD_ERR_OVERFLOW, "skip should return overflow");
    return true;
}

//-///////////////////////////////////////////////////////////////////////////
// Tests from the second file (test2.c)
//-///////////////////////////////////////////////////////////////////////////

#define RUN_PRIMITIVE_TEST(type_name, c_type, generator) \
bool test_##type_name(void) { \
    c_type src = generator(); \
    c_type dst = 0; \
    size_t size = bstd_size_##type_name(); \
    uint8_t* buf = (uint8_t*)malloc(size); \
    TEST_ASSERT(buf != NULL, "malloc failed"); \
    size_t offset = 0; \
    bstd_status status = bstd_marshal_##type_name(buf, size, &offset, src); \
    TEST_ASSERT(status == BSTD_OK, "Marshal failed"); \
    offset = 0; \
    status = bstd_unmarshal_##type_name(buf, size, &offset, &dst); \
    TEST_ASSERT(status == BSTD_OK, "Unmarshal failed"); \
    TEST_ASSERT(src == dst, "Value mismatch"); \
    free(buf); \
    return true; \
}

#define RUN_VARINT_TEST(type_name, c_type, generator) \
bool test_##type_name(void) { \
    c_type src = generator(); \
    c_type dst = 0; \
    size_t size = bstd_size_##type_name(src); \
    uint8_t* buf = (uint8_t*)malloc(size); \
    TEST_ASSERT(buf != NULL, "malloc failed"); \
    size_t offset = 0; \
    bstd_status status = bstd_marshal_##type_name(buf, size, &offset, src); \
    TEST_ASSERT(status == BSTD_OK, "Marshal failed"); \
    offset = 0; \
    status = bstd_unmarshal_##type_name(buf, size, &offset, &dst); \
    TEST_ASSERT(status == BSTD_OK, "Unmarshal failed"); \
    TEST_ASSERT(src == dst, "Value mismatch"); \
    free(buf); \
    return true; \
}

// Generate test functions for all primitive types
RUN_PRIMITIVE_TEST(bool, bool, generate_bool)
RUN_PRIMITIVE_TEST(uint8, uint8_t, generate_uint8)
RUN_PRIMITIVE_TEST(uint16, uint16_t, generate_uint16)
RUN_PRIMITIVE_TEST(uint32, uint32_t, generate_uint32)
RUN_PRIMITIVE_TEST(uint64, uint64_t, generate_uint64)
RUN_PRIMITIVE_TEST(int8, int8_t, generate_int8)
RUN_PRIMITIVE_TEST(int16, int16_t, generate_int16)
RUN_PRIMITIVE_TEST(int32, int32_t, generate_int32)
RUN_PRIMITIVE_TEST(int64, int64_t, generate_int64)
RUN_VARINT_TEST(uint, uintptr_t, generate_uint)
RUN_VARINT_TEST(int, intptr_t, generate_int)

bool test_float32(void) {
    float src = generate_float32();
    float dst = 0;
    uint8_t buf[4];
    size_t offset = 0;
    bstd_status status = bstd_marshal_float32(buf, sizeof(buf), &offset, src);
    TEST_ASSERT(status == BSTD_OK, "Marshal failed");
    offset = 0;
    status = bstd_unmarshal_float32(buf, sizeof(buf), &offset, &dst);
    TEST_ASSERT(status == BSTD_OK, "Unmarshal failed");
    TEST_ASSERT(src == dst, "Value mismatch");
    return true;
}

bool test_float64(void) {
    double src = generate_float64();
    double dst = 0;
    uint8_t buf[8];
    size_t offset = 0;
    bstd_status status = bstd_marshal_float64(buf, sizeof(buf), &offset, src);
    TEST_ASSERT(status == BSTD_OK, "Marshal failed");
    offset = 0;
    status = bstd_unmarshal_float64(buf, sizeof(buf), &offset, &dst);
    TEST_ASSERT(status == BSTD_OK, "Unmarshal failed");
    TEST_ASSERT(src == dst, "Value mismatch");
    return true;
}

bool test_string(void) {
    char* src = generate_string_alloc();
    char* dst = NULL;
    size_t src_len = strlen(src);
    size_t size = bstd_size_string(src, src_len);
    uint8_t* buf = (uint8_t*)malloc(size);
    TEST_ASSERT(buf != NULL, "malloc failed");
    size_t offset = 0;
    bstd_status status = bstd_marshal_string(buf, size, &offset, src, src_len);
    TEST_ASSERT(status == BSTD_OK, "Marshal failed");
    offset = 0;
    status = bstd_unmarshal_string_alloc(buf, size, &offset, &dst);
    TEST_ASSERT(status == BSTD_OK, "Unmarshal failed");
    TEST_ASSERT(compare_string(src, dst), "String mismatch");
    free(src);
    free(dst);
    free(buf);
    return true;
}

bool test_bytes(void) {
    size_t src_len;
    uint8_t* src = generate_bytes_alloc(&src_len);
    uint8_t* dst = NULL;
    size_t dst_len = 0;
    size_t size = bstd_size_bytes(src, src_len);
    uint8_t* buf = (uint8_t*)malloc(size);
    TEST_ASSERT(buf != NULL, "malloc failed");
    size_t offset = 0;
    bstd_status status = bstd_marshal_bytes(buf, size, &offset, src, src_len);
    TEST_ASSERT(status == BSTD_OK, "Marshal failed");
    offset = 0;
    status = bstd_unmarshal_bytes_alloc(buf, size, &offset, &dst, &dst_len);
    TEST_ASSERT(status == BSTD_OK, "Unmarshal failed");
    TEST_ASSERT(compare_bytes(src, src_len, dst, dst_len), "Bytes mismatch");
    free(src);
    free(dst);
    free(buf);
    return true;
}

bool test_string_view(void) {
    char* src = generate_string_alloc();
    bstd_string_view dst_view;
    size_t src_len = strlen(src);
    size_t size = bstd_size_string(src, src_len);
    uint8_t* buf = (uint8_t*)malloc(size);
    TEST_ASSERT(buf != NULL, "malloc failed");
    size_t offset = 0;
    bstd_marshal_string(buf, size, &offset, src, src_len);
    offset = 0;
    bstd_unmarshal_string_view(buf, size, &offset, &dst_view);
    TEST_ASSERT(src_len == dst_view.len, "Length mismatch");
    TEST_ASSERT(memcmp(src, dst_view.data, src_len) == 0, "Content mismatch");
    free(src);
    free(buf);
    return true;
}

// bstd function pointers for a slice of int32
static size_t sizer_int32(const void* el) { (void)el; return bstd_size_int32(); }
static bstd_status marshal_int32_ptr(uint8_t* buf, size_t buf_len, size_t* offset, const void* el) { return bstd_marshal_int32(buf, buf_len, offset, *(const int32_t*)el); }
static bstd_status unmarshal_int32_ptr(const uint8_t* buf, size_t buf_len, size_t* offset, void* out_el) { return bstd_unmarshal_int32(buf, buf_len, offset, (int32_t*)out_el); }

bool test_slice_of_int32(void) {
    size_t src_count;
    int32_t* src = (int32_t*)generate_slice_alloc(&src_count, sizeof(int32_t), generate_int32_generic);
    size_t size = bstd_size_slice(src, src_count, sizeof(int32_t), sizer_int32);
    uint8_t* buf = (uint8_t*)malloc(size);
    TEST_ASSERT(buf != NULL, "malloc failed");
    size_t offset = 0;
    bstd_status status = bstd_marshal_slice(buf, size, &offset, src, src_count, sizeof(int32_t), marshal_int32_ptr);
    TEST_ASSERT(status == BSTD_OK, "Marshal slice failed");
    size_t dst_count = 0;
    int32_t* dst = NULL;
    offset = 0;
    status = bstd_unmarshal_slice_alloc(buf, size, &offset, (void**)&dst, &dst_count, sizeof(int32_t), unmarshal_int32_ptr);
    TEST_ASSERT(status == BSTD_OK, "Unmarshal slice failed");
    TEST_ASSERT(src_count == dst_count, "Slice count mismatch");
    TEST_ASSERT(compare_slice(src, dst, src_count, sizeof(int32_t), compare_int32_generic), "Slice content mismatch");
    free(src);
    free(dst);
    free(buf);
    return true;
}

// bstd function pointers for the complex TestStruct
static size_t sizer_string_ptr(const void* el) { return bstd_size_string(*(const char**)el, strlen(*(const char**)el)); }
static bstd_status marshal_string_ptr(uint8_t* buf, size_t buf_len, size_t* offset, const void* el) { const char* str = *(const char**)el; return bstd_marshal_string(buf, buf_len, offset, str, strlen(str)); }
static bstd_status unmarshal_string_ptr_alloc(const uint8_t* buf, size_t buf_len, size_t* offset, void* out_el) { return bstd_unmarshal_string_alloc(buf, buf_len, offset, (char**)out_el); }
static size_t sizer_uint64(const void* el) { (void)el; return bstd_size_uint64(); }
static bstd_status marshal_uint64_ptr(uint8_t* buf, size_t buf_len, size_t* offset, const void* el) { return bstd_marshal_uint64(buf, buf_len, offset, *(uint64_t*)el); }
static bstd_status unmarshal_uint64_ptr(const uint8_t* buf, size_t buf_len, size_t* offset, void* out_el) { return bstd_unmarshal_uint64(buf, buf_len, offset, (uint64_t*)out_el); }

static size_t sizer_test_struct(const void* el) {
    const TestStruct* ts = (const TestStruct*)el;
    size_t size = 0;
    size += bstd_size_int32();
    size += bstd_size_string(ts->name, strlen(ts->name));
    size += bstd_size_pointer(ts->optional_value, sizer_uint64);
    size += bstd_size_bytes(ts->blob, ts->blob_count);
    size += bstd_size_slice(ts->tags, ts->tags_count, sizeof(char*), sizer_string_ptr);
    size += bstd_size_map(ts->map_keys, ts->map_values, ts->map_count, sizeof(int32_t), sizeof(char*), sizer_int32, sizer_string_ptr);
    return size;
}

static bstd_status marshal_test_struct(uint8_t* buf, size_t buf_len, size_t* offset, const void* el) {
    const TestStruct* ts = (const TestStruct*)el;
    bstd_status s;
    s = bstd_marshal_int32(buf, buf_len, offset, ts->id); if (s != BSTD_OK) return s;
    s = bstd_marshal_string(buf, buf_len, offset, ts->name, strlen(ts->name)); if (s != BSTD_OK) return s;
    s = bstd_marshal_pointer(buf, buf_len, offset, ts->optional_value, marshal_uint64_ptr); if (s != BSTD_OK) return s;
    s = bstd_marshal_bytes(buf, buf_len, offset, ts->blob, ts->blob_count); if (s != BSTD_OK) return s;
    s = bstd_marshal_slice(buf, buf_len, offset, ts->tags, ts->tags_count, sizeof(char*), marshal_string_ptr); if (s != BSTD_OK) return s;
    s = bstd_marshal_map(buf, buf_len, offset, ts->map_keys, ts->map_values, ts->map_count, sizeof(int32_t), sizeof(char*), marshal_int32_ptr, marshal_string_ptr); if (s != BSTD_OK) return s;
    return BSTD_OK;
}

static bstd_status unmarshal_test_struct(const uint8_t* buf, size_t buf_len, size_t* offset, void* out_el) {
    TestStruct* ts = (TestStruct*)out_el;
    size_t initial_offset = *offset;
    memset(ts, 0, sizeof(TestStruct));
    bstd_status s;
    s = bstd_unmarshal_int32(buf, buf_len, offset, &ts->id); if (s != BSTD_OK) goto fail;
    s = bstd_unmarshal_string_alloc(buf, buf_len, offset, &ts->name); if (s != BSTD_OK) goto fail;
    s = bstd_unmarshal_pointer_alloc(buf, buf_len, offset, (void**)&ts->optional_value, sizeof(uint64_t), unmarshal_uint64_ptr); if (s != BSTD_OK) goto fail;
    s = bstd_unmarshal_bytes_alloc(buf, buf_len, offset, &ts->blob, &ts->blob_count); if (s != BSTD_OK) goto fail;
    s = bstd_unmarshal_slice_alloc_ex(buf, buf_len, offset, (void**)&ts->tags, &ts->tags_count, sizeof(char*), unmarshal_string_ptr_alloc, free_string_element); if (s != BSTD_OK) goto fail;
    s = bstd_unmarshal_map_alloc_ex(buf, buf_len, offset, (void**)&ts->map_keys, (void**)&ts->map_values, &ts->map_count, sizeof(int32_t), sizeof(char*), unmarshal_int32_ptr, unmarshal_string_ptr_alloc, NULL, free_string_element); if (s != BSTD_OK) goto fail;
    return BSTD_OK;
fail:
    free_test_struct(ts);
    memset(ts, 0, sizeof(*ts));
    *offset = initial_offset;
    return s;
}

bool test_complex_struct(void) {
    TestStruct src, dst;
    generate_test_struct(&src);
    size_t size = sizer_test_struct(&src);
    uint8_t* buf = (uint8_t*)malloc(size);
    TEST_ASSERT(buf != NULL, "malloc failed");
    size_t offset = 0;
    bstd_status status = marshal_test_struct(buf, size, &offset, &src);
    TEST_ASSERT(status == BSTD_OK, "Marshal struct failed with code %d", status);
    offset = 0;
    status = unmarshal_test_struct(buf, size, &offset, &dst);
    TEST_ASSERT(status == BSTD_OK, "Unmarshal struct failed with code %d", status);
    TEST_ASSERT(compare_test_struct(&src, &dst), "Struct content mismatch");
    free_test_struct(&src);
    free_test_struct(&dst);
    free(buf);
    return true;
}

//-///////////////////////////////////////////////////////////////////////////
// Main Test Runner
//-///////////////////////////////////////////////////////////////////////////

int main(void) {
    srand((unsigned int)time(NULL));
    printf("--- Starting benc test suite ---\n");

    // Basic data type and error condition tests
    RUN_TEST(test_data_types);
    RUN_TEST(test_err_buf_too_small);
    RUN_TEST(test_varint_overflow);
    RUN_TEST(test_time);

    // Generated primitive tests
    RUN_TEST(test_bool);
    RUN_TEST(test_uint8);
    RUN_TEST(test_uint16);
    RUN_TEST(test_uint32);
    RUN_TEST(test_uint64);
    RUN_TEST(test_int8);
    RUN_TEST(test_int16);
    RUN_TEST(test_int32);
    RUN_TEST(test_int64);
    RUN_TEST(test_float32);
    RUN_TEST(test_float64);
    RUN_TEST(test_uint);
    RUN_TEST(test_int);

    // String, bytes, and view tests
    RUN_TEST(test_string);
    RUN_TEST(test_string_view);
    RUN_TEST(test_bytes);
    
    // Generic container tests
    RUN_TEST(test_pointer);
    RUN_TEST(test_slices);
    RUN_TEST(test_slice_of_int32);
    RUN_TEST(test_maps);
    
    // Complex struct test
    RUN_TEST(test_complex_struct);

    printf("--- Test suite finished ---\n");
    printf("Result: %d passed, %d failed.\n", tests_passed, tests_failed);

    return tests_failed == 0 ? 0 : 1;
}
