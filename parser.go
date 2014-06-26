package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// Returns (packageName, []structs, error)
func GetFileInfo(inputPath string) (string, []string, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, inputPath, nil, 0)

	if err != nil {
		return "", nil, err
	}

	packageName := f.Name.String()
	structs := []string{}

	for name, obj := range f.Scope.Objects {
		if obj.Kind == ast.Typ {
			ts, ok := obj.Decl.(*ast.TypeSpec)
			if !ok {
				return "", nil, fmt.Errorf("Unknown type without TypeSpec: %v", obj)
			}

			_, ok = ts.Type.(*ast.StructType)
			if !ok {
				// Not a struct, so skip it
				continue
			}

			structs = append(structs, name)
		}
	}

	return packageName, structs, nil
}
