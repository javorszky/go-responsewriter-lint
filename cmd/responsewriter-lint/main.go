package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/javorszky/go-responsewriter-lint/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.New())
}
