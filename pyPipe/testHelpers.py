"""
    Funcs to help with unit testing modules
"""
import copy
import core

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

