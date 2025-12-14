// resetgen-analyzer checks that Reset() is called before sync.Pool.Put()
package main

import (
	"github.com/flaticols/resetgen/cmd/resetgen-analyzer/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
