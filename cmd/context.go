package cmd

import (
	"fmt"
	"os"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

func GetFunctions() error {
	fset := token.NewFileSet()

	node, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		return err
	}

	for _, pkg := range node {
		ast.Inspect(pkg, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true // continue traversal
			}

			ast.Inspect(fn, func(n ast.Node) bool {
				cl, ok := n.(*ast.CallExpr)
				if ok {
					fmt.Println()
					printer.Fprint(os.Stdout, fset, cl)
					fmt.Println()
				}
				return true
			})

			fmt.Println()
			printer.Fprint(os.Stdout, fset, fn)
			fmt.Println()
			return false // don't continue traversal of this AST node
		})
	}
	return nil
}

// func ListPackages(dir string) ([]string, error) {
// 	entries, err := os.ReadDir(dir)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read directory: %v", err)
// 	}

// 	packageNames := make(map[string]bool)

// 	for _, entry := range entries {
// 		if entry.IsDir() {
// 			continue
// 		}
// 		name := entry.Name()
// 		ext := filepath.Ext(name)
// 		if ext != ".go" {
// 			continue
// 		}

// 		filename := filepath.Join(dir, name)
// 		fset := token.NewFileSet()

// 		node, err := parser.ParseFile(fset, filename, nil, 0)
// 		if err != nil {
// 			log.Printf("failed to parse file %s: %v", filename, err)
// 			continue
// 		}

// 		for _, decl := range node.Decls {
// 			pkgDecl, ok := decl.(*ast.PackageDecl)
// 			if !ok {
// 				continue
// 			}
// 			packageName := pkgDecl.Name.String()
// 			packageNames[packageName] = true
// 			break
// 		}
// 	}

// 	var result []string
// 	for name := range packageNames {
// 		result = append(result, name)
// 	}

// 	return result, nil
// }
