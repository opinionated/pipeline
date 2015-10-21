package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
)

type TaxonomyModule struct {
	in  chan Story
	out chan Story

	err     chan error
	closing chan chan error

	mainArticle analyzer.Analyzable
}

func (m *TaxonomyModule) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (error, bool) {

	if main.Name == " " {
		fmt.Println("empty name")
		return nil, true
	}

	// TODO: put taxonomy analyze code in here
	return nil, true
}

func (m *TaxonomyModule) Setup() {
	m.out = make(chan Story)
	m.closing = make(chan chan error)
}

func (m *TaxonomyModule) Close() error {
	errc := make(chan error)
	m.closing <- errc
	return <-errc
}

func (m *TaxonomyModule) SetInputChan(inc chan Story) {
	m.in = inc
}

func (m *TaxonomyModule) GetOutputChan() chan Story {
	return m.out
}

func (m *TaxonomyModule) SetErrorPropogateChan(errc chan error) {
	m.err = errc
}

func (m *TaxonomyModule) getErrorPropogateChan() chan error {
	return m.err
}

func (m *TaxonomyModule) getInputChan() chan Story {
	return m.in
}

func (m *TaxonomyModule) getClose() chan chan error {
	return m.closing
}

// check that was compiled properly
var _ Module = (*TaxonomyModule)(nil)
