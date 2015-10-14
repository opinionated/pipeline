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
	Analyze(string, string) bool
}

type PipelineModule struct {
	modID     string         //ID of current module (for log purposes)
	AOI       string         //File name of Article Of Interest (AOI) in system
	relatives chan string    //File names (FOR NOW, LATER DB IDS PLS) of relatives to AOI in system
	analyzer  CustomAnalyzer //CustomAnalyzer object to run

	//Create with database link?/xml files link? !!!!LINK TO PREPROCESSED DATA!!!!
}
