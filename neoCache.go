package pipeline

import (
	"bytes"
)

// Neo4jCache implements a cache for neo4j requests. Its key
// setup makes (a,b) == (b,a). The cache is (naively) implemented
// as a hash map.
// TODO: determine if this is useful outside of the taxonomy
// TODO: make some form of cache expiration so that we don't end up
//	caching the whole graph
type Neo4jCache struct {
	cache map[string]float64
}

// Setup creates an empty cache.
func (n *Neo4jCache) Setup() {
	n.cache = make(map[string]float64)
}

// Clear empties the cache by overwriting it with an empty cache.
func (n *Neo4jCache) Clear() {
	n.Setup()
}

// Get gets a cached value. The boolean return value is true
// if the value is in the cache.
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

// Add puts a value in the cache, overwriting the previous value if
// it exists.
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
