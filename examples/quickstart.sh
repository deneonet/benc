#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
repo="$PWD"
scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
cp testdata/conformance/sources/model.go "$scratch/model.go"

case "${1:-js}" in
  js)
    go run ./cmd -lang=js "$scratch/model.go"
    cp std/javascript/std.js "$scratch/"
    (cd "$scratch" && node model_benc_test.js)
    ;;
  py)
    go run ./cmd -lang=py "$scratch/model.go"
    cp std/python/bstd.py "$scratch/"
    (cd "$scratch" && python3 model_benc_test.py)
    ;;
  c)
    go run ./cmd -lang=c "$scratch/model.go"
    gcc -std=c17 -DBSTD_IMPLEMENTATION -x c -c "$repo/std/c/benc.h" -o "$scratch/bstd.o"
    gcc -std=c17 -I"$repo/std/c" "$scratch/model_benc.c" \
      "$scratch/model_benc_test.c" "$scratch/bstd.o" -o "$scratch/consumer"
    "$scratch/consumer"
    ;;
  cpp)
    go run ./cmd -lang=cpp "$scratch/model.go"
    g++ -std=c++20 -I"$repo/std/c++" "$scratch/model_benc_test.cpp" -o "$scratch/consumer"
    "$scratch/consumer"
    ;;
  java)
    go run ./cmd -lang=java "$scratch/model.go"
    java -cp "$scratch" com.sun.tools.javac.Main -d "$scratch" \
      "$repo"/std/java/src/main/java/main/com/benc/*.java \
      "$scratch/model_benc.java" "$scratch/model_benc_test.java"
    java -cp "$scratch" BencGeneratedTest
    ;;
  *) echo "usage: $0 [js|py|c|cpp|java]" >&2; exit 2 ;;
esac
echo "Generated consumer passed: Go AST -> ${1:-js}"
