package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func run(pth string) error {
	fset := token.NewFileSet()

	node, err := parser.ParseDir(fset, pth, nil, 0)
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

			// fmt.Printf("Package: %s, Function Name: %s\n", name, fn.Name)
			fmt.Println()
			printer.Fprint(os.Stdout, fset, fn)
			fmt.Println()
			return false // don't continue traversal of this AST node
		})
	}
	return nil
}

func main() {
	err := run(".")
	if err != nil {
		panic(err)
	}
}
