#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
flags=(-O1 -g -fno-omit-frame-pointer -fsanitize=address,undefined)
# LeakSanitizer cannot inspect processes under some container ptrace policies;
# ASan memory-safety and UBSan undefined-behaviour checks still run.
export ASAN_OPTIONS="${ASAN_OPTIONS:-detect_leaks=0}"
gcc -std=c17 -Wall -Wextra -Werror "${flags[@]}" std/c/test.c -o "$scratch/c-test"
"$scratch/c-test"
gcc -std=c17 -Wall -Wextra -Werror "${flags[@]}" std/c/hardening_test.c -o "$scratch/c-hardening"
"$scratch/c-hardening"
g++ -std=c++20 -Wall -Wextra -Werror "${flags[@]}" std/c++/test.cpp -o "$scratch/cpp-test"
"$scratch/cpp-test"
