package pipeline

import (
	"bufio"
	"fmt"
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/pipeline"
	"os"
	"testing"
)

type CustomAnalyzer interface {
	Analyze(analyzer.Analyzable, chan analyzer.Analyzable) chan analyzer.Analyzable
}

type Stage struct {
	modID     string                   //ID of current module (for log purposes)
	AOI       analyzer.Analyzable      //File name of Article Of Interest (AOI) in system
	relatives chan analyzer.Analyzable //File names (FOR NOW, LATER DB IDS PLS) of relatives to AOI in system
	analyzer  CustomAnalyzer           //CustomAnalyzer object to run

}
