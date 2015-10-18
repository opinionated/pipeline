package pipeline

import (
	"github.com/opinionated/analyzer-core/analyzer"
)

// Type send through the pipeline
type AnalyzableStory struct {
	MainArticle     analyzer.Analyzable
	RelatedArticles chan analyzer.Analyzable
}

// A stage of the pipeline
type Module interface {
	// Do any stage setup
	Setup()

	// Close the stage cleanly
	// Gets any error currently in the stage and returns it
	Close() error

	// actually run the analysis
	Analyze(chan analyzer.Analyzable, chan analyzer.Analyzable, chan bool)

	SetInputChan(chan AnalyzableStory)
	GetOutputChan() chan AnalyzableStory

	// To let upstream mods know when to get out
	SetErrorPropogateChan(chan error)

	// Acually run the code
	Run()
}
