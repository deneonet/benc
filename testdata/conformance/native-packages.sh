#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
for language in c 'c++'; do
  cmake -S "std/$language" -B "$scratch/$language-build"
  cmake --build "$scratch/$language-build"
  ctest --test-dir "$scratch/$language-build" --output-on-failure
  cmake --install "$scratch/$language-build" --prefix "$scratch/$language-install"
done
test -s "$scratch/c-install/include/benc/benc.h"
test -s "$scratch/c++-install/include/benc/std.hpp"
