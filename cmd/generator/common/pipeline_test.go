package common_test

import (
	"fmt"
	"github.com/banditmoscow1337/benc/cmd/generator/c"
	"github.com/banditmoscow1337/benc/cmd/generator/common"
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
	"go/ast"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

const modelSource = `package models
import "time"
type Root struct {
 S string
 I int
 U uint
 Count uint8
 Fixed [3]byte
 Numbers [2]int32
 Raw []byte
 Children []Child
 Groups [][]byte
 P *Child
 Dict map[string]Child
 Named Labels
 When time.Time
}
type Labels map[string]int32
type Child struct { ID int64; Text string; Scores []uint8 }
`

func run(t *testing.T, dir, program string, args ...string) string {
	t.Helper()
	cmd := exec.Command(program, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", program, args, err, out)
	}
	return string(out)
}
func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}
func parse(t *testing.T, path string) *common.Context {
	t.Helper()
	ctx := common.NewContext(path)
	var err error
	switch filepath.Ext(path) {
	case ".go":
		err = golang.Parse(ctx)
	case ".h":
		err = c.Parse(ctx)
	case ".hpp":
		err = cpp.Parse(ctx)
	case ".js":
		err = javascript.Parse(ctx)
	case ".ts":
		err = typescript.Parse(ctx)
	case ".py":
		err = python.Parse(ctx)
	case ".rs":
		err = rust.Parse(ctx)
	case ".java":
		err = java.Parse(ctx)
	case ".kt":
		err = kotlin.Parse(ctx)
	case ".cs":
		err = csharp.Parse(ctx)
	case ".swift":
		err = swift.Parse(ctx)
	case ".scala":
		err = scala.Parse(ctx)
	default:
		t.Fatalf("unknown source %s", path)
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Validate(); err != nil {
		t.Fatal(err)
	}
	return ctx
}
func canonical(ctx *common.Context, e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		if x.Name == "byte" {
			return "uint8"
		}
		if x.Name == "rune" {
			return "int32"
		}
		return x.Name
	case *ast.ArrayType:
		if x.Len != nil {
			return "[" + ctx.ExprToString(x.Len) + "]" + canonical(ctx, x.Elt)
		}
		return "[]" + canonical(ctx, x.Elt)
	case *ast.MapType:
		return "map[" + canonical(ctx, x.Key) + "]" + canonical(ctx, x.Value)
	case *ast.StarExpr:
		return "*" + canonical(ctx, x.X)
	default:
		return ctx.ExprToString(e)
	}
}
func signature(ctx *common.Context) map[string][]string {
	result := map[string][]string{}
	for _, ts := range ctx.Types {
		if st, ok := ts.Type.(*ast.StructType); ok {
			for _, f := range st.Fields.List {
				if ctx.ShouldIgnoreField(f) {
					continue
				}
				for _, n := range f.Names {
					result[ts.Name.Name] = append(result[ts.Name.Name], n.Name+":"+canonical(ctx, f.Type))
				}
			}
			if len(st.Fields.List) == 0 {
				result[ts.Name.Name] = []string{}
			}
		} else {
			result[ts.Name.Name] = []string{canonical(ctx, ts.Type)}
		}
	}
	return result
}
func generate(t *testing.T, ctx *common.Context) {
	t.Helper()
	for _, gen := range []common.Generator{golang.New(ctx), c.New(ctx), cpp.New(ctx), javascript.New(ctx), typescript.New(ctx), python.New(ctx), rust.New(ctx), java.New(ctx), kotlin.New(ctx), csharp.New(ctx), swift.New(ctx), scala.New(ctx)} {
		if err := gen.Generate(); err != nil {
			t.Fatal(err)
		}
		if err := gen.Tests(); err != nil {
			t.Fatal(err)
		}
	}
}

func generatedFiles(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "model_benc*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("generator produced no files")
	}
	files := make(map[string][]byte, len(paths))
	for _, path := range paths {
		files[filepath.Base(path)], err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

func requireTools(t *testing.T, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			if os.Getenv("CI") == "true" {
				t.Fatalf("required compiler %s unavailable: %v", name, err)
			}
			t.Skipf("%s unavailable", name)
		}
	}
}

