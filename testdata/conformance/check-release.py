"""Keep published runtime versions and compatibility notes in sync."""
import json
from pathlib import Path
import re
import tomllib
from xml.etree import ElementTree as ET

root = Path(__file__).resolve().parents[2]
version = "2.0.0"

def read(path):
    return (root / path).read_text(encoding="utf-8")

versions = {
    "generator": re.search(r'const version = "([^"]+)"', read("cmd/main.go")).group(1),
    "npm JS/TS": json.loads(read("std/javascript/package.json"))["version"],
    "Python": tomllib.loads(read("std/python/pyproject.toml"))["project"]["version"],
    "Rust": tomllib.loads(read("std/rust/Cargo.toml"))["package"]["version"],
    "Rust lock": next(package["version"] for package in tomllib.loads(read("std/rust/Cargo.lock"))["package"] if package["name"] == "benc"),
    "Java": ET.fromstring(read("std/java/pom.xml")).findtext("{*}version"),
    "Kotlin": re.search(r'^version = "([^"]+)"', read("std/kotlin/build.gradle.kts"), re.M).group(1),
    "Scala": re.search(r'^version := "([^"]+)"', read("std/scala/build.sbt"), re.M).group(1),
    "C#": ET.fromstring(read("std/c#/benc/benc.csproj")).findtext(".//Version"),
    "C": re.search(r'project\(benc_c VERSION ([0-9.]+)', read("std/c/CMakeLists.txt")).group(1),
    "C++": re.search(r'project\(benc VERSION ([0-9.]+)', read("std/c++/CMakeLists.txt")).group(1),
}
for name, actual in versions.items():
    if actual != version:
        raise SystemExit(f"{name}: version {actual!r}, expected {version}")

for path in ["std/javascript/LICENSE", "std/python/LICENSE", "std/rust/LICENSE",
             "std/swift/LICENSE", "std/java/src/main/resources/META-INF/LICENSE",
             "std/kotlin/src/main/resources/META-INF/LICENSE",
             "std/scala/src/main/resources/META-INF/LICENSE"]:
    if read(path) != read("LICENSE"):
        raise SystemExit(f"{path}: license differs from root")
if f"Release: `{version}`" not in read("spec/compatibility.md"):
    raise SystemExit("compatibility release version missing")
print(f"Benc {version} / wire v1 metadata: OK")
