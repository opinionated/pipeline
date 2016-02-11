package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/jmcvetta/neoism"
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"os"
	"strconv"
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

	// used to rank taxonomies against the main
	taxonomyEngine TaxonomyEngine
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
			score, err := m.getTaxonomyRelationStrength(mainTax.Label, relatedTax.Label)
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

// Requests relation strength between two taxonomies from the db.
func (m *TaxonomyModule) getTaxonomyRelationStrength(main, related string) (float64, error) {
	// check if this request is cached
	// saves sending a net request
	// is now only a minor time saver, will get bigger with large data sets
	if v, isCached := m.cache.Get(main, related); isCached {
		return v, nil
	}

	cq := neoism.CypherQuery{
		Statement: `MATCH (a)-[r]-(b) 
		WHERE a.name={main} AND b.name={related} 
		RETURN r.cost`,
		Parameters: neoism.Props{"main": main, "related": related},
		Result:     []struct{}{},
	}

	// neoism seems to be broken so build request manually
	// these are the structures neoism uses internally
	type cypherRequest struct {
		Query      string                 `json:"query"`
		Parameters map[string]interface{} `json:"params"`
	}

	type cypherResult struct {
		Columns []string
		Data    [][]*json.RawMessage
	}

	result := cypherResult{}
	payload := cypherRequest{
		Query:      cq.Statement,
		Parameters: cq.Parameters,
	}

	ne := neoism.NeoError{}
	url := m.db.HrefCypher // get URL through db driver
	// send request
	resp, err := m.db.Session.Post(url, &payload, &result, &ne)
	if err != nil {
		fmt.Println("resp is:", resp)
		panic(err)
	}

	// if the results are filled, parse them
	// TODO: do more legitimate parsing, this is probably unsafe
	if len(result.Data) > 0 {
		// read in the first (and only) element
		rawMessage := result.Data[0][0]
		r, err := rawMessage.MarshalJSON()
		if err != nil {
			panic(err)
		}

		// be lazy and convert it to string then atoi the string
		s := string(r[:])
		score, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}

		// add this query to the cache (wouldn't hit here if it was in cache)
		m.cache.Add(main, related, float64(score))

		return float64(score), nil
	}

	// if results empty then there is no relation
	m.cache.Add(main, related, 0)

	return 0, nil
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
	db, err := neoism.Connect("http://neo4j:root@localhost:7474/db/data/")
	if err != nil {
		fmt.Println("error in setup, db is:", db)
		panic(err)
	}

	m.db = db

	m.cache.Setup()
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
