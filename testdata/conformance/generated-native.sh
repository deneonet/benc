#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
repo="$PWD"
scratch="$(mktemp -d)"
trap 'rm -rf "$scratch" || true' EXIT
cp testdata/conformance/sources/model.go "$scratch/model.go"
language="${1:?target language}"
source="${2:-go}"
if [[ "$source" != go ]]; then
  case "$source" in
    c) ext=h ;;
    cpp) ext=hpp ;;
    ts) ext=d.ts ;;
    js|py|rs|java|kt|cs|swift|scala) ext="$source" ;;
    *) echo "unknown source: $source" >&2; exit 2 ;;
  esac
  go run ./cmd -lang="$source" "$scratch/model.go"
  # The generated declaration is parsed again as a source file. Each source
  # therefore exercises its parser -> Go AST -> selected target generator.
  cp "$scratch/model_benc.$ext" "$scratch/model.${ext##*.}"
  source_path="$scratch/model.${ext##*.}"
else
  source_path="$scratch/model.go"
fi
expected='034100420dac02c9010203feffffff050000000200ff01f7ffffffffffffff0178020405010101010102000101010101010300000000000000017001060101640400000000000000017a01070101010101016efbffffff010101017b9274a54b9bb601'
check_wire() {
  local actual
  actual=$("$@" | sed -n 's/.*WIRE=\([0-9a-f]*\).*/\1/p' | tail -n 1)
  if [[ "$actual" != "$expected" ]]; then
    echo "wire mismatch: $actual" >&2
    exit 1
  fi
  echo "Nonzero wire matches Go"
}

case "$language" in
  rust)
    go run ./cmd -lang=rs "$source_path"
    mkdir -p "$scratch/rust/src"
    cp "$scratch/model_benc.rs" "$scratch/rust/src/model_benc.rs"
    cp "$scratch/model_benc_test.rs" "$scratch/rust/src/main.rs"
    mkdir -p "$scratch/rust/src/bin"
    cp testdata/conformance/values/value.rs "$scratch/rust/src/bin/value.rs"
    cat > "$scratch/rust/Cargo.toml" <<EOF
[package]
name = "benc-generated-smoke"
version = "0.1.0"
edition = "2024"
[dependencies]
benc = { path = "$repo/std/rust" }
chrono = "0.4.42"
EOF
    cargo run --manifest-path "$scratch/rust/Cargo.toml" --bin benc-generated-smoke
    check_wire cargo run --manifest-path "$scratch/rust/Cargo.toml" --bin value
    ;;
  kotlin)
    go run ./cmd -lang=kt "$source_path"
    mkdir -p "$scratch/kotlin/src/main/kotlin"
    cp std/kotlin/src/main/kotlin/Bstd.kt "$scratch/kotlin/src/main/kotlin/"
    cp "$scratch/model_benc.kt" "$scratch/kotlin/src/main/kotlin/"
    cp testdata/conformance/values/value.kt "$scratch/kotlin/src/main/kotlin/"
    cat > "$scratch/kotlin/settings.gradle.kts" <<'EOF'
pluginManagement { repositories { gradlePluginPortal() } }
rootProject.name = "generated-smoke"
EOF
    cat > "$scratch/kotlin/build.gradle.kts" <<'EOF'
plugins {
    kotlin("jvm") version "1.9.21"
    application
}
repositories { mavenCentral() }
application { mainClass.set("ValueKt") }
kotlin { jvmToolchain(21) }
EOF
    check_wire gradle -p "$scratch/kotlin" run --no-daemon
    ;;
  csharp)
    go run ./cmd -lang=cs "$source_path"
    mkdir -p "$scratch/csharp"
    cp 'std/c#/benc/benc.cs' "$scratch/csharp/"
    cp "$scratch/model_benc.cs" "$scratch/csharp/"
    cp testdata/conformance/values/value.cs "$scratch/csharp/"
    cat > "$scratch/csharp/smoke.csproj" <<'EOF'
<Project Sdk="Microsoft.NET.Sdk"><PropertyGroup><OutputType>Exe</OutputType><TargetFramework>net9.0</TargetFramework><Nullable>enable</Nullable><ImplicitUsings>enable</ImplicitUsings></PropertyGroup></Project>
EOF
    check_wire dotnet run --project "$scratch/csharp/smoke.csproj"
    ;;
  swift)
    go run ./cmd -lang=swift "$source_path"
    mkdir -p "$scratch/swift/Sources/Bstd" "$scratch/swift/Sources/Smoke"
    cp std/swift/Sources/Bstd/Bstd.swift "$scratch/swift/Sources/Bstd/"
    cp "$scratch/model_benc.swift" "$scratch/swift/Sources/Smoke/"
    cp testdata/conformance/values/value.swift "$scratch/swift/Sources/Smoke/"
    cat > "$scratch/swift/Package.swift" <<'EOF'
// swift-tools-version: 6.0
import PackageDescription
let package = Package(name: "GeneratedSmoke", targets: [
  .target(name: "Bstd"),
  .executableTarget(name: "Smoke", dependencies: ["Bstd"])
])
EOF
    check_wire swift run --package-path "$scratch/swift" Smoke
    ;;
  scala)
    go run ./cmd -lang=scala "$source_path"
    mkdir -p "$scratch/scala/src/main/scala"
    cp std/scala/src/main/scala/dev/benc/benc.scala "$scratch/scala/src/main/scala/"
    cp "$scratch/model_benc.scala" "$scratch/scala/src/main/scala/"
    cp testdata/conformance/values/value.scala "$scratch/scala/src/main/scala/"
    printf 'scalaVersion := "3.4.2"\n' > "$scratch/scala/build.sbt"
    check_wire bash -c 'cd "$1" && sbt "runMain BencValue"' _ "$scratch/scala"
    ;;
  typescript)
    go run ./cmd -lang=ts "$source_path"
    cp std/javascript/std.d.ts "$scratch/std.d.ts"
    (cd "$scratch" && npx --yes --package typescript@5.9.3 tsc --strict --noEmit --target es2020 model_benc.d.ts)
    ;;
  *) echo "unknown language: $language" >&2; exit 2 ;;
esac
