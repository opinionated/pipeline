package pipeline

import (
	"fmt"
	"github.com/opinionated/pipeline/analyzer/dbInterface"
	"github.com/sajari/word2vec"
	"strings"
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
	client       word2vec.Client
}

// Setup the connection to the neo db
func (na *WordVecAnalyzer) Setup() error {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		panic(err)
	}

	na.client = word2vec.Client{Addr: "localhost:1234"}

	return nil
}

// replaces spaces in a string with underscores
func fixString(str string) string {
	return strings.Replace(str, " ", "_", -1)
}

func (na WordVecAnalyzer) getFixedStrs(article Article) ([]string, error) {
	strs, err := relationDB.GetRelations(
		article.Name(),
		na.MetadataType,
		0.0)
	if err != nil {
		return nil, err
	}

	return strs, err
}

func (na WordVecAnalyzer) getScore(main []string, related []string) (float32, error) {

	var totalScore float32
	totalScore = 0.0
	for i := range main {
		for j := range related {

			if main[i] == related[j] {
				continue
			}

			a := word2vec.Expr{}
			a.Add(1, main[i])

			b := word2vec.Expr{}
			b.Add(1, related[j])

			score, err := na.client.Cos(a, b)
			if err != nil {
				continue
			}
			if score > 0.3 {
				fmt.Println("similarity:", main[i], related[j])
				totalScore += score
			}
		}
	}

	return totalScore, nil
}

// Analyze the  relation between two articles with the score func
func (na WordVecAnalyzer) Analyze(main Article,
	related *Article) (bool, error) {
	//na.tryInv(main, related)
	mainStrs, err := na.getFixedStrs(main)
	if err != nil {
		return false, err
	}

	relStrs, err := na.getFixedStrs(*related)
	if err != nil {
		return false, err
	}

	score, err := na.getScore(mainStrs, relStrs)
	fmt.Println("related:", related.name, "totalScore:", score)

	return true, nil
}

var _ StandardModuleAnalyzer = (*WordVecAnalyzer)(nil)

func (na WordVecAnalyzer) tryInv(main Article, related *Article) {
	strs, err := relationDB.GetRelations(
		main.Name(),
		na.MetadataType,
		0.0)
	if err != nil {
		panic("bad bad")
	}

	for i := range strs {
		strs[i] = fixString(strs[i])
	}
}
