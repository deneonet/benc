/**
 * @file utils.h
 * @brief Declarations for test utilities for the benc serialization library.
 *
 * This header provides the public interface for random data generators,
 * comparison functions, and memory management for complex test structures.
 */

#ifndef BENC_UTILS_H
#define BENC_UTILS_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h> // FIXED: Added for strcmp and memcmp

// This is a dependency for the TestStruct and other functions.
#include "benc.h"


#ifdef __cplusplus
extern "C" {
#endif

//-///////////////////////////////////////////////////////////////////////////
// Type Definitions for Complex Structures
//-///////////////////////////////////////////////////////////////////////////

// A struct designed to test a variety of types, including nested ones.
typedef struct {
    int32_t id;
    char* name;                 // Dynamically allocated string
    uint64_t* optional_value;   // Pointer to a value
    uint8_t* blob;              // Slice of bytes
    size_t blob_count;
    char** tags;                // Slice of strings
    size_t tags_count;
    // For simplicity, a map from int32_t to a string.
    int32_t* map_keys;
    char** map_values;
    size_t map_count;
} TestStruct;


//-///////////////////////////////////////////////////////////////////////////
// Generic Function Pointer Types
//-///////////////////////////////////////////////////////////////////////////

/**
 * @brief A generic function that generates a random value in place.
 * @param out_element Pointer to the memory where the generated element should be written.
 */
typedef void (*generate_fn)(void* out_element);

/**
 * @brief A generic function that compares two elements.
 * @param a Pointer to the first element.
 * @param b Pointer to the second element.
 * @return True if elements are equal, false otherwise.
 */
typedef bool (*compare_fn)(const void* a, const void* b);


//-///////////////////////////////////////////////////////////////////////////
// Public API - Generators & Comparison Functions
//-///////////////////////////////////////////////////////////////////////////

// --- Generators for Primitives ---
bool generate_bool(void);
intptr_t generate_int(void);
int8_t generate_int8(void);
int16_t generate_int16(void);
int32_t generate_int32(void);
int64_t generate_int64(void);
uintptr_t generate_uint(void);
uint8_t generate_uint8(void);
uint16_t generate_uint16(void);
uint32_t generate_uint32(void);
uint64_t generate_uint64(void);
float generate_float32(void);
double generate_float64(void);
char* generate_string_alloc(void); // Returns a heap-allocated string
uint8_t* generate_bytes_alloc(size_t* out_len); // Returns heap-allocated bytes

// --- Generic Generators ---
void* generate_slice_alloc(size_t* out_count, size_t element_size, generate_fn element_generator);
void generate_map_alloc(void** out_keys, void** out_values, size_t* out_count, size_t key_size, size_t value_size, generate_fn key_generator, generate_fn value_generator);
void* generate_pointer_alloc(size_t element_size, generate_fn element_generator);

// --- Generators for TestStruct ---
void generate_test_struct(TestStruct* ts);

// --- Comparison Functions ---
bool compare_string(const char* a, const char* b);
bool compare_bytes(const uint8_t* a, size_t a_len, const uint8_t* b, size_t b_len);
bool compare_slice(const void* a, const void* b, size_t count, size_t element_stride, compare_fn element_comparer);
bool compare_map(const void* a_keys, const void* a_values, size_t a_count, const void* b_keys, const void* b_values, size_t b_count, size_t key_size, size_t value_size, compare_fn key_comparer, compare_fn value_comparer);
bool compare_pointer(const void* a, const void* b, size_t element_size, compare_fn element_comparer);
bool compare_test_struct(const TestStruct* a, const TestStruct* b);

// --- Memory Freeing Functions ---
void free_string_slice(char** slice, size_t count);
void free_map_of_strings(int32_t* keys, char** values, size_t count);
void free_test_struct(TestStruct* ts);

#ifdef __cplusplus
}
#endif

#endif // BENC_UTILS_H


//-///////////////////////////////////////////////////////////////////////////
// Private Helper Functions
//-///////////////////////////////////////////////////////////////////////////

// Returns a random number of elements for slices/maps (0-3)
static size_t random_count() {
    return rand() % 4;
}

// Generates a random string of a given length. The caller must free the result.
static char* random_string_alloc(size_t len) {
    const char charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    char* str = (char*)malloc(len + 1);
    if (!str) return NULL;
    for (size_t i = 0; i < len; ++i) {
        str[i] = charset[rand() % (sizeof(charset) - 1)];
    }
    str[len] = '\0';
    return str;
}

