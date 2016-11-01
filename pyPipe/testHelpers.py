"""
    Funcs to help with unit testing modules
"""
import copy
import Queue
import core
import threading

# runs articles thru the modules
# use this to maybe test pipes?
class ModuleTestRunner():

    def __init__(self, module, mainArticle, relatedArticles):
        # TODO: do we care about the type of the main article?
        assert isinstance(module, core.Module)

        self.module = module
        self.mainArticle = mainArticle
        self.relatedArticles = relatedArticles
        self.outputArticles = []

    # send all the articles thru the module
    def Run(self):
        for article in self.relatedArticles:
            assert isinstance(article, core.Article)

            # make a copy so we don't mess up our original data
            articleCopy = copy.deepcopy(article)

            # TODO: error check this
            self.module.Analyze(self.mainArticle, articleCopy)
            self.outputArticles.append(articleCopy)

    def GetResults(self):
        return self.outputArticles

# for sending articles between modules
# wraps a queue of TestArticlePairs
class TestCommChannel():
    def __init__(self):
        self.queue = Queue.Queue()
    
    def Push(self, pair):
        assert isinstance(pair, TestArticlePair)
        self.queue.put(pair)

    def Has(self):
        return not self.queue.empty()
    
    def Get(self):
        return self.queue.get()

# main, related article pair, send thru test comm channels
class TestArticlePair():
    def __init__(self, main, related):
        assert isinstance(main, core.Article)
        assert isinstance(related, core.Article)
        self.main = main
        self.related = related

# wraps around a module, handles sending articles around
class TestModuleWrapper():
    def __init__(self, module, inChan, outChan):
        assert isinstance(inChan, TestCommChannel)
        assert isinstance(outChan, TestCommChannel)
        self.inChan = inChan
        self.outChan = outChan

        assert isinstance(module, core.Module)
        self.module = module
        
        self.isProcessing = False

    def SetOutChan(self, chan):
        assert isinstance(outChan, TestCommChannel)
        self.outChan = outChan

    def SetInChan(self, chan):
        assert isinstance(inChan, TestCommChannel)
        self.inChan = chan

    def GetInChan(self):
        return self.inChan

    def GetOutChan(self):
        return self.outChan

    # module empty when not processing and no articles on recv chan
    def Empty(self):
        #TODO: this may be unsafe if we end up threading
        return not self.inChan.Has() and not self.isProcessing

    def Start(self):
        self.running = True
        self.thread = threading.Thread(target=self._run)
        self.thread.start()

    def Stop(self):
        self.running = False
        self.thread.join()

    def _run(self):
        while self.running:
            if(self.inChan.Has()):
                pair = self.inChan.Get()
                self.module.Analyze(pair.main, pair.related)
                self.outChan.Push(pair)

            # don't burn the cpu
            time.sleep(0.01)

class TestPipeline():
    def __init__(self):
        self.modules = []
        self.running = False

    def AddModule(self, module):
        # build teh wrapper and wire it up
        inChan = None
        if len(self.modules) == 0:
            inChan = TestCommChannel()
        else:
            last = self._lastModule()
            inChan = last.GetOutChan()

        wrap = TestModuleWrapper(module, inChan, TestCommChannel())
        wrap.Start()

        self.modules.append(wrap)
    
    def _lastModule(self):
        n = len(self.modules)
        return self.modules[n - 1]

    def _firstModule(self):
        if len(self.modules) == 0:
            return None
        return self.modules[0]

    def PushPair(self, main, related):
        pair = TestArticlePair(main, related)
        first = self._firstModule()
        first.GetInChan().Push(pair)

    def PullPair(self):
        last = self._lastModule()
        outc = last.GetOutChan()
        if not outc.Has():
            return None
        else:
            return outc.Get()

    def AllThrough(self):
        for module in self.modules:
            if not module.Empty():
                return False
        return True

    def Close(self):
        for module in self.modules:
            module.Stop()

import time
# build and run a pipeline
# queues the output
class PipelineTestRunner():
    def __init__(self, modules):
        self.pipeline = TestPipeline()
        for module in modules:
            self.pipeline.AddModule(module)

    def Run(self, mainArticle, relatedArticles, timeout=30):
        for related in relatedArticles:
            self.pipeline.PushPair(mainArticle, related)

        
        # run through in 0.5 s steps
        for i in xrange(2 * timeout):
            if self.pipeline.AllThrough():
                return
            time.sleep(0.5)

        # shouldn't ever reach here
        assert False

    # build an array of the output articles
    def Get(self):
        assert self.pipeline.AllThrough()

        related = []
        while True:
            pair = self.pipeline.PullPair()
            if pair is None:
                break

            related.append(pair.related)

        return related

    def Close(self):
        self.pipeline.Close()
