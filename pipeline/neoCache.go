package pipeline

import (
	"bytes"
)

// implements a two way cache for neo4j requests
// the biggest part about it is that it is bi-directional so order
// doesn't matter ie (a,b) == (b,a)
// TODO: determine if this is useful outside of the taxonomy
// TODO: make some form of cache experation so that we don't end up
// 	caching the whole graph
type Neo4jCache struct {
	cache map[string]float64
}

func (n *Neo4jCache) Setup() {
	n.cache = make(map[string]float64)
}

func (n *Neo4jCache) Clear() {
	n.Setup()
}

// checks if the keys are in the cache or not
func (n *Neo4jCache) Get(a, b string) (float64, bool) {
	var buffer bytes.Buffer
	if a < b {
		buffer.WriteString(a)
		buffer.WriteString(b)
	} else {
		buffer.WriteString(b)
		buffer.WriteString(a)
	}
	v, contains := n.cache[buffer.String()]
	return v, contains
}

// adds value, overwriting any currently in the cache
func (n *Neo4jCache) Add(a, b string, value float64) {
	var buffer bytes.Buffer
	if a < b {
		buffer.WriteString(a)
		buffer.WriteString(b)
	} else {
		buffer.WriteString(b)
		buffer.WriteString(a)
	}

	n.cache[buffer.String()] = value
}
