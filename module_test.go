package pipeline_test

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline"
)

type testSet struct {
	mainArticle      string
	relatedArticles  []string
	expectedArticles []string
}

var neoTestSet = testSet{
	mainArticle: "Guns, Anger and Nonsense in Oregon",
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

	case err := <-pipe.Error():
		cerr := pipe.Close()
		if cerr != nil {
			err = fmt.Errorf("%s\n%s", err, cerr)
		}
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

		case err := <-pipe.Error():
			cerr := pipe.Close()
			if cerr != nil {
				err = fmt.Errorf("%s\n%s", err, cerr)
			}
			return nil, err
		}
	}
}
