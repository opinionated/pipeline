package pipeline

import (
	"fmt"
	"sync"
)

// Score represents a single score in the pipeline.
type Score interface {
	// convert the scores into an array of floats to make combining easy
	Serialize() []float32
}

// Article is used to track an article score in teh pipeline
type Article struct {
	name   string
	scores map[string]Score
	mux    *sync.Mutex // need to serialize access to map TODO: I think we can drop this, not 100% tho
}

// NewArticle with name
func NewArticle(n string) Article {
	article := Article{
		name:   n,
		scores: make(map[string]Score),
		mux:    new(sync.Mutex)}

	return article
}

// Name of the article
func (article Article) Name() string {
	return article.name
}

// AddScore adds a generic score by name
func (article *Article) AddScore(name string, score Score) error {
	article.mux.Lock()

	var err error
	if _, ok := article.scores[name]; ok {
		err = fmt.Errorf("%s is already in the map\n", name)
	} else {
		article.scores[name] = score
	}

	article.mux.Unlock()

	return err
}

// GetScore by name
func (article *Article) GetScore(name string) (Score, error) {

	article.mux.Lock()

	var err error
	val, ok := article.scores[name]
	article.mux.Unlock()

	if !ok {
		err = fmt.Errorf("%s isn't in the map\n", name)
	}

	return val, err
}

// AllScores as an array
func (article *Article) AllScores() []Score {
	article.mux.Lock()

	arr := make([]Score, len(article.scores))
	i := 0
	for _, score := range article.scores {
		arr[i] = score
		i++
	}

	article.mux.Unlock()
	return arr
}

// Keys as array
func (article *Article) Keys() []string {
	article.mux.Lock()

	arr := make([]string, len(article.scores))
	i := 0
	for key := range article.scores {
		arr[i] = key
		i++
	}

	article.mux.Unlock()
	return arr
}
