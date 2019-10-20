package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/maratori/idescriptive/pkg/idescriptive"
)

func main() {
	singlechecker.Main(idescriptive.NewAnalyzer())
}
