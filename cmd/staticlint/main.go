/*
Static analyzer multichecker

Multichecher contains several code analyzers from external packages and its own implementation.

Usage:

	go vet -vettool=[path ...]

Multichecker consists of:
  - golang.org/x/tools/go/analysis/passes/assign - checks for assignments to struct fields inside methods;
  - golang.org/x/tools/go/analysis/passes/atomic - checks for common mistakes using the sync/atomic package;
  - golang.org/x/tools/go/analysis/passes/bools - detects common mistakes involving boolean operators;
  - golang.org/x/tools/go/analysis/passes/copylock - checks for locks erroneously passed by value;
  - golang.org/x/tools/go/analysis/passes/errorsas - checks that the second argument to errors;
  - golang.org/x/tools/go/analysis/passes/fieldalignment - detects structs that would use less memory if their fields were sorted;
  - golang.org/x/tools/go/analysis/passes/httpresponse - checks for mistakes using HTTP responses;
  - golang.org/x/tools/go/analysis/passes/loopclosure - checks for references to enclosing loop variables from within nested functions;
  - golang.org/x/tools/go/analysis/passes/lostcancel - checks for failure to call a context cancellation function;
  - golang.org/x/tools/go/analysis/passes/nilfunc - checks for useless comparisons against nil;
  - golang.org/x/tools/go/analysis/passes/printf - checks consistency of Printf format strings and arguments;
  - golang.org/x/tools/go/analysis/passes/shadow - checks for shadowed variables;
  - golang.org/x/tools/go/analysis/passes/shift - checks for shifts that exceed the width of an integer;
  - golang.org/x/tools/go/analysis/passes/sigchanyzer - detects misuse of unbuffered signal as argument to signal.Notify;
  - golang.org/x/tools/go/analysis/passes/sortslice - checks for calls to sort.Slice that do not use a slice type as first argument;
  - golang.org/x/tools/go/analysis/passes/stdmethods - checks for misspellings in the signatures of methods similar to well-known interfaces;
  - golang.org/x/tools/go/analysis/passes/stringintconv - flags type conversions from integers to strings;
  - golang.org/x/tools/go/analysis/passes/structtag - checks struct field tags are well formed;
  - golang.org/x/tools/go/analysis/passes/tests - checks for common mistaken usages of tests and examples;
  - golang.org/x/tools/go/analysis/passes/unmarshal - checks for passing non-pointer or non-interface types to unmarshal and decode functions;
  - golang.org/x/tools/go/analysis/passes/unreachable - checks for unreachable code;
  - golang.org/x/tools/go/analysis/passes/unsafeptr - checks for invalid conversions of uintptr to unsafe.Pointer;
  - golang.org/x/tools/go/analysis/passes/unusedresult - checks for unused results of calls to certain pure functions;
  - golang.org/x/tools/go/analysis/passes/unusedwrite - checks for unused writes to the elements of a struct or array object;
  - honnef.co/go/tools/simple - contains analyzes that simplify code;
  - honnef.co/go/tools/staticcheck - contains analyzes that find bugs and performance issues;
  - honnef.co/go/tools/stylecheck - contains analyzes that enforce style rules;
  - go-shortener-url/pkg/osexitcheckanalyzer - prevents the use of a direct call to os.Exit in function main of package main.
*/
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"go-shortener-url/pkg/osexitcheckanalyzer"
)

func main() {
	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		osexitcheckanalyzer.OsExitCheckAnalyzer,
	}

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range stylecheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	multichecker.Main(
		checks...,
	)
}
