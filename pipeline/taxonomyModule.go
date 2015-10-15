package pipeline

import (
	"github.com/opinionated/analyzer-core/alchemy"
)

type TaxonomyModule struct {
	in  chan AnalyzableStory
	out chan AnalyzableStory

	err     chan error
	closing chan chan error
}

func (m *TaxonomyModule) Run() {
	// local mutable variables for the loop

	var err error // store any errors that come up in here so they can be reported

	// we need to worry about blocking calls here because they might make us miss
	// an error or a close
	// avoid blocking by moving variables through this stage in buffered pipes
	// this problem becomes a bit of a cluster because we pass a pipe through a pipe

	// take advantage of how reads from nil pipes hang to control flow
	// create a bunch of nil-able tmp vars to stand in for the actual chans
	var inc = m.in
	var outc chan AnalyzableStory

	// vars for the story, we have an out value and an in value
	var istory AnalyzableStory = AnalyzableStory{nil, nil}
	var storyInc chan analyzer.Analyzable

	var ostory AnalyzableStory = AnalyzableStory{nil, nil}
	// make a tmp for the related articles so we don't need to create 100s of these
	var tmpSOc chan analyzer.Analyzable = make(chan analyzer.Analyzable, 1)
	var storyOutc chan analyzer.Analyzable

	for {
		// declare vars that get run by the loop here
		select {
		case istory = <-inc:
			// try to read a story in
			// once you have a story, build the output and set the inc to nil
			// so that we don't get another story too soon
			inc = nil
			storyInc = nil

			// build the output story
			// TODO: need to make sure we won't have issues with this kind of copy
			ostory.MainArticle = istory.MainArticle
			ostory.RelatedArticles = make(chan analyzer.Analyzable)

			// buffer so we can send right away
			outc = make(chan AnalyzableStory, 1)
			outc <- ostory

		case m.out <- <-outc:
			// wait to start processing the next story until you can pass it down the line
			outc = nil

			// once we know we have someone to read, we can send down the line
			storyInc = istory.RelatedArticles

		case next := <-storyInc:
			// process each story as it comes out of the main
			storyInc = nil // no mater what, don't go get the next article yet

			if next == nil {
				// we are at the end of the loop
				// read on closed chan returns zero value (null in this case)
				close(ostory.RelatedArticles)

				// start looking for the next story in the stream
				inc = m.in
				break // get out of this if statement
			}

			// process story here....

			// make sure to send this in a non-blocking way
			storyOutc = tmpSOc
			storyOutc <- next // processed variable

		case ostory.RelatedArticles <- <-storyOutc:
			// send the result down the related stream
			storyOutc = nil
			storyInC = istory.RelatedArticles

		case errc := <-m.closing:
			// send the error back
			errc <- err

			// close anything open
			close(m.err)
			close(m.closeing)
			close(m.out)

			// end the go routine
			return

		case bigErr := <-m.err:
			// what to do when a big error comes along
			// propogate and return
			m.err <- bigErr
			err = bigErr

			// ignore any values that will come down the line
			m.in = nil
		}
	}
}

func (m *TaxonomyModule) Setup() {
	out := make(chan analyzer.Analyzable)
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
