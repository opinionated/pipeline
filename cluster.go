package pipeline

import (
	"fmt"
	"github.com/biogo/cluster/meanshift"
	"github.com/opinionated/word2vec"
	"math"
)

const (
	// mainArticleID id
	mainArticleID = 1

	// relatedArticleID id
	relatedArticleID = 2
)

// Feature a
type Feature struct {
	id        string
	data      []float64
	relevance float32

	// which article it belongs to
	which int
}

// Features of the data
type Features []Feature

// Len as
func (f Features) Len() int {
	return len(f)
}

// Values a
func (f Features) Values(i int) []float64 {
	return []float64(f[i].data)
}

// Weight a
func (f Features) Weight(i int) float64 {
	return float64(f[i].relevance)
}

func printVec(vec []float64) {
	for i := range vec {
		fmt.Printf("%6.3f, ", vec[i])
		if i%10 == 0 {
			fmt.Println("")
		}
		if i%100 == 0 {
			fmt.Println("")
		}
	}
}

// sum two vectors
func addVecs(a, b word2vec.Vector) []float64 {
	result := make([]float64, len(a))
	for i := range a {
		result[i] = float64(a[i] - b[i])
	}
	return result
}

// calculate teh manhattan distance (L1) between two vectors
func manDist(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += math.Abs(a[i] - b[i])
	}
	return sum

}

// square all the elements in the vector
func squareVec(a word2vec.Vector) []float64 {
	result := make([]float64, len(a))
	for i := range a {
		result[i] = float64(a[i] * a[i])
	}
	return result
}

// cast a vector ([]float32) to []float64
func doCastVec(a word2vec.Vector) []float64 {
	result := make([]float64, len(a))
	for i := range a {
		result[i] = float64(a[i])
	}
	return result
}

// takes in a features vec and converts to a sort of covariance thing
func toDistVec(main Features) Features {

	ret := make(Features, len(main))

	for i := range main {
		// slot for each related
		slot := make([]float64, len(main))
		for j := range main {
			slot[j] = dotVecs(main[i].data, main[j].data)
			if i == j {
				slot[j] = 0.5
			}
		}

		ret[i] = Feature{id: main[i].id, data: slot, relevance: main[i].relevance, which: main[i].which}
	}

	return ret
}

func buildFeatureArray(words map[string]word2vec.Vector, relevances map[string]float32, which int) (Features, error) {
	ret := make(Features, len(words))
	idx := 0
	for key, vec := range words {
		vec.Normalise()
		castVec := doCastVec(vec)

		ret[idx] = Feature{id: key, data: castVec, relevance: relevances[key], which: which}
		idx++
	}

	return ret, nil
}

func dotVecs(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

// ClusterOverlap for the two articles.
// Care about the # of overlapping clusers and the "strength" of the overlapping clusters.
// Calc strength based on relevance of the words in the cluster and the properties of the cluster its self.
func ClusterOverlap(main, related map[string]word2vec.Vector, mainRelevance, relatedRelevance map[string]float32) (float32, int) {
	mainVecs, _ := buildFeatureArray(main, mainRelevance, mainArticleID)
	relatedVecs, _ := buildFeatureArray(related, relatedRelevance, relatedArticleID)

	allVecs := make(Features, len(mainVecs)+len(relatedVecs))
	var totalMainRel float32
	for i := range mainVecs {
		allVecs[i] = mainVecs[i]
		totalMainRel += mainVecs[i].relevance
	}

	var totalRelatedRel float32
	for i := range relatedVecs {
		allVecs[i+len(mainVecs)] = relatedVecs[i]
		totalRelatedRel += relatedVecs[i].relevance
	}

	// need to have some kind of total relevance
	if totalRelatedRel < 0.0001 || totalMainRel < 0.0001 {
		return 0.0, 0
	}

	// TODO: look into adaptive bandwidth stuff
	features := allVecs
	shifter := meanshift.NewTruncGauss(0.60, 2.6010)
	clusterer := meanshift.New(features, shifter, 0.01, 10)
	err := clusterer.Cluster()

	if err != nil {
		fmt.Println("err:", err)
		return 0.0, 0
	}

	numOverlaps := 0
	var score float32

	for _, c := range clusterer.Centers() {

		numMains := 0
		numRels := 0
		var mainQuality float32
		var relatedQuality float32

		for _, i := range c.Members() {
			f := features[i]

			if f.which == mainArticleID {
				numMains++
				mainQuality += f.relevance
			} else {
				numRels++
				relatedQuality += f.relevance
			}
		}

		// found at least one of each article in the cluster
		if numMains > 0 && numRels > 0 {
			numOverlaps++

			// how much of each story this cluster "captures"
			mainSignificance := mainQuality / totalMainRel
			relSignificance := relatedQuality / totalRelatedRel
			if totalRelatedRel < 0.001 || totalMainRel < 0.001 {
				panic("sig too low!!!")
			}

			// find cluster strength by doing a cluster covariance
			// clusterstrength is always [0:1]
			var clusterStrength float32
			for _, i := range c.Members() {
				f := features[i]
				var dsum float32
				for _, j := range c.Members() {
					if i == j {
						continue
					}
					ff := features[j]
					dsum += float32(dotVecs(f.data, ff.data))
				}
				clusterStrength += dsum
			}

			if clusterStrength < 0 {
				panic("oh nose!!! expected to only have pos vals")
			}

			// denom is num itrs run, sub len c b/c we don't mul by ourselves
			// len(c) can't be one, so denom calc is OK
			denom := float32(len(c.Members())*len(c.Members()) - len(c.Members()))
			clusterStrength = float32(math.Sqrt(float64(clusterStrength / denom)))

			// link from a => b = %rel(a) * avg rel(b)
			relMain := relSignificance * (mainQuality / float32(numMains))
			mainRel := mainSignificance * (relatedQuality / float32(numRels))
			score += (relMain + mainRel) * clusterStrength
		}
	}

	return score, numOverlaps
}
