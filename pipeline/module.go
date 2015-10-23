package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
)

// type send through the pipeline
type Story struct {
	MainArticle     analyzer.Analyzable
	RelatedArticles chan analyzer.Analyzable
}

// a stage in the pipeline
// the important thing here is to override the Analyze function with
// your own analyze code
type Module interface {
	Setup()

	// compare one related article to the main article
	// return err, useAgain and moding the analyzable by reference
	Analyze(analyzer.Analyzable, *analyzer.Analyzable) (error, bool)

	// TODO: handle when the story changes

	// up to each module to make sure the close happens properly
	Close() error

	// story stream accessors
	SetInputChan(chan Story)
	getInputChan() chan Story

	GetOutputChan() chan Story

	// gets the closing stream
	getClose() chan chan error

	// for significant errors
	// talked to by the pipeline manager
	SetErrorPropogateChan(chan error)
	getErrorPropogateChan() chan error
}

// manages input stream for a module
// calls the module's Analyze function
// pass the module in by reference
func Run(m Module) {

	// dummy error for now
	// TODO: decide if we want/need this
	var err error

	// tmp chans for story in/out stream
	inputStream := m.getInputChan()
	var outputStream chan Story

	var currentStory Story
	var relatedIn chan analyzer.Analyzable // input related articles tmp chan

	var analyzedStory Story
	var relatedOut chan analyzer.Analyzable // output related article tmp chan

	// tells the analyzer to exit early
	cancelAnalyze := make(chan bool, 1)

	// tmp state holder
	var analyzedArticle analyzer.Analyzable

	// response type from article
	type AnalyzedResponse struct {
		article analyzer.Analyzable
		use     bool
	}

	analyzed := make(chan AnalyzedResponse) // analyzed related articles

	// wraps a call to analyze so we can do it async
	doAnalyze := func(main, related analyzer.Analyzable) {

		// TODO: let this kick things out of the pipe, give errors
		// TODO: handle errors
		err, use := m.Analyze(main, &related)

		// TODO: set up to use err
		if err != nil {
			panic(err)
		}

		if !use {
			fmt.Println("WARNING: if upstream write depends on downstream read this will break")
		}

		response := AnalyzedResponse{related, use}
		// return or handle error/terminate
		select {
		case <-cancelAnalyze:
			fmt.Println("closing out of doAnalyze early")
			// TODO: make sure all the mem gets released and all processes stop
			return
		case analyzed <- response:
		}
	}

	for {

		select {
		case nextStory, isOpen := <-inputStream:
			// check if there is a new story

			if !isOpen {
				fmt.Println("cancelAnalyze reading from input chan")
				// if the line closed
				// TODO: decide if we want it to stay open or go for close here...

				// stop reading from the input story chan
				inputStream = nil

				// let it loop until told to close
				// TODO: make some way to handle it
				break
			}

			currentStory = nextStory

			// don't get a new story until we are cancelAnalyze with this one
			inputStream = nil

			// build the output story
			analyzedStory.MainArticle = currentStory.MainArticle
			analyzedStory.RelatedArticles = make(chan analyzer.Analyzable)

			// enable story output stream
			outputStream = m.GetOutputChan()

		case outputStream <- analyzedStory:
			// write the story down stream

			relatedIn = currentStory.RelatedArticles
			outputStream = nil

		case next, isOpen := <-relatedIn:
			// read and analyze each related article in the current story

			if !isOpen {

				// close this story up
				close(analyzedStory.RelatedArticles)
				analyzedStory.RelatedArticles = nil

				// look for next story
				inputStream = m.getInputChan()
				relatedIn = nil

				break
			}

			// don't read a new article until we analyze this one
			relatedIn = nil

			// TODO: look into staging this all better
			go doAnalyze(currentStory.MainArticle, next)

		case analyzedResponse := <-analyzed:

			if analyzedResponse.use {

				// only write the article if the stream lets us
				relatedOut = analyzedStory.RelatedArticles
				analyzedArticle = analyzedResponse.article

			} else {

				// go run to related
				relatedIn = currentStory.RelatedArticles
			}

		case relatedOut <- analyzedArticle:

			// read the next related article
			relatedIn = currentStory.RelatedArticles

			relatedOut = nil

		case errc := <-m.getClose():
			// stop the run function
			// its up to the module's Close func to close
			// the closing, err, output chans

			if cancelAnalyze != nil {
				select {
				case cancelAnalyze <- true:
					fmt.Println("sent cancel analyze")
					// try to send cancelAnalyze down the line
					// TODO: test this
				default:
					fmt.Println("warning, could not send cancelAnalyze down to chan")
					close(cancelAnalyze)
				}
			}

			// close the chans we created
			if analyzedStory.RelatedArticles != nil {
				close(analyzedStory.RelatedArticles)
			}

			// send the error down stream and exit
			errc <- err
			return

		} // end select
	} // end for
}
