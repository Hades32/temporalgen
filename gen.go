package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	typeNameP   = flag.String("type", "", "name of type to generate stubs for")
	execSuffix  = flag.String("execSuffix", "Exec", "suffix of the generated 'Exec' methods")
	startSuffix = flag.String("startSuffix", "Start", "suffix of the generated 'Start' methods")
	dryRun      = flag.Bool("dry", false, "just print to stdout")
)

func main() {
	flag.Parse()
	typeName := *typeNameP
	starTypeName := "*" + typeName
	if typeName == "" {
		flag.Usage()
		fmt.Println("Expected to be used via 'go generate'. Place a comment like this in your code")
		fmt.Println("//go:generate go run github.com/Hades32/temporalgen -type ActivitiesStruct")
		os.Exit(1)
	}
	run(typeName, starTypeName)
}

func run(typeName string, starTypeName string) {
	pkg, err := packages.Load(&packages.Config{
		Mode: packages.NeedModule | packages.NeedName | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo, // | packages.NeedDeps,
	})
	if err != nil {
		panic(err)
	}
	out := os.Stdout
	if !*dryRun {
		out, err = os.OpenFile(strings.ToLower(typeName)+".gen.go", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o660)
		if err != nil {
			panic(err)
		}
		defer out.Sync()
		defer out.Close()
	}
	for _, p := range pkg {
		fmt.Fprintf(out, "// Code generated by \"temporal-gen -type=%s\"; DO NOT EDIT.\n\n", typeName)
		fmt.Fprintf(out, "package %s\n\n", p.Name)
		fmt.Fprintln(out, "import (")
		imports := append(usedImports(p, starTypeName), "go.temporal.io/sdk/workflow")
		sort.Strings(imports)
		for _, imp := range imports {
			fmt.Fprint(out, "\t\"", imp, "\"\n")
		}
		fmt.Fprintln(out, ")")
		fmt.Fprintln(out)
		fmt.Fprintf(out, "type %sStub struct {\n\ta *%s\n}\n", typeName, typeName)
		printStubs(p, starTypeName, out, typeName)
	}
}

func usedImports(p *packages.Package, starTypeName string) (imports []string) {
	pkgPathSet := map[string]bool{}
	// name -> pkgPath
	pkgLookup := map[string]string{}
	for _, p := range p.Imports {
		pkgLookup[p.Name] = p.PkgPath
	}
	for _, f := range p.Syntax {
		for _, decl := range f.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil {
				continue
			}
			if getTypeName(funcDecl.Recv.List[0].Type) != starTypeName {
				continue
			}
			firstParam := true
			for _, field := range funcDecl.Type.Params.List {
				if firstParam {
					// the first param _must_ be context.Context, but that's implicit in Temporal's API
					firstParam = false
					continue
				}
				for _, t := range getTypePkg(pkgLookup, field.Type) {
					pkgPathSet[t] = true
				}
			}
			for _, field := range funcDecl.Type.Results.List {
				for _, t := range getTypePkg(pkgLookup, field.Type) {
					pkgPathSet[t] = true
				}
			}
		}
	}
	for pkgPath := range pkgPathSet {
		if pkgPath == "" {
			continue
		}
		imports = append(imports, pkgPath)
	}
	return imports
}