// Generates random bytes of a given length. The caller must free the result.
static uint8_t* random_bytes_alloc(size_t len) {
    uint8_t* buf = (uint8_t*)malloc(len);
    if (!buf) return NULL;
    for (size_t i = 0; i < len; ++i) {
        buf[i] = (uint8_t)(rand() % 256);
    }
    return buf;
}

//-///////////////////////////////////////////////////////////////////////////
// Generator Implementations
//-///////////////////////////////////////////////////////////////////////////

bool generate_bool(void) { return rand() % 2 == 1; }
intptr_t generate_int(void) { return (intptr_t)rand() - (RAND_MAX / 2); }
int8_t generate_int8(void) { return (int8_t)((rand() % 256) - 128); }
int16_t generate_int16(void) { return (int16_t)((rand() % 65536) - 32768); }
int32_t generate_int32(void) { return (int32_t)rand() - (RAND_MAX / 2); }
int64_t generate_int64(void) { return ((int64_t)rand() << 32) | rand(); }
uintptr_t generate_uint(void) { return (uintptr_t)rand(); }
uint8_t generate_uint8(void) { return (uint8_t)(rand() % 256); }
uint16_t generate_uint16(void) { return (uint16_t)(rand() % 65536); }
uint32_t generate_uint32(void) { return (uint32_t)rand(); }
uint64_t generate_uint64(void) { return ((uint64_t)rand() << 32) | rand(); }
float generate_float32(void) { return (float)rand() / (float)RAND_MAX; }
double generate_float64(void) { return (double)rand() / (double)RAND_MAX; }
char* generate_string_alloc(void) { return random_string_alloc(5 + (rand() % 15)); }
uint8_t* generate_bytes_alloc(size_t* out_len) {
    *out_len = 3 + (rand() % 7);
    return random_bytes_alloc(*out_len);
}

// Helper generators for use with generic functions
static void generate_int32_generic(void* out) { *(int32_t*)out = generate_int32(); }
static void generate_string_alloc_generic(void* out) { *(char**)out = generate_string_alloc(); }
static void generate_uint64_generic(void* out) { *(uint64_t*)out = generate_uint64(); }

void* generate_slice_alloc(size_t* out_count, size_t element_size, generate_fn element_generator) {
    *out_count = random_count();
    if (*out_count == 0) return NULL;
    uint8_t* slice = (uint8_t*)malloc(*out_count * element_size);
    if (!slice) abort(); // In tests, better to abort on alloc failure
    for (size_t i = 0; i < *out_count; ++i) {
        element_generator(slice + i * element_size);
    }
    return slice;
}

void generate_map_alloc(void** out_keys, void** out_values, size_t* out_count, size_t key_size, size_t value_size, generate_fn key_generator, generate_fn value_generator) {
    *out_count = random_count();
    if (*out_count == 0) {
        *out_keys = NULL;
        *out_values = NULL;
        return;
    }
    *out_keys = malloc(*out_count * key_size);
    *out_values = malloc(*out_count * value_size);
    if (!*out_keys || !*out_values) abort();

    uint8_t* keys_ptr = (uint8_t*)*out_keys;
    uint8_t* values_ptr = (uint8_t*)*out_values;

    for (size_t i = 0; i < *out_count; ++i) {
        key_generator(keys_ptr + i * key_size);
        value_generator(values_ptr + i * value_size);
    }
}

void* generate_pointer_alloc(size_t element_size, generate_fn element_generator) {
    if (rand() % 4 == 0) { // 25% chance of being NULL
        return NULL;
    }
    void* ptr = malloc(element_size);
    if (!ptr) abort();
    element_generator(ptr);
    return ptr;
}

void generate_test_struct(TestStruct* ts) {
    ts->id = generate_int32();
    ts->name = generate_string_alloc();
    ts->optional_value = (uint64_t*)generate_pointer_alloc(sizeof(uint64_t), generate_uint64_generic);
    ts->blob = generate_bytes_alloc(&ts->blob_count);
    ts->tags = (char**)generate_slice_alloc(&ts->tags_count, sizeof(char*), generate_string_alloc_generic);
    generate_map_alloc((void**)&ts->map_keys, (void**)&ts->map_values, &ts->map_count, sizeof(int32_t), sizeof(char*), generate_int32_generic, generate_string_alloc_generic);
}

//-///////////////////////////////////////////////////////////////////////////
// Comparison Function Implementations
//-///////////////////////////////////////////////////////////////////////////

bool compare_string(const char* a, const char* b) {
    if (a == NULL && b == NULL) return true;
    if (a == NULL || b == NULL) return false;
    return strcmp(a, b) == 0;
}

