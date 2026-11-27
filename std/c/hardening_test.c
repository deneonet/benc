// Run with: gcc -std=c17 -Wall -Wextra -Werror -fsanitize=address,undefined
//             -fno-omit-frame-pointer hardening_test.c -o hardening_test
#include <assert.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define BSTD_IMPLEMENTATION
#include "benc.h"

static bstd_status skip_byte(const uint8_t *b, size_t len, size_t *off) {
    return bstd_skip_byte(b, len, off);
}
static bstd_status skip_pointer_recursive(const uint8_t *b, size_t len, size_t *off) {
    return bstd_skip_pointer(b, len, off, skip_pointer_recursive);
}
static void free_pointer_recursive(void *out) {
    bstd_free_pointer(*(void **)out, free_pointer_recursive);
}
static bstd_status read_pointer_recursive(const uint8_t *b, size_t len, size_t *off, void *out) {
    return bstd_unmarshal_pointer_alloc_ex(b, len, off, (void **)out, sizeof(void *), read_pointer_recursive, free_pointer_recursive);
}
static bstd_status read_byte(const uint8_t *b, size_t len, size_t *off, void *out) {
    return bstd_unmarshal_byte(b, len, off, (uint8_t *)out);
}

static bstd_status owned_string(const uint8_t *b, size_t len, size_t *off, void *out) {
    return bstd_unmarshal_string_alloc(b, len, off, (char **)out);
}

static int freed;
static void free_owned_string(void *out) {
    if (*(char **)out) { free(*(char **)out); ++freed; }
}

// The callback may fail *after* allocating; the _ex caller must clean the
// partially initialized element as well as all preceding elements.
static bstd_status fail_after_alloc(const uint8_t *b, size_t len, size_t *off, void *out) {
    bstd_status s = owned_string(b, len, off, out);
    return s == BSTD_OK ? BSTD_ERR_INVALID_ARGUMENT : s;
}

