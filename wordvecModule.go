package pipeline

import (
	"github.com/opinionated/pipeline/analyzer/dbInterface"
	"github.com/sajari/word2vec"
)

// WordVecScore is flow, # connections
type WordVecScore struct {
	similarity float32
}

// Serialize the neo score
func (score WordVecScore) Serialize() []float32 {
	arr := make([]float32, 1)
	arr[0] = score.similarity

	return arr
}

// WordVecAnalyzer does a standard op on a neo connection
// Used as a "standard module analyzer
type WordVecAnalyzer struct {
	MetadataType string
	model        word2vec.Model
}

// Setup the connection to the neo db
func (na *WordVecAnalyzer) Setup() error {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		panic(err)
	}

	return nil
}

// Analyze the  relation between two articles with the score func
func (na WordVecAnalyzer) Analyze(main Article,
	related *Article) (bool, error) {

	return true, nil
}

var _ StandardModuleAnalyzer = (*WordVecAnalyzer)(nil)
