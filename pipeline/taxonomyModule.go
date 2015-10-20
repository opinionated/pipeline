package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
)

type TaxonomyModule struct {
	in  chan AnalyzableStory
	out chan AnalyzableStory

	err     chan error
	closing chan chan error

	mainArticle analyzer.Analyzable
}

func (m *TaxonomyModule) Analyze(main, related analyzer.Analyzable) analyzer.Analyzable {
	if main.Name == " " {
		fmt.Println("empty name")
		return related
	}

	// TODO: put taxonomy analyze code in here

	return related
}

func (m *TaxonomyModule) Setup() {
	m.out = make(chan AnalyzableStory)
	m.closing = make(chan chan error)
}

func (m *TaxonomyModule) Close() error {
	errc := make(chan error)
	m.closing <- errc
	return <-errc
}

func (m *TaxonomyModule) SetInputChan(inc chan AnalyzableStory) {
	m.in = inc
}

func (m *TaxonomyModule) GetOutputChan() chan AnalyzableStory {
	return m.out
}

func (m *TaxonomyModule) SetErrorPropogateChan(errc chan error) {
	m.err = errc
}

func (m *TaxonomyModule) getErrorPropogateChan() chan error {
	return m.err
}

func (m *TaxonomyModule) getInputChan() chan AnalyzableStory {
	return m.in
}

func (m *TaxonomyModule) getClose() chan chan error {
	return m.closing
}

// check that was compiled properly
var _ Module = (*TaxonomyModule)(nil)
