package noosexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const doc = `
Package noosexit реализует статический анализатор, запрещающий
использование прямого вызова os.Exit в функции main пакета main.

Анализатор применяется только к исполняемым пакетам:
  - github.com/Nekrasov-Sergey/metrics-collector/cmd/server
  - github.com/Nekrasov-Sergey/metrics-collector/cmd/agent

Использование os.Exit в main считается нежелательным по следующим причинам:
  - defer-функции не выполняются
  - нарушается корректное завершение приложения
  - усложняется модульное и интеграционное тестирование

Рекомендуемый подход:
  - основная логика приложения должна быть вынесена в отдельную функцию
    (например, run)
  - функция main должна вызывать эту логику и обрабатывать возвращаемую ошибку
    централизованно, без использования os.Exit.
`

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  doc,
	Run:  run,
}

var allowedMainPackages = map[string]struct{}{
	"github.com/Nekrasov-Sergey/metrics-collector/cmd/server": {},
	"github.com/Nekrasov-Sergey/metrics-collector/cmd/agent":  {},
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	if _, ok := allowedMainPackages[pass.Pkg.Path()]; !ok {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				obj := pass.TypesInfo.Uses[sel.Sel]
				fnObj, ok := obj.(*types.Func)
				if !ok || fnObj.Pkg() == nil {
					return true
				}

				if fnObj.Pkg().Path() == "os" && fnObj.Name() == "Exit" {
					pass.Reportf(call.Pos(), "запрещено использовать os.Exit в функции main")
				}

				return true
			})
		}
	}

	return nil, nil
}
