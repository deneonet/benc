package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"go.kine.bz/benc/cmd/bencgen/codegens"
	"go.kine.bz/benc/cmd/bencgen/parser"
)

var iFlag = flag.String("in", "", "the input .benc file")
var oFlag = flag.String("out", "", "the output directory")
var nFlag = flag.String("file", "", "the name of the output file (no extension)")
var fFlag = flag.Bool("force", false, "disables the breaking-change detector")
var lFlag = flag.String("lang", "", "the language of the code that should be generated")

func printError(m string) {
	errorMessage := "\n\033[1;31m[bencgen] Error:\033[0m\n"
	errorMessage += fmt.Sprintf("    \033[1;37mMessage:\033[0m %s\n", m)
	fmt.Println(errorMessage)
	os.Exit(-1)
}

func main() {
	flag.Usage = func() {
		fmt.Println("Usage for bencgen available here: https://github.com/deneonet/benc/cmd/bencgen#usage")
	}
	flag.Parse()

	if len(*iFlag) == 0 {
		printError("Missing input file: --in ...")
		return
	}

	if len(*lFlag) == 0 {
		printError("Missing language: --lang ...")
		return
	}

	lang := codegens.GeneratorLanguage(*lFlag)
	generator := codegens.NewGenerator(lang, *iFlag)
	if generator == nil {
		printError("Unknown language provided.")
		return
	}

	start := time.Now()

	file, err := os.Open(*iFlag)
	if err != nil {
		panic(err)
	}

	content, err := os.ReadFile(*iFlag)
	if err != nil {
		panic(err)
	}

	parser := parser.NewParser(file, string(content))
	nodes := parser.Parse()

	/*bcd := bcd.NewBcd(nodes, *iFlag)
	bcd.Analyze(*fFlag)*/

	outCode := codegens.Generate(generator, nodes)

	dn := "out"
	if len(*oFlag) != 0 {
		dn = *oFlag
	}

	fn := *iFlag
	if len(*nFlag) != 0 {
		fn = *nFlag
	}

	if strings.Contains(fn, "/") {
		out := strings.Split(fn, "/")
		fn = out[len(out)-1]
	}

	ffn := fn + "." + *lFlag

	err = os.MkdirAll(dn, os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(dn+"/"+ffn, []byte(outCode), os.ModePerm)
	if err != nil {
		panic(err)
	}

	elapsed := time.Since(start)

	successMessage := "\n\033[1;92m[bencgen] Success:\033[0m\n"
	successMessage += fmt.Sprintf("    \033[37mCompiled into `%s`, in %dms, saved at: %s !\033[0m\n", *lFlag, elapsed.Milliseconds(), dn+"/"+ffn)
	fmt.Println(successMessage)
}
