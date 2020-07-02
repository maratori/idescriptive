package idescriptive

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NewAnalyzer returns Analyzer that reports obscure interfaces.
func NewAnalyzer() *analysis.Analyzer {
	r := runner{
		strict: false,
	}
	fs := flag.NewFlagSet("", flag.PanicOnError)
	fs.BoolVar(&r.strict, "strict", r.strict, "Require all parameters to have names in interface declaration")

	return &analysis.Analyzer{
		Name:     "idescriptive",
		Doc:      "report obscure interfaces",
		Flags:    *fs,
		Run:      r.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

// runner is necessary to encapsulate flags with logic.
type runner struct {
	strict bool
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
				r.reportIssuesForMethod(pass, method.Names[0].Name, funcType)
			}
		}
	})

	return nil, nil
}

func (r *runner) reportIssuesForMethod(pass *analysis.Pass, methodName string, funcType *ast.FuncType) {
	var issues []issue

	switch r.strict {
	case true:
		issues = analyseMethodStrict(funcType)
	case false:
		issues = analyseMethod(funcType)
	}

	for _, i := range issues {
		pass.Reportf(i.Pos, "missing name for incoming parameter %s in method %s%s",
			types.ExprString(i.Type),
			methodName,
			strings.TrimPrefix(types.ExprString(funcType), "func"),
		)
	}
}

type issue struct {
	Pos  token.Pos
	Type ast.Expr
}

func analyseMethodStrict(funcType *ast.FuncType) []issue {
	if funcType.Params.NumFields() == 0 {
		return nil
	}

	for _, param := range funcType.Params.List {
		// no need to check other params because either all params have names either all not
		return issueIfDoNotHaveName(param)
	}

	return nil
}

func analyseMethod(funcType *ast.FuncType) []issue {
	if funcType.Params.NumFields() == 0 {
		return nil
	}

	issues := []issue{}

	for _, param := range funcType.Params.List {
		if typeIsSelfDescribing(param.Type) {
			continue
		}

		issues = append(issues, issueIfDoNotHaveName(param)...)
	}

	return issues
}

func issueIfDoNotHaveName(param *ast.Field) []issue {
	if len(param.Names) == 0 {
		return []issue{{Pos: param.Pos(), Type: param.Type}}
	}

	for _, name := range param.Names {
		if name == nil {
			return []issue{{Pos: param.Pos(), Type: param.Type}} // should never happen
		}

		if name.Name == "" {
			return []issue{{Pos: name.Pos(), Type: param.Type}} // should never happen
		}
	}

	return nil
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
