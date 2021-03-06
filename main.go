package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/andrew-d/sleepywolf/common"
)

var (
	extractFnameRe = regexp.MustCompile(`(.*)(\.go)$`)

	verbose       = flag.Bool("v", false, "print information while generating")
	keepGenerated = flag.Bool("keep", false, "keep the generated temp files")
	prefix        = flag.String("prefix", "/api", "prefix for generated URLs")
	writeToStdout = flag.Bool("stdout", false, "write the output to stdout instead of a file")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\t%s [options] [input_file]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s generates Go code to link up resources with Goji\n\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

// Get the name of the function to call to register the given handler
func RegisterFuncFor(funcName string) (string, error) {
	mapping := map[string]string{
		"DeleteOne":  "Delete",
		"DeleteMany": "Delete",
		"GetMany":    "Get",
		"GetOne":     "Get",
		"Patch":      "Patch",
		"Post":       "Post",
		"Put":        "Put",
	}

	s, ok := mapping[funcName]
	if !ok {
		return "", fmt.Errorf("unknown function name: %s", funcName)
	}
	return s, nil
}

// Helper function to generate a URL for a given resource / function pair
func UrlFor(structName, funcName string) (string, error) {
	mapping := map[string]string{
		"DeleteOne":  "%s/:id",
		"DeleteMany": "%s",
		"GetMany":    "%s",
		"GetOne":     "%s/:id",
		"Patch":      "%s/:id",
		"Post":       "%s",
		"Put":        "%s/:id",
	}

	format, ok := mapping[funcName]
	if !ok {
		return "", fmt.Errorf("unknown function name: %s", funcName)
	}

	// Remove any trailing "Resource" and lower-case
	structName = strings.TrimSuffix(structName, "Resource")
	structName = strings.ToLower(structName)

	// TODO: inflect the name of the resource to generate a real url

	return fmt.Sprintf(format, strings.ToLower(structName)), nil
}

// Helper function that, given the name of a "Before" function and a handler name,
// returns whether or not the handler should execute the Before function.
func HasBeforeType(funcName, beforeType string) (bool, error) {
	if beforeType == "BeforeAll" {
		return true, nil
	}

	mapping := map[string]string{
		"DeleteOne":  "BeforeOne",
		"DeleteMany": "BeforeMany",
		"GetMany":    "BeforeMany",
		"GetOne":     "BeforeOne",
		"Patch":      "BeforeOne",
		"Post":       "BeforeMany",
		"Put":        "BeforeOne",
	}

	f, ok := mapping[funcName]
	if !ok {
		return false, fmt.Errorf("unknown function name: %s", funcName)
	}

	return f == beforeType, nil
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		usage()
	}

	inputPath := filepath.ToSlash(args[0])
	outputPath := extractFnameRe.ReplaceAllString(args[0], `${1}_goji.go`)
	if *writeToStdout {
		fmt.Fprint(os.Stderr, "Output File   : STDOUT\n")
	} else {
		fmt.Fprintf(os.Stderr, "Output File   : %s\n", outputPath)
	}

	// Step 1: obtain information about the input file
	packageName, structs, err := GetFileInfo(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting file info: %s\n", err)
		return
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Package Name  : %s\n", packageName)
		for _, s := range structs {
			fmt.Fprintf(os.Stderr, "  Struct      : %s\n", s)
		}
	}

	// Step 2: Find the import path of this file
	importPath, err := getImportPath(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't find input file in $GOPATH\n")
		return
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Import Path   : %s\n", importPath)
	}

	// Step 3: Generate a template that will extract information about each of
	// the structs we've already found.
	tmpl := template.Must(template.New("gather_gen.go").Parse(gatherTemplate))
	gatherFile := bytes.Buffer{}
	err = tmpl.Execute(&gatherFile, struct {
		ImportPath  string
		PackageName string
		StructNames []string
	}{importPath, packageName, structs})

	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't execute template: %s\n", err)
		return
	}

	// Step 4: Create a temporary file and write the formatted code to it.
	tmpFile, err := common.TempFileWithSuffix("", "gather_gen", ".go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't create temp file: %s\n", err)
		return
	}
	// Order is LIFO, so we remove and then close, so the order is Close then
	// remove.
	if !*keepGenerated {
		defer os.Remove(tmpFile.Name())
	}
	defer tmpFile.Close()

	if *verbose {
		fmt.Fprintf(os.Stderr, "Temp File     : %s\n", tmpFile.Name())
	}

	err = common.GoFmt(tmpFile, &gatherFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't format generated code: %s\n", err)
		return
	}

	// Step 5: Run this file
	structInfoBuff := bytes.Buffer{}
	errBuff := bytes.Buffer{}
	cmd := exec.Command("go", "run", "-a", tmpFile.Name())
	cmd.Stdout = &structInfoBuff
	cmd.Stderr = &errBuff
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't run gather code: %s\n", err)
		return
	}

	// Step 6: Deserialize the struct info from the gather file.
	structInfos := []common.StructInfo{}
	err = json.NewDecoder(&structInfoBuff).Decode(&structInfos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't decode json from gather: %s\n", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Valid Structs : %d\n", len(structInfos))
		for _, s := range structInfos {
			fmt.Fprintf(os.Stderr, "  Struct '%s'\n", s.StructName)

			fmt.Fprintf(os.Stderr, "    Handlers   : ")
			for i, handler := range s.Handlers {
				if i > 0 {
					fmt.Fprintf(os.Stderr, ", ")
				}
				fmt.Fprintf(os.Stderr, "%s/%d", handler.Name, handler.Params)
			}
			fmt.Fprintf(os.Stderr, "\n")

			fmt.Fprintf(os.Stderr, "    BeforeOne  : %t\n", s.BeforeOne != nil)
			fmt.Fprintf(os.Stderr, "    BeforeMany : %t\n", s.BeforeMany != nil)
			fmt.Fprintf(os.Stderr, "    BeforeAll  : %t\n", s.BeforeAll != nil)

			if len(s.Warnings) > 0 {
				fmt.Fprintf(os.Stderr, "    Warnings   :\n")
				for _, w := range s.Warnings {
					fmt.Fprintf(os.Stderr, "      - %s\n", w)
				}
			}
		}
	}

	// Step 7a: Optionally open the output file.
	var outFile io.Writer
	if *writeToStdout {
		outFile = os.Stdout
	} else {
		f, err := os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't create output file: %s\n", err)
			return
		}
		defer f.Close()

		outFile = f
	}

	// Step 7b: Generate the final output
	funcMap := template.FuncMap{
		"RegisterFuncFor": RegisterFuncFor,
		"UrlFor":          UrlFor,
		"HasBeforeType":   HasBeforeType,
	}
	tmpl = template.Must(template.New("gather_gen.go").
		Funcs(funcMap).
		Parse(finalTemplate))

	finalBuff := bytes.Buffer{}
	err = tmpl.Execute(&finalBuff, struct {
		PackageName string
		Structs     []common.StructInfo
		UrlPrefix   string
	}{packageName, structInfos, *prefix})

	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't execute template: %s\n", err)
		return
	}

	err = common.GoFmt(outFile, &finalBuff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't format final code: %s\n", err)
		return
	}
}
