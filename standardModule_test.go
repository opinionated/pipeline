package pipeline_test

import (
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/analyzer-core/dbInterface"
	"github.com/opinionated/pipeline"
	"testing"
)

func TestStandardTaxonomy(t *testing.T) {

	taxModule := pipeline.NeoModule{}
	taxModule.SetParams("Taxonomy", pipeline.ScoreSimpleMul)

	taxpipe := pipeline.NewPipeline()
	taxpipe.AddStage(&taxModule)

	taxStory := storyFromSet(neoTestSet)
	taxresults, err := storyDriver(taxpipe, taxStory)
	if err != nil {
		t.Error(err)
	}

	// now go use the func pipe
	analyzeFunc := func(
		main analyzer.Analyzable, related *analyzer.Analyzable) (bool, error) {
		flow, count, err :=
			relationDB.StrengthBetween(main.FileName,
				related.FileName, "Taxonomy")
		if err != nil {
			return false, err
		}
		related.Score += float64(flow * float32(count))
		return true, nil
	}

	setupFunc := func() error {
		return relationDB.Open("http://localhost:7474")
	}

	funcModule := pipeline.StandardModule{}
	funcModule.SetFuncs(analyzeFunc, setupFunc)

	pipe := pipeline.NewPipeline()
	pipe.AddStage(&funcModule)

	story := storyFromSet(neoTestSet)
	results, err := storyDriver(pipe, story)
	if err != nil {
		t.Error(err)
	}

	// now go check
	if len(results) != len(taxresults) {
		panic("bad!")
	}

	if len(results) != len(neoTestSet.relatedArticles) {
		panic("bad!")
	}

	for i := range results {
		if results[i].FileName != taxresults[i].FileName {
			t.Errorf("out of order")
		}
		if results[i].Score != taxresults[i].Score {
			t.Error("scores not equal!")
		}
	}

}
