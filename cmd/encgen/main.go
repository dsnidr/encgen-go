package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
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
	showVersion := flag.Bool("version", false, "Outputs the version of encgen and exits")
	flag.Parse()

	if *showVersion {
		version := "unknown"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			version = info.Main.Version
		}

		fmt.Fprintf(os.Stdout, "encgen version: %s\n", version)
		return 0
	}

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