bool compare_bytes(const uint8_t* a, size_t a_len, const uint8_t* b, size_t b_len) {
    if (a_len != b_len) return false;
    if (a_len == 0 && b_len == 0) return true;
    if (a == NULL || b == NULL) return false; // Should not happen if lens are > 0, but good practice
    return memcmp(a, b, a_len) == 0;
}

bool compare_slice(const void* a, const void* b, size_t count, size_t element_stride, compare_fn element_comparer) {
    if (count == 0) return true;
    if (a == NULL || b == NULL) return false;

    const uint8_t* pa = (const uint8_t*)a;
    const uint8_t* pb = (const uint8_t*)b;
    for (size_t i = 0; i < count; ++i) {
        if (!element_comparer(pa + i * element_stride, pb + i * element_stride)) {
            return false;
        }
    }
    return true;
}

bool compare_map(const void* a_keys, const void* a_values, size_t a_count, const void* b_keys, const void* b_values, size_t b_count, size_t key_size, size_t value_size, compare_fn key_comparer, compare_fn value_comparer) {
    if (a_count != b_count) return false;
    if (a_count == 0) return true;
    if (a_keys == NULL || b_keys == NULL) return false;

    const uint8_t* a_k_ptr = (const uint8_t*)a_keys;
    const uint8_t* a_v_ptr = (const uint8_t*)a_values;
    const uint8_t* b_k_ptr = (const uint8_t*)b_keys;
    const uint8_t* b_v_ptr = (const uint8_t*)b_values;

    // This is O(n^2), but for small test maps it's fine.
    // A real implementation might sort or use a hash map for verification.
    for (size_t i = 0; i < a_count; ++i) {
        const void* key_a = a_k_ptr + i * key_size;
        const void* val_a = a_v_ptr + i * value_size;
        bool found = false;
        for (size_t j = 0; j < b_count; ++j) {
            const void* key_b = b_k_ptr + j * key_size;
            if (key_comparer(key_a, key_b)) {
                const void* val_b = b_v_ptr + j * value_size;
                if (!value_comparer(val_a, val_b)) return false;
                found = true;
                break;
            }
        }
        if (!found) return false;
    }
    return true;
}

bool compare_pointer(const void* a, const void* b, size_t element_size, compare_fn element_comparer) {
    // FIXED: Mark unused parameter to silence compiler warnings.
    (void)element_size;
    if (a == NULL && b == NULL) return true;
    if (a == NULL || b == NULL) return false;
    return element_comparer(a, b);
}

// Comparison helpers for generic functions
static bool compare_int32_generic(const void* a, const void* b) { return *(int32_t*)a == *(int32_t*)b; }
static bool compare_string_ptr_generic(const void* a, const void* b) { return compare_string(*(char**)a, *(char**)b); }
static bool compare_uint64_generic(const void* a, const void* b) { return *(uint64_t*)a == *(uint64_t*)b; }

bool compare_test_struct(const TestStruct* a, const TestStruct* b) {
    if (a->id != b->id) return false;
    if (!compare_string(a->name, b->name)) return false;
    if (!compare_pointer(a->optional_value, b->optional_value, sizeof(uint64_t), compare_uint64_generic)) return false;
    if (!compare_bytes(a->blob, a->blob_count, b->blob, b->blob_count)) return false;
    if (a->tags_count != b->tags_count) return false;
    if (!compare_slice(a->tags, b->tags, a->tags_count, sizeof(char*), compare_string_ptr_generic)) return false;
    if (a->map_count != b->map_count) return false;
    if (!compare_map(a->map_keys, a->map_values, a->map_count, b->map_keys, b->map_values, b->map_count, sizeof(int32_t), sizeof(char*), compare_int32_generic, compare_string_ptr_generic)) return false;
    return true;
}

//-///////////////////////////////////////////////////////////////////////////
// Memory Freeing Function Implementations
//-///////////////////////////////////////////////////////////////////////////

void free_string_slice(char** slice, size_t count) {
    if (!slice) return;
    for (size_t i = 0; i < count; ++i) {
        free(slice[i]);
    }
    free(slice);
}

void free_map_of_strings(int32_t* keys, char** values, size_t count) {
    if (!values) { // Keys don't need freeing, but values are strings
        free(keys);
        return;
    }
    for (size_t i = 0; i < count; ++i) {
        free(values[i]);
    }
    free(keys);
    free(values);
}

void free_test_struct(TestStruct* ts) {
    if (!ts) return;
    free(ts->name);
    free(ts->optional_value);
    free(ts->blob);
    free_string_slice(ts->tags, ts->tags_count);
    free_map_of_strings(ts->map_keys, ts->map_values, ts->map_count);
}