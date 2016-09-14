package pipeline_test

import (
	"container/heap"
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
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
	story.MainArticle = analyzer.Analyzable{FileName: set.mainArticle}
	story.RelatedArticles = make(chan analyzer.Analyzable)

	go func() {

		for i := range set.relatedArticles {
			story.RelatedArticles <- analyzer.Analyzable{
				FileName: set.relatedArticles[i],
			}
		}

		close(story.RelatedArticles)

	}()

	return story
}

// manages running a story
func storyDriver(
	pipe *pipeline.Pipeline,
	story pipeline.Story) ([]analyzer.Analyzable, error) {

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

	related := make([]analyzer.Analyzable, 0)

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

// only gets the top #num articles from the articles list
func heapFilter(articles []analyzer.Analyzable, num int) []analyzer.Analyzable {
	mheap := make(analyzer.Heap, 0)
	if mheap == nil {
		panic("oh nose, nil heap!")
	}

	fmt.Println("made it!")
	heap.Init(&mheap)

	for i := range articles {
		if mheap.Len() == num {
			if articles[i].Score > mheap.Peek().Score {
				heap.Pop(&mheap)
				heap.Push(&mheap, &articles[i])
			}
		} else {
			heap.Push(&mheap, &articles[i])
		}
	}

	ret := make([]analyzer.Analyzable, num)
	for i := 0; i < num; i++ {
		ret[i] = *heap.Pop(&mheap).(*analyzer.Analyzable)
	}

	return ret
}

// for the full test
type neoAnalyzer struct {
	metadataType string
	weight       float64
}

func (na neoAnalyzer) Setup() error {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		panic(err)
	}
	return nil
}

func (na neoAnalyzer) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	flow, count, err := relationDB.StrengthBetween(
		main.FileName,
		related.FileName,
		na.metadataType)

	if err != nil {
		return false, err
	}

	related.Score += na.weight * float64(flow) * float64(count*count)
	return true, nil

}

// better than the square for things with lots of connections
type neoAverageAnalyzer struct {
	metadataType string
	weight       float64
}

func (na neoAverageAnalyzer) Setup() error {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		panic(err)
	}
	return nil
}

func (na neoAverageAnalyzer) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	flow, count, err := relationDB.StrengthBetween(
		main.FileName,
		related.FileName,
		na.metadataType)

	if err != nil {
		return false, err
	}

	if count > 0 {
		related.Score += na.weight * float64(flow) / float64(count)
	}
	return true, nil

}

type threshAnalyzer struct {
	threshhold float64
}

func (ta threshAnalyzer) Setup() error { return nil }

func (ta threshAnalyzer) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	if related.Score > ta.threshhold {
		return true, nil
	}
	return false, nil
}

func TestFull(t *testing.T) {

	taxFunc := neoAnalyzer{metadataType: "Taxonomy", weight: 3.0}
	taxModule := pipeline.StandardModule{}
	taxModule.SetFuncs(taxFunc)

	threshFunc := threshAnalyzer{threshhold: 1.0}
	threshModule := pipeline.StandardModule{}
	threshModule.SetFuncs(threshFunc)

	conceptsFunc := neoAnalyzer{metadataType: "Concept", weight: 4.0}
	conceptsModule := pipeline.StandardModule{}
	conceptsModule.SetFuncs(conceptsFunc)

	lastThreshFunc := threshAnalyzer{threshhold: 1.0}
	lastThreshModule := pipeline.StandardModule{}
	lastThreshModule.SetFuncs(lastThreshFunc)

	keyFunc := neoAverageAnalyzer{metadataType: "Keyword", weight: 3.0}
	keyModule := pipeline.StandardModule{}
	keyModule.SetFuncs(&keyFunc)

	entityFunc := neoAverageAnalyzer{metadataType: "Entity", weight: 2.0}
	entityModule := pipeline.StandardModule{}
	entityModule.SetFuncs(&entityFunc)

	// build the pipe
	pipe := pipeline.NewPipeline()

	// do coarse methods
	pipe.AddStage(&taxModule)
	pipe.AddStage(&conceptsModule)

	// thresh then do finer methods
	pipe.AddStage(&keyModule)
	pipe.AddStage(&entityModule)

	// build the story
	assert.Nil(t, relationDB.Open("http://localhost:7474"))
	articles, err := relationDB.GetAll()

	assert.Nil(t, err)
	assert.True(t, len(articles) > 150)

	set := testSet{
		mainArticle:     "The Horror in San Bernardino",
		relatedArticles: articles,
	}

	story := storyFromSet(set)
	fmt.Println(story.MainArticle.FileName)

	raw, err := storyDriver(pipe, story)
	data := heapFilter(raw, 20)

	// only get the top couple of articles

	assert.Nil(t, err)
	fmt.Println("main:", story.MainArticle.FileName)
	for i := range data {
		fmt.Println(data[i].FileName, "score:", data[i].Score)
	}
}
