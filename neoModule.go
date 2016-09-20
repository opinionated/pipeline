package pipeline

import (
	"github.com/opinionated/pipeline/analyzer/dbInterface"
)

// NeoScore is flow, # connections
type NeoScore struct {
	Flow  float32
	Count int
}

// Serialize the neo score
func (score NeoScore) Serialize() []float32 {
	arr := make([]float32, 2)
	arr[0] = score.Flow
	arr[1] = float32(score.Count)
	return arr
}

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
func (na NeoAnalyzer) Analyze(main Article,
	related *Article) (bool, error) {

	flow, count, err := relationDB.StrengthBetween(
		main.Name(),
		related.Name(),
		na.MetadataType)

	if err != nil {
		return false, err
	}

	neoScore := NeoScore{flow, count}
	err = related.AddScore("neo_"+na.MetadataType, neoScore)

	return err == nil, err
}

var _ StandardModuleAnalyzer = (*NeoAnalyzer)(nil)