// Each source parser must produce the same Go AST and feed every target.
// Generated tests compile and execute against the actual runtime libraries.
func TestPipelineMatrix(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	goTool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if _, err := os.Stat(goTool); err != nil {
		t.Skip("Go toolchain unavailable")
	}
	requireTools(t, "gcc", "g++", "node", "python3", "java")
	fixture := t.TempDir()
	original := filepath.Join(fixture, "model.go")
	write(t, original, modelSource)
	source := parse(t, original)
	want := signature(source)
	generate(t, source)
	jsRuntime, err := os.ReadFile(filepath.Join(root, "std", "javascript", "std.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct{ name, path string }{{"go", original}, {"c", filepath.Join(fixture, "model_benc.h")}, {"cpp", filepath.Join(fixture, "model_benc.hpp")}, {"js", filepath.Join(fixture, "model_benc.js")}, {"ts", filepath.Join(fixture, "model_benc.d.ts")}, {"py", filepath.Join(fixture, "model_benc.py")}, {"rs", filepath.Join(fixture, "model_benc.rs")}, {"java", filepath.Join(fixture, "model_benc.java")}, {"kt", filepath.Join(fixture, "model_benc.kt")}, {"cs", filepath.Join(fixture, "model_benc.cs")}, {"swift", filepath.Join(fixture, "model_benc.swift")}, {"scala", filepath.Join(fixture, "model_benc.scala")}} {
		t.Run(input.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "model"+filepath.Ext(input.path))
			data, err := os.ReadFile(input.path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			ctx := parse(t, path)
			if got := signature(ctx); !reflect.DeepEqual(got, want) {
				t.Fatalf("Go AST mismatch\ngot:  %#v\nwant: %#v", got, want)
			}
			generate(t, ctx)
			first := generatedFiles(t, dir)
			generate(t, ctx)
			if second := generatedFiles(t, dir); !reflect.DeepEqual(first, second) {
				t.Fatal("generated files changed on repeated generation")
			}
			if err := os.WriteFile(filepath.Join(dir, "std.js"), jsRuntime, 0644); err != nil {
				t.Fatal(err)
			}
			pyRuntime, err := os.ReadFile(filepath.Join(root, "std", "python", "bstd.py"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "bstd.py"), pyRuntime, 0644); err != nil {
				t.Fatal(err)
			}
			run(t, dir, "python3", "model_benc_test.py")
			run(t, dir, "node", "model_benc_test.js")
			javaSources := []string{"java", "com.sun.tools.javac.Main", "-d", dir}
			for _, file := range []string{"Bstd.java", "BencException.java", "BencBufferTooSmallException.java", "BencOverflowException.java"} {
				javaSources = append(javaSources, filepath.Join(root, "std", "java", "src", "main", "java", "main", "com", "benc", file))
			}
			javaSources = append(javaSources, "model_benc.java", "model_benc_test.java")
			run(t, dir, javaSources[0], javaSources[1:]...)
			run(t, dir, "java", "-cp", dir, "BencGeneratedTest")
			run(t, dir, "gcc", "-std=c17", "-Wall", "-Wextra", "-Werror", "-DBSTD_IMPLEMENTATION", "-x", "c", "-c", filepath.Join(root, "std", "c", "benc.h"), "-o", "bstd.o")
			run(t, dir, "gcc", "-std=c17", "-Wall", "-Wextra", "-Werror", "-I"+filepath.Join(root, "std", "c"), "model_benc.c", "model_benc_test.c", "bstd.o", "-o", "c-test")
			run(t, dir, filepath.Join(dir, "c-test"))
			run(t, dir, "g++", "-std=c++20", "-Wall", "-Wextra", "-Werror", "-I"+filepath.Join(root, "std", "c++"), "model_benc_test.cpp", "-o", "cpp-test")
			run(t, dir, filepath.Join(dir, "cpp-test"))
			write(t, filepath.Join(dir, "go.mod"), fmt.Sprintf("module generated\n\ngo 1.25.6\nrequire github.com/banditmoscow1337/benc v0.0.0\nreplace github.com/banditmoscow1337/benc => %s\n", root))
			goDir := filepath.Join(dir, "gopkg")
			if err := os.Mkdir(goDir, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"model_benc.go", "model_benc_test.go"} {
				data, err := os.ReadFile(filepath.Join(dir, file))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(goDir, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if input.name == "go" {
				write(t, filepath.Join(goDir, "model.go"), modelSource)
			}
			run(t, dir, goTool, "test", "-mod=mod", "./gopkg")
		})
	}
}

// Compare a nonzero value byte-for-byte across all four generated encoders.
// It covers embedded NUL, varints, arrays, bytes, nested collections, named
// maps and a time value with nanoseconds beyond Date's millisecond precision.
func TestGeneratedWireValues(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	goTool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if _, err := os.Stat(goTool); err != nil {
		t.Skip("Go toolchain unavailable")
	}
	requireTools(t, "gcc", "g++", "node", "python3", "java")
	dir := t.TempDir()
	write(t, filepath.Join(dir, "model.go"), modelSource)
	generate(t, parse(t, filepath.Join(dir, "model.go")))
	for _, ext := range []string{"c", "cpp", "js", "py", "java"} {
		data, err := os.ReadFile(filepath.Join(root, "testdata", "conformance", "values", "value."+ext))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "value."+ext), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	jsRuntime, err := os.ReadFile(filepath.Join(root, "std", "javascript", "std.js"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "std.js"), jsRuntime, 0644); err != nil {
		t.Fatal(err)
	}
	pyRuntime, err := os.ReadFile(filepath.Join(root, "std", "python", "bstd.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bstd.py"), pyRuntime, 0644); err != nil {
		t.Fatal(err)
	}
	goDir := filepath.Join(dir, "gopkg")
	if err := os.Mkdir(goDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"model.go", "model_benc.go", "value_test.go"} {
		var data []byte
		if file == "value_test.go" {
			data, err = os.ReadFile(filepath.Join(root, "testdata", "conformance", "values", file))
		} else {
			data, err = os.ReadFile(filepath.Join(dir, file))
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(goDir, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(t, filepath.Join(dir, "go.mod"), fmt.Sprintf("module generated\n\ngo 1.25.6\nrequire github.com/banditmoscow1337/benc v0.0.0\nreplace github.com/banditmoscow1337/benc => %s\n", root))
	goOutput := run(t, dir, goTool, "test", "-mod=mod", "-run", "TestWireValue", "-v", "./gopkg")
	var want string
	for _, line := range strings.Split(goOutput, "\n") {
		if strings.HasPrefix(line, "WIRE=") {
			want = strings.TrimPrefix(line, "WIRE=")
		}
	}
	if want == "" {
		t.Fatalf("Go reference emitted no wire: %s", goOutput)
	}
	run(t, dir, "gcc", "-std=c17", "-Wall", "-Wextra", "-Werror", "-DBSTD_IMPLEMENTATION", "-x", "c", "-c", filepath.Join(root, "std", "c", "benc.h"), "-o", "bstd.o")
	run(t, dir, "gcc", "-std=c17", "-Wall", "-Wextra", "-Werror", "-I"+filepath.Join(root, "std", "c"), "model_benc.c", "value.c", "bstd.o", "-o", "value-c")
	run(t, dir, "g++", "-std=c++20", "-Wall", "-Wextra", "-Werror", "-I"+filepath.Join(root, "std", "c++"), "value.cpp", "-o", "value-cpp")
	javaSources := []string{"com.sun.tools.javac.Main", "-d", dir}
	for _, file := range []string{"Bstd.java", "BencException.java", "BencBufferTooSmallException.java", "BencOverflowException.java"} {
		javaSources = append(javaSources, filepath.Join(root, "std", "java", "src", "main", "java", "main", "com", "benc", file))
	}
	javaSources = append(javaSources, "model_benc.java", "value.java")
	run(t, dir, "java", javaSources...)
	for name, got := range map[string]string{
		"C":      run(t, dir, filepath.Join(dir, "value-c")),
		"C++":    run(t, dir, filepath.Join(dir, "value-cpp")),
		"JS":     run(t, dir, "node", "value.js"),
		"Python": run(t, dir, "python3", "value.py"),
		"Java":   run(t, dir, "java", "-cp", dir, "BencValue"),
	} {
		if strings.TrimSpace(got) != want {
			t.Errorf("%s wire %q; Go wire %q", name, strings.TrimSpace(got), want)
		}
	}
}

func TestUnsupportedDeclarationsFailBeforeGeneration(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"unknown", "type X struct { Field Mystery }", "X.Field"},
		{"interface", "type X struct { Field any }", "X.Field"},
		{"embedded", "type X struct { Y }; type Y struct { A int }", "embedded fields"},
		{"recursive", "type X struct { P *X }", "recursive type"},
		{"badmap", "type X struct { M map[float64]int }", "floating map keys"},
		{"empty", "type X struct {}", "empty structs"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "bad.go")
			write(t, path, "package bad\n"+tc.body)
			ctx := common.NewContext(path)
			if err := golang.Parse(ctx); err != nil {
				t.Fatal(err)
			}
			if err := ctx.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate() = %v; want %q", err, tc.want)
			}
			if err := cpp.New(ctx).Generate(); err == nil {
				t.Fatal("generator accepted invalid AST")
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), "bad_benc.hpp")); !os.IsNotExist(err) {
				t.Fatalf("invalid AST produced output: %v", err)
			}
		})
	}
	for _, tc := range []struct{ extension, body string }{
		{".js", "class X {\n constructor() {\n this.X = [];\n }\n}\n"},
		{".h", "typedef struct { char* Name; } X;"},
	} {
		path := filepath.Join(t.TempDir(), "bad"+tc.extension)
		write(t, path, tc.body)
		ctx := common.NewContext(path)
		var err error
		if tc.extension == ".js" {
			err = javascript.Parse(ctx)
		} else {
			err = c.Parse(ctx)
		}
		if err == nil {
			t.Fatalf("%s: ambiguous field accepted", tc.extension)
		}
	}
	for _, tc := range []struct{ extension, body string }{
		{".h", "typedef struct {\n bool Flag;\n Mystery Bad;\n} Record;"},
		{".js", "class Record {\n constructor() {\n this.Flag = true;\n this.Bad = null; // Mystery\n }\n}"},
	} {
		path := filepath.Join(t.TempDir(), "record"+tc.extension)
		write(t, path, tc.body)
		ctx := common.NewContext(path)
		var err error
		if tc.extension == ".h" {
			err = c.Parse(ctx)
		} else {
			err = javascript.Parse(ctx)
		}
		if err != nil {
			t.Fatal(err)
		}
		err = ctx.Validate()
		if err == nil || !strings.Contains(err.Error(), "Record.Bad") || !strings.Contains(err.Error(), ":3:") && tc.extension == ".h" || !strings.Contains(err.Error(), ":4:") && tc.extension == ".js" {
			t.Fatalf("%s: expected positioned field error, got %v", tc.extension, err)
		}
	}
}

func TestExplicitIgnoreOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "record.go")
	write(t, path, "package example\ntype Record struct {\n ID int32\n Bad chan int //benc:ignore\n}\n")
	ctx := parse(t, path)
	fields := ctx.GetSupportedFields(ctx.Types[0])
	if len(fields) != 1 || fields[0].Names[0].Name != "ID" {
		t.Fatalf("fields after explicit ignore = %#v", fields)
	}
}

// A field that a source parser cannot map must report its location instead of
// silently disappearing from the AST and corrupting the wire field order.
func TestNewSourceParsersRejectUnknownField(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "model.go")
	write(t, input, modelSource)
	generate(t, parse(t, input))
	for _, tc := range []struct{ ext, old, replacement string }{
		{"hpp", "std::string Text;", "Mystery Text;"},
		{"d.ts", "S: string; // string", "S: string; // Mystery"},
		{"py", "Text: str = \"\" # string", "Text: str = \"\" # Mystery"},
		{"rs", "pub Text: String, // string", "pub Text: String, // Mystery"},
		{"java", "public String Text = \"\"; // string", "public String Text = \"\"; // Mystery"},
		{"kt", "var Text: String = \"\" // string", "var Text: String = \"\" // Mystery"},
		{"cs", "public string Text = \"\"; // string", "public string Text = \"\"; // Mystery"},
		{"swift", "public var Text: String = \"\" // string", "public var Text: String = \"\" // Mystery"},
		{"scala", "Text: String = \"\", // string", "Text: String = \"\", // Mystery"},
	} {
		t.Run(tc.ext, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, "model_benc."+tc.ext))
			if err != nil {
				t.Fatal(err)
			}
			body := strings.Replace(string(data), tc.old, tc.replacement, 1)
			if body == string(data) {
				t.Fatalf("fixture marker missing: %q", tc.old)
			}
			path := filepath.Join(t.TempDir(), "invalid."+tc.ext)
			write(t, path, body)
			ctx := common.NewContext(path)
			var parseErr error
			switch tc.ext {
			case "hpp":
				parseErr = cpp.Parse(ctx)
			case "d.ts":
				parseErr = typescript.Parse(ctx)
			case "py":
				parseErr = python.Parse(ctx)
			case "rs":
				parseErr = rust.Parse(ctx)
			case "java":
				parseErr = java.Parse(ctx)
			case "kt":
				parseErr = kotlin.Parse(ctx)
			case "cs":
				parseErr = csharp.Parse(ctx)
			case "swift":
				parseErr = swift.Parse(ctx)
			case "scala":
				parseErr = scala.Parse(ctx)
			}
			if parseErr == nil {
				parseErr = ctx.Validate()
			}
			if parseErr == nil || !strings.Contains(parseErr.Error(), "Mystery") || !strings.Contains(parseErr.Error(), "invalid.") {
				t.Fatalf("expected positioned unknown field diagnostic, got %v", parseErr)
			}
		})
	}
}
