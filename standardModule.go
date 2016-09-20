package pipeline

import ()

// StandardModuleAnalyzer can be plugged into the standard module
type StandardModuleAnalyzer interface {
	Analyze(Article, *Article) (bool, error)
	Setup() error
}

// StandardModule takes functions as parameters and uses those to adjust behavior.
// For most analysis this should be totally fine
type StandardModule struct {
	in      chan Story
	out     chan Story
	err     chan error
	closing chan chan error

	sma StandardModuleAnalyzer
}

// SetFuncs used by the standard module
func (m *StandardModule) SetFuncs(sma StandardModuleAnalyzer) {
	m.sma = sma
}

// Analyze falls through to the analyze fun.
func (m *StandardModule) Analyze(main Article,
	related *Article) (bool, error) {

	return m.sma.Analyze(main, related)
}

// Setup does normal setup and calls the setup func.
func (m *StandardModule) Setup() {
	m.out = make(chan Story, 1)
	m.closing = make(chan chan error)

	if err := m.sma.Setup(); err != nil {
		panic(err)
	}
}

// Close stops the module and cleans up any open connections.
func (m *StandardModule) Close() error {
	errc := make(chan error)
	m.closing <- errc
	return <-errc
}

// SetInputChan sets the module's input channel.
func (m *StandardModule) SetInputChan(inc chan Story) {
	m.in = inc
}

// GetOutputChan returns the modules output channel.
func (m *StandardModule) GetOutputChan() chan Story {
	return m.out
}

// SetErrorPropogateChan sets the channel for errors to propagate out
// of this module.
func (m *StandardModule) SetErrorPropogateChan(errc chan error) {
	m.err = errc
}

// remaining methods are used internally by run methods

func (m *StandardModule) getErrorPropogateChan() chan error {
	return m.err
}

func (m *StandardModule) getInputChan() chan Story {
	return m.in
}

func (m *StandardModule) getClose() chan chan error {
	return m.closing
}

// check that the module was compiled properly
var _ Module = (*StandardModule)(nil)
