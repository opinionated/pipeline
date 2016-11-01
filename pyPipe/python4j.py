from py2neo import Graph, Path, Node, Relationship
from collections import namedtuple

#private database for all the requests
db = None

# Open a connection to the database if one doesn't already exist
def dbOpen(uri, pw):
	global db
	if(db is None):
		db = Graph(uri, password=pw)
		print(db)
		return None
	else:	return None


# Close the database
# Note: The go implementation only returns nil and there is no function to close the database
# This is only here to mirror the previous implementation, I don't think a close is necessary
def dbClose():
	return None
	
# getByUUID gets an article by its UUID
# Will return full list of results, return type is list of dict
# Should only be of length 1, if length > 1 then there is a non-unique UUID.
# returns None if no such article exists
def getByUUID(articleID):
	result = db.data(
				statement="MATCH (n {Identifier: {Identifier} }) RETURN n, n.Identifier",
				parameters={"Identifier": articleID}
			 )
	if len(result) > 0 :
		return result
	else:
		print "Nothing found by that UUID!"
		return None
	
# Stores an article in the database
# articleID should be a UUID for the article
# Will verify that articleID is actually a UUID
def store(articleID):
	result = getByUUID(articleID)
	if(result is not None):
		print("UUID not unique!")
		return
	newArticle = Node("Article", Identifier=articleID)
	db.create(newArticle)
	#print '{0} is currently in the database: {1}'.format(articleID, db.exists(newArticle))
	
# insertRelations takes in an array of {String, double} pairs that represent a relation (Text) and its strength (Relevance).
# These relations are then added to the corresponding article in the database
# assumes that there are values for all key/value pairs
def insertRelations(articleID, keyword, values):

	statement = "match (start:Article {Identifier: {articleID}}) "\
				"unwind {relations} as relations "\
				"foreach (relation in relations | "\
				"merge (end:" + keyword + " {Text: relation.Text}) "\
				"create unique (start)-[:Relation {Relevance: relation.Relevance}]->(end)"
				
	
	db.data(statement, parameters={"articleID": articleID, "keyword": keyword, "relations": values})
	
	
#####################################################################
#test code goes here
#####################################################################
dbOpen("http://localhost:7474", "root")

TextRelevance = namedtuple("TextRelevance", "Text Relevance")

values = [TextRelevance("Obama", 2)]

result = insertRelations("Megasatan", "keywords", values)

print result

#graph.run("UNWIND range(1, 10) AS n RETURN n, n * n as n_sq").dump()

#print(graph.data("MATCH (a:Person) RETURN a.name LIMIT 4"))