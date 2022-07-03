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

	return nil, nil // nolint:nilnil // linter doesn't return a result
}

func (r *runner) reportIssuesForMethod(pass *analysis.Pass, methodName string, funcType *ast.FuncType) {
	var issues []issue

	switch r.strict {
	case true:
		issues = analyseMethodStrict(funcType)
	case false:
		issues = analyseMethod(pass.TypesInfo, funcType)
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

func analyseMethod(info *types.Info, funcType *ast.FuncType) []issue {
	if funcType.Params.NumFields() <= 1 {
		// Mostly single parameter of a method is evident
		return nil
	}

	type Seen struct {
		first []issue
		count int
	}

	issues := []issue{}
	selfDescribingTypes := map[string]Seen{}

	for _, param := range funcType.Params.List {
		if typeIsSelfDescribing(info, param.Type) {
			tStr := types.ExprString(param.Type)
			seen := selfDescribingTypes[tStr]

			switch seen.count {
			case 0:
				seen.first = issueIfDoNotHaveName(param)
			case 1:
				issues = append(issues, seen.first...)
			}

			selfDescribingTypes[tStr] = Seen{
				first: seen.first,
				count: seen.count + 1,
			}

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

func typeIsSelfDescribing(info *types.Info, paramType ast.Expr) bool {
	switch t := paramType.(type) {
	case *ast.Ident:
		switch o := info.ObjectOf(t).(type) {
		case *types.TypeName:
			switch o.Type().(type) {
			case *types.Named:
				return true
			default:
				return o.IsAlias()
			}
		case *types.Builtin, *types.Const, *types.Func, *types.Label, *types.Nil, *types.PkgName, *types.Var:
			panic(fmt.Sprintf("unexpected object %T %#v", o, o)) // should never happen
		default:
			panic(fmt.Sprintf("unknown object %T %#v", o, o))
		}
	case *ast.StarExpr:
		return typeIsSelfDescribing(info, t.X)
	case *ast.Ellipsis:
		return typeIsSelfDescribing(info, t.Elt)
	case *ast.ArrayType:
		return typeIsSelfDescribing(info, t.Elt)
	case *ast.MapType:
		return typeIsSelfDescribing(info, t.Key) && typeIsSelfDescribing(info, t.Value)
	case *ast.ParenExpr:
		return typeIsSelfDescribing(info, t.X)
	case *ast.ChanType:
		switch t.Dir {
		case ast.RECV: // <-chan
			return false
		case ast.SEND: // chan<-
			return typeIsSelfDescribing(info, t.Value)
		case ast.RECV | ast.SEND:
			return false
		default:
			panic(fmt.Sprintf("unknown chan direction %#v", t.Dir))
		}
	case *ast.StructType:
		// Empty struct{} is ok because it's just a flag (used in channels and maps).
		// Non-empty struct is ok because it's self-describing.
		return true
	case *ast.FuncType:
		return true // may be false?
	case *ast.SelectorExpr: // http.Response
		// It has a name that describes what is it.
		return true // should unsafe.Pointer be exception?
	case *ast.InterfaceType:
		// If interface has method(s) it's self-describing.
		return t.Methods.NumFields() > 0 // non-empty interface
	case *ast.BadExpr, *ast.BasicLit, *ast.BinaryExpr, *ast.CallExpr, *ast.CompositeLit, *ast.FuncLit,
		*ast.IndexExpr, *ast.KeyValueExpr, *ast.SliceExpr, *ast.TypeAssertExpr, *ast.UnaryExpr:
		panic(fmt.Sprintf("unexpected type %T %#v", t, t)) // should never happen
	default:
		panic(fmt.Sprintf("unknown type %T %#v", t, t))
	}
}
