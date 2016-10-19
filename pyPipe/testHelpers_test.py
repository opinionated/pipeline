
import testHelpers
import core
import unittest

class AddModule(core.Module):
  def __init__(self, name):
    self.counter = 0
    self.name = name
  
  def Analyze(self, main, related):
    assert isinstance(related, core.Article)
    related.AddScore(self.name, self.counter)
    self.counter += 1

class TestAddModule(unittest.TestCase):
  def test_simple(self):

    articleSet = []
    articleSet.append(core.Article("a"))
    articleSet.append(core.Article("b"))
    articleSet.append(core.Article("c"))

    addModule = AddModule("add")

    runner = testHelpers.ModuleTestRunner(addModule, "", articleSet)
    runner.Run()
    results = runner.GetResults()

    self.assertTrue(len(articleSet) == len(results))

    for i in xrange(len(articleSet)):
      assert(articleSet[i].GetUUID() == results[i].GetUUID())
      assert(results[i].GetScore("add") == i)

class TestPipeline(unittest.TestCase):
  def test_simple(self):
    articleSet = []
    articleSet.append(core.Article("a"))
    articleSet.append(core.Article("b"))
    articleSet.append(core.Article("c"))
    
    addModule = AddModule("add")
    secondAddModule = AddModule("2nd")

    runner = testHelpers.PipelineTestRunner([addModule, secondAddModule])
    runner.Run(core.Article("main"), articleSet)

    results = runner.Get()
    self.assertTrue(len(articleSet) == len(results))

    for i in xrange(len(articleSet)):
      assert(articleSet[i].GetUUID() == results[i].GetUUID())
      assert(results[i].GetScore("add") == i)
      assert(results[i].GetScore("2nd") == i)

    runner.Close()


if __name__=="__main__":
  unittest.main()

