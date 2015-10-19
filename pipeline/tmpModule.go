package pipeline

import (
	"github.com/opinionated/analyzer-core/analyzer"
)

type tmpModule struct {
	in          chan AnalyzableStory
	out         chan AnalyzableStory
	err         chan error
	closing     chan chan error
	mainArticle analyzer.Analyzable
}

func (m *tmpModule) Analyze(in chan analyzer.Analyzable,
	out chan analyzer.Analyzable,
	done chan bool) {

	var article analyzer.Analyzable
	intmp := in
	var outtmp chan analyzer.Analyzable

	for {
		select {

		case <-done:
			return

		case article = <-intmp:
			//your code here...
			outtmp = out
			intmp = nil

		case outtmp <- article:
			outtmp = nil
			intmp = in

		}
	}
}

func (m *tmpModule) Run() {

	var err error

	// this is basically just a lot of code to do this:
	// for( story : pipe){
	// 	for( related : story){
	// 		analyze(related)
	// 		writeToPipe(related)
	// 	}
	// }
	//
	// however, if the next stage breaks while we are trying
	// to write to it this stage will enter deadlock
	//
	// error handlinging, managing the story stream and managing
	// each story's related articles without deadlock is a bit
	// messy
	//
	// this function manages a related article stream that gets
	// fed into Analyze so you don't need to worry as much about
	// the details of how this all gets handled

	// tmp chans for in/out stream
	var inc = m.in
	var outc chan AnalyzableStory
	var istory AnalyzableStory           // input story
	var storyc chan analyzer.Analyzable  // input related articles tmp chan
	var ostory AnalyzableStory           // output story
	var results chan analyzer.Analyzable // output related article tmp chan
	// input to the analyze function
	analyze_in := make(chan analyzer.Analyzable)
	var analyze chan analyzer.Analyzable
	// output from the analyze function
	analyze_out := make(chan analyzer.Analyzable)
	finishAnalyzer := make(chan bool, 1) // tell the analyzer to finish
	var freshArticle analyzer.Analyzable
	var processedArticle analyzer.Analyzable
	// now spin up the analyzer
	go m.Analyze(analyze_in, analyze_out, finishAnalyzer)
	for {
		select {
		case nextStory, isOpen := <-inc:
			// check if there is a new story
			if !isOpen {
				// if the line closed
				// TODO: decide if we want it to stay open or go for close here...
				// stop the analyze task
				finishAnalyzer <- true
				// stop reading from the in chan
				inc = nil
				// continue on
				break
			}
			// set the current story
			istory = nextStory
			// once you have a story, build the output and set the inc to nil
			// so that we don't get another story too soon
			inc = nil
			// build the output story
			ostory.MainArticle = istory.MainArticle
			ostory.RelatedArticles = make(chan analyzer.Analyzable)
			// enable sending to out
			outc = m.out
		case outc <- ostory:
			// send the output story down the line
			// wait to start processing the next story until you can pass it down the line
			outc = nil
			// once we know we have someone to read, we can send down the line
			storyc = istory.RelatedArticles
			// set this up for the analyze function
			m.mainArticle = istory.MainArticle
		case next, isOpen := <-storyc:
			// read stories from upstream until stream closes
			storyc = nil // no mater what, don't go get the next article yet
			if !isOpen {
				// we are at the end of the current story's line
				// read on closed chan returns zero value (null in this case)
				// TODO: sync this with the analyze func so we don't write on closed stream
				close(ostory.RelatedArticles)
				ostory.RelatedArticles = nil
				// start looking for the next story in the stream
				inc = m.in
				break // get out of this if statement
			}

			// send what we just read over to the analyzer
			freshArticle = next
			analyze = analyze_in
		case analyze <- freshArticle:
			// send the most recent article over to the analyzer
			analyze = nil
		case processedArticle = <-analyze_out:
			// read the result from the analyzer
			results = ostory.RelatedArticles
		case results <- processedArticle:
			// write the article to the next stage
			results = nil
			// read the next article in
			storyc = istory.RelatedArticles
		case errc := <-m.closing:
			// send the error back
			// close anything open
			m.in = nil
			close(m.err)
			close(m.closing)
			close(m.out)
			// close the chans we created
			close(analyze_in)
			if ostory.RelatedArticles != nil {
				close(ostory.RelatedArticles)
			}
			finishAnalyzer <- true
			close(finishAnalyzer)
			errc <- err
			return
		case bigErr := <-m.err:
			// what to do when a big error comes along
			// propogate and return
			m.err <- bigErr
			err = bigErr
			// ignore any values that will come down the line
			// TODO: make all these nil
			m.in = nil
		}
	}
}
