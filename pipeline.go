package pipeline

import (
	"fmt"
)

// Pipeline manages a series of pipeline modules.
type Pipeline struct {
	modules     []Module
	errc        chan error
	closeSignal chan struct{} // treat this like a signal
}

// NewPipeline builds a new pipeline with the given number of stages.
func NewPipeline() *Pipeline {
	p := &Pipeline{
		modules:     make([]Module, 0, 1),
		errc:        make(chan error),
		closeSignal: make(chan struct{}),
	}

	return p
}

// AddStage adds a new module, assumes that the module has not been set up yet
func (p *Pipeline) AddStage(m Module) {
	m.Setup()

	size := len(p.modules)

	if size == 0 {
		m.SetInputChan(make(chan Story))
	} else {
		m.SetInputChan(p.GetOutput())
	}

	// modules propagate errors up to the pipeline
	m.SetErrorPropogateChan(p.errc)

	p.modules = append(p.modules, m)
}

// PushStory to the pipeline
// blocks until the story hits the pipeline (or the buffer)
func (p *Pipeline) PushStory(story Story) {
	p.modules[0].getInputChan() <- story
}

// GetOutput returns the final output chan of the pipeline
// not thread safe
func (p *Pipeline) GetOutput() chan Story {
	return p.modules[len(p.modules)-1].GetOutputChan()
}

// Start the pipeline
func (p *Pipeline) Start() {

	for i := range p.modules {
		go Run(p.modules[i])

	}

	go p.run()
}

func (p *Pipeline) run() {

	// wait until pipeline gets closed or a module has a big error
	select {
	case err := <-p.errc:
		fmt.Println("pipeline is not set up to handle big errors yet")
		panic(err)

	case <-p.closeSignal:
	}
}

// Error on the pipeline
func (p *Pipeline) Error() <-chan error {
	return p.errc
}

// Close the pipeline
func (p *Pipeline) Close() error {
	close(p.closeSignal)

	// go close all the individual modules
	var err error
	for i := range p.modules {
		merr := p.modules[i].Close()

		if merr != nil {
			if err != nil {
				err = fmt.Errorf("Error(s) closing pipeline:")
			}

			err = fmt.Errorf("%s\n\t%s", err.Error(), merr.Error())
		}
	}

	return err
}
