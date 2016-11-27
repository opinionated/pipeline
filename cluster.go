package pipeline

import (
	"fmt"
	"github.com/biogo/cluster/cluster"
	"github.com/biogo/cluster/kmeans"
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
	id   string
	data []float64

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

func buildFeatureArray(words map[string]word2vec.Vector, which int) (Features, error) {
	ret := make(Features, len(words))
	idx := 0
	for key, vec := range words {
		vec.Normalise()
		castVec := make([]float64, len(vec))
		for i := range vec {
			castVec[i] = float64(vec[i])
		}
		//castVec := squareVec(vec)

		ret[idx] = Feature{id: key, data: castVec, which: which}
		idx++
	}

	return ret, nil
}

func doKMean(words map[string]word2vec.Vector) {
	fmt.Println("===========================")
	features, _ := buildFeatureArray(words, 0)
	ms, _ := kmeans.New(features)
	ms.Seed(2)
	err := ms.Cluster()
	if err != nil {
		panic(err)
	}

	for _, c := range ms.Centers() {
		fmt.Println("")
		for _, i := range c.Members() {
			f := features[i]
			fmt.Println(f.id)

			if len(c.Members()) == 1 {
				//printVec(f.data)
			}
		}
	}
}

func doCluster(words map[string]word2vec.Vector, shifter meanshift.Shifter) {
	fmt.Println("===========================")
	features, _ := buildFeatureArray(words, 0)
	ms := meanshift.New(features, shifter, 0.2, 25)
	err := ms.Cluster()
	if err != nil {
		panic(err)
	}

	var mainc cluster.Center
	for _, c := range ms.Centers() {
		fmt.Println("")
		for _, i := range c.Members() {
			f := features[i]
			fmt.Println(f.which, f.id)

			if len(c.Members()) == 1 {
				//fmt.Println(c.V())
			} else {
				mainc = c
			}
		}
	}

	fmt.Println("going for dist")
	centers := ms.Centers()
	for i := range centers {
		if len(centers[i].Members()) == 1 {
			idx := centers[i].Members()[0]
			mem := features[idx]
			fmt.Println(mem.id, "to main:", manDist(centers[i].V(), mainc.V()))
			for j := i + 1; j < len(centers); j++ {
				jfIdx := centers[j].Members()[0]
				jFeature := features[jfIdx]
				centerDist := manDist(centers[i].V(), centers[j].V())
				fmt.Println(mem.id, "to", jFeature.id, ":", centerDist)
			}
		}
	}
}

func doClusterOverlap(features Features, shifter meanshift.Shifter) {
	fmt.Println("===========================")
	ms := meanshift.New(features, shifter, 0.01, 10)
	err := ms.Cluster()
	if err != nil {
		panic(err)
	}

	for _, c := range ms.Centers() {
		fmt.Println("")
		for _, i := range c.Members() {
			f := features[i]
			fmt.Println(f.which, f.id)

			if len(c.Members()) == 1 {
				//fmt.Println(c.V())
			}
		}
	}
}

// ClusterOverlap for the two articles
func ClusterOverlap(main, related map[string]word2vec.Vector) {
	mainVecs, _ := buildFeatureArray(main, mainArticleID)
	relatedVecs, _ := buildFeatureArray(related, relatedArticleID)

	allVecs := make(Features, len(mainVecs)+len(relatedVecs))
	for i := range mainVecs {
		allVecs[i] = mainVecs[i]
	}

	for i := range relatedVecs {
		allVecs[i+len(mainVecs)] = relatedVecs[i]
	}

	//doClusterOverlap(allVecs, meanshift.NewUniform(1.15))
	doClusterOverlap(allVecs, meanshift.NewTruncGauss(0.65, 3.0))

	dot(main, main, "Psychiatry", "Psychology")
	dot(main, main, "Psychiatry", "Mental_health")
	dot(main, main, "Psychology", "Mental_health")

}

func dot(main, related map[string]word2vec.Vector, a, b string) {
	fmt.Println(a, "dot", b, "=", main[a].Dot(related[b]))
}

// Cluster similar words
func Cluster(words map[string]word2vec.Vector) {
	fmt.Println("\n\n*******************************************\n******************************************\narticle:")
	//ushifter = []meanshift.Shifter{meanshift.NewUniform(1.0), meanshift.NewUniform(1.5), meanshift.NewUniform(2.0)}
	//gshifter = []dd
	//fmt.Println("\n*******median:")
	//doCluster(words, meanshift.NewUniform(0.57))
	//doCluster(words, meanshift.NewUniform(1.06))
	doKMean(words)
	//doCluster(words, meanshift.NewUniform(6.2))

	doCluster(words, meanshift.NewTruncGauss(0.60, 2.5))
	//doCluster(words, meanshift.NewTruncGauss(1.20, 2.30))

	rob := words["Psychiatry"]
	rob = words[""]
	//roy := words["Roy_Blunt"]
	roy := words["Joe_Manchin"]
	rob.Normalise()
	roy.Normalise()
	roy.Dot(rob)
	fmt.Println("dot:", roy.Dot(rob))

	a := words["Maine"]
	a.Normalise()
	b := words["Ohio"]
	b.Normalise()
	fmt.Println("dot:", a.Dot(b))
	//doCluster(m, main, related, meanshift.NewUniform(3.0))
	//fmt.Println("\n*******gaus:")
	//doCluster(m, main, related, meanshift.NewTruncGauss(0.55, 4.0))
	//doCluster(m, main, related, meanshift.NewTruncGauss(0.6, 4.0))
	//doCluster(m, main, related, meanshift.NewTruncGauss(0.65, 4.0))
}
