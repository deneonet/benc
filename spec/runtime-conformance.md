# Runtime conformance for wire v1

This file documents tests of the runtimes, not a schema language. Source
declarations still pass through Go AST before target generation.

`testdata/conformance/wire-v1.json` is the independent byte corpus. Decimal
strings represent all 64-bit numbers; `hex` is the expected wire value. An
invalid vector has a broad error class (`truncated`, `overflow`/`limit`,
`invalid data`/`invalid terminator`). No decoder may return a successful
cursor for an invalid vector. Exact bytes are required for the valid vectors
except maps with more than one entry, whose key iteration order is not
canonical. The normal map profile rejects duplicate decoded keys; generic
skip functions cannot compare keys.

Decoder budgets per field are 64 MiB and 1,048,576 elements. C and Go offer
explicit depth-aware composite APIs; callers with recursive declarations must
forward their depth state. The other runtimes bound a single collection and
payload but do not yet expose a common recursive depth parameter.

| Runtime | Shared JSON corpus run here | Other tests | Caveat |
| --- | --- | --- | --- |
| Go | yes, `std/golang/hardening_test.go` | `go test ./...` | `int`/`uint` are native width |
| C | targeted strict vectors | C tests and hardening test | Generic map owns parallel key/value arrays; key equality is caller-specific |
| C++ | targeted strict vectors | C++20 test | Native-width varints |
| JavaScript | yes, Node built-in test | 35 cases | Exact timestamp uses `Time` |
| TypeScript | yes, generated JS via same Node test | 35 cases | Rebuild checked with `build.js --check` |
| Python | yes, `unittest` | existing unit tests | Exact timestamp uses `*_time_nanos` |
| Rust | targeted strict vectors | `cargo test`, generated native smoke | Native CI requires Cargo/rustc |
| Java | targeted strict vectors | `mvn test`, generated Java pipeline | Maven package built in CI |
| Kotlin | targeted strict vectors | `gradle test`, generated native smoke | JVM toolchain required |
| Scala | targeted strict vectors | `sbt test`, generated native smoke | Scala 3 required |
| Swift | targeted strict vectors | `swift test`, generated native smoke | Swift 6 required |
| C# | targeted strict vectors | `dotnet test`, generated native smoke | .NET 9 SDK required |

Commands in an environment with all toolchains:

```sh
go test ./...
gcc -std=c17 -Wall -Wextra -Werror std/c/test.c -o /tmp/benc-c && /tmp/benc-c
gcc -std=c17 -Wall -Wextra -Werror std/c/hardening_test.c -o /tmp/benc-hardening && /tmp/benc-hardening
g++ -std=c++20 -Wall -Wextra -Werror std/c++/test.cpp -o /tmp/benc-cpp && /tmp/benc-cpp
python3 -m unittest discover -s std/python -p '*test.py'
npm --prefix std/javascript test
bash testdata/conformance/package-smoke.sh javascript
cargo test --manifest-path std/rust/Cargo.toml
mvn -f std/java/pom.xml test
gradle -p std/kotlin test
(cd std/scala && sbt test)
swift test --package-path std/swift
dotnet test 'std/c#/benc.sln'
```

CI runs these commands on every PR and on main, including installed package
smoke checks for JS/TS and Python and C/C++ ASan/UBSan checks. The manual and
weekly native workflow compiles each source-language parser into every native
target. The shared JSON corpus is executable in Go, JavaScript/TypeScript and
Python. C and C++ exercise selected corpus failures through native tests;
the other packages have targeted regression cases, not a complete JSON-driven
runner yet. CI results are the authority for toolchains unavailable locally.

JavaScript is the executable source shared with TypeScript. The single
`std/javascript` npm package contains `std.js` and `std.d.ts`. `npm run build`
regenerates the declarations from the JS export list; tests run the JS against
the independent vectors, `--check` rejects stale declarations, and generated
TypeScript output is checked by `tsc`. No second wire implementation or
separate TypeScript package is maintained.

This changes strict decode of malformed data: booleans other than `00/01`,
wrong collection trailers, invalid UTF-8 strings and duplicate map keys
produce errors. Raw arbitrary bytes remain legal as `[]byte`. Producers
should use exact nanosecond APIs when the host date type has less than
nanosecond precision. A timestamp outside signed-64 nanoseconds is rejected,
not silently replaced by epoch. The canonical byte order for multi-entry
maps and a legacy decoder are not part of this stage.
