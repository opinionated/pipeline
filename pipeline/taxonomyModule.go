package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
)

type TaxonomyModule struct {
	in  chan AnalyzableStory
	out chan AnalyzableStory

	err     chan error
	closing chan chan error

	mainArticle analyzer.Analyzable
}

// actually runs the analysis
func (m *TaxonomyModule) Analyze(in chan analyzer.Analyzable,
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
			//fmt.Println("reading file:", article.Name)
			outtmp = out
			intmp = nil
		case outtmp <- article:
			outtmp = nil
			intmp = in
		}
	}
	fmt.Println("oh nose!!!")
}

func (m *TaxonomyModule) Run() {
	var err error
	// this function manages the input chan so Analyze is only fed new values
	// when the value can be written

	// we need to worry about blocking calls here because they might make us miss
	// an error or a close
	// avoid blocking by moving variables through this stage in buffered pipes
	// this problem becomes a bit of a cluster because we pass a pipe through a pipe

	// take advantage of how reads from nil pipes hang to control flow
	// create a bunch of nil-able tmp vars to stand in for the actual chans

	var inc = m.in
	var outc chan AnalyzableStory

	// vars for the story, we have an out value and an in value
	var istory AnalyzableStory
	var storyc chan analyzer.Analyzable

	var ostory AnalyzableStory

	// set up the actual analyze function
	analyze_in := make(chan analyzer.Analyzable)
	var analyze chan analyzer.Analyzable

	analyze_out := make(chan analyzer.Analyzable) // buffed to save one var
	var results chan analyzer.Analyzable

	finishAnalyzer := make(chan bool, 1)

	var freshArticle analyzer.Analyzable
	var processedArticle analyzer.Analyzable

	// now spin up the analyzer
	go m.Analyze(analyze_in, analyze_out, finishAnalyzer)

	for {
		// declare vars that get run by the loop here
		select {

		// 1) READ THE NEXT STORY ON THE LINE
		case nextStory, isOpen := <-inc:
			// if the line closes, close the module
			if !isOpen {
				//fmt.Println("closing pipe")
				// TODO: decide how to handle this
				// stop the analyze task
				finishAnalyzer <- true

				// don't let it read from here again
				inc = nil

				// handle this how you normally would with closing
				break
			}
			istory = nextStory

			// once you have a story, build the output and set the inc to nil
			// so that we don't get another story too soon
			inc = nil

			// build the output story
			// TODO: need to make sure we won't have issues with this kind of copy
			ostory.MainArticle = istory.MainArticle
			ostory.RelatedArticles = make(chan analyzer.Analyzable)

			// enable sending to out
			outc = m.out

		case outc <- ostory:
			// wait to start processing the next story until you can pass it down the line
			outc = nil

			// once we know we have someone to read, we can send down the line
			storyc = istory.RelatedArticles

			// once this has been written, now set up the analyzable
			m.mainArticle = istory.MainArticle

		case next, isOpen := <-storyc:
			// process each story as it comes out of the main
			storyc = nil // no mater what, don't go get the next article yet

			if !isOpen {
				// we are at the end of the loop
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
			// try to read the next relevant article
			analyze = nil

		case processedArticle = <-analyze_out:
			results = ostory.RelatedArticles

		case results <- processedArticle:
			results = nil
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
			//fmt.Println("done with close")
			// end the go routine
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

func (m *TaxonomyModule) Setup() {
	m.out = make(chan AnalyzableStory, 1)
	m.closing = make(chan chan error)
}

func (m *TaxonomyModule) Close() error {
	errc := make(chan error)
	m.closing <- errc
	return <-errc
}

func (m *TaxonomyModule) SetInputChan(inc chan AnalyzableStory) {
	m.in = inc
}

func (m *TaxonomyModule) GetOutputChan() chan AnalyzableStory {
	return m.out
}

func (m *TaxonomyModule) SetErrorPropogateChan(errc chan error) {
	m.err = errc
}

// check that was compiled properly
var _ Module = (*TaxonomyModule)(nil)
