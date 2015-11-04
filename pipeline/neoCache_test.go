package pipeline_test

import (
	"github.com/opinionated/pipeline/pipeline"
	"testing"
)

func TestGetHash(t *testing.T) {
	a := "one"
	b := "two"

	cache := pipeline.Neo4jCache{}
	cache.Setup()

	if _, has := cache.Get(a, b); has {
		t.Errorf("got something out of empty cache")
	}

	cache.Add(a, b, 1.0)
	cache.Add(a, a, 2.0)
	cache.Add(b, b, 3.0)

	if _, has := cache.Get(b, a); !has {
		t.Errorf("cache failed to bidirectionalize")
	}

	cache.Add(b, a, 4.0)
	if v, _ := cache.Get(a, b); v != 4.0 {
		t.Errorf("cache did not go bidirectional!")
	}

}
