// Staticlint multichecker — набор статических анализаторов для проверки качества и корректности Go-кода проекта metrics-collector.
//
// Запуск:
//
//	make staticlint
//
// Multichecker запускает несколько групп анализаторов:
//
//  1. Стандартные анализаторы пакета golang.org/x/tools/go/analysis/passes.
//     Они выявляют типовые ошибки, не обнаруживаемые компилятором.
//
//     - printf
//     Проверяет соответствие форматных строк и аргументов в функциях
//     форматированного вывода (fmt.Printf, log.Printf и др.).
//
//     - shadow
//     Обнаруживает затенение переменных во вложенных областях видимости,
//     что часто приводит к логическим ошибкам.
//
//     - structtag
//     Проверяет корректность struct-тегов (json, db, yaml и т.д.).
//
//  2. Дополнительные публичные анализаторы.
//
//     - copylock
//     Запрещает копирование структур, содержащих примитивы синхронизации
//     (sync.Mutex, sync.Once и др.), предотвращая race condition.
//
//     - lostcancel
//     Обнаруживает утечки контекста при отсутствии вызова cancel-функции,
//     возвращаемой context.WithCancel / context.WithTimeout.
//
//  3. Анализаторы staticcheck класса SA (security / correctness).
//
//     Все SA-анализаторы подключаются автоматически и направлены на
//     обнаружение реальных ошибок выполнения: nil dereference, неверной
//     синхронизации, бесполезного кода и других критических проблем.
//
//  4. Анализатор другого класса staticcheck.
//
//     - S1000 (simple)
//     Обнаруживает select с единственным case и рекомендует
//     заменить его на прямую операцию чтения или записи в канал.
//
//  5. Собственный анализатор noosexit.
//
//     Запрещает использование прямого вызова os.Exit в функции main
//     исполняемых пакетов проекта. Анализатор предотвращает пропуск defer,
//     улучшает тестируемость и обеспечивает корректное завершение приложения.
//
// Multichecker завершает выполнение с ненулевым кодом возврата,
// если хотя бы один анализатор обнаружил нарушение.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/simple/s1000"

	// стандартные анализаторы
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"

	// дополнительные анализаторы
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/lostcancel"

	"honnef.co/go/tools/staticcheck"

	// собственный анализатор
	"github.com/Nekrasov-Sergey/metrics-collector/cmd/staticlint/noosexit"
)

func main() {
	analyzers := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers)+7)

	// SA-анализаторы staticcheck
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	analyzers = append(analyzers,
		// стандартные анализаторы
		printf.Analyzer,    // проверка форматов printf
		shadow.Analyzer,    // обнаружение затенения переменных
		structtag.Analyzer, // проверка корректности struct-тегов

		// дополнительные анализаторы
		copylock.Analyzer,   // защита от копирования структур с mutex
		lostcancel.Analyzer, // обнаружение потерянного cancel у context

		// собственный анализатор
		noosexit.Analyzer,

		// один анализатор другого класса staticcheck (simple)
		s1000.SCAnalyzer.Analyzer,
	)

	multichecker.Main(analyzers...)
}
