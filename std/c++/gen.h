#pragma once

#include "std.hpp"
#include <random>
#include <string>
#include <vector>
#include <map>
#include <optional>
#include <complex>
#include <algorithm>
#include <sstream>
#include <iomanip>
#include <functional>
#include <concepts>

// --- Compiler/Language Standard Check ---
#if __cplusplus < 202002L
    #error "bstd.hpp requires a C++20 compliant compiler. Please use the -std=c++20 flag."
#endif

namespace bstd::gen {

// Constants
constexpr int MaxDepth = 2;

// --- Helper Functions ---

// Concept for Random Number Generator
template<typename T>
concept URBG = std::uniform_random_bit_generator<std::remove_reference_t<T>>;

template<URBG Gen>
int RandomCount(Gen& g) {
    std::uniform_int_distribution<int> dist(1, 3);
    return dist(g);
}

template<URBG Gen>
std::string RandomString(Gen& g, int length) {
    static const char charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    std::uniform_int_distribution<size_t> dist(0, sizeof(charset) - 2);
    std::string s;
    s.reserve(length);
    for (int i = 0; i < length; ++i) {
        s.push_back(charset[dist(g)]);
    }
    return s;
}

template<URBG Gen>
std::vector<std::byte> RandomBytes(Gen& g, int length) {
    std::vector<std::byte> b(length);
    std::uniform_int_distribution<int> dist(0, 255);
    for (int i = 0; i < length; ++i) {
        b[i] = static_cast<std::byte>(dist(g));
    }
    return b;
}

template<URBG Gen>
bstd::time_point RandomTime(Gen& g) {
    // Current time + random seconds (0 to 1,000,000)
    auto now = std::chrono::system_clock::now();
    std::uniform_int_distribution<int64_t> dist(0, 1000000);
    return std::chrono::time_point_cast<std::chrono::nanoseconds>(now + std::chrono::seconds(dist(g)));
}

// --- Primitive Generators ---

template<URBG Gen>
bool GenerateBool(Gen& g, int) {
    std::bernoulli_distribution dist(0.5);
    return dist(g);
}

template<URBG Gen>
int GenerateInt(Gen& g, int) {
    std::uniform_int_distribution<int> dist;
    return dist(g);
}

template<URBG Gen>
int8_t GenerateInt8(Gen& g, int) {
    std::uniform_int_distribution<int> dist(-128, 127);
    return static_cast<int8_t>(dist(g));
}

template<URBG Gen>
int16_t GenerateInt16(Gen& g, int) {
    std::uniform_int_distribution<int> dist(-32768, 32767);
    return static_cast<int16_t>(dist(g));
}

template<URBG Gen>
int32_t GenerateInt32(Gen& g, int) {
    std::uniform_int_distribution<int32_t> dist;
    return dist(g);
}

template<URBG Gen>
int64_t GenerateInt64(Gen& g, int) {
    std::uniform_int_distribution<int64_t> dist;
    return dist(g);
}

template<URBG Gen>
unsigned int GenerateUint(Gen& g, int) {
    std::uniform_int_distribution<unsigned int> dist;
    return dist(g);
}

template<URBG Gen>
uint8_t GenerateUint8(Gen& g, int) {
    std::uniform_int_distribution<int> dist(0, 255);
    return static_cast<uint8_t>(dist(g));
}

template<URBG Gen>
uint16_t GenerateUint16(Gen& g, int) {
    std::uniform_int_distribution<int> dist(0, 65535);
    return static_cast<uint16_t>(dist(g));
}

template<URBG Gen>
uint32_t GenerateUint32(Gen& g, int) {
    std::uniform_int_distribution<uint32_t> dist;
    return dist(g);
}

template<URBG Gen>
uint64_t GenerateUint64(Gen& g, int) {
    std::uniform_int_distribution<uint64_t> dist;
    return dist(g);
}

template<URBG Gen>
uintptr_t GenerateUintptr(Gen& g, int) {
    std::uniform_int_distribution<uintptr_t> dist;
    return dist(g);
}

template<URBG Gen>
float GenerateFloat32(Gen& g, int) {
    std::uniform_real_distribution<float> dist;
    return dist(g);
}

template<URBG Gen>
double GenerateFloat64(Gen& g, int) {
    std::uniform_real_distribution<double> dist;
    return dist(g);
}

template<URBG Gen>
std::complex<float> GenerateComplex64(Gen& g, int) {
    std::uniform_real_distribution<float> dist;
    return {dist(g), dist(g)};
}

template<URBG Gen>
std::complex<double> GenerateComplex128(Gen& g, int) {
    std::uniform_real_distribution<double> dist;
    return {dist(g), dist(g)};
}

template<URBG Gen>
std::string GenerateString(Gen& g, int) {
    std::uniform_int_distribution<int> lenDist(5, 20);
    return RandomString(g, lenDist(g));
}

template<URBG Gen>
std::byte GenerateByte(Gen& g, int) {
    std::uniform_int_distribution<int> dist(0, 255);
    return static_cast<std::byte>(dist(g));
}

template<URBG Gen>
std::vector<std::byte> GenerateBytes(Gen& g, int) {
    std::uniform_int_distribution<int> lenDist(3, 10);
    return RandomBytes(g, lenDist(g));
}

template<URBG Gen>
int32_t GenerateRune(Gen& g, int) {
    return GenerateInt32(g, 0);
}

template<URBG Gen>
bstd::time_point GenerateTime(Gen& g, int) {
    return RandomTime(g);
}

// --- Container Generators ---

// GenerateSlice
template<typename T, URBG Gen, typename GeneratorFunc>
std::vector<T> GenerateSlice(Gen& g, int depth, GeneratorFunc generator) {
    int count = RandomCount(g);
    std::vector<T> s;
    s.reserve(count);
    for (int i = 0; i < count; ++i) {
        s.push_back(generator(g, depth));
    }
    return s;
}

// GenerateSliceSlice
template<typename T, URBG Gen, typename GeneratorFunc>
std::vector<std::vector<T>> GenerateSliceSlice(Gen& g, int depth, GeneratorFunc generator) {
    int count = RandomCount(g);
    std::vector<std::vector<T>> s;
    s.reserve(count);
    for (int i = 0; i < count; ++i) {
        s.push_back(GenerateSlice<T>(g, depth, generator));
    }
    return s;
}

// GeneratePointer (maps to std::optional in bstd)
template<typename T, URBG Gen, typename GeneratorFunc>
std::optional<T> GeneratePointer(Gen& g, int depth, GeneratorFunc generator) {
    std::uniform_int_distribution<int> dist(0, 3);
    // 1 in 4 chance of being null
    if (dist(g) == 0) {
        return std::nullopt;
    }
    return generator(g, depth);
}

// GenerateMap
template<typename K, typename V, URBG Gen, typename KGen, typename VGen>
std::map<K, V> GenerateMap(Gen& g, int depth, KGen keyGen, VGen valGen) {
    int count = RandomCount(g);
    std::map<K, V> m;
    for (int i = 0; i < count; ++i) {
        m.emplace(keyGen(g, depth), valGen(g, depth));
    }
    return m;
}

// --- Comparison Functions ---

// Return type is std::optional<std::string>. std::nullopt means success (no error).
using CompareResult = std::optional<std::string>;

inline CompareResult CompareField(std::string_view name, std::function<CompareResult()> cmp) {
    if (auto err = cmp()) {
        std::ostringstream oss;
        oss << name << ": " << *err;
        return oss.str();
    }
    return std::nullopt;
}

template<typename T>
CompareResult ComparePrimitive(const T& a, const T& b) {
    if constexpr (std::is_floating_point_v<T>) {
        // Simple epsilon check could be added here if needed, 
        // but exact match is requested by the port.
        if (a != b) {
            std::ostringstream oss;
            oss << "mismatch: " << a << " != " << b;
            return oss.str();
        }
    } else {
        if (a != b) {
            std::ostringstream oss;
            // Handle types that might not stream directly (like std::byte)
            if constexpr (std::is_same_v<T, std::byte>) {
                oss << "mismatch: " << static_cast<int>(a) << " != " << static_cast<int>(b);
            } else {
                oss << "mismatch: " << a << " != " << b;
            }
            return oss.str();
        }
    }
    return std::nullopt;
}

inline CompareResult CompareBytes(std::span<const std::byte> a, std::span<const std::byte> b) {
    if (std::equal(a.begin(), a.end(), b.begin(), b.end())) {
        return std::nullopt;
    }
    std::ostringstream oss;
    oss << "bytes mismatch (len " << a.size() << " vs " << b.size() << ")";
    return oss.str();
}

template<typename T, typename ElemCmp>
CompareResult CompareSlice(const std::vector<T>& a, const std::vector<T>& b, ElemCmp elemCmp) {
    if (a.size() != b.size()) {
        std::ostringstream oss;
        oss << "length mismatch: " << a.size() << " != " << b.size();
        return oss.str();
    }
    for (size_t i = 0; i < a.size(); ++i) {
        if (auto err = elemCmp(a[i], b[i])) {
            std::ostringstream oss;
            oss << "index [" << i << "]: " << *err;
            return oss.str();
        }
    }
    return std::nullopt;
}

template<typename K, typename V, typename ValCmp>
CompareResult CompareMap(const std::map<K, V>& a, const std::map<K, V>& b, ValCmp valCmp) {
    if (a.size() != b.size()) {
        std::ostringstream oss;
        oss << "length mismatch: " << a.size() << " != " << b.size();
        return oss.str();
    }
    for (const auto& [k, v1] : a) {
        auto it = b.find(k);
        if (it == b.end()) {
            std::ostringstream oss;
            oss << "key missing in b"; // Simplify key printing for generic types
            return oss.str();
        }
        if (auto err = valCmp(v1, it->second)) {
            std::ostringstream oss;
            oss << "key value mismatch: " << *err;
            return oss.str();
        }
    }
    return std::nullopt;
}

template<typename T, typename ElemCmp>
CompareResult ComparePointer(const std::optional<T>& a, const std::optional<T>& b, ElemCmp elemCmp) {
    if (!a.has_value() && !b.has_value()) return std::nullopt;
    if (!a.has_value() && b.has_value()) return "mismatch: a is null, b is not";
    if (a.has_value() && !b.has_value()) return "mismatch: a is not null, b is null";
    return elemCmp(*a, *b);
}

} // namespace bstd::gen