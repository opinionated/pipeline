package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
)

// Story is passed through the pipeline.
type Story struct {
	MainArticle     analyzer.Analyzable
	RelatedArticles chan analyzer.Analyzable
}

// Module is a stage in the pipeline.
type Module interface {
	Setup()

	// Analyze is run on each article in a Story's related articles. It ranks
	// the related article against the main article. This is the meat of each
	// module.
	Analyze(analyzer.Analyzable, *analyzer.Analyzable) (bool, error)

	// up to each module to make sure the close happens properly
	Close() error

	// story stream accessors
	SetInputChan(chan Story)
	getInputChan() chan Story

	// get the story output channel
	GetOutputChan() chan Story

	// gets the closing stream
	getClose() chan chan error

	// for significant errors
	// talked to by the pipeline manager
	SetErrorPropogateChan(chan error)
	getErrorPropogateChan() chan error
}

type analyzedResponse struct {
	article analyzer.Analyzable
	use     bool
}

// Run wraps story stream for a module, calling module.Analyze
// on each related article in a story.
func Run(m Module) {

	fmt.Println("in run")
	// dummy error for now
	// TODO: decide if we want/need this
	var merr error

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

	analyzed := make(chan analyzedResponse) // analyzed related articles
	var analyzedtmp chan analyzedResponse
	analyzedtmp = analyzed

	// wraps a call to analyze so we can do it async
	doAnalyze := func(main, related analyzer.Analyzable) {

		// TODO: let this kick things out of the pipe, give errors
		// TODO: handle errors
		use, err := m.Analyze(main, &related)

		// TODO: set up to use err
		if err != nil {
			merr = err

			select {
			case m.getErrorPropogateChan() <- err:
			case <-cancelAnalyze:
			}
			return
		}

		if !use {
			fmt.Println("WARNING: if upstream write depends on downstream read this will break")
		}

		response := analyzedResponse{related, use}
		// return or handle error/terminate
		select {
		case <-cancelAnalyze:
			fmt.Println("closing out of doAnalyze early")
			// TODO: make sure all the mem gets released and all processes stop
			return
		case analyzed <- response:
		}
	}

	// nil it so we don't read unless we actually have something to read
	analyzed = nil

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
			analyzed = analyzedtmp

			// TODO: look into staging this all better
			go doAnalyze(currentStory.MainArticle, next)

		case analyzedResponse := <-analyzed:
			analyzed = nil

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
				fmt.Println("closing related")
				close(analyzedStory.RelatedArticles)
			}

			// send the error down stream and exit
			errc <- merr
			return

		} // end select
	} // end for
}
