package utils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// AddAppToInternalMain uses AST parsing to add a new app to internal/main.go
func AddAppToInternalMain(path, projectName, appName string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			newImport := &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`"%s/internal/%s"`, projectName, appName),
				},
			}
			gd.Specs = append(gd.Specs, newImport)
			return false
		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			if kv, ok := cl.Type.(*ast.MapType); ok {
				if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "string" {
					newAppEntry := &ast.KeyValueExpr{
						Key: &ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf(`"%s"`, appName),
						},
						Value: &ast.CompositeLit{
							Type: &ast.SelectorExpr{
								X:   ast.NewIdent(appName),
								Sel: ast.NewIdent("App"),
							},
						},
					}
					cl.Elts = append(cl.Elts, newAppEntry)
					return false
				}
			}
		}
		return true
	})

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// AddModuleToAppMain uses AST parsing to add a new module to an app's main file.
func AddModuleToAppMain(path, projectName, appName, moduleName string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			newImport := &ast.ImportSpec{
				Name: ast.NewIdent(moduleName),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`"%s/internal/%s/%s"`, projectName, appName, moduleName),
				},
			}
			gd.Specs = append(gd.Specs, newImport)
			return false
		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if se, ok := ce.Fun.(*ast.SelectorExpr); ok {
				if x, ok := se.X.(*ast.Ident); ok && x.Name == "core" && se.Sel.Name == "New" {
					newModuleEntry := &ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(moduleName),
							Sel: ast.NewIdent(strings.Title(moduleName) + "Module"),
						},
					}
					ce.Args = append(ce.Args, newModuleEntry)
					return false
				}
			}
		}
		return true
	})

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}
