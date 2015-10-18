package pipeline_test

import (
	"fmt"
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/pipeline"
	"os"
	"testing"
)

func TestFilterTaxonomys(t *testing.T) {
	articles := make(chan analyzer.Analyzable)

	go func() {
		for i := 1; i < len(simpleTaxonomySet); i++ {
			article := analyzer.BuildAnalyzable()
			article.FileName = "testData/simpleTaxonomy/" + simpleTaxonomySet[i]
			article.Name = simpleTaxonomySet[i]
			articles <- article
		}
		close(articles)
	}()

	top := analyzer.BuildAnalyzable()
	top.FileName = "testData/simpleTaxonomy/" + simpleTaxonomySet[0]
	file, err := os.Open(top.FileName + "_taxonomy.xml")
	if err != nil {
		panic(err)
	}

	err = alchemy.ToXML(file, &top.Taxonomys)
	if err != nil {
		panic(err)
	}

	filtered := pipeline.FilterTaxonomy(top, articles)
	found := <-filtered
	if found.Name != simpleTaxonomySet[1] {
		t.Errorf("expected article:", simpleTaxonomySet[1], "got:", found.Name)
	}
}

func TestLoadTaxonomys(t *testing.T) {
	articles := make(chan analyzer.Analyzable)
	withTaxonomy := pipeline.LoadTaxonomy(articles)

	go func() {
		for _, link := range simpleTaxonomySet {
			article := analyzer.BuildAnalyzable()
			article.FileName = "testData/simpleTaxonomy/" + link
			articles <- article
		}
		close(articles)
	}()

	for i := 0; i < len(simpleTaxonomySet); i++ {
		analyzable := <-withTaxonomy
		fmt.Println("analyzable:", analyzable)
	}
}

func TestBuildData(t *testing.T) {
	t.Skip("only run this when you need to set up a new set")
	// build a test set
	articles := simpleTaxonomySet
	path := "testData/simpleTaxonomy/"
	for _, link := range articles {
		article, err := alchemy.ParseArticle(path + link + ".txt")
		if err != nil {
			panic(err)
		}

		keywords := alchemy.Keywords{}
		err = alchemy.GetKeywords(article, &keywords)
		if err != nil {
			panic(err)
		}
		err = alchemy.MarshalToFile(path+link+"_keywords.xml", keywords)
		if err != nil {
			panic(err)
		}

		taxonomy := alchemy.Taxonomys{}
		err = alchemy.GetTaxonomy(article, &taxonomy)
		if err != nil {
			panic(err)
		}
		err = alchemy.MarshalToFile(path+link+"_taxonomy.xml", taxonomy)
		if err != nil {
			panic(err)
		}
	}
}
