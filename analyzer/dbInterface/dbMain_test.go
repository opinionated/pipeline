package relationDB

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func finishTest() {
	clear()
	Close()
}

type Relation struct {
	Text      string `json:"Text"`
	Relevance float32
}

var testGraph = []struct {
	node      string
	relations []Relation
}{
	{
		"gunsRgreat",
		[]Relation{
			{"guns", 3.0},
			{"knife massacres", 2.0},
			{"carrots", 0.5},
			{"terroists", 1.5},
			{"hunting", 2.5},
			{"2nd ammendment", 3.0},
		},
	},
	{
		"gunsBad",
		[]Relation{
			{"guns", 2.0},
			{"kids", 4.0},
			{"carrots", 1.5},
		},
	},
	{
		"obama hitler",
		[]Relation{
			{"kenya", 4.0},
			{"hunting", 1.0},
			{"guns", 1.5},
			{"muslim", 3.0},
			{"terrorists", 2.5},
			{"2nd ammendment", 2.0},
		},
	},
	{
		"disconnected from everything",
		[]Relation{},
	},
}

func setupTestGraph(t *testing.T) {
	clear()
	for i := range testGraph {
		article := testGraph[i]
		assert.Nil(t, Store(article.node))
		assert.Nil(t, InsertRelations(article.node, "keywords", article.relations))
	}
}

func TestShortestPath(t *testing.T) {
	t.Skip("explicitly enable, this will clear")
	assert.Nil(t, Open("http://localhost:7474/"))
	setupTestGraph(t)
	//defer finishTest()

	info, err := GetByUUID("gunsBad")
	assert.Nil(t, err)
	assert.Equal(t, "gunsBad", info.Identifier)

	score, count, err := StrengthBetween("gunsRgreat", "obama hitler", "keywords")
	assert.Nil(t, err)
	assert.EqualValues(t, 13.0, score)
	assert.EqualValues(t, 3, count)

	score, count, err = StrengthBetween("gunsRgreat", "disconnected from everything", "keywords")
	assert.Nil(t, err)
	assert.EqualValues(t, 0.0, score)
	assert.EqualValues(t, 0, count)

	score, count, err = StrengthBetween("gunsRgreat", "obama hitler", "taxonomy")
	assert.Nil(t, err)
	assert.EqualValues(t, 0.0, score)
	assert.EqualValues(t, 0, count)
}

func TestMultiInsert(t *testing.T) {
	t.Skip("explicitly enable, this will clear")
	assert.Nil(t, Open("http://localhost:7474/"))
	defer finishTest()
	assert.Nil(t, clear())

	var _ = []string{
		"z", "1.0",
	}
	var relations = []Relation{
		{"x", 1.0},
		{"y", 1.0},
	}

	assert.Nil(t, Store("n"))
	assert.Nil(t, InsertRelations("n", "keywords", relations))
}

func TestInsert(t *testing.T) {
	t.Skip("explicitly enable, this will clear")
	// hits get by uuid and insert
	assert.Nil(t, Open("http://localhost:7474/"))
	assert.Nil(t, clear())
	defer finishTest()

	var tests = []struct {
		in  string
		err string
	}{
		{"hey", ""},
		{"hey", "uuid not unique"},
		{"hey hey", ""},
	}

	for i := range tests {
		err := Store(tests[i].in)
		errStr := tests[i].err
		if err != nil {
			assert.Equal(t, errStr, err.Error())
		}
	}
}

func TestInvRelations(t *testing.T) {
	assert.Nil(t, Open("http://localhost:7474/"))
	articles, err := GetRelationsInv("Mr. Obama", "", 0.5)

	assert.Nil(t, err)
	assert.Nil(t, articles)
	assert.True(t, len(articles) > 2)
}
