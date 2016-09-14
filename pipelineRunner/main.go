package main

import (
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline"
	"github.com/opinionated/pipeline/analyzer/dbInterface"
)

// for running the pipe
func buildStory(article string) (pipeline.Story, error) {

	story := pipeline.Story{}
	story.MainArticle = analyzer.Analyzable{FileName: article}
	story.RelatedArticles = make(chan analyzer.Analyzable)

	articles, err := relationDB.GetAll()
	if err != nil {
		return story, err
	}

	go func(relatedArticles []string, story pipeline.Story) {

		for i := range relatedArticles {
			story.RelatedArticles <- analyzer.Analyzable{
				FileName: relatedArticles[i],
			}
		}

		close(story.RelatedArticles)

	}(articles, story)

	return story, nil
}

// run the article through the pipeline
func runArticle(pipe *pipeline.Pipeline, article string) ([]analyzer.Analyzable, error) {
	story, err := buildStory(article)
	if err != nil {
		panic(err)
	}

	pipe.Start()
	pipe.PushStory(story)

	var result pipeline.Story

	select {
	case result = <-pipe.GetOutput():
		break

	case <-pipe.Error():
		// go get the error when you actually close the pipe
		err := pipe.Close()
		return nil, err
	}

	var related []analyzer.Analyzable

	for {
		select {
		case analyzed, open := <-result.RelatedArticles:
			if !open {
				return related, pipe.Close()
			}
			related = append(related, analyzed)

		case <-pipe.Error():
			// get the error when you close the pipe
			err := pipe.Close()
			return nil, err
		}
	}
}

// functions for building the pipeline
func getCountSqLambdaWithWeight(weight float64) func(float32, int) float64 {
	return func(flow float32, count int) float64 {
		return weight * float64(flow) * float64(count*count)
	}
}

func getTaxonomyModule() (pipeline.Module, error) {
	weight := 1.0
	scoreFunc := getCountSqLambdaWithWeight(weight)
	taxFunc := pipeline.NeoAnalyzer{MetadataType: "Taxonomy",
		ScoreFunc: scoreFunc}

	taxModule := new(pipeline.StandardModule)
	taxModule.SetFuncs(&taxFunc) // is safe to ref local value

	return taxModule, nil
}

func getConceptModule() (pipeline.Module, error) {
	weight := 1.0
	scoreFunc := getCountSqLambdaWithWeight(weight)
	neoFunc := pipeline.NeoAnalyzer{MetadataType: "Concept",
		ScoreFunc: scoreFunc}

	module := new(pipeline.StandardModule)
	module.SetFuncs(&neoFunc)

	return module, nil
}

func getEntityModule() (pipeline.Module, error) {
	weight := 1.0
	scoreFunc := getCountSqLambdaWithWeight(weight)
	neoFunc := pipeline.NeoAnalyzer{MetadataType: "Entity",
		ScoreFunc: scoreFunc}

	module := new(pipeline.StandardModule)
	module.SetFuncs(&neoFunc)

	return module, nil
}

func getKeywordModule() (pipeline.Module, error) {
	weight := 1.0
	scoreFunc := getCountSqLambdaWithWeight(weight)
	neoFunc := pipeline.NeoAnalyzer{MetadataType: "Keyword",
		ScoreFunc: scoreFunc}

	module := new(pipeline.StandardModule)
	module.SetFuncs(&neoFunc)

	return module, nil
}

func buildPipeline() (*pipeline.Pipeline, error) {
	pipe := pipeline.NewPipeline()
	taxModule, err := getTaxonomyModule()
	if err != nil {
		return nil, err
	}

	pipe.AddStage(taxModule)

	conceptModule, err := getConceptModule()
	pipe.AddStage(conceptModule)

	entityModule, err := getEntityModule()
	pipe.AddStage(entityModule)

	keywordModule, err := getKeywordModule()
	pipe.AddStage(keywordModule)

	return pipe, nil
}

func main() {
	startServer(":8005")

}
