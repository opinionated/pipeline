package pipeline

import (
	"fmt"
	"sync"
)

// Pipeline manages a series of pipeline modules.
type Pipeline struct {
	modules      []Module
	size         int
	in           chan Story
	errc         chan error
	closeModules chan bool // treat this like a signal

	// count how many running modules we have
	// use to make sure all modules get closed cleanly
	wg sync.WaitGroup
}

// NewPipeline builds a new pipeline with the given number of stages.
func NewPipeline(numStages int) *Pipeline {
	p := &Pipeline{
		modules:      make([]Module, numStages, numStages),
		size:         0,
		errc:         make(chan error),
		closeModules: make(chan bool),
	}
	return p
}

// SetInput sets the pipeline input chan
// expected to run this before the thing is set up
func (p *Pipeline) SetInput(inc chan Story) {
	if len(p.modules) > 0 {
		//panic("going to set input before pipeline gets built")
	}

	p.in = inc
}

// AddStage adds a new module, assumes that the module has not been set up yet
func (p *Pipeline) AddStage(m Module) {
	if p.size == cap(p.modules) {
		panic("added more modules than expected... going beyond cap")
	}

	m.Setup()

	if p.size == 0 {
		// if first element
		m.SetInputChan(p.in)
	} else {
		m.SetInputChan(p.GetOutput())
	}

	// modules propogate errors up to the pipeline
	m.SetErrorPropogateChan(p.errc)

	p.modules[p.size] = m
	p.size = p.size + 1
}

// GetOutput returns the final output chan of the pipeline
func (p *Pipeline) GetOutput() chan Story {
	if p.size == 0 {
		panic("tried getting output of nil")
	}

	return p.modules[p.size-1].GetOutputChan()
}

// Start the pipeline
func (p *Pipeline) Start() {
	if p.size == 0 {
		panic("tring to run an empty pipe")
	}

	for _, m := range p.modules {

		p.wg.Add(1)
		go p.runStage(m, p.closeModules)
	}

	go p.run()
}

func (p *Pipeline) run() {
	// wait until pipeline gets closed or a module has a big error
	select {
	case err := <-p.errc:
		fmt.Println("pipeline is not set up to handle big errors yet")
		panic(err)
	case <-p.closeModules:

	}
}

// Stop the pipeline
func (p *Pipeline) Stop() {
	// use the closeModules chan as a signal
	close(p.closeModules)

	// don't finish until all modules close
	p.wg.Wait()
}

// run a stage of the module, called by the pipeline
// m is the module to be run and is already all hooked up, is a reference
// terminate stops the stage from running
func (p *Pipeline) runStage(m Module, terminate chan bool) {
	go Run(m)

	// wait for close signal
	<-terminate
	m.Close()

	// decrease open module count
	p.wg.Done()
}
