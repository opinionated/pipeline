package pipeline_test

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/pipeline"
	"testing"
)

var simpleTaxonomySet = []string{
	"thisIsWhatEscalationLooksLike",
	"isValdimirPutinTryingToTeachTheWestALessonInSyria",
	"alabamaPutsUpMoreHurdlesForVoters",
}

func buildStory(name string, output chan analyzer.Analyzable) pipeline.Story {
	story := pipeline.Story{}

	mainArticle := analyzer.Analyzable{}
	mainArticle.Name = "main"
	story.MainArticle = mainArticle

	story.RelatedArticles = output

	return story
}

func RunPipeFull(quit chan bool) {
	// build story
	related := make(chan analyzer.Analyzable, 4)
	story := buildStory("test", related)

	// build module
	module := pipeline.TaxonomyModule{}
	module.Setup()

	i := make(chan pipeline.Story)
	module.SetInputChan(i)

	errc := make(chan error)
	module.SetErrorPropogateChan(errc)
	// send story to module
	go pipeline.Run(&module)

	i <- story
	close(i)
	ostr := <-module.GetOutputChan()

	for i := 0; i < 5000000; i++ {
		toSend := analyzer.BuildAnalyzable()

		related <- toSend

		<-ostr.RelatedArticles

	}
	close(related)

	module.Close()

	fmt.Println("all done!")

	quit <- true
	close(quit)
}

func BenchmarkModularPipeline(b *testing.B) {
	// build story
	related := make(chan analyzer.Analyzable)
	story := buildStory("test", related)

	// build module
	module := pipeline.TaxonomyModule{}
	module.Setup()

	i := make(chan pipeline.Story)
	module.SetInputChan(i)

	errc := make(chan error)
	module.SetErrorPropogateChan(errc)
	// send story to module
	go pipeline.Run(&module)
	i <- story
	close(i)

	ostr := <-module.GetOutputChan()
	j := 0
	for ; j < b.N; j++ {
		toSend := analyzer.BuildAnalyzable()

		related <- toSend

		<-ostr.RelatedArticles
	}
	module.Close()
}

func BenchmarkModularPipeline1(b *testing.B) { BenchmarkModularPipeline(b) }

func TestTaxonomyModuleSimple(t *testing.T) {
	q := make(chan bool)
	go RunPipeFull(q)
	<-q
}
