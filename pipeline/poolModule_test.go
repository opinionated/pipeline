package pipeline_test

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/pipeline"
	"testing"
)

func TestPoolCuttoff(t *testing.T) {

	// analyzable factory
	m := func(name string, score float64) analyzer.Analyzable {
		analyzable := analyzer.BuildAnalyzable()
		analyzable.Name = name
		analyzable.Score = score
		return analyzable
	}

	// in/out sets
	// the heap will only grab the highest values
	inputArticles := []analyzer.Analyzable{m("d", 4.0), m("a", 0.1), m("b", 2.0), m("c", 0.2)}
	outputArticles := []analyzer.Analyzable{m("b", 0.0), m("d", 0.0)}

	// build the cut module
	cut := pipeline.PoolModule{}
	cut.Setup()
	cut.SetCapacity(2)

	inc := make(chan pipeline.Story)
	cut.SetInputChan(inc)
	cut.SetErrorPropogateChan(make(chan error))

	// needs a special run because it behaves differently
	go pipeline.RunPool(&cut)

	// build the input story
	inputStory := pipeline.Story{}
	inputStory.MainArticle = analyzer.Analyzable{}
	inputStory.MainArticle.Name = "test"

	inputStory.RelatedArticles = make(chan analyzer.Analyzable)

	inc <- inputStory

	done := make(chan error)

	// send input articles
	go func() {
		for _, article := range inputArticles {
			fmt.Println("sending:", article.Name)
			inputStory.RelatedArticles <- article
		}
		close(inputStory.RelatedArticles)
	}()

	// read output articles
	go func() {
		outStory := <-cut.GetOutputChan()

		i := 0 // tracks # articles read
		for _, expected := range outputArticles {

			result := <-outStory.RelatedArticles
			if result.Name != expected.Name {
				done <- fmt.Errorf("cutoff error, expected: %s got: %s", expected.Name, result.Name)
			}
			i++
		}

		if i != len(outputArticles) {
			done <- fmt.Errorf("pool error, did not get all expected articles back!")
		}

		// signal done
		close(done)
	}()

	// print any errors
	for err := range done {
		t.Errorf("%s\n", err)
	}

	cut.Close()

}
