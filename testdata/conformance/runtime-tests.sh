#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
case "${1:?runtime}" in
  python)
    python3 -m unittest discover -s std/python -p '*test.py'
    bash testdata/conformance/package-smoke.sh python
    ;;
  javascript)
    npm --prefix std/javascript test
    bash testdata/conformance/package-smoke.sh javascript
    ;;
  java)
    mvn -B -f std/java/pom.xml test package
    ;;
  rust)
    cargo test --locked --manifest-path std/rust/Cargo.toml
    (cd std/rust && cargo package --locked --allow-dirty)
    ;;
  kotlin)
    gradle -p std/kotlin test jar --no-daemon
    ;;
  csharp)
    dotnet test 'std/c#/benc.sln'
    dotnet pack 'std/c#/benc/benc.csproj' --configuration Release
    ;;
  swift)
    swift test --package-path std/swift
    swift build --package-path std/swift --configuration release
    ;;
  scala)
    (cd std/scala && sbt test package)
    ;;
  *) echo "unknown runtime: $1" >&2; exit 2 ;;
esac
