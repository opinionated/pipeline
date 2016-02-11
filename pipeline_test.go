package pipeline_test

import (
	"github.com/opinionated/pipeline/pipeline"
	"testing"
)

func TestDoubleStage(t *testing.T) {
	p := pipeline.NewPipeline(2)

	inc := make(chan pipeline.Story)
	p.SetInput(inc)

	p.AddStage(&pipeline.TaxonomyModule{})
	p.AddStage(&pipeline.TaxonomyModule{})

	go p.Start()

	errc := make(chan error)
	_ = BuildStoryFromFile("simpleTaxonomy", "testSets/simpleTaxonomy.json")
	go StoryDriver(errc, inc, p.GetOutput(), "simpleTaxonomy")

	err := <-errc
	if err != nil {
		t.Errorf("%s", err)
		close(errc)
	}
	p.Stop()

}
