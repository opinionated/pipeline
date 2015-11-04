import sys

####################################################
"""
    NOTE: NEED TO REWORK THIS TO MAKE IT MORE REUSABLE

"""
####################################################

def main():
    print "\n\nBROKEN FOR NOW\n\n"

    print "\n\twelcome to the pipeline module creator"
    print "\tthis tool makes generating new pipe modules super duper easy"
    print "\tto generate a module, just enter the module name"
    print '\texample input: "taxonomy" would create a new module called "taxonomyModule"'
    print '\tin the file "taxonomyModule.go"\n'
    print '\tonce you are done generating the module, run go fmt __new file__ to make the code look pretty\n'
    pipe_name = raw_input("enter pipe module name: ")
    
    ##CreateSimplePipe(pipe_name)


def CreateSimplePipe(name):
    fd = open(name + "Module.go", "w")
    
    WriteHeader(fd, name)
    WriteStruct(fd, name)

    WriteAnalyze(fd, name)
    #WriteRun(fd, name)
    WriteSetup(fd, name)
    WriteClose(fd, name)
    WriteSetInputChan(fd, name)
    WriteGetOutputChan(fd, name)
    WriteSetErrorChan(fd, name)
    WriteFooter(fd, name)
  
def WriteAnalyze(fd, name):
    WriteFuncHeader(fd, name, 'Analyze',  'in chan analyzer.Analyzable,\n' +
            'out chan analyzer.Analyzable,\n' +
            'done chan bool')
    fd.write('\nvar article analyzer.Analyzable')
    fd.write('\n')
    fd.write('intmp := in')
    fd.write('\n')
    fd.write('var outtmp chan analyzer.Analyzable')
    fd.write('\n\n')
    fd.write('for {')
    fd.write('\n')
    fd.write('select {')
    fd.write('\n\n')
    fd.write('case <-done:')
    fd.write('\n')
    fd.write('return')
    fd.write('\n\n')
    fd.write('case article = <-intmp:')
    fd.write('\n')
    fd.write('//your code here...\n')
    fd.write('outtmp = out')
    fd.write('\n')
    fd.write('intmp = nil')
    fd.write('\n\n')
    fd.write('case outtmp <- article:')
    fd.write('\n')
    fd.write('outtmp = nil')
    fd.write('\n')
    fd.write('intmp = in')
    fd.write('\n\n}\n}\n}\n')
    fd.write('\n\n')


