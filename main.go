package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/andrew-d/sleepywolf/common"
)

var _ = common.StructInfo{}
var _ = template.Must

const gatherTemplate = `
// DO NOT EDIT!!!
// This code was generated by github.com/andrew-d/sleepywolf

package main

import (
	"os"

	"github.com/andrew-d/sleepywolf/gather"

	// This is the package we're introspecting
	"{{.ImportPath}}"
)

{{$pn := .PackageName}}

func main() {
	g := gather.NewInfoGatherer()
{{range .StructNames}}
	g.Register(&{{$pn}}.{{.}}{})
{{end}}
	g.Run(os.Stdout)
}
`

var (
	extractFnameRe = regexp.MustCompile(`(.*)(\.go)$`)
	verbose        = flag.Bool("v", false, "print information while generating")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\t%s [options] [input_file]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s generates Go code to link up resources with Goji\n\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		usage()
	}

	inputPath := filepath.ToSlash(args[0])
	outputPath := extractFnameRe.ReplaceAllString(args[0], `${1}_goji.go`)
	_ = outputPath

	// Step 1: obtain information about the input file
	packageName, structs, err := GetFileInfo(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting file info: %s\n", err)
		return
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Package Name : %s\n", packageName)
		for _, s := range structs {
			fmt.Fprintf(os.Stderr, "  Struct     : %s\n", s)
		}
	}

	// Step 2: Find the import path of this file
	importPath, err := getImportPath(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't find input file in $GOPATH")
		return
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Import Path  : %s\n", importPath)
	}

	// Step 3: Generate a template that will extract information about each of
	// the structs we've already found.
	tmpl := template.Must(template.New("gather_gen.go").Parse(gatherTemplate))
	err = tmpl.Execute(os.Stdout, struct {
		ImportPath  string
		PackageName string
		StructNames []string
	}{importPath, packageName, structs})

	// Step 4: Run this file
}
