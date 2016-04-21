package pipeline_test

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline"
	"github.com/stretchr/testify/assert"
	"testing"
)

var simpleSet = testSet{
	mainArticle: "a",
	relatedArticles: []string{
		"b",
		"c",
		"d",
	},
}

type addAnalyzer struct {
	howmuch int
}

func (a addAnalyzer) Analyze(
	_ analyzer.Analyzable, related *analyzer.Analyzable) (bool, error) {
	related.Score += float64(a.howmuch)
	return true, nil
}

func (a addAnalyzer) Setup() error {
	return nil
}

// error for the article at count
type errorAnalyzer struct {
	count int
	when  int
}

func (a *errorAnalyzer) Analyze(
	_ analyzer.Analyzable, related *analyzer.Analyzable) (bool, error) {

	a.count++
	if a.count == a.when {
		return false, fmt.Errorf("ok bump!")
	}

	return true, nil
}

type bumpAnalyzer struct {
	count int
	when  int
}

func (a errorAnalyzer) Setup() error {
	return nil
}

func (a *bumpAnalyzer) Analyze(
	_ analyzer.Analyzable, related *analyzer.Analyzable) (bool, error) {

	a.count++
	if a.count == a.when {
		return false, nil
	}

	return true, nil
}

func (a bumpAnalyzer) Setup() error {
	return nil
}

func TestStandardAdd(t *testing.T) {

	add := addAnalyzer{howmuch: 1}
	funcModule := pipeline.StandardModule{}
	funcModule.SetFuncs(add)

	pipe := pipeline.NewPipeline()
	pipe.AddStage(&funcModule)

	story := storyFromSet(simpleSet)
	data, err := storyDriver(pipe, story)

	assert.Nil(t, err)
	assert.Len(t, data, 3)

	for i := range data {
		assert.Equal(t, 1.0, data[i].Score)
	}
}

func TestBump(t *testing.T) {

	add := addAnalyzer{howmuch: 1}
	addModule := pipeline.StandardModule{}
	addModule.SetFuncs(add)

	bump := bumpAnalyzer{when: 1, count: 0}
	bumpModule := pipeline.StandardModule{}
	bumpModule.SetFuncs(&bump)

	pipe := pipeline.NewPipeline()
	pipe.AddStage(&addModule)
	pipe.AddStage(&bumpModule)

	story := storyFromSet(simpleSet)
	data, err := storyDriver(pipe, story)

	assert.Nil(t, err)
	assert.Len(t, data, 2)

	for i := range data {
		assert.Equal(t, 1.0, data[i].Score)
	}
}

func TestError(t *testing.T) {

	errFunc := errorAnalyzer{when: 2}
	funcModule := pipeline.StandardModule{}
	funcModule.SetFuncs(&errFunc)

	pipe := pipeline.NewPipeline()
	pipe.AddStage(&funcModule)

	story := storyFromSet(simpleSet)
	data, err := storyDriver(pipe, story)

	assert.NotNil(t, err)
	assert.Len(t, data, 0)
}
