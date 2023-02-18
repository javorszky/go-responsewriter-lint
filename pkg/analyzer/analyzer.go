package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "responsewriterlint",
		Doc:      "Checks the order of method calls on http.ResponseWriter to flag potential bugs from out of order calls",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	// pass.ResultOf[inspect.Analyzer] will be set if we've added inspect.Analyzer to Requires.
	inspectorInstance := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{ // filter needed nodes: visit only them
		(*ast.FuncDecl)(nil),
	}

	// I need to look into all the FuncDecl, grab the param list, take note of the name of the response writer in the
	// params, look into the body, and then check inside for more info.
	inspectorInstance.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		// Take note of the function name for reporting purposes.
		rwVarNames := make([]string, 0)
		funcName := funcDecl.Name.Name
		lr := NewLinterResult(funcName)

		// Look at the incoming parameters to the function, and take note of the var names that are of type
		// http.ResponseWriter, and save them in the rwVarNames slice.
		for _, f := range funcDecl.Type.Params.List {
			paramType, ok := f.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			xIdent, ok := paramType.X.(*ast.Ident)
			if !ok {
				continue
			}

			if xIdent.Name != "http" {
				continue
			}

			if paramType.Sel.Name != "ResponseWriter" {
				continue
			}

			for _, names := range f.Names {
				rwVarNames = append(rwVarNames, names.Name)
			}
		}

		// there were no variables that are http.responsewriter
		if len(rwVarNames) == 0 {
			return
		}

		var recursiveF func(ast.Expr)
		recursiveF = func(x ast.Expr) {
			callX, ok := x.(*ast.CallExpr)
			if ok {
				fX, ok := callX.Fun.(*ast.SelectorExpr)
				if ok {
					_, ok := fX.X.(*ast.CallExpr)
					if ok {
						recursiveF(fX.X)
						return
					}

					x, ok := fX.X.(*ast.Ident)
					if ok {
						isVarOnHttpWriter := false
						for _, v := range rwVarNames {
							if x.Name == v {
								isVarOnHttpWriter = true
								break
							}
						}

						if !isVarOnHttpWriter {
							return
						}

						switch fX.Sel.Name {
						case FuncHeader:
							p := pass.Fset.Position(fX.Sel.NamePos)
							lr.Record(FuncHeader, p.Line, fX.Sel.Pos())
						case FuncWrite:
							p := pass.Fset.Position(fX.Sel.NamePos)
							lr.Record(FuncWrite, p.Line, fX.Sel.Pos())
						case FuncWriteHeader:
							p := pass.Fset.Position(fX.Sel.NamePos)
							lr.Record(FuncWriteHeader, p.Line, fX.Sel.Pos())
						default:
						}
					}
				}
			}
		}

		for _, listItem := range funcDecl.Body.List {
			expression, ok := listItem.(*ast.ExprStmt)
			if !ok {
				continue
			}

			recursiveF(expression.X)
		}

		for _, e := range lr.Errors() {
			pass.Reportf(e.pos, "function %s: %s", funcName, e.message)
		}
	})

	return nil, nil
}
