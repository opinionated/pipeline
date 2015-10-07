package pipeline

import (
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/scraper-core/article"
)

func StartPipe(article scraper.Article) {
	mainArticle := analyzer.BuildAnalyzable()
	err := alchemy.GetTaxonomy(article.GetData(), &mainArticle.Taxonomys)
}
