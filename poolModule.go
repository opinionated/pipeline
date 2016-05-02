package pipeline

import (
	"container/heap"
	"fmt"
	"github.com/opinionated/analyzer-core/analyzer"
)

// PoolModule pools stories into a min heap.
// Total op takes n*log(n) where n is related articles. In
// practice it will be longer because it needs to read and
// process all the articles (at nlogn) then send them down
// stream at n.

// This is not staged as well as it could be. Buffered chans
// may help. Examples are not processing articles when waiting
// for a reader, not analyzing while waiting for pipe to open
// up (this is only an issue in the other pipeline). These all
// seem to have fairly straight forwards fixes.
type PoolModule struct {
	in          chan Story
	out         chan Story
	err         chan error
	closing     chan chan error
	mainArticle analyzer.Analyzable

	// for pooling articles
	heap     analyzer.Heap
	capacity int
}

func (m *PoolModule) SetCapacity(capacity int) {

	m.capacity = capacity
	m.heap = make(analyzer.Heap, 0)
	fmt.Println("made heap")
	heap.Init(&m.heap)
	fmt.Println("done with cap")
}

// RunPool is a special run method for this module.
// TODO: make sure this stays up to date with the normal run
func RunPool(m *PoolModule) {

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

	analyzed := make(chan analyzedResponse) // analyzed related articles
	analyzedTmp := analyzed
	analyzed = nil

	// pump values out
	pipeOut := func() {
		fmt.Println("going to pipe out")
		for m.heap.Len() > 0 {
			relatedArticle := heap.Pop(&m.heap).(*analyzer.Analyzable)

			select {
			case analyzed <- analyzedResponse{*relatedArticle, true}:

			case <-cancelAnalyze:
				return

			}
		}
		select {
		case analyzed <- analyzedResponse{analyzer.Analyzable{}, false}:

		case <-cancelAnalyze:
			return

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
			// enabled by inputStream when there is a new story

			// write the story down stream
			// when the story is read a module is scanning for related articles
			fmt.Println("sent output story")

			relatedIn = currentStory.RelatedArticles
			outputStream = nil

		case next, isOpen := <-relatedIn:
			// enabled when a story is being read downstream
			// does not flip with anything here

			// read and analyze each related article in the current story

			if !isOpen {
				fmt.Println("not open")

				// when this story closes, pipe everything along
				go pipeOut()
				analyzed = analyzedTmp // start looking at analyzed

				relatedIn = nil // stop reading from here

				break
			}

			// add items to the heap
			if m.heap.Len() == m.capacity {
				if m.heap.Peek().Score < next.Score {
					// only add at capacity if this element is bigger
					// than the lowest in the heap
					heap.Pop(&m.heap)
					heap.Push(&m.heap, &next)
				}
			} else {
				// if not at capacity, always add
				heap.Push(&m.heap, &next)
			}

		case analyzedResponse := <-analyzed:
			// first enabled when all articles have been read
			// then flips with relatedOut to send all the articles
			// down stream

			if analyzedResponse.use {

				// only write the article if the stream lets us
				relatedOut = analyzedStory.RelatedArticles
				analyzedArticle = analyzedResponse.article

				// don't read next until this one is sent
				// gets de-nil'd by relatedOut
				analyzed = nil

			} else {

				// use.false is signal that piping the related stories is done
				close(analyzedStory.RelatedArticles)
				analyzedStory.RelatedArticles = nil

				// look for next story
				inputStream = m.getInputChan()
			}

		case relatedOut <- analyzedArticle:
			// enabled by analyzed, flips with analyzed to send related
			// articles down stream

			// read the next related article
			analyzed = analyzedTmp

			// don't try to write to related out until there is a new article
			// de-nil'd by analyzed
			relatedOut = nil

		case errc := <-m.getClose():
			fmt.Println("hit an error")
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
			errc <- err
			return

		} // end select
	} // end for
}

// Setup builds the heap.
func (m *PoolModule) Setup() {
	m.out = make(chan Story, 1)
	m.closing = make(chan chan error)
}

// Analyze pools articles into a min heap.
func (m *PoolModule) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	return true, nil
}

// Close stops the module and cleans up any open connections.
func (m *PoolModule) Close() error {
	errc := make(chan error)
	m.closing <- errc
	return <-errc
}

func (m *PoolModule) getClose() chan chan error {
	return m.closing
}

// SetInputChan sets the module's input channel.
func (m *PoolModule) SetInputChan(inc chan Story) {
	m.in = inc
}

func (m *PoolModule) getInputChan() chan Story {
	return m.in
}

// GetOutputChan returns the modules output channel.
func (m *PoolModule) GetOutputChan() chan Story {
	return m.out
}

// SetErrorPropogateChan sets the channel for errors to propagate out
// of this module.
func (m *PoolModule) SetErrorPropogateChan(errc chan error) {
	m.err = errc
}

func (m *PoolModule) getErrorPropogateChan() chan error {
	return m.err
}

// missing getClose, getErrorPropogateChan, getInput

// check that the module was compiled properly
var _ Module = (*PoolModule)(nil)