func printStubs(p *packages.Package, starTypeName string, out *os.File, typeName string) {
	for _, f := range p.Syntax {
		for _, decl := range f.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil {
				continue
			}
			if getTypeName(funcDecl.Recv.List[0].Type) != starTypeName {
				continue
			}
			fmt.Fprintln(out)
			fmt.Fprintf(out, "func (s *%sStub) %s%s(ctx workflow.Context", typeName, funcDecl.Name, *execSuffix)
			if len(funcDecl.Type.Params.List) > 0 {
				fmt.Fprint(out, ", ")
				printParams(out, funcDecl.Type.Params, true, false)
			}
			fmt.Fprint(out, ") ")
			if len(funcDecl.Type.Results.List) > 1 || len(funcDecl.Type.Results.List[0].Names) > 0 {
				fmt.Fprint(out, "(")
			}
			printParams(out, funcDecl.Type.Results, false, false)
			if len(funcDecl.Type.Results.List) > 1 || len(funcDecl.Type.Results.List[0].Names) > 0 {
				fmt.Fprint(out, ")")
			}
			fmt.Fprint(out, " {\n")
			fmt.Fprintf(out, "\tf := workflow.ExecuteActivity(ctx, s.a.%s", funcDecl.Name)
			if len(funcDecl.Type.Params.List) > 0 {
				fmt.Fprint(out, ", ")
				printParams(out, funcDecl.Type.Params, true, true)
			}
			fmt.Fprint(out, ")\n")
			if len(funcDecl.Type.Results.List) > 1 {
				fmt.Fprintf(out, "\tvar _res %s\n", getTypeName(funcDecl.Type.Results.List[0].Type))
				fmt.Fprint(out, "\treturn _res, f.Get(ctx, &_res)\n")
			} else {
				fmt.Fprint(out, "\treturn f.Get(ctx, nil)\n")
			}
			fmt.Fprint(out, "}\n\n")

			fmt.Fprintf(out, "func (s *%sStub) %s%s(ctx workflow.Context", typeName, funcDecl.Name, *startSuffix)
			if len(funcDecl.Type.Params.List) > 0 {
				fmt.Fprint(out, ", ")
				printParams(out, funcDecl.Type.Params, true, false)
			}
			fmt.Fprint(out, ") workflow.Future {\n")
			fmt.Fprintf(out, "\tf := workflow.ExecuteActivity(ctx, s.a.%s", funcDecl.Name)
			if len(funcDecl.Type.Params.List) > 0 {
				fmt.Fprint(out, ", ")
				printParams(out, funcDecl.Type.Params, true, true)
			}
			fmt.Fprint(out, ")\n")
			fmt.Fprint(out, "\treturn f\n")
			fmt.Fprint(out, "}\n")
		}
	}
}

func getTypeName(t ast.Expr) (typeName string) {
	if ident, ok := t.(*ast.Ident); ok {
		typeName = ident.Name
	} else if typ, ok := t.(*ast.StarExpr); ok {
		typeName = "*" + getTypeName(typ.X)
	} else if sel, ok := t.(*ast.SelectorExpr); ok {
		typeName = getTypeName(sel.X) + "." + sel.Sel.Name
	} else if mp, ok := t.(*ast.MapType); ok {
		return fmt.Sprintf("map[%s]%s", getTypeName(mp.Key), getTypeName(mp.Value))
	} else if arr, ok := t.(*ast.ArrayType); ok {
		typeName = "[]" + getTypeName(arr.Elt)
	} else {
		fmt.Printf("what is this %t", t)
	}
	return typeName
}

func getTypePkg(imports map[string]string, t ast.Expr) (pgkNames []string) {
	if _, ok := t.(*ast.Ident); ok {
		return []string{""}
	} else if typ, ok := t.(*ast.StarExpr); ok {
		return getTypePkg(imports, typ.X)
	} else if sel, ok := t.(*ast.SelectorExpr); ok {
		return []string{imports[getTypeName(sel.X)]}
	} else if mp, ok := t.(*ast.MapType); ok {
		return []string{imports[getTypeName(mp.Key)], imports[getTypeName(mp.Value)]}
	} else if arr, ok := t.(*ast.ArrayType); ok {
		return []string{imports[getTypeName(arr.Elt)]}
	} else {
		panic(fmt.Sprintf("unexpected node when searching type package: %s", t))
	}
}

func printParams(out *os.File, params *ast.FieldList, skipFirst bool, onlyNames bool) {
	skip := skipFirst
	hasPrevious := false
	for _, field := range params.List {
		if skip {
			skip = false
			continue
		}
		var names []string
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
		if hasPrevious {
			fmt.Fprint(out, ", ")
		} else {
			hasPrevious = true
		}
		if len(names) > 0 {
			fmt.Fprint(out, strings.Join(names, ", "))
			if onlyNames {
				continue
			}
			fmt.Fprint(out, " ")
		}
		fmt.Fprint(out, getTypeName(field.Type))
	}
}
