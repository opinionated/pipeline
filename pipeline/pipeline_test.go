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

	done := make(chan bool)
	_ = BuildStoryFromFile("simpleTaxonomy", "testSets/simpleTaxonomy.json")
	go StoryDriver(t, inc, p.GetOutput(), "simpleTaxonomy", done)
	<-done

	p.Stop()
}
