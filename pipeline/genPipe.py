import sys


def main():
    print "\n\twelcome to the pipeline module creator"
    print "\tthis tool makes generating new pipe modules super duper easy"
    print "\tto generate a module, just enter the module name"
    print '\texample input: "taxonomy" would create a new module called "TaxonomyModule"'
    print '\tin the file "taxonomyModule.go"\n'
    print '\tonce you are done generating the module, run go fmt __new file__ to make the code look pretty\n'
    pipe_name = raw_input("enter pipe module name: ")
    
    CreateSimplePipe(pipe_name)


def CreateSimplePipe(name):
    name = name[:1].lower() + name[1:] 
    fd = open(name + "Module.go", "w")
    
    name = name[:1].upper() + name[1:] 
    WriteHeader(fd, name)
    WriteStruct(fd, name)

    WriteAnalyze(fd, name)
    WriteSetup(fd, name)
    WriteClose(fd, name)
    WriteSetInputChan(fd, name)
    WriteGetOutputChan(fd, name)
    WriteSetErrorChan(fd, name)
    WriteFooter(fd, name)
  
def WriteAnalyze(fd, name):
    fd.write('// Analyze _____.\n')
    fd.write('func (m *' + name + ') Analyze (main analyzer.Analyzable,\n' +
            'related *analyzer.Analyzable) (bool, error) {\n')
    fd.write("\n return true, nil\n")
    fd.write('}\n\n')

def WriteSetup(fd, name):
    fd.write("// Setup _______.\n")
    WriteFuncHeader(fd, name, 'Setup', '')
    fd.write('m.out = make(chan AnalyzableStory, 1)\n')
    fd.write('m.closing = make(chan chan error)\n')
    fd.write('}\n\n')

def WriteClose(fd, name):
    fd.write('// Close stops the module and cleans up any open connections.\n')
    fd.write('func (m *' + name + 'Module) Close() error {\n')
    fd.write('errc := make(chan error)\n')
    fd.write('m.closing <- errc\n')
    fd.write('return <-errc\n')
    fd.write('}\n\n')

def WriteSetInputChan(fd, name):
    fd.write("// SetInputChan sets the module's input channel.\n")
    WriteFuncHeader(fd, name, 'SetInputChan', 'inc chan AnalyzableStory')
    fd.write('m.in = inc\n')
    fd.write('}\n\n')

def WriteGetOutputChan(fd, name):
    fd.write('// GetOutputChan returns the modules output channel.\n')
    fd.write('func (m *' + name + 'Module) GetOutputChan() chan AnalyzableStory {\n')
    fd.write('return m.out\n')
    fd.write('}\n\n')

def WriteSetErrorChan(fd, name):
    fd.write('// SetErrorPropogateChan sets the channel for errors to propagate out\n')
    fd.write('// of this module.\n')
    WriteFuncHeader(fd, name, 'SetErrorPropogateChan', 'errc chan error')
    fd.write('m.err = errc\n')
    fd.write('}\n\n')

def WriteEndMethods(fd, name):
  fd.write('// remaining methods are used internally by run methods\n\n'+

    'func (m *' + name + 'Module) getErrorPropogateChan() chan error { \n'+
    '\treturn m.err\n'+
    '}\n'+
    '\n'+
    '\n'+
    'func (m *' + name + 'Module) getInputChan() chan Story { \n'+
    '\treturn m.in \n'+
    '}\n'+
    '\n'+
    'func (m *' + name + 'Module) getClose() chan chan error { \n'+
    '\treturn m.closing\n'+
    '}\n')
 

def WriteFooter(fd, name):
    fd.write('// check that the module was compiled properly\n')
    fd.write('var _ Module = (*' + name + 'Module)(nil)\n')

def WriteHeader(fd, name):
    fd.write('package pipeline\n')
    fd.write('import (\n"github.com/opinionated/analyzer-core/analyzer"\n)\n')

def WriteStruct(fd, name):
    fd.write('// ' + name + 'Module _______.\n')
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

