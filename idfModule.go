package pipeline

import (
	"github.com/opinionated/pipeline/analyzer/dbInterface"
)

// IDFScore is flow, # connections
type IDFScore struct {
	Counts []int
}

// Serialize the neo score
func (score IDFScore) Serialize() []float32 {
	length := len(score.Counts)
	arr := make([]float32, length)
	for i := range score.Counts {
		arr[i] = float32(score.Counts[i])
	}

	return arr
}

// IDFAnalyzer does a standard op on a neo connection
// Used as a "standard module analyzer
type IDFAnalyzer struct {
	MetadataType string
}

// Setup the connection to the neo db
func (na IDFAnalyzer) Setup() error {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		panic(err)
	}
	return nil
}

func (na IDFAnalyzer) tryInv(main Article, related *Article) {
	strs, err := relationDB.GetRelations(
		main.Name(),
		na.MetadataType,
		0.0)
	if err != nil {
		panic("bad bad")
	}

	for _ = range strs {
		panic("good idf")
	}
}

// Analyze the  relation between two articles with the score func
func (na IDFAnalyzer) Analyze(main Article,
	related *Article) (bool, error) {

	counts, err := relationDB.GetFauxIDF(
		main.Name(),
		related.Name(),
		na.MetadataType)

	if err != nil {
		return false, err
	}

	neoScore := IDFScore{Counts: counts}
	err = related.AddScore("idf_"+na.MetadataType, neoScore)

	return err == nil, err
}

var _ StandardModuleAnalyzer = (*IDFAnalyzer)(nil)
