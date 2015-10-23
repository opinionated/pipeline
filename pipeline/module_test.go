package pipeline_test

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/pipeline"
	"github.com/opinionated/utils/config"
	"os"
	"testing"
)

// functions to help with testing pipeline stages
// does not have any actual tests of the modular pipeline
// TODO: write tests for the run() function

func BuildStoryFromFile(name, file string) pipeline.Story {
	f, err := os.Open(file)
	defer f.Close()

	if err != nil {
		panic(err)
	}

	config.InitConfig()
	err = config.ReadFile(name, f)

	if err != nil {
		panic(err)
	}

	story := pipeline.Story{}
	story.MainArticle = analyzer.Analyzable{}
	story.MainArticle.Name = "main"
	_, ok := config.From(name).Nested("inputSet").GetArray("related")
	if !ok {
		panic("could not convert to array")

	}

	return story
}

// manages testing of a story, given the input and expected output
// load the story you want it to drive, then build it from the file
func StoryDriver(t *testing.T, inc chan pipeline.Story, output chan pipeline.Story, name string, done chan bool) {

	// build the story to send down the pipe
	story := pipeline.Story{}

	story.MainArticle = analyzer.Analyzable{}
	input := config.From(name).Nested("inputSet")
	mainName, ok := input.Get("main").(string)
	if !ok {
		panic("could not read main article name")
	}
	story.MainArticle.Name = mainName
	story.MainArticle.FileName = "testData/" + mainName

	story.RelatedArticles = make(chan analyzer.Analyzable)

	// send it down k
	inc <- story

	// go feed the stories into the pipe
	go func() {
		// build the inputs
		arr, ok := input.GetArray("related")

		if !ok {
			panic("could not convert input to array")
		}

		for _, link := range arr {
			related := analyzer.Analyzable{}
			str, ok := link.(string)
			if !ok {
				panic("error, could not convert type!")
			}
			related.Name = str
			related.FileName = "testData/" + str
			fmt.Println("sending:", related.Name)
			story.RelatedArticles <- related
		}
		defer close(story.RelatedArticles)
	}()

	// read the actual ouput and compare it to the expected
	quit := make(chan bool)
	go func() {
		ostory := <-output
		arr, ok := config.From(name).GetArray("output")

		if !ok {
			panic("could not convert output to array")
		}

		i := 0 // count along with expected

		for article := range ostory.RelatedArticles {
			if i > len(arr) {
				// make sure it doesn't go out of range
				t.Errorf("unexpected output for set %s: article %s is beyond test set",
					name,
					article.Name)
			}

			fmt.Println("from pipe:", article.Name)
			// convert arr to str
			str, ok := arr[i].(string)
			if !ok {
				panic("error, could not convert output to string")
			}

			if article.Name != str {
				// compare expected and actual sets
				t.Errorf("unexpected output for set %s: expected %s but got %s",
					name,
					str,
					article.Name)
			}

			i++
		}

		if i != len(arr) {
			t.Errorf("failed to read all the expected inputs out of the pipe")
		}
		// finish up
		close(quit)
	}()

	<-quit

	close(done)
}
