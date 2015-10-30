package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/jmcvetta/neoism"
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
	db            *neoism.Database
}

// TODO: think about switching the order of err + bool
func (m *TaxonomyModule) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (error, bool) {

	// reload taxonomies
	// TODO: make a compare so we can check when main article changes
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

	if m.mainTaxonomys[0].Label == relatedTax[0].Label {
		//		fmt.Println("done here")
		//		return nil, true
	}

	str, err := m.rankTaxonomyAgainsMain(relatedTax)
	fmt.Println("main", main.Name, "related", related.Name, "score:", str)

	if err != nil {
		fmt.Println("got an err:", err)
		return nil, false
	}
	fmt.Println("got str:", str)
	if str == 0 {
		return nil, false
	}

	// TODO: put taxonomy analyze code in here
	return nil, true
}

func (m *TaxonomyModule) rankTaxonomyAgainsMain(related []alchemy.Taxonomy) (float64, error) {
	totalScore := 0.0
	fmt.Println("main:", m.mainTaxonomys, "related:", related)
	for _, mainTax := range m.mainTaxonomys {
		for _, relatedTax := range related {
			fmt.Println("comparing", mainTax.Label, "against", relatedTax.Label)
			if mainTax.Label == relatedTax.Label {
				fmt.Println("are equivalent")
				totalScore += 5.0
				continue
			}
			score, err := m.getRelationStrength(mainTax.Label, relatedTax.Label)
			if err != nil {
				return 0.0, err
			}
			totalScore += float64(score)
		}
	}
	return totalScore, nil
}

func (m *TaxonomyModule) getRelationStrength(main, related string) (float64, error) {
	res := []struct {
		strength float64 `json:"r.cost"`
		rType    string  `json:"type(r)"`
	}{}

	cq := neoism.CypherQuery{
		Statement: `MATCH (a)-[r]-(b) WHERE a.name={main} AND b.name={related} RETURN r.cost,type(r)`,
		//Statement:  `MATCH (a)-[r]-(b) WHERE a.name={main} AND b.name={related} RETURN a`,
		Parameters: neoism.Props{"main": main, "related": related},
		Result:     &res,
	}
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
	url := m.db.HrefCypher
	resp, err := m.db.Session.Post(url, &payload, &result, &ne)
	if len(result.Data) > 0 {
		//tmpLen := len(result.Data[0][1])
		j := result.Data[0][0]
		r, err := j.MarshalJSON()
		s := string(r[:len(r)])
		fmt.Println("body is: ", s, "err:", err)
		fmt.Println("resp raw is:", result)
		fmt.Println("resp is:", result.Data[0][1])
	}
	//err := m.db.Cypher(&cq)
	if err != nil {
		fmt.Println("resp is:", resp)
		panic(err)
	}

	if len(res) == 0 {
		return 0, err
	}

	fmt.Println("result is:", cq.Result)
	return 0, nil
	//	return float64(res[0].strength), err
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
