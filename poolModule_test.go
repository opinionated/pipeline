package pipeline_test

import (
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPoolCuttoff(t *testing.T) {
	poolModule := pipeline.PoolModule{}
	poolModule.SetCapacity(2)

	pipe := pipeline.NewPipeline()
	pipe.AddStage(&poolModule)

	story := pipeline.Story{
		MainArticle:     analyzer.Analyzable{Name: "main"},
		RelatedArticles: make(chan analyzer.Analyzable, 5),
	}

	story.RelatedArticles <- analyzer.Analyzable{Name: "a", Score: 1.5}
	story.RelatedArticles <- analyzer.Analyzable{Name: "b", Score: 2.5}
	story.RelatedArticles <- analyzer.Analyzable{Name: "c", Score: 3.5}
	story.RelatedArticles <- analyzer.Analyzable{Name: "d", Score: 4.5}
	close(story.RelatedArticles)

	data, err := storyDriver(pipe, story)

	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 2)

	assert.Equal(t, "d", data[0].Name)
	assert.Equal(t, "c", data[1].Name)
}
