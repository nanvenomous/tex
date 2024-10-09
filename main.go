package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()

	node, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for name, pkg := range node {
		ast.Inspect(pkg, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true // continue traversal
			}
			fmt.Printf("Package: %s, Function Name: %s\n", name, fn.Name)
			return false // don't continue traversal of this AST node
		})
	}
}
