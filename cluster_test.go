package pipeline

import (
	"github.com/biogo/cluster/meanshift"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
type Feature struct {
	id   string
	data []float64
}

// Len of the set
func (f Feature) Len() int {
	return len(f.data)
}

// Get the ith
func (f Feature) Get(i int) float32 {
	return f.data[i]
}

type Features []Feature

func (f Features) Len() int {
	return len(f)
}

func (f Features) Values(i int) []float64 {
	return []float64(f[i].data)
}

*/
func TestClusterSimple(t *testing.T) {
	t.Log("hey hey")

	// first 3 group 1
	// 2nd 3 group 2
	// 3rd 4 outliers
	vectors := [][]float64{{1.0, 2.0}, {1.1, 1.2}, {1.3, 1.1},
		{-1.0, 0.9}, {-0.8, 1.1}, {-1.1, 1.0},
		{0.0, 0.0}, {1.0, -1.0}, {-1.0, -1.0}}

	// build data
	features := make(Features, len(vectors))
	for i := 0; i < len(vectors); i++ {
		features[i] = Feature{id: "0", data: vectors[i]}
	}

	t.Log(vectors)

	ms := meanshift.New(features, meanshift.NewTruncGauss(.60, 3), 0.1, 5)
	err := ms.Cluster()

	assert.Nil(t, err)

	for _, c := range ms.Centers() {
		t.Log("new group")
		for _, i := range c.Members() {
			f := features[i]
			t.Log(f)
		}
	}
}