int main(void) {
    const uint8_t huge[] = {255,255,255,255,255,255,255,255,255,1};
    const uint8_t overflow[] = {255,255,255,255,255,255,255,255,255,2};
    const uint8_t empty[] = {0};
    const uint8_t wrong[] = {0,0,0,0,0};
    const uint8_t valid[] = {0,1,1,1,1};
    const uint8_t one_string[] = {1,1,'x',1,1,1,1};
    const uint8_t wrong_string[] = {1,1,'x',1,1,1,0};
    const uint8_t map_string[] = {1,1,'k',1,'v',1,1,1,0};
    const uint8_t pointer_string[] = {1,1,'x'};
    size_t off = 0, count = 77, len = 77;
    uintptr_t u = 0;
    const uint8_t *view = (const uint8_t *)1;
    void *elements = (void *)1;
    void *keys = (void *)1, *values = (void *)1;

    assert(bstd_unmarshal_bytes_view(huge, sizeof(huge), &off, &view, &len) == BSTD_ERR_DECODE_LIMIT);
    assert(off == 0 && view == (const uint8_t *)1 && len == 77);
    assert(bstd_unmarshal_slice_alloc(huge, sizeof(huge), &off, &elements, &count, 64, owned_string) == BSTD_ERR_DECODE_LIMIT);
    assert(off == 0 && elements == (void *)1 && count == 77);
    assert(bstd_unmarshal_map_alloc(huge, sizeof(huge), &off, &keys, &values, &count, 64, 64, owned_string, owned_string) == BSTD_ERR_DECODE_LIMIT);
    assert(off == 0 && keys == (void *)1 && values == (void *)1);
    assert(bstd_unmarshal_uint(overflow, sizeof(overflow), &off, &u) == BSTD_ERR_OVERFLOW && off == 0);
    assert(bstd_skip_varint(overflow, sizeof(overflow), &off) == BSTD_ERR_OVERFLOW && off == 0);

    assert(bstd_skip_slice(empty, sizeof(empty), &off, skip_byte) == BSTD_ERR_BUF_TOO_SMALL && off == 0);
    assert(bstd_unmarshal_slice_alloc(empty, sizeof(empty), &off, &elements, &count, 1, owned_string) == BSTD_ERR_BUF_TOO_SMALL && off == 0);
    assert(bstd_skip_map(empty, sizeof(empty), &off, skip_byte, skip_byte) == BSTD_ERR_BUF_TOO_SMALL && off == 0);
    assert(bstd_skip_slice(wrong, sizeof(wrong), &off, skip_byte) == BSTD_ERR_INVALID_TERMINATOR && off == 0);
    assert(bstd_skip_map(wrong, sizeof(wrong), &off, skip_byte, skip_byte) == BSTD_ERR_INVALID_TERMINATOR && off == 0);
    assert(bstd_skip_slice(valid, sizeof(valid), &off, skip_byte) == BSTD_OK && off == sizeof(valid));

    off = SIZE_MAX;
    assert(bstd_skip_int32(valid, sizeof(valid), &off) == BSTD_ERR_BUF_TOO_SMALL && off == SIZE_MAX);
    assert(bstd_unmarshal_uint(valid, sizeof(valid), &off, &u) == BSTD_ERR_BUF_TOO_SMALL && off == SIZE_MAX);
    off = 0;
    const uint8_t invalid_bool[] = {2};
    assert(bstd_skip_bool(invalid_bool, sizeof(invalid_bool), &off) == BSTD_ERR_INVALID_BOOL && off == 0);
    const uint8_t invalid_utf8[] = {1, 0xff};
    bstd_string_view invalid_view = {0};
    assert(bstd_skip_string(invalid_utf8, sizeof(invalid_utf8), &off) == BSTD_ERR_INVALID_UTF8 && off == 0);
    assert(bstd_unmarshal_string_view(invalid_utf8, sizeof(invalid_utf8), &off, &invalid_view) == BSTD_ERR_INVALID_UTF8 && off == 0);

    uint8_t int_buf[10];
    intptr_t decoded = 0;
    assert(bstd_marshal_int(int_buf, sizeof(int_buf), &off, INTPTR_MIN) == BSTD_OK);
    size_t int_len = off;
    off = 0;
    assert(bstd_unmarshal_int(int_buf, int_len, &off, &decoded) == BSTD_OK && decoded == INTPTR_MIN);
    off = 0;
    assert(bstd_marshal_int32(int_buf, sizeof(int_buf), &off, INT32_MIN) == BSTD_OK);
    off = 0;
    int32_t decoded_fixed = 0;
    assert(bstd_unmarshal_int32(int_buf, 4, &off, &decoded_fixed) == BSTD_OK && decoded_fixed == INT32_MIN);

    uint8_t deep[BSTD_MAX_DECODE_DEPTH + 2];
    memset(deep, 1, sizeof(deep));
    off = 0;
    assert(bstd_skip_pointer(deep, sizeof(deep), &off, skip_pointer_recursive) == BSTD_ERR_DECODE_LIMIT && off == 0);
    elements = (void *)1;
    assert(bstd_unmarshal_pointer_alloc_ex(deep, sizeof(deep), &off, &elements, sizeof(void *), read_pointer_recursive, free_pointer_recursive) == BSTD_ERR_DECODE_LIMIT);
    assert(off == 0 && elements == (void *)1);
    const uint8_t shallow[] = {1, 0};
    assert(bstd_skip_pointer(shallow, sizeof(shallow), &off, skip_pointer_recursive) == BSTD_OK && off == sizeof(shallow));

    freed = 0; off = 0; elements = (void *)1;
    assert(bstd_unmarshal_slice_alloc_ex(one_string, sizeof(one_string), &off, &elements, &count, sizeof(char *), fail_after_alloc, free_owned_string) == BSTD_ERR_INVALID_ARGUMENT);
    assert(off == 0 && elements == (void *)1 && freed == 1);
    freed = 0;
    assert(bstd_unmarshal_slice_alloc_ex(wrong_string, sizeof(wrong_string), &off, &elements, &count, sizeof(char *), owned_string, free_owned_string) == BSTD_ERR_INVALID_TERMINATOR);
    assert(off == 0 && elements == (void *)1 && freed == 1);
    freed = 0;
    assert(bstd_unmarshal_map_alloc_ex(map_string, sizeof(map_string), &off, &keys, &values, &count, sizeof(char *), sizeof(char *), owned_string, owned_string, free_owned_string, free_owned_string) == BSTD_ERR_INVALID_TERMINATOR);
    assert(off == 0 && keys == (void *)1 && values == (void *)1 && freed == 2);
    freed = 0;
    assert(bstd_unmarshal_pointer_alloc_ex(pointer_string, sizeof(pointer_string), &off, &elements, sizeof(char *), fail_after_alloc, free_owned_string) == BSTD_ERR_INVALID_ARGUMENT);
    assert(off == 0 && elements == (void *)1 && freed == 1);

    // A bounded malformed corpus must never trap or attempt large allocations.
    uint32_t seed = 1;
    for (size_t i = 0; i < 2000; ++i) {
        uint8_t input[8];
        for (size_t j = 0; j < sizeof(input); ++j) {
            seed = seed * 1664525u + 1013904223u;
            input[j] = (uint8_t)(seed >> 24);
        }
        off = 0;
        (void)bstd_skip_slice(input, sizeof(input), &off, skip_byte);
        off = 0;
        (void)bstd_skip_map(input, sizeof(input), &off, skip_byte, skip_byte);
        off = 0;
        if (bstd_unmarshal_slice_alloc(input, sizeof(input), &off, &elements, &count, 1, read_byte) == BSTD_OK) free(elements);
    }
    puts("C hardening regressions passed");
    return 0;
}
