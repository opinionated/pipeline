package pipeline

import (
	"github.com/opinionated/pipeline/analyzer/dbInterface"
	"github.com/opinionated/word2vec"
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
	cache        []relationDB.KeywordRelation
	cacheArt     relationDB.KeywordRelation
}

// Setup the connection to the neo db
func (na *WordVecAnalyzer) Setup() error {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		panic(err)
	}

	na.client = word2vec.Client{Addr: "localhost:1234"}
	na.cache = nil

	return nil
}

// replaces spaces in a string with underscores
func fixString(str string) string {
	return strings.Replace(str, " ", "_", -1)
}

func (na WordVecAnalyzer) getFixedStrs(article Article) ([]relationDB.KeywordRelation, error) {
	if na.cacheArt.Identifier == article.Name() {
		return na.cache, nil
	}

	strs, err := relationDB.GetRelations(
		article.Name(),
		na.MetadataType,
		0.0)
	if err != nil {
		return nil, err
	}

	for i := range strs {
		strs[i].Identifier = fixString(strs[i].Identifier)
	}

	return strs, err
}

func (na WordVecAnalyzer) getScore(main, related []relationDB.KeywordRelation) (float32, int, error) {

	var totalScore float32
	totalScore = 0.0
	totalCount := 0

	// build maps for the data
	// TODO: refactor and cleanup
	mainStrs := make([]string, len(main))
	relatedStrs := make([]string, len(related))
	mainMap := make(map[string]float32)
	relatedMap := make(map[string]float32)

	for i := range main {
		mainStrs = append(mainStrs, main[i].Identifier)
		mainMap[main[i].Identifier] = main[i].Relevance
	}
	for i := range related {
		relatedStrs = append(relatedStrs, related[i].Identifier)
		relatedMap[related[i].Identifier] = related[i].Relevance
	}

	// TODO: clean up err handling
	mainVecs, err := na.client.Vectors(mainStrs)
	if err != nil {
		panic(err)
	}

	relatedVecs, err := na.client.Vectors(relatedStrs)
	if err != nil {
		panic(err)
	}

	totalScore, totalCount = ClusterOverlap(mainVecs, relatedVecs, mainMap, relatedMap)

	return totalScore, totalCount, nil
}

// Analyze the  relation between two articles with the score func
func (na *WordVecAnalyzer) Analyze(main Article,
	related *Article) (bool, error) {
	mainStrs, err := na.getFixedStrs(main)
	if err != nil {
		return false, err
	}

	if main.Name() != na.cacheArt.Identifier {
		na.cache = mainStrs
	}

	relStrs, err := na.getFixedStrs(*related)
	if err != nil {
		return false, err
	}

	score, count, err := na.getScore(mainStrs, relStrs)

	scoreStruct := NeoScore{score, count}
	err = related.AddScore("wordvec_"+na.MetadataType, scoreStruct)
	return true, err
}

var _ StandardModuleAnalyzer = (*WordVecAnalyzer)(nil)