def WriteRun(fd, name):

    WriteFuncHeader(fd, name, 'Run', '')
    fd.write('\n')    
    fd.write('	var err error\n\n' + 
    	'// this is basically just a lot of code to do this:\n'+
	'// for( story : pipe){\n'+
	'// 	for( related : story){\n'+
	'// 		analyze(related)\n'+
	'// 		writeToPipe(related)\n'+
	'// 	}\n'+
	'// }\n'+
	'//\n'+
	'// however, if the next stage breaks while we are trying\n'+
	'// to write to it this stage will enter deadlock\n'+
	'//\n'+
	'// error handlinging, managing the story stream and managing\n'+
	"// each story's related articles without deadlock is a bit\n"+
	'// messy\n'+
	'//\n'+
	'// this function manages a related article stream that gets\n'+
	"// fed into Analyze so you don't need to worry as much about\n"+
	'// the details of how this all gets handled\n')


    fd.write('\n' + 
    '// tmp chans for in/out stream\n'+
    'var inc = m.in\n'+
    'var outc chan AnalyzableStory\n'+
    '\n' +
    'var istory AnalyzableStory          // input story\n'+
    'var storyc chan analyzer.Analyzable // input related articles tmp chan\n'+
    '\n' +
    'var ostory AnalyzableStory           // output story\n'+
    'var results chan analyzer.Analyzable // output related article tmp chan\n'+
    '\n' +
    '// input to the analyze function\n'+
    'analyze_in := make(chan analyzer.Analyzable)\n'+
    'var analyze chan analyzer.Analyzable\n'+
    '\n' +
    '// output from the analyze function\n'+
    'analyze_out := make(chan analyzer.Analyzable)\n'+
    '\n' +
    'finishAnalyzer := make(chan bool, 1) // tell the analyzer to finish\n'+
    '\n' +
    'var freshArticle analyzer.Analyzable\n'+
    'var processedArticle analyzer.Analyzable\n'+
    '\n' +
    '// now spin up the analyzer\n'+
    'go m.Analyze(analyze_in, analyze_out, finishAnalyzer)\n'+
    '\n' +
    'for {\n'+
    '	select {\n'+
    '\n' +
    '	case nextStory, isOpen := <-inc:\n'+
    '		// check if there is a new story\n'+
    '\n' +
    '		if !isOpen {\n'+
    '			// if the line closed\n'+
    '			// TODO: decide if we want it to stay open or go for close here...\n'+
    '			// stop the analyze task\n'+
    '			finishAnalyzer <- true\n'+
    '\n' +
    '			// stop reading from the in chan\n'+
    '			inc = nil\n'+
    '\n' +
    '			// continue on\n'+
    '			break\n'+
    '		}\n'+
    '\n' +
    '		// set the current story\n'+
    '		istory = nextStory\n'+
    '\n' +
    '		// once you have a story, build the output and set the inc to nil\n'+
    "		// so that we don't get another story too soon\n"+
    '		inc = nil\n'+
    '\n' +
    '		// build the output story\n'+
    '		ostory.MainArticle = istory.MainArticle\n'+
    '		ostory.RelatedArticles = make(chan analyzer.Analyzable)\n'+
    '\n' +
    '		// enable sending to out\n'+
    '		outc = m.out\n'+
    '\n' +
    '	case outc <- ostory:\n'+
    '		// send the output story down the line\n'+
    '\n' +
    '		// wait to start processing the next story until you can pass it down the line\n'+
    '		outc = nil\n'+
    '\n' +
    '		// once we know we have someone to read, we can send down the line\n'+
    '		storyc = istory.RelatedArticles\n'+
    '\n' +
    '		// set this up for the analyze function\n'+
    '		m.mainArticle = istory.MainArticle\n'+
    '\n' +
    '	case next, isOpen := <-storyc:\n'+
    '		// read stories from upstream until stream closes\n'+
    '\n' +
    "		storyc = nil // no mater what, don't go get the next article yet\n"+
    '\n' +
    '		if !isOpen {\n'+
    "			// we are at the end of the current story's line\n"+
    '			// read on closed chan returns zero value (null in this case)\n'+
    "			// TODO: sync this with the analyze func so we don't write on closed stream\n"+
    '			close(ostory.RelatedArticles)\n'+
    '			ostory.RelatedArticles = nil\n'+
    '\n' +
    '			// start looking for the next story in the stream\n'+
    '			inc = m.in\n'+
    '			break // get out of this if statement\n'+
    '		}\n'+
    '\n' +
'\n'+
    '		// send what we just read over to the analyzer\n'+
    '		freshArticle = next\n'+
    '		analyze = analyze_in\n'+
    '\n' +
    '	case analyze <- freshArticle:\n'+
    '		// send the most recent article over to the analyzer\n'+
    '		analyze = nil\n'+
    '\n' +
    '	case processedArticle = <-analyze_out:\n'+
    '		// read the result from the analyzer\n'+
    '		results = ostory.RelatedArticles\n'+
    '\n' +
    '	case results <- processedArticle:\n'+
    '		// write the article to the next stage\n'+
    '		results = nil\n'+
    '\n' +
    '		// read the next article in\n'+
    '		storyc = istory.RelatedArticles\n'+
    '\n' +
    '	case errc := <-m.closing:\n'+
    '		// send the error back\n'+
    '\n' +
    '		// close anything open\n'+
    '		m.in = nil\n'+
    '		close(m.err)\n'+
    '		close(m.closing)\n'+
    '		close(m.out)\n'+
    '\n' +
    '		// close the chans we created\n'+
    '		close(analyze_in)\n'+
    '		if ostory.RelatedArticles != nil {\n'+
    '			close(ostory.RelatedArticles)\n'+
    '		}\n'+
    '\n' +
    '		finishAnalyzer <- true\n'+
    '		close(finishAnalyzer)\n'+
    '\n' +
    '		errc <- err\n'+
    '\n' +
    '		return\n'+
    '\n' +
    '	case bigErr := <-m.err:\n'+
    '		// what to do when a big error comes along\n'+
    '		// propogate and return\n'+
    '		m.err <- bigErr\n'+
    '		err = bigErr\n'+
    '\n' +
    '		// ignore any values that will come down the line\n'+
    '		// TODO: make all these nil\n'+
    '		m.in = nil\n'+
    '\n' +
    '	}\n'+
    '}\n')
    fd.write('}\n\n')


def WriteSetup(fd, name):
    WriteFuncHeader(fd, name, 'Setup', '')
    fd.write('m.out = make(chan AnalyzableStory, 1)\n')
    fd.write('m.closing = make(chan chan error)\n')
    fd.write('}\n\n')

def WriteClose(fd, name):
    fd.write('func (m *' + name + 'Module) Close() error {\n')
    fd.write('errc := make(chan error)\n')
    fd.write('m.closing <- errc\n')
    fd.write('return <-errc\n')
    fd.write('}\n\n')

def WriteSetInputChan(fd, name):
    WriteFuncHeader(fd, name, 'SetInputChan', 'inc chan AnalyzableStory')
    fd.write('m.in = inc\n')
    fd.write('}\n\n')

def WriteGetOutputChan(fd, name):
    fd.write('func (m *' + name + 'Module) GetOutputChan() chan AnalyzableStory {\n')
    fd.write('return m.out\n')
    fd.write('}\n\n')

def WriteSetErrorChan(fd, name):
    WriteFuncHeader(fd, name, 'SetErrorPropogateChan', 'errc chan error')
    fd.write('m.err = errc\n')
    fd.write('}\n\n')

def WriteFooter(fd, name):
    fd.write('// check that the module was compiled properly\n')
    fd.write('var _ Module = (*' + name + 'Module)(nil)\n')
def WriteHeader(fd, name):
    fd.write('package pipeline\n')
    fd.write('import (\n"github.com/opinionated/analyzer-core/analyzer"\n)\n')

def WriteStruct(fd, name):
    fd.write('type ' + name + 'Module struct {\n')
    fd.write('in\tchan\tAnalyzableStory\n')
    fd.write('out\tchan\tAnalyzableStory\n')

    fd.write('err\tchan\terror\n')

    fd.write('closing\tchan\tchan\terror\n')
    fd.write('mainArticle analyzer.Analyzable\n')
    fd.write('}\n')

def WriteFuncHeader(fd, name, func, params):
    fd.write('func (m *' + name + 'Module) ' + func + '(' + params + ') {\n')

if __name__ == '__main__':
    main()

