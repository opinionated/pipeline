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
// example taxonomy might be "government and politics/elections".
type TaxonomyModule struct {
	in  chan Story
	out chan Story

	err     chan error
	closing chan chan error

	mainTaxonomys []alchemy.Taxonomy

	db    *neoism.Database // db driver for neo4j graph db
	cache Neo4jCache       // simple cache to store neo requests
}

// Analyze ranks articles by taxonomy.
func (m *TaxonomyModule) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	// reload taxonomies
	// TODO: make a compare so we can check when main article changes
	if len(m.mainTaxonomys) == 0 {
		tax, err := m.getArticleTaxonomy(main)
		if err != nil {
			return false, err
		}

		m.mainTaxonomys = tax
	}

	// load the related article taxonomies
	relatedTax, err := m.getArticleTaxonomy(*related)
	if err != nil {
		return false, err
	}

	// score taxonomies
	str, err := m.rankTaxonomyAgainsMain(relatedTax)
	if err != nil {
		fmt.Println("got an err:", err)
		return false, err
	}

	// update related score
	related.Score += str

	// TODO: put taxonomy analyze code in here
	return true, nil
}

// actually does scoring, returns score
func (m *TaxonomyModule) rankTaxonomyAgainsMain(related []alchemy.Taxonomy) (float64, error) {

	totalScore := 0.0

	for _, mainTax := range m.mainTaxonomys {
		for _, relatedTax := range related {

			if mainTax.Label == relatedTax.Label {
				// give equlivalent taxonomies a high score
				totalScore += 5.0 * float64(mainTax.Score*relatedTax.Score)
				continue
			}

			// fetch from DB
			score, err := m.getRelationStrength(mainTax.Label, relatedTax.Label)
			if err != nil {
				// TODO: handle this error better
				return 0.0, err
			}

			// factor in the taxonomy weights
			score *= float64(mainTax.Score * relatedTax.Score)
			totalScore += score
		}
	}

	return totalScore, nil
}

// sends off DB request
// TODO: clean and speed this up
func (m *TaxonomyModule) getRelationStrength(main, related string) (float64, error) {
	// check if this request is cached
	// saves sending a net reqeust
	if v, isCached := m.cache.Get(main, related); isCached {
		// is now only a minor time saver, will get bigger with large data sets
		return v, nil
	}

	cq := neoism.CypherQuery{
		Statement: `MATCH (a)-[r]-(b) 
		WHERE a.name={main} AND b.name={related} 
		RETURN r.cost`,
		Parameters: neoism.Props{"main": main, "related": related},
		Result:     []struct{}{},
	}

	// neoism seems to be broken so build request manualy
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

		fmt.Println("connected main:", main, "related:", related, "score:", score)

		// add this query to the cache (wouldn't hit here if it was in cache)
		m.cache.Add(main, related, float64(score))

		return float64(score), nil
	}
	fmt.Println("failed to connect main:", main, "related:", related)

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
