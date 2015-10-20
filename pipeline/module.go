package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
)

// Type send through the pipeline
type AnalyzableStory struct {
	MainArticle     analyzer.Analyzable
	RelatedArticles chan analyzer.Analyzable
}

type Module interface {

	// Do any stage setup
	Setup()

	// all your code goes here
	Analyze(analyzer.Analyzable, analyzer.Analyzable) analyzer.Analyzable

	// up to each module to make sure the close happens properly
	Close() error

	// story stream accessors
	SetInputChan(chan AnalyzableStory)
	getInputChan() chan AnalyzableStory

	GetOutputChan() chan AnalyzableStory

	// gets the closing stream
	getClose() chan chan error

	// for significant errors
	SetErrorPropogateChan(chan error)
	getErrorPropogateChan() chan error
}

// manages input stream for a module
// calls the module's Analyze function
// pass the module in by reference (check if we actually need to)
func Run(m Module) {

	// dummy error for now
	var err error

	// tmp chans for story in/out stream
	inc := m.getInputChan()
	var outc chan AnalyzableStory

	var istory AnalyzableStory          // input story
	var storyc chan analyzer.Analyzable // input related articles tmp chan

	var ostory AnalyzableStory           // output story
	var results chan analyzer.Analyzable // output related article tmp chan

	// input to the analyze function

	// output from the analyze function
	analyzed := make(chan analyzer.Analyzable)

	// tells the analyzer to exit early
	done := make(chan bool, 1) // buffered so we can write directly

	// tmp holder variables
	var processedArticle analyzer.Analyzable
	var mainArticle analyzer.Analyzable

	// this func is run async to process data
	// captures analyzed, module references to actual objs
	// TODO: think about making this set-able
	// 	Problem with making it setable is it doesn't get to capture....
	manageAnalyze := func(main, related analyzer.Analyzable) {

		// process result
		// TODO: let this kick things out of the pipe, give errors
		result := m.Analyze(main, related)

		// return or handle error/terminate
		select {
		case <-done:
			fmt.Println("closing out of manageAnalyze early")
			// TODO: make sure all the mem gets closed
			return
		case analyzed <- result:
			// this case is all good
		}
	}

	// run the loop
	for {

		select {
		case nextStory, isOpen := <-inc:
			// check if there is a new story

			if !isOpen {
				fmt.Println("done reading from input chan")
				// if the line closed
				// TODO: decide if we want it to stay open or go for close here...

				// stop reading from the input story chan
				inc = nil

				// let it loop until told to close
				// TODO: make some way to handle it
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

			// enable story out chan
			outc = m.GetOutputChan()

		case outc <- ostory:
			// send the output story to the next stage

			// disable story out chan
			outc = nil

			// story is connected to next stage, start reading input and set main article
			storyc = istory.RelatedArticles
			mainArticle = istory.MainArticle

		case next, isOpen := <-storyc:
			// read related article stream

			// wait to read next article
			storyc = nil

			if !isOpen {
				// we are at the end of the current story's line
				// read on closed chan returns zero value (null in this case)

				// close the story's related stream
				close(ostory.RelatedArticles)
				ostory.RelatedArticles = nil

				// enable reading the main article stream
				inc = m.getInputChan()

				break // get out of this if statement
			}

			// analyze the story
			// TODO: look into staging this all better
			go manageAnalyze(mainArticle, next)

		case processedArticle = <-analyzed:
			// read the analyzed result
			// TODO: look into handling kicking values out of the stream

			// enable writing the analyzed related article to the output stream
			results = ostory.RelatedArticles

		case results <- processedArticle:
			// write the analyzed related article to the output stream

			// disable writing to the output stream
			results = nil

			// enable reading the input story's related article stream
			storyc = istory.RelatedArticles

		case errc := <-m.getClose():
			// stop the run function
			// its up to the module's Close func to close
			// the closing, err, output chans

			if done != nil {
				select {
				case done <- true:
					fmt.Println("send done OK")
					// try to send done down the line
					// TODO: test this
				default:
					fmt.Println("warning, could not send done down to chan")
					close(done)
				}
			}

			// stop reading analyzed

			// close the chans we created
			if ostory.RelatedArticles != nil {
				close(ostory.RelatedArticles)
			}

			// send the error down stream and exit
			errc <- err
			return

		} // end select
	} // end for
}
