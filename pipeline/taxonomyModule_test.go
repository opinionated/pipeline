package pipeline_test

import (
	"github.com/opinionated/pipeline/pipeline"
	"testing"
)

func TestTaxonomySimple(t *testing.T) {
	// run a small, simple test set through the pipe
	BuildStoryFromFile("test", "testSets/simpleTaxonomy.json")

	pipe := pipeline.TaxonomyModule{}
	pipe.Setup()

	inc := make(chan pipeline.Story)
	pipe.SetInputChan(inc)
	pipe.SetErrorPropogateChan(make(chan error))
	go pipeline.Run(&pipe)

	done := make(chan bool)

	go StoryDriver(t, inc, pipe.GetOutputChan(), "test", done)

	<-done
	pipe.Close()
}
