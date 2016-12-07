package pipeline

import (
	"fmt"
	"github.com/biogo/cluster/cluster"
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

func addVecs(a, b word2vec.Vector) []float64 {
	result := make([]float64, len(a))
	//a.Normalise()
	//b.Normalise()
	for i := range a {
		result[i] = float64(a[i] - b[i])
	}
	return result
}

func manDist(a, b []float64) float64 {
	sum := 0.0

	for i := range a {
		sum += math.Abs(a[i] - b[i])
	}

	return sum

}
func squareVec(a word2vec.Vector) []float64 {
	result := make([]float64, len(a))
	//a.Normalise()
	for i := range a {
		result[i] = float64(a[i] * a[i])
		//result[i] = math.Sqrt(float64(a[i]))
	}
	return result
}

func doCastVec(a word2vec.Vector) []float64 {
	result := make([]float64, len(a))
	//a.Normalise()
	for i := range a {
		result[i] = float64(a[i])
		//result[i] = math.Sqrt(float64(a[i]))
	}
	return result
}

// takes in a features vec and converts to dist
func toDistVec(main Features) Features {

	ret := make(Features, len(main))

	for i := range main {
		// slot for each related
		slot := make([]float64, len(main))
		for j := range main {
			slot[j] = dotVecs(main[i].data, main[j].data)
			//slot[j] = slot[j] * slot[j]
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
		/*
			castVec := make([]float64, len(vec))
			for i := range vec {
				castVec[i] = float64(vec[i])
			}
		*/
		castVec := doCastVec(vec)
		//castVec := squareVec(vec)

		ret[idx] = Feature{id: key, data: castVec, relevance: relevances[key], which: which}
		idx++
	}

	return ret, nil
}

func doClusterOverlap(features Features, shifter meanshift.Shifter) (cluster.Clusterer, error) {
	//fmt.Println("===========================")
	ms := meanshift.New(features, shifter, 0.10, 10)
	//ms, _ := kmeans.New(features)
	/*
		if len(features) > 10 {
			ms.Seed(4)
		} else {
			ms.Seed(3)
		}
	*/
	err := ms.Cluster()
	return ms, err
}

func dotVecs(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

// ClusterOverlap for the two articles.
// care about the # of overlapping clusers and the "strength" of the overlapping clusters
// strength is how compact the cluster is
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

	if totalRelatedRel < 0.0001 || totalMainRel < 0.0001 {
		return 0.0, 0
	}

	//doClusterOverlap(allVecs, meanshift.NewUniform(1.15))

	//features := toDistVec(allVecs)
	features := allVecs
	shifter := meanshift.NewTruncGauss(0.80, 0.1)
	//shifter := meanshift.NewTruncGauss(0.15, 0.1)
	//shifter := meanshift.NewUniform(0.95)
	clusterer := meanshift.New(features, shifter, 0.01, 10)
	err := clusterer.Cluster()

	if err != nil {
		fmt.Println("err:", err)
		return 0.0, 0
	}

	numOverlaps := 0
	var score float32
	//	fmt.Println("==========================")

	withinArr := clusterer.Within()
	for whichCluster, c := range clusterer.Centers() {

		hasMain := false
		hasRel := false
		//		fmt.Println("")

		numMains := 0
		numRels := 0
		var mainQuality float32
		var relatedQuality float32
		var mainSquaredQuality float32
		var relatedSquaredQuality float32
		mainQuality = 0
		relatedQuality = 0
		for _, i := range c.Members() {
			f := features[i]
			//			fmt.Println(f.which, f.id)

			if f.which == mainArticleID {
				hasMain = true
				numMains++
				mainQuality += f.relevance
				mainSquaredQuality += f.relevance * f.relevance
			} else {
				hasRel = true
				numRels++
				relatedQuality += f.relevance
				relatedSquaredQuality += f.relevance * f.relevance
			}
		}

		if hasMain && hasRel {
			numOverlaps++

			// how much of each story this cluster "captures"
			mainSignificance := mainQuality / totalMainRel
			relSignificance := relatedQuality / totalRelatedRel
			if totalRelatedRel < 0.001 || totalMainRel < 0.001 {
				panic("sig too low!!!")
			}

			clusterStrength := float32(withinArr[whichCluster])
			if clusterStrength > 0.000000000001 {
				fmt.Println(clusterStrength, numMains, numRels, relSignificance)
			}

			// calc the dist b/w the vecs
			//score += mainSignificance + relSignificance
			//score += (mainSquaredQuality + relatedSquaredQuality) / float32(numRels)+numMains)
			//score += (relatedQuality / float32(numRels)) + (relatedQuality / float32(numMains))
			//score += ((relatedSquaredQuality / float32(numRels)) + (mainSquaredQuality / float32(numMains))) / 2
			//score += (1 * relSignificance * float32(numRels)) + (1 * mainSignificance * float32(numMains))
			score += ((0 * relSignificance * float32(1)) + (0*mainSignificance*float32(1))*clusterStrength*0) / 2.0
			//score += (1 * relSignificance * float32(relatedQuality)) + (1 * mainSignificance * float32(mainQuality))
			//score += float32(numRels) + float32(numMains)
			//score += relatedQuality + mainQuality
			// how relevant a cluster is to its artice * how strong the connections are
			score += (relSignificance*(mainQuality/float32(numMains)) + mainSignificance*(relatedQuality/float32(numRels))) * (1 - clusterStrength)
			//score += (relSignificance*(mainQuality/float32(numMains)) + mainSignificance*(relatedQuality/float32(numRels)))
			//score += mainSquaredQuality + relatedSquaredQuality

			/*
				if mainSignificance < relSignificance {
					score += relatedScore
				} else {
					score += mainScore
				}
			*/

			//score += float32(math.Max(float64(mainSignificance), float64(relSignificance)))
			//score += float32(len(c.Members()))
		}
	}

	return score, numOverlaps
}

func dot(main, related map[string]word2vec.Vector, a, b string) {
	/*
		av := main[a]
		bv := related[b]

		avc := doCastVec(av)
		bvc := doCastVec(bv)
	*/
	fmt.Println(a, "dot", b, "=", main[a].Dot(related[b])) //, dotVecs(avc, bvc))
}
