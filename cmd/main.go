package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/banditmoscow1337/benc/cmd/generator/c"
	"github.com/banditmoscow1337/benc/cmd/generator/cpp"
	"github.com/banditmoscow1337/benc/cmd/generator/csharp"
	"github.com/banditmoscow1337/benc/cmd/generator/golang"
	"github.com/banditmoscow1337/benc/cmd/generator/java"
	"github.com/banditmoscow1337/benc/cmd/generator/javascript"
	"github.com/banditmoscow1337/benc/cmd/generator/kotlin"
	"github.com/banditmoscow1337/benc/cmd/generator/python"
	"github.com/banditmoscow1337/benc/cmd/generator/rust"
	"github.com/banditmoscow1337/benc/cmd/generator/scala"
	"github.com/banditmoscow1337/benc/cmd/generator/swift"
	"github.com/banditmoscow1337/benc/cmd/generator/typescript"

	"github.com/banditmoscow1337/benc/cmd/generator/common"
)

const version = "2.0.0"

func main() {
	langFlag := flag.String("lang", "go", "Comma separated target generators: go, js, ts, py, rs, java, kt, cs, swift, scala, c, cpp")
	listFlag := flag.Bool("list-languages", false, "List the implemented parsers and generators")
	versionFlag := flag.Bool("version", false, "Print generator and wire version")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("benc %s (wire v1)\n", version)
		return
	}
	if *listFlag {
		fmt.Println("Parsers: .go, .c, .h, .cpp, .hpp, .js, .ts, .py, .rs, .java, .kt, .cs, .swift, .scala\nGenerators: go, js, ts, py, rs, java, kt, cs, swift, scala, c, cpp\nRun go test ./... and the generated native smoke workflow for compiler and wire validation.")
		return
	}

	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Usage: benc -lang=go,js,ts,py,rs,java,kt,cs,swift,scala,c,cpp <input_file.go|.c|.h|.cpp|.hpp|.js|.ts|.py|.rs|.java|.kt|.cs|.swift|.scala>")
	}
	langs := strings.Split(*langFlag, ",")
	for i, lang := range langs {
		langs[i] = strings.TrimSpace(lang)
		switch langs[i] {
		case "go", "js", "ts", "py", "rs", "java", "kt", "cs", "swift", "scala", "c", "cpp":
		default:
			log.Fatalf("unknown generator %q (available: go, js, ts, py, rs, java, kt, cs, swift, scala, c, cpp)", lang)
		}
	}
	switch filepath.Ext(args[0]) {
	case ".go", ".c", ".h", ".cpp", ".hpp", ".js", ".ts", ".py", ".rs", ".java", ".kt", ".cs", ".swift", ".scala":
	default:
		log.Fatalf("unsupported input extension %q", filepath.Ext(args[0]))
	}
	if _, err := os.Stat(args[0]); err != nil {
		log.Fatal(err)
	}

	ctx := common.NewContext(args[0])

	// Detect Input Type
	if strings.HasSuffix(ctx.InputFile, ".scala") {
		if err := scala.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".swift") {
		if err := swift.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".cs") {
		if err := csharp.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".kt") {
		if err := kotlin.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".java") {
		if err := java.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".rs") {
		if err := rust.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".py") {
		if err := python.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".ts") {
		if err := typescript.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".js") {
		if err := javascript.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".cpp") || strings.HasSuffix(ctx.InputFile, ".hpp") {
		if err := cpp.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(ctx.InputFile, ".c") || strings.HasSuffix(ctx.InputFile, ".h") {
		if err := c.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := golang.Parse(ctx); err != nil {
			log.Fatal(err)
		}
	}

	if err := ctx.Validate(); err != nil {
		log.Fatal(err)
	}

	for _, lang := range langs {
		var generator common.Generator
		switch lang {
		case "go":
			generator = golang.New(ctx)
		case "js":
			generator = javascript.New(ctx)
		case "ts":
			generator = typescript.New(ctx)
		case "py":
			generator = python.New(ctx)
		case "rs":
			generator = rust.New(ctx)
		case "java":
			generator = java.New(ctx)
		case "kt":
			generator = kotlin.New(ctx)
		case "cs":
			generator = csharp.New(ctx)
		case "swift":
			generator = swift.New(ctx)
		case "scala":
			generator = scala.New(ctx)
		case "c":
			generator = c.New(ctx)
		case "cpp":
			generator = cpp.New(ctx)
		}

		if err := generator.Generate(); err != nil {
			log.Fatalf("%s generation failed: %v", lang, err)
		}

		if err := generator.Tests(); err != nil {
			log.Fatalf("%s generation test failed: %v", lang, err)
		}
	}
}
