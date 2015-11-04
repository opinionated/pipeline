package pipeline_test

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/pipeline"
	"testing"
)

// only allow through articles with score >= limit
func pipeCutoff(in, out chan pipeline.Story, limit float64) {
	for story := range in {

		ostory := pipeline.Story{}
		ostory.MainArticle = story.MainArticle
		ostory.RelatedArticles = make(chan analyzer.Analyzable)

		out <- ostory
		for related := range story.RelatedArticles {
			if related.Score >= limit {
				ostory.RelatedArticles <- related
			}
		}

		close(ostory.RelatedArticles)
	}
	close(out)
	fmt.Println("done with cutoff")
}

func TestTaxonomySimple(t *testing.T) {
	// run a small, simple test set through the pipe
	BuildStoryFromFile("test", "testSets/simpleTaxonomy.json")

	taxModule := pipeline.TaxonomyModule{}
	taxModule.Setup()

	inc := make(chan pipeline.Story)
	taxModule.SetInputChan(inc)
	taxModule.SetErrorPropogateChan(make(chan error))

	// start pipeline
	go pipeline.Run(&taxModule)

	// add a cutoff
	cutOut := make(chan pipeline.Story)
	go pipeCutoff(taxModule.GetOutputChan(), cutOut, 2.0)

	errc := make(chan error)

	go StoryDriver(errc, inc, cutOut, "test")

	err := <-errc
	if err != nil {
		t.Errorf("%s", err)
		close(errc)
	}

	taxModule.Close()
}
