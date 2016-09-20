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
	mux    *sync.Mutex // need to serialize access to map
}

// Name of the article
func (article Article) Name() string {
	return article.name
}

// AddScore adds a generic score by name
func (article *Article) AddScore(name string, score PipelineScore) error {
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
	var val Score

	if val, ok := article.scores[name]; !ok {
		err = fmt.Errorf("%s isn't in the map\n", name)
	}

	article.mux.Unlock()

	return val, err
}
