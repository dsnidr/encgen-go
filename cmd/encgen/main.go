package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dsnidr/encgen-go/generator"
	"github.com/dsnidr/encgen-go/parser"
)

func main() {
	os.Exit(run())
}

func run() int {
	name := flag.String("name", "", "The name of the input struct")
	inpath := flag.String("inpath", ".", "The location of the input struct")
	outpath := flag.String("outpath", ".", "The output location you want the generator code written to")
	flag.Parse()

	if strings.TrimSpace(*name) == "" {
		fmt.Fprintln(os.Stderr, "the --name flag is required")
		return 1
	}

	parsedStruct, err := parser.ParseStruct(*inpath, *name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "struct parse failed: %v\n", err)
		return 1
	}

	if err := generator.GenerateEncoder(parsedStruct, *outpath); err != nil {
		fmt.Fprintf(os.Stderr, "error generating encoder: %v\n", err)
		return 1
	}

	return 0
}
