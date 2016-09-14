package pipeline

import (
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/analyzer/dbInterface"
)

// NeoAnalyzer does a standard op on a neo connection
// Used as a "standard module analyzer
type NeoAnalyzer struct {
	MetadataType string
	ScoreFunc    func(flow float32, count int) float64
}

// Setup the connection to the neo db
func (na NeoAnalyzer) Setup() error {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		panic(err)
	}
	return nil
}

// Analyze the  relation between two articles with the score func
func (na NeoAnalyzer) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	flow, count, err := relationDB.StrengthBetween(
		main.FileName,
		related.FileName,
		na.MetadataType)

	if err != nil {
		return false, err
	}

	related.Score += na.ScoreFunc(flow, count)
	return true, nil
}

var _ StandardModuleAnalyzer = (*NeoAnalyzer)(nil)
