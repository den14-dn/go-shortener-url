package osexitcheckanalyzer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"go-shortener-url/pkg/osexitcheckanalyzer"
)

func TestMyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), osexitcheckanalyzer.OsExitCheckAnalyzer, "./pkg1")
}
