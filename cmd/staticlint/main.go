// Package main implements a custom static analyzer multichecker for the URL shortener service.
//
// The multichecker combines multiple static analyzers into a single tool:
//
//  1. Standard Go analyzers from golang.org/x/tools/go/analysis/passes
//     - asmdecl: reports mismatches between assembly files and Go declarations
//     - assign: detects useless assignments
//     - atomic: checks for common mistakes using sync/atomic
//     - and others...
//
//  2. All SA category analyzers from staticcheck.io
//     These analyze common coding mistakes and issues
//
//  3. Selected analyzers from other staticcheck.io categories:
//     - ST1000: checks for missing package documentation
//     - ST1003: checks naming conventions
//
//  4. Additional public analyzers:
//     - errcheck: checks for unchecked errors
//     - bodyclose: ensures HTTP response bodies are closed
//
//  5. Custom analyzer:
//     - exitcheck: prohibits direct os.Exit calls in main function
//
// Usage:
//
//	Build:
//	    cd cmd/staticlint
//	    go build
//
//	Run:
//	    ./staticlint              # checks all packages above current directory
//	    ./staticlint ./...        # checks current directory and subdirectories
//	    ./staticlint path/to/pkg  # checks specific package
//
// The exitcheck analyzer helps ensure proper error handling in main packages
// by prohibiting direct calls to os.Exit, encouraging the use of proper
// error handling patterns instead.
package main

import (
	"os"
	"path/filepath"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"

	"github.com/Eorthus/shorturl/cmd/staticlint/custom_analyzers"
)

func main() {
	// Если аргументы не переданы, проверяем все пакеты выше текущей директории
	if len(os.Args) == 1 {
		// filepath.Join корректно соединяет части пути для любой ОС
		path := filepath.Join("..", "...")
		os.Args = append(os.Args, path)
	}

	// Стандартные анализаторы
	standardAnalyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	// Анализаторы SA класса из staticcheck
	saAnalyzers := make([]*analysis.Analyzer, 0)
	for _, a := range staticcheck.Analyzers {
		if a.Analyzer.Name[:2] == "SA" {
			saAnalyzers = append(saAnalyzers, a.Analyzer)
		}
	}

	// Дополнительные анализаторы из других классов staticcheck
	extraAnalyzers := []*analysis.Analyzer{}
	for _, a := range staticcheck.Analyzers {
		if a.Analyzer.Name == "ST1000" || a.Analyzer.Name == "ST1003" {
			extraAnalyzers = append(extraAnalyzers, a.Analyzer)
		}
	}

	// Публичные анализаторы
	publicAnalyzers := []*analysis.Analyzer{
		errcheck.Analyzer,  // Проверяет обработку ошибок
		bodyclose.Analyzer, // Проверяет закрытие тел HTTP-ответов
	}

	// Собственный анализатор
	customAnalyzers := []*analysis.Analyzer{
		custom_analyzers.ExitCheckAnalyzer,
	}

	// Объединяем все анализаторы
	allAnalyzers := append(standardAnalyzers, saAnalyzers...)
	allAnalyzers = append(allAnalyzers, extraAnalyzers...)
	allAnalyzers = append(allAnalyzers, publicAnalyzers...)
	allAnalyzers = append(allAnalyzers, customAnalyzers...)

	multichecker.Main(
		allAnalyzers...,
	)
}
