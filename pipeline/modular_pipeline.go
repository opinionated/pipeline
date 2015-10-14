package pipeline

import (
	"fmt"
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"github.com/opinionated/pipeline/pipeline"
	"os"
	"testing"
)

type Pipeline interface {
	run_pipeline(string) chan string //article ids
	parallel_run() chan string       //article ids
}

type OpinionatedPipeline struct {
	analysis_modules []PipelineModule
	AOIs             chan string
	//DATABASE (temp xml) LINK HERE? !!!!!LINK TO PREPROCESSED DATA!!!!!
}

/**

Runs Opinionated Pipeline

*/
func (op OpinionatedPipeline) run_pipeline(current string) chan string {

	var new_relatives chan string = nil

	for index, module := range op.analysis_modules { //loop over each module

		module.AOI = current                      //set AOI to current AOI
		new_relatives = module.analyzer.Analyze() //Run analyzer

		if index != len(op.analysis_modules)-2 { //unless you are before the last module, set output of prev. as input of next
			op.analysis_modules[index+1].relatives = new_relatives
		}

	}

	return new_relatives //return final output of final module

}

/**

Run Pipeline as a Pipeline

*/

func (op OpinionatedPipeline) parallel_run() chan string {

	go func() {

		curr := <-AOIs
		op.run_pipeline(curr)

	}()

	return nil
}
