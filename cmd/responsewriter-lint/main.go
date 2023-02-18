package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/javorszky/responsewriter-linter/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.New())
}
