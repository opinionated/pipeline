package pipeline

import (
	"github.com/opinionated/analyzer-core/alchemy"
	"github.com/opinionated/analyzer-core/analyzer"
	"os"
)

// TODO: make error chans

// filter by taxonomy, this is just a rought for now
func FilterTaxonomy(top analyzer.Analyzable,
	in chan analyzer.Analyzable) chan analyzer.Analyzable {

	out := make(chan analyzer.Analyzable)
	// get a chan with the taxonomys already loaded
	withTaxonomy := LoadTaxonomy(in)

	go func() {
		for toAnalyze := range withTaxonomy {
			// TODO: make a filtering function to score them
			if toAnalyze.Taxonomys.Taxonomys[0].Label == top.Taxonomys.Taxonomys[0].Label {
				out <- toAnalyze
			}
		}
		close(out)
	}()

	return out
}

// Gets taxonomy data from file and puts it into the chan
func LoadTaxonomy(in chan analyzer.Analyzable) chan analyzer.Analyzable {

	out := make(chan analyzer.Analyzable)

	names := make(chan string)
	go func() {
		files := openFile(names)
		for toAnalyze := range in {
			names <- toAnalyze.FileName + "_taxonomy.xml"
			toRead := <-files
			defer toRead.Close() // need to close b/c open file won't do it for you
			err := alchemy.ToXML(toRead, &toAnalyze.Taxonomys)
			if err != nil {
				panic(err)
			}
			out <- toAnalyze
		}
		close(out)
	}()

	return out
}

// NOTE: does not close file, you will need to do that later
func openFile(in chan string) chan *os.File {
	out := make(chan *os.File)

	go func() {
		for name := range in {
			file, err := os.Open(name)
			if err != nil {
				panic(err)
			}
			out <- file
		}
		close(out)
	}()
	return out
}
