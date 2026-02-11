# Benc (Polyglot Fork)

<!-- Update with your repo details -->

<!-- Update with your repo details -->

  

**Benc** is a high-performance binary serialization format.

This project is a **major fork** of [deneonet/benc](https://github.com/deneonet/benc). While the original project focuses on Go efficient serialization, this fork introduces a comprehensive **Code Generator** and **Runtime Libraries (`std`)** for a wide variety of languages, turning Go structs into a universal Interface Definition Language (IDL).

Unlike the original implementation, this repository enables cross-language data exchange by generating native serialization code for Go, C, C++, C#, Java, Kotlin, Scala, Rust, Python, Swift, TypeScript, JavaScript.

## 🚀 Getting Started

### 1. Installation

Install the Benc CLI generator tool:

```
go install github.com/yourusername/benc/cmd@latest
```

### 2. Define your Schema (IDL)

Benc uses **Go structs** as the schema definition. Create a simple `.go` file describing your data types.
```
package schema

// MyData will be serialized across languages
type MyData struct {
    ID      uint64
    Name    string
    Tags    []string
    Payload []byte
}
```

### 3. Generate Code

Use the CLI to generate serialization logic for your target languages.

```
# Generate C++ code
benc --lang cpp --out ./cpp_client/generated.hpp --input ./schema/schema.go

# Generate TypeScript code
benc --lang ts --out ./web_client/generated.ts --input ./schema/schema.go
```

### 4. Integration
Import the generated code and the Benc standard library (`std`) for your specific language.

**Example (C++):**

```
#include "std/c++/benc.hpp"
#include "generated.hpp"

int main() {
    MyData data;
    data.id = 12345;
    data.name = "Benc Fork";
    
    // Serialize
    std::vector<uint8_t> buffer = benc::serialize(data);
    
    // Deserialize
    MyData result = benc::deserialize<MyData>(buffer);
}
```

##  Project Structure

-   **`cmd/`**: Contains the CLI tool source code. This includes the parsers (for Go structs) and the generators for all supported target languages.
-   **`std/`**: The runtime standard libraries required by the generated code. These are minimal, efficient implementations of the Benc encoding spec for every supported language.
-   **`test/`**: Integration tests ensuring compatibility between the Go implementation and generated ports.
    

## Contributing

We welcome contributions, especially for adding new language ports or optimizing existing generators.

1.  Check the [CONTRIBUTING.md](CONTRIBUTING.md) file.
2.  Ensure you run tests for the specific language `std` lib you are modifying.
3.  If modifying the protocol, update the `cmd/generator` logic to reflect changes across all languages.
    

## License

[MIT](LICENSE)