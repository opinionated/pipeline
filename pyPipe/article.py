"""
    Core structures for python    

    Article is the structure that gets passed around
    It stores "scores" and provides getters / setters for these
    
    Modules "build" article info by computing scores then adding them to articles
    Each module should only provide a "score" func that accepts a main article and a related article

    Pipelines manage running modules and shuffling the data around
    We should be able to change the pipeline without changing the article or modules

    This file defines article and module base classes
"""

class Article():
    def __init__(uuid):
        self.uuid = uuid
        self.scores = {}
    
    # TODO: should we define a class for scores? We probably want them to be serializable?
    # TODO: how should we handle the error cases?
    def AddScore(name, val):
        if name in self.scores:
            # error case
            return False

        self.scores[name] = val

    def GetScore(name):
        if name not in self.scores:
            # error case
            return False

        return scores[name]

    def GetUUID():
        return self.uuid

class Module():
    def Analyze(mainArticle, relatedArticle):
        return False
