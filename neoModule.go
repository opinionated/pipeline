package pipeline

import (
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/analyzer-core/dbInterface"
)

// NeoModule uses neo4j to analyze articles
type NeoModule struct {
	in          chan Story
	out         chan Story
	err         chan error
	closing     chan chan error
	mainArticle analyzer.Analyzable

	scoreFunc    func(float32, int) float32
	metadataType string
}

// ScoreSimpleMul just multiplies two commands together
func ScoreSimpleMul(flow float32, count int) float32 {
	return flow * float32(count)
}

// SetParams sets the variables used for the module
// TODO: think about any more params
func (m *NeoModule) SetParams(metadataType string, scoreFunc func(float32, int) float32) {
	m.metadataType = metadataType
	m.scoreFunc = scoreFunc
}

// Analyze via the relation database.
func (m *NeoModule) Analyze(main analyzer.Analyzable,
	related *analyzer.Analyzable) (bool, error) {

	flow, count, err := relationDB.StrengthBetween(main.FileName, related.FileName, m.metadataType)
	if err != nil {
		return false, err
	}

	related.Score += float64(m.scoreFunc(flow, count))
	return true, nil
}

// Setup tries to open the relation db.
func (m *NeoModule) Setup() {
	err := relationDB.Open("http://localhost:7474")
	if err != nil {
		// TODO: better error handling
		panic(err)
	}

	m.out = make(chan Story, 1)
	m.closing = make(chan chan error)
}

// Close stops the module and cleans up any open connections.
func (m *NeoModule) Close() error {
	errc := make(chan error)
	m.closing <- errc
	return <-errc
}

// SetInputChan sets the module's input channel.
func (m *NeoModule) SetInputChan(inc chan Story) {
	m.in = inc
}

// GetOutputChan returns the modules output channel.
func (m *NeoModule) GetOutputChan() chan Story {
	return m.out
}

// SetErrorPropogateChan sets the channel for errors to propagate out
// of this module.
func (m *NeoModule) SetErrorPropogateChan(errc chan error) {
	m.err = errc
}

// remaining methods are used internally by run methods

func (m *NeoModule) getErrorPropogateChan() chan error {
	return m.err
}

func (m *NeoModule) getInputChan() chan Story {
	return m.in
}

func (m *NeoModule) getClose() chan chan error {
	return m.closing
}

// check that the module was compiled properly
var _ Module = (*NeoModule)(nil)
