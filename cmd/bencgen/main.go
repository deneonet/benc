package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deneonet/benc/cmd/bencgen/bcd"
	"github.com/deneonet/benc/cmd/bencgen/codegens"
	"github.com/deneonet/benc/cmd/bencgen/parser"
)

var iFlag = flag.String("in", "", "comma-separated list of input .benc files")
var oFlag = flag.String("out", "", "the output directory")
var nFlag = flag.String("file", "", "custom output filename (use '...' to replace with input filename)")
var fFlag = flag.Bool("force", false, "disables the breaking-change detector")
var lFlag = flag.String("lang", "", "the language of the code that should be generated")
var dFlag = flag.String("import-dir", "", "comma-separated list of import directories")

func printError(m string) {
	errorMessage := "\n\033[1;31m[bencgen] Error:\033[0m\n"
	errorMessage += fmt.Sprintf("    \033[1;37mMessage:\033[0m %s\n", m)
	fmt.Println(errorMessage)
	os.Exit(-1)
}

func processFile(inputFile string, outputDir string, filenamePattern string, lang codegens.GenLang, importDirs []string) {
	generator := codegens.NewGen(lang, inputFile)
	if generator == nil {
		printError("Unknown language provided.")
	}

	start := time.Now()

	content, err := os.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	parser := parser.NewParser(strings.NewReader(string(content)), string(content))
	nodes := parser.Parse()

	outCode := codegens.Generate(generator, nodes, importDirs)

	bcd := bcd.NewBcd(nodes, inputFile)
	bcd.Analyze(*fFlag)

	fileBase := strings.TrimSuffix(filepath.Base(inputFile), ".benc")
	outputFilename := strings.ReplaceAll(filenamePattern, "...", fileBase) + "." + *lFlag

	outputDir = strings.ReplaceAll(outputDir, "...", fileBase)

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic(err)
	}

	outputPath := filepath.Join(outputDir, outputFilename)
	if err := os.WriteFile(outputPath, []byte(outCode), os.ModePerm); err != nil {
		panic(err)
	}

	elapsed := time.Since(start)

	successMessage := "\n\033[1;92m[bencgen] Success:\033[0m\n"
	successMessage += fmt.Sprintf("    \033[37mCompiled %s into `%s`, in %dms, saved at: %s !\033[0m\n", inputFile, *lFlag, elapsed.Milliseconds(), outputPath)
	fmt.Println(successMessage)
}

func main() {
	flag.Usage = func() {
		fmt.Println("Usage for bencgen available here: https://github.com/deneonet/benc/tree/main/cmd/bencgen#usage")
	}
	flag.Parse()

	if len(*iFlag) == 0 {
		printError("Missing input file(s): --in ...")
	}

	if len(*lFlag) == 0 {
		printError("Missing language: --lang ...")
	}

	var importDirs []string
	if len(*dFlag) > 0 {
		importDirs = strings.Split(*dFlag, ",")
	}

	lang := codegens.GenLang(*lFlag)
	outputDir := "out"
	if len(*oFlag) != 0 {
		outputDir = *oFlag
	}

	filenamePattern := "..."
	if len(*nFlag) != 0 {
		filenamePattern = *nFlag
	}

	inputFiles := strings.Split(*iFlag, ",")

	for _, inputFile := range inputFiles {
		inputFile = strings.TrimSpace(inputFile)
		if inputFile == "" {
			continue
		}
		processFile(inputFile, outputDir, filenamePattern, lang, importDirs)
	}
}
