package idescriptive

import (
	"flag"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NewAnalyzer returns Analyzer that reports interfaces without named arguments.
func NewAnalyzer() *analysis.Analyzer {
	runner := runner{
		allTypes: false,
	}
	fs := flag.NewFlagSet("", flag.PanicOnError)
	fs.BoolVar(&runner.allTypes, "all-types", runner.allTypes, "All parameters should be named regardless of type")

	return &analysis.Analyzer{
		Name:     "idescriptive",
		Doc:      "report interfaces without named arguments",
		Flags:    *fs,
		Run:      runner.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

// runner is necessary to encapsulate flags with logic.
type runner struct {
	allTypes bool
}

func (r *runner) run(pass *analysis.Pass) (interface{}, error) {
	traverse := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) // nolint:errcheck // let's panic
	filter := []ast.Node{
		(*ast.InterfaceType)(nil),
	}
	traverse.Preorder(filter, func(node ast.Node) {
		interfaceType := node.(*ast.InterfaceType) // nolint:errcheck // let's panic
		if interfaceType.Incomplete || interfaceType.Methods == nil {
			return
		}
		for _, method := range interfaceType.Methods.List {
			if funcType, ok := method.Type.(*ast.FuncType); ok {
				r.checkMethod(pass, funcType)
			}
		}
	})

	return nil, nil
}

func (r *runner) checkMethod(pass *analysis.Pass, funcType *ast.FuncType) {
	if funcType.Params == nil {
		return
	}

	for _, param := range funcType.Params.List {
		if !r.allTypes && !needToCheckType(param.Type) {
			continue
		}

		if len(param.Names) == 0 {
			pass.Reportf(param.Pos(), "missing incoming parameter name")
		}

		for _, name := range param.Names {
			if name.Name == "" {
				pass.Reportf(name.Pos(), "missing incoming parameter name")
			}
		}
	}
}

// nolint:gocyclo // it's ok to have huge switch
func needToCheckType(paramType ast.Expr) bool {
	switch t := paramType.(type) {
	case *ast.Ident:
		return needToCheckBuiltin(t.Name)
	case *ast.StarExpr:
		return needToCheckType(t.X)
	case *ast.Ellipsis:
		return needToCheckType(t.Elt)
	case *ast.ArrayType:
		return needToCheckType(t.Elt)
	case *ast.MapType:
		return needToCheckType(t.Key) || needToCheckType(t.Value)
	case *ast.ParenExpr:
		return needToCheckType(t.X)
	case *ast.ChanType:
		switch t.Dir {
		case ast.RECV: // <-chan
			return true
		case ast.SEND: // chan<-
			return needToCheckType(t.Value)
		case ast.RECV | ast.SEND:
			return true
		default:
			panic(fmt.Sprintf("unknown chan direction %#v", t.Dir))
		}
	case *ast.FuncType:
		return false
	case *ast.StructType:
		return true
	case *ast.InterfaceType:
		return t.Methods == nil || len(t.Methods.List) == 0 // empty interface
	case *ast.SelectorExpr:
		return false
	default:
		panic(fmt.Sprintf("unknown type %#v", t))
	}
}

func needToCheckBuiltin(name string) bool {
	for _, n := range []string{
		"bool",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"int8",
		"int16",
		"int32",
		"int64",
		"float32",
		"float64",
		"complex64",
		"complex128",
		"string",
		"int",
		"uint",
		"uintptr",
		"byte",
		"rune",
	} {
		if name == n {
			return true
		}
	}

	return false
}
