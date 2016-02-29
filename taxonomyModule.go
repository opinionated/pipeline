package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"os"
)

// TaxonomyModule ranks articles by taxonomy, a high level grouping. An
// example taxonomy might be "government and politics/elections". This
// is a fairly coarse method.
type TaxonomyModule struct {
	in  chan Story
	out chan Story

	err     chan error
	closing chan chan error

	// hold the current main article
	mainTaxonomys  []alchemy.Taxonomy // holds main article's taxonomies
	mainIdentifier string             // to check when article changes

	// interface with DB
	engine alchemy.TaxonomyEngine
}

// Analyze ranks articles by taxonomy.
func (m *TaxonomyModule) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	// reload when we get a new main article
	if len(m.mainTaxonomys) == 0 || m.mainIdentifier != main.FileName {
		tax, err := m.getArticleTaxonomy(main)
		if err != nil {
			return false, err
		}

		m.mainIdentifier = main.FileName
		m.mainTaxonomys = tax
	}

	// load related article data
	relatedTax, err := m.getArticleTaxonomy(*related)
	if err != nil {
		return false, err
	}

	// calculate score
	str, err := m.rankTaxonomyAgainstMain(relatedTax)
	if err != nil {
		fmt.Println("got an err:", err)
		return false, err
	}

	// update related score
	related.Score += str

	// TODO: put taxonomy analyze code in here
	return true, nil
}

// Scores related taxonomies against the main taxonomies.
// Each main taxonomy is scored against each related taxonomy.
// A relation score for each taxonomy pair is calculated from
// the graph, then weighted by the taxonomy/article strength.
func (m *TaxonomyModule) rankTaxonomyAgainstMain(related []alchemy.Taxonomy) (float64, error) {

	totalScore := 0.0

	for _, mainTax := range m.mainTaxonomys {
		for _, relatedTax := range related {

			if mainTax.Label == relatedTax.Label {
				// give equivalent taxonomies a high score
				totalScore += 5.0 * float64(mainTax.Score*relatedTax.Score)
				continue
			}

			// fetch from DB
			score, err := m.engine.GetRelationStrength(mainTax.Label, relatedTax.Label)
			if err != nil {
				// TODO: handle this error better
				return 0.0, err
			}

			// factor in the taxonomy/article strengths
			score *= float64(mainTax.Score * relatedTax.Score)
			totalScore += score
		}
	}

	return totalScore, nil
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

// Setup connects to the graph server and sets up the cache
func (m *TaxonomyModule) Setup() {
	m.out = make(chan Story)
	m.closing = make(chan chan error)

	m.engine.Start()
}

// Close stops the module and cleans up any open connections.
func (m *TaxonomyModule) Close() error {
	errc := make(chan error)
	m.closing <- errc
	return <-errc
}

// SetInputChan sets the module's input channel.
func (m *TaxonomyModule) SetInputChan(inc chan Story) {
	m.in = inc
}

// GetOutputChan returns the modules output channel.
func (m *TaxonomyModule) GetOutputChan() chan Story {
	return m.out
}

// SetErrorPropogateChan sets the channel for errors to propagate out
// of this module.
func (m *TaxonomyModule) SetErrorPropogateChan(errc chan error) {
	m.err = errc
}

// remaining methods are used internally by run methods

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
