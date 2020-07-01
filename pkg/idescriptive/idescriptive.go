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
	r := runner{
		allTypesShouldHaveName: false,
	}
	fs := flag.NewFlagSet("", flag.PanicOnError)
	fs.BoolVar(
		&r.allTypesShouldHaveName,
		"all-types",
		r.allTypesShouldHaveName,
		"All parameters should be named regardless of type",
	)

	return &analysis.Analyzer{
		Name:     "idescriptive",
		Doc:      "report interfaces without named arguments",
		Flags:    *fs,
		Run:      r.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

// runner is necessary to encapsulate flags with logic.
type runner struct {
	allTypesShouldHaveName bool
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
			// *ast.Ident - embedded interface from current package
			// *ast.SelectorExpr - embedded interface from different package
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
		if !r.allTypesShouldHaveName && typeIsSelfDescribing(param.Type) {
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
func typeIsSelfDescribing(paramType ast.Expr) bool {
	switch t := paramType.(type) {
	case *ast.Ident:
		return !builtinTypeShouldHaveName(t.Name)
	case *ast.StarExpr:
		return typeIsSelfDescribing(t.X)
	case *ast.Ellipsis:
		return typeIsSelfDescribing(t.Elt)
	case *ast.ArrayType:
		return typeIsSelfDescribing(t.Elt)
	case *ast.MapType:
		return typeIsSelfDescribing(t.Key) && typeIsSelfDescribing(t.Value)
	case *ast.ParenExpr:
		return typeIsSelfDescribing(t.X)
	case *ast.ChanType:
		switch t.Dir {
		case ast.RECV: // <-chan
			return false
		case ast.SEND: // chan<-
			return typeIsSelfDescribing(t.Value)
		case ast.RECV | ast.SEND:
			return false
		default:
			panic(fmt.Sprintf("unknown chan direction %#v", t.Dir))
		}
	case *ast.FuncType:
		return true
	case *ast.StructType:
		return false
	case *ast.InterfaceType:
		return t.Methods != nil && len(t.Methods.List) > 0 // non-empty interface
	case *ast.SelectorExpr:
		return true
	default:
		panic(fmt.Sprintf("unknown type %#v", t))
	}
}

func builtinTypeShouldHaveName(name string) bool {
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
