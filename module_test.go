package pipeline_test

import (
	"container/heap"
	"fmt"
	"github.com/opinionated/pipeline"
	"github.com/opinionated/pipeline/analyzer/dbInterface"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testSet struct {
	mainArticle      string
	relatedArticles  []string
	expectedArticles []string
}

var terrorTestSet = testSet{
	mainArticle: "Ted ‘Carpet-Bomb’ Cruz",
	relatedArticles: []string{
		"The Horror in San Bernardino",
		"A BAD CALL ON THE BERGDAHL COURT-MARTIAL",
		"A BETTER SAFEGUARD AGAINST THREATS FROM ABROAD",
		"A COLLEGE EDUCATION FOR PRISONERS",
		"A CONSTITUTIONAL STANDOFF IN VENEZUELA",
		"A Fearful Congress Sits Out the War Against ISIS",
		"Agony and Starvation in the Syrian War",
		"A Maid’s Peaceful Rebellion in Colombia",
		"America and Its Fellow Executioners",
		"America’s Empty Embassies",
		"A Misguided Plan for Carriage Horses",
		"An Important Win in the Supreme Court for Class Actions",
		"An Opening for States to Restrict Guns",
		"A Pause to Weigh Risks of Gene Editing",
		"A Safer World, Thanks to the Iran Deal ",
		"A Shameful Round-Up of Refugees",
		"Canada’s Warm Embrace of Refugees",
		"Connecticut’s Second-Chance Society",
		"Despair Over Gun Deaths Is Not an Option",
		"Doubts About Saudi Arabia’s Antiterrorism Coalition",
		"Extradite El Chapo Guzmán",
		"France’s Diminished Liberties",
		"France’s State of Emergency",
		"Gov. Cuomo’s Burden on Ethics",
		"Gunmakers’ War Profiteering on the Home Front",
		"Guns, Anger and Nonsense in Oregon",
		"How to Help the Syrians Who Want to Return Home",
		"India and Pakistan Try Again",
		"In France, the Political Fruits of Fear",
		"In Venezuela, a Triumph for the Opposition",
		"Iran’s Hard-Liners Cling to the Past",
		"Iran’s Other Scary Weapons Program",
		"Iraq and the Kurds Are Going Broke",
		"Is Warfare in Our Bones?",
		"Justice Antonin Scalia’s Supreme Court Legacy",
		"Keeping the Lights On During a Dark Time",
		"Missteps in Europe’s Online Privacy Bill",
		"Mr. Obama’s Wise Call on a Prisoner Swap",
		"New Tensions Over the Iran Nuclear Deal",
		"New York City Policing, by the Numbers",
		"Poland Deviates From Democracy",
		"President Obama’s Call to America’s Better Nature",
		"President Obama’s Tough, Calming Talk on Terrorism",
		"Saudi Arabia’s Barbaric Executions",
		"Saudi Arabia’s Execution Spree",
		"Thailand's Fear of Free Speech",
		"The Importance of Retaking Ramadi",
		"The Paris Climate Pact Will Need Strong Follow-Up",
		"The Pentagon’s Insubordination on Guantánamo",
		"The Supreme Court, the Nativists and Immigrants",
		"The Tarnished Trump Brand",
		"The Trump Effect, and How It Spreads",
		"The Unfair Treatment of Ebola Workers",
		"The Urgent Need for Peace in Yemen",
		"Toward a Stronger European Border",
		"Two Sides of Ted Cruz: Tort Reformer and Personal Injury Lawyer",
		"What France's Vote Means",
		"What It Will Take to Bankrupt ISIS ",
		"What Narendra Modi Can Do in Paris",
		"What Went Wrong With Navy SEALs",
	},
}
var neoTestSet = testSet{
	mainArticle: "A College Education for Prisoners",
	relatedArticles: []string{
		"Gov. Christie Leaves Gun Controls Behind in New Jersey",
		"A Bad Call on the Bergdahl Court-Martial",
		"Agony and Starvation in the Syrian War",
		"America’s Empty Embassies",
		"An Appalling Silence on Gun Control",
		"A New Cuban Exodus",
		"A Pause to Weigh Risks of Gene Editing",
		"A Shameful Round-Up of Refugees",
		"At the Supreme Court, a Big Threat to Unions",
		"Candidates’ Children in the Peanut Gallery",
		"Connecticut’s Second-Chance Society",
		"Course Correction for School Testing",
		"Donald Trump Drags Bill Clinton’s Baggage Out",
		"Don’t Change the Legal Rule on Intent",
		"Extradite El Chapo Guzmán",
		"For Grieving Families, Each Gun Massacre Echoes the Last",
		"France’s Diminished Liberties",
		"France’s State of Emergency",
		"Getting Rid of Big Currency Notes Could Help Fight Crime",
		"Gov. Cuomo’s Push on Justice Reform",
		"Guns, Anger and Nonsense in Oregon",
		"Hillary Clinton Should Just Say Yes to a $15 Minimum Wage",
		"Horror Stories From New York State Prisons",
		"Iraq and the Kurds Are Going Broke",
		"Is Warfare in Our Bones?",
		"Justice Antonin Scalia’s Supreme Court Legacy",
		"Keep Guns Away From Abusers",
		"Keeping the Lights On During a Dark Time",
		"Kentucky’s Bizarre Attack on Health Reform",
		"Making Choices in Iowa",
		"Michigan’s Failure to Protect Flint",
		"New Minimum Wages in the New Year",
		"New Tensions Over the Iran Nuclear Deal",
		"New York City Policing, by the Numbers",
		"New York’s Humane Retreat From Solitary Confinement",
		"New York’s ID Card Deserves Respect",
		"No Justification for High Drug Prices",
		"Pass Sentencing Reform",
		"Put Reforms Into State Prison Guards’ Contract",
		"The Counterfeit High School Diploma",
	},
}

func storyFromSet(set testSet) pipeline.Story {

	story := pipeline.Story{}
	story.MainArticle = pipeline.NewArticle(set.mainArticle)
	story.RelatedArticles = make(chan pipeline.Article)

	go func() {

		for i := range set.relatedArticles {
			story.RelatedArticles <- pipeline.NewArticle(set.relatedArticles[i])
		}

		close(story.RelatedArticles)

	}()

	return story
}

// manages running a story
func storyDriver(
	pipe *pipeline.Pipeline,
	story pipeline.Story) ([]pipeline.Article, error) {

	pipe.Start()
	pipe.PushStory(story)

	var result pipeline.Story

	select {
	case result = <-pipe.GetOutput():
		break

	case <-pipe.Error():
		// go get the error when you actually close the pipe
		err := pipe.Close()
		return nil, err
	}

	related := make([]pipeline.Article, 0)

	for {
		select {
		case analyzed, open := <-result.RelatedArticles:
			if !open {
				return related, pipe.Close()
			}
			related = append(related, analyzed)

		case <-pipe.Error():
			// get the error when you close the pipe
			err := pipe.Close()
			return nil, err
		}
	}
}

type articleScorePair struct {
	score   float32
	article pipeline.Article
}

// min heap, lowest value on end
type articleScoreHeap []articleScorePair

func (h articleScoreHeap) Len() int {
	return len(h)
}
func (h articleScoreHeap) Less(i, j int) bool {
	return h[i].score < h[j].score
}

func (h articleScoreHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *articleScoreHeap) Push(x interface{}) {
	item := x.(articleScorePair)
	*h = append(*h, item)
}

func (h *articleScoreHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (h articleScoreHeap) Min() float32 {
	return h[0].score
}

var _ heap.Interface = (*articleScoreHeap)(nil)

// only gets the top #num articles from the articles list
func heapFilter(articles []pipeline.Article,
	scoreFuncs map[string]func(pipeline.Score) float32,
	weightMap map[string]float32, num int) []pipeline.Article {

	mheap := &articleScoreHeap{}
	if mheap == nil {
		panic("oh nose, nil heap!")
	}

	fmt.Println("made it!")
	heap.Init(mheap)

	for _, article := range articles {
		score := scoreArticle(&article, scoreFuncs, weightMap)

		// need to make sure the size stays the same

		//mheap.Push(articleScorePair{score, article})
		heap.Push(mheap, articleScorePair{score, article})
		if mheap.Len() > num {
			// going onto the heap
			//mheap.Pop()
			heap.Pop(mheap)
		}

	}

	fmt.Println("through loop!")
	ret := make([]pipeline.Article, 0, num)
	for mheap.Len() > 0 {
		item := heap.Pop(mheap).(articleScorePair)
		fmt.Println(item.article.Name(), item.score)
		ret = append(ret, item.article)
	}
	return ret
}

func ScoreAverage(neo pipeline.Score) float32 {
	score, ok := neo.(pipeline.NeoScore)
	if !ok {
		panic("failed to convert neo score!")
	}

	if score.Count > 0 {
		return score.Flow * float32(score.Count)
	}
	return 0
}

func SquareFlow(neo pipeline.Score) float32 {
	score, ok := neo.(pipeline.NeoScore)
	if !ok {
		panic("failed to convert neo score!")
	}

	//return score.Flow * score.Flow * float32(score.Count)
	if score.Count > 0 {
		//val := score.Flow / float32(score.Count)
		val := score.Flow
		return val
	}
	return 0
}

func SquareCount(neo pipeline.Score) float32 {
	score, ok := neo.(pipeline.NeoScore)
	if !ok {
		panic("failed to convert neo score!")
	}

	return score.Flow * float32(score.Count)
}

func IDFAverage(idf pipeline.Score) float32 {
	score, ok := idf.(pipeline.IDFScore)
	if !ok {
		panic("failed to convert neo score!")
	}

	if len(score.Counts) == 0 {
		return 0.0
	}

	var sum float32
	for i := range score.Counts {
		sum += 1.0 / float32(score.Counts[i]*score.Counts[i])
	}

	if sum == 0 {
		return 0.0
	}

	return sum
}

func smoothScoreArticle(article *pipeline.Article, funcs map[string]func(pipeline.Score) float32, weights map[string]float32) float32 {
	keys := article.Keys()

	var articleScore float32
	for _, key := range keys {
		scoreFunc, fok := funcs[key]
		weight, wok := weights[key]
		if !fok || !wok {
			panic("key " + key + "is not in funcs")
		}

		score, _ := article.GetScore(key)
		articleScore += scoreFunc(score) * weight
	}

	return articleScore
}

func scoreArticle(article *pipeline.Article, funcs map[string]func(pipeline.Score) float32, weights map[string]float32) float32 {
	keys := article.Keys()

	var articleScore float32
	for _, key := range keys {
		scoreFunc, fok := funcs[key]
		weight, wok := weights[key]
		if !fok || !wok {
			panic("key " + key + "is not in funcs")
		}

		score, _ := article.GetScore(key)
		articleScore += scoreFunc(score) * weight
	}

	return articleScore
}

type threshAnalyzer struct {
	threshhold float32
	analyzers  map[string]func(pipeline.Score) float32
	weights    map[string]float32
}

func (ta threshAnalyzer) Setup() error { return nil }

func (ta threshAnalyzer) Analyze(main pipeline.Article,
	related *pipeline.Article) (bool, error) {

	score := scoreArticle(related, ta.analyzers, ta.weights)
	if score > ta.threshhold {
		return true, nil
	}

	return false, nil
}

func TestFull(t *testing.T) {

	taxFunc := pipeline.NeoAnalyzer{MetadataType: "Taxonomy"}
	taxModule := pipeline.StandardModule{}
	taxModule.SetFuncs(taxFunc)

	conceptsFunc := pipeline.NeoAnalyzer{MetadataType: "Concept"}
	conceptsModule := pipeline.StandardModule{}
	conceptsModule.SetFuncs(conceptsFunc)

	keyFunc := pipeline.NeoAnalyzer{MetadataType: "Keyword"}
	keyModule := pipeline.StandardModule{}
	keyModule.SetFuncs(&keyFunc)

	entityFunc := pipeline.NeoAnalyzer{MetadataType: "Entity"}
	entityModule := pipeline.StandardModule{}
	entityModule.SetFuncs(&entityFunc)

	// idf funcs
	keyIDFFunc := pipeline.IDFAnalyzer{MetadataType: "Keyword"}
	keyIDFModule := pipeline.StandardModule{}
	keyIDFModule.SetFuncs(&keyIDFFunc)

	entityIDFFunc := pipeline.IDFAnalyzer{MetadataType: "Entity"}
	entityIDFModule := pipeline.StandardModule{}
	entityIDFModule.SetFuncs(&entityIDFFunc)

	conceptIDFFunc := pipeline.IDFAnalyzer{MetadataType: "Concept"}
	conceptIDFModule := pipeline.StandardModule{}
	conceptIDFModule.SetFuncs(&conceptIDFFunc)

	// word2vec
	entityWVFunc := pipeline.WordVecAnalyzer{MetadataType: "Entity"}
	entityWVModule := pipeline.StandardModule{}
	entityWVModule.SetFuncs(&entityWVFunc)

	conceptWVFunc := pipeline.WordVecAnalyzer{MetadataType: "Concept"}
	conceptWVModule := pipeline.StandardModule{}
	conceptWVModule.SetFuncs(&conceptWVFunc)

	keyWVFunc := pipeline.WordVecAnalyzer{MetadataType: "Keyword"}
	keyWVModule := pipeline.StandardModule{}
	keyWVModule.SetFuncs(&keyWVFunc)

	scoreFuncs := make(map[string]func(pipeline.Score) float32)
	scoreFuncs["neo_Taxonomy"] = SquareCount //SquareFlow
	scoreFuncs["neo_Concept"] = SquareCount
	scoreFuncs["neo_Keyword"] = ScoreAverage
	scoreFuncs["neo_Entity"] = ScoreAverage
	scoreFuncs["idf_Keyword"] = IDFAverage
	scoreFuncs["idf_Entity"] = IDFAverage
	scoreFuncs["idf_Concept"] = IDFAverage
	scoreFuncs["wordvec_Concept"] = SquareFlow
	scoreFuncs["wordvec_Keyword"] = SquareFlow
	scoreFuncs["wordvec_Entity"] = SquareFlow

	weightMap := make(map[string]float32)
	weightMap["neo_Taxonomy"] = 3.0
	weightMap["neo_Concept"] = 3.0
	weightMap["neo_Keyword"] = 3.0
	weightMap["neo_Entity"] = 3.0
	weightMap["idf_Keyword"] = 10.0
	weightMap["idf_Entity"] = 10.0
	weightMap["idf_Concept"] = 10.0
	weightMap["wordvec_Taxonomy"] = 10.0
	weightMap["wordvec_Concept"] = 15.0
	weightMap["wordvec_Keyword"] = 10.0
	weightMap["wordvec_Entity"] = 10.0

	threshFunc := threshAnalyzer{0.0, scoreFuncs, weightMap}
	threshModule := pipeline.StandardModule{}
	threshModule.SetFuncs(threshFunc)

	lastThreshFunc := threshAnalyzer{0.0, scoreFuncs, weightMap}
	lastThreshModule := pipeline.StandardModule{}
	lastThreshModule.SetFuncs(lastThreshFunc)

	// build the pipe
	pipe := pipeline.NewPipeline()

	// 1.1 seems to do it for words

	// do coarse methods
	//	pipe.AddStage(&taxModule)
	//pipe.AddStage(&conceptsModule)
	//pipe.AddStage(&keyIDFModule)
	//pipe.AddStage(&entityIDFModule)
	//pipe.AddStage(&conceptIDFModule)
	//pipe.AddStage(&threshModule)
	pipe.AddStage(&entityWVModule)
	pipe.AddStage(&conceptWVModule)
	//pipe.AddStage(&lastThreshModule)
	pipe.AddStage(&keyWVModule)

	// thresh then do finer methods
	//pipe.AddStage(&keyModule)
	//pipe.AddStage(&entityModule)

	// build the story
	assert.Nil(t, relationDB.Open("http://localhost:7474"))
	articles, err := relationDB.GetAll()

	assert.Nil(t, err)
	//assert.True(t, len(articles) > 150)

	set := testSet{
		//mainArticle: "The Horror in San Bernardino",
		mainArticle: "Fear Ignorance, Not Muslims",
		//mainArticle:     "Ted ‘Carpet-Bomb’ Cruz",
		//mainArticle: "Deregulating Corporate America",
		//mainArticle: "Course Correction for School Testing",
		//mainArticle: "If New York Really Wants to Help the Homeless",
		//mainArticle: "Social Security in an Election Year",
		//mainArticle: "The Reproductive Rights Rollback of 2015",
		//mainArticle: "Strong Unions, Strong Democracy",
		//mainArticle: "Voter Fatigue in New York",
		//mainArticle: "Depraved Indifference Toward Flint",

		relatedArticles: articles,
	}

	story := storyFromSet(set)
	fmt.Println(story.MainArticle.Name())

	raw, err := storyDriver(pipe, story)
	fmt.Println("len of data comming out:", len(raw))
	data := heapFilter(raw, scoreFuncs, weightMap, 10)

	// only get the top couple of articles

	assert.Nil(t, err)
	fmt.Println("main:", story.MainArticle.Name())
	for i := range data {
		fmt.Println(i, data[i].Name())
		printArticle(data[i])
		fmt.Println("total score:", scoreArticle(&data[i], scoreFuncs, weightMap))
		fmt.Println()
	}
}

func printArticle(article pipeline.Article) {

	keys := article.Keys()
	for _, key := range keys {
		score, _ := article.GetScore(key)
		fmt.Println(key, "is:", score)
	}
}
