// Package relationDB is used to find relations between articles.
package relationDB

// Graph structure:
//
// article nodes contain a neo4j id and a uuid for locating the actual file.
// articles are related to metadata nodes through weighted edges. Articles
// are related to each other through metadata
//
// metadata nodes are labelled with the kind of metadata that they contain, eg "taxonomy"
// the data for a metadata node is stored in the Text member
//
// weighted, undirected edges connect articles to metadata. Now they
// reflect the relevance of a piece of metadata to an article. Edge weights are
// stored in the Relevance member
//
// a relation between two articles might look like (in psuedo neo4j querry language):
//
// (:Article "usa")-[2.0]       [5.0]-(:Article "#1 country")
//                       \     /
//                   (:Taxonomy "merica")
//                       /
// (:Article "bad")-[0.5]
//

import (
	"fmt"
	// use neoism
	"gopkg.in/jmcvetta/neoism.v1"
	"strings"
)

// DBKeyword comment
type DBKeyword struct {
	Relevance float32 `json:"Relevance"`
	Text      string  `json:"Text"`
}

// ArticleInfo comment
type ArticleInfo struct {
	// assumes that this is universally unique
	Identifier string `json:"n.Identifier"`
}

// IDError when bad id
type IDError struct {
	uuid    string
	message string
}

// Error from IDError
func (e *IDError) Error() string {
	return fmt.Sprintf("%s - %s", e.uuid, e.message)
}

// private database for all the requests
var db *neoism.Database

// Open a connection to the DB if one isn't already open
// you should turn off auth by settind dbms.security.auth_enabled = false
// in neo4j/data/dbms/auth
func Open(where string) error {
	if db != nil {
		return nil
	}

	tmp, err := neoism.Connect(where)
	if err != nil {
		return err
	}

	db = tmp
	return nil
}

// Close the db
func Close() error {
	return nil
}

// Store an article in the DB
// articleID should be a uuid for the article
// goes and checks that this is actually a uuid to prevent double inserts
func Store(articleID string) error {
	if info, err := GetByUUID(articleID); err != nil {
		return fmt.Errorf("bad uuid: %s", err.Error())
	} else if info.Identifier != "" {
		return fmt.Errorf("uuid not unique")
	}

	cq := neoism.CypherQuery{
		Statement:  `create (:Article {Identifier:{Identifier}})`,
		Parameters: neoism.Props{"Identifier": articleID},
		Result:     nil,
	}
	return db.Cypher(&cq)
}

// GetByUUID gets an article by its uuid
// ArticleInfo empty if no such article exists
func GetByUUID(articleID string) (ArticleInfo, error) {
	result := []ArticleInfo{}

	cq := neoism.CypherQuery{
		Statement:  `match (n {Identifier: {Identifier} }) return n, n.Identifier`,
		Parameters: neoism.Props{"Identifier": articleID},
		Result:     &result,
	}

	err := db.Cypher(&cq)
	if err != nil {
		return ArticleInfo{}, err
	}

	if len(result) > 1 {
		return result[0], fmt.Errorf("too many articles returned!\n")
	}
	if len(result) > 0 {
		return result[0], nil
	}

	// nothing
	return ArticleInfo{}, nil

}

// StrengthBetween find how closely two nodes are related by some label
// finds all paths between and sums up the output
// returns "flow" or sum of relevances between and how many common keyword or edges
// they share
func StrengthBetween(startID string, endID string, label string) (float32, int, error) {
	result := []struct {
		Score float32 `json:"total"`
		Count int     `json:"count"`
	}{}

	statementStr := `
	match (start:Article {Identifier: {startID}}),(end:Article {Identifier: {endID}}) 
	match p = (start)-[rel_s]-(mid:MetadataType)-[rel_e]-(end) with collect(p) as paths

	return reduce(o_s = 0, path in paths 
	| o_s + reduce(s = 0, rel in relationships(path) | s + rel.Relevance)) as total, length(paths) as count
						`
	cq := neoism.CypherQuery{
		Statement:  fixLabel(statementStr, label),
		Parameters: neoism.Props{"startID": startID, "endID": endID, "label": label},
		Result:     &result,
	}

	err := db.Cypher(&cq)
	if err != nil {
		return 0, 0, err
	}
	if len(result) != 1 {
		return 0, 0, fmt.Errorf("result is too long")
	}

	return result[0].Score, result[0].Count, err
}

// InsertRelations inserts an array of relations named by keyword
// assumes that values has Text, Relevance
func InsertRelations(articleID string, keyword string, values interface{}) error {

	statementStr := `
			match (start:Article {Identifier: {articleID}})
			unwind {relations} as relations
			foreach (relation in relations | 
			merge (end:MetadataType {Text: relation.Text}) 
			create unique (start)-[:Relation {Relevance: relation.Relevance}]->(end)
			)
	`
	cq := neoism.CypherQuery{
		Statement:  fixLabel(statementStr, keyword),
		Parameters: neoism.Props{"articleID": articleID, "Leyword": keyword, "relations": values},
	}

	err := db.Cypher(&cq)
	return err
}

func fixLabel(statement string, label string) string {
	return strings.Replace(statement, "MetadataType", label, 1)
}

// clear deletes all nodes from teh db, used most for testing
func clear() error {
	cq := neoism.CypherQuery{
		Statement: `
		match (node) optional match (node)-[edge]-() 
		delete node, edge`,
	}

	return db.Cypher(&cq)
}

// GetAll returns all the articles
func GetAll() ([]string, error) {
	result := []struct {
		Article string `json:"article"`
	}{}

	statementStr := `
		match (start:Article) return start.Identifier as article 
		`
	cq := neoism.CypherQuery{
		Statement: statementStr,
		Result:    &result,
	}

	err := db.Cypher(&cq)
	if err != nil {
		return nil, err
	}

	ret := make([]string, len(result))
	for i := range result {
		ret[i] = result[i].Article
	}

	return ret, nil

}

// GetRelations gets the metadata types comming out of an article.
func GetRelations(article string, metadataType string, thresh float64) ([]string, error) {
	result := []struct {
		Identifier string `json:"metadata"`
	}{}

	statementStr := `
		match (start:Article)-[r]-(key)
		where r.Relevance > {thresh} and start.Identifier={article} 
		return key.Text as metadata 
		`
	cq := neoism.CypherQuery{
		Statement:  statementStr,
		Parameters: neoism.Props{"thresh": thresh, "article": article},
		Result:     &result,
	}

	err := db.Cypher(&cq)
	if err != nil {
		return nil, err
	}

	ret := make([]string, len(result))
	for i := range result {
		ret[i] = result[i].Identifier
	}

	return ret, nil

}

// GetRelationsInv gets articles related to some metadata item.
func GetRelationsInv(keyword string, metadataType string, thresh float64) ([]string, error) {
	result := []struct {
		Identifier string `json:"Identifier"`
	}{}

	statementStr := `
		match (start:Article)-[r]-(key)
		where r.Relevance > {thresh} and key.Text={text} 
		return start.Identifier as Identifier
		`
	cq := neoism.CypherQuery{
		Statement:  statementStr,
		Parameters: neoism.Props{"thresh": thresh, "text": keyword},
		Result:     &result,
	}

	err := db.Cypher(&cq)
	if err != nil {
		return nil, err
	}

	ret := make([]string, len(result))
	for i := range result {
		ret[i] = result[i].Identifier
	}

	return ret, nil
}
