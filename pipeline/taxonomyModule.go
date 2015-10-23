package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"os"
)

type TaxonomyModule struct {
	in  chan Story
	out chan Story

	err     chan error
	closing chan chan error

	mainTaxonomys []alchemy.Taxonomy
}

// TODO: think about switching the order of err + bool
func (m *TaxonomyModule) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (error, bool) {

	// TODO: reload when the taxonomies close
	if len(m.mainTaxonomys) == 0 {
		tax, err := m.getArticleTaxonomy(main)
		if err != nil {
			return err, false
		}

		m.mainTaxonomys = tax
	}

	// load the related article taxonomies
	relatedTax, err := m.getArticleTaxonomy(*related)
	if err != nil {
		return err, false
	}

	// TODO: make a helper function so we can change how
	// the scoring is done
	if m.mainTaxonomys[0].Label == relatedTax[0].Label {
		return nil, true
	}

	// TODO: put taxonomy analyze code in here
	return nil, false
}

// helper function to load taxonomies from file
// made a member so it doesn't clutter package
func (m *TaxonomyModule) getArticleTaxonomy(article analyzer.Analyzable) ([]alchemy.Taxonomy, error) {

	// open tax file
	file, err := os.Open(article.FileName + "_taxonomy.xml")
	defer file.Close()
	if err != nil {
		fmt.Println("oh nsoe, error opening file")
		return []alchemy.Taxonomy{}, err
	}

	// read taxonomies from file
	ret := alchemy.Taxonomys{}
	err = alchemy.ToXML(file, &ret)
	if err != nil {
		fmt.Println("oh nose, error reading file")
		return []alchemy.Taxonomy{}, err
	}

	return ret.Taxonomys, err
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
