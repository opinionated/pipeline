
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

if __name__=="__main__":
  unittest.main()

