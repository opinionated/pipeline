package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/opinionated/analyzer-core/analyzer"
	"net/http"
	"net/url"
)

type articleStore struct {
	start *articleNode
	limit int
}

func (store *articleStore) add(ID int32, data map[string]interface{}) {
	node := new(articleNode)
	node.ID = ID
	node.data = data
	node.next = store.start

	store.start = node

	itr := node
	for i := 0; i < store.limit-1; i++ {
		if itr.next == nil {
			// hit end of list before we hit limit
			return
		}

		itr = itr.next
	}

	// moved up to limit, then nil the next
	itr.next = nil
}

func (store *articleStore) get(ID int32) (data map[string]interface{}) {
	itr := store.start
	for itr != nil {
		if itr.ID == ID {
			return itr.data
		}

		itr = itr.next
	}

	return make(map[string]interface{})
}

type articleNode struct {
	next *articleNode
	ID   int32
	data map[string]interface{} // data from a run
}

func serializePipelineResult(articles []analyzer.Analyzable) ([]byte, error) {
	//var resultMap map[string]interface{}
	resultMap := make(map[string]interface{})
	for _, article := range articles {
		resultMap[article.FileName] = article.Score
	}

	return json.Marshal(resultMap)
}

// send an article to run thru the pipeline
// returns a reference to the run and stores the debug info
func handleDebugRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	article := vars["article"]

	cleanedArticle, err := url.QueryUnescape(article)
	if err != nil {
		panic(err)
	}

	pipe, err := buildPipeline()
	if err != nil {
		panic(err)
	}

	result, err := runArticle(pipe, cleanedArticle)
	if err != nil {
		panic(err)
	}

	bytes, err := serializePipelineResult(result)
	if err != nil {
		panic(err)
	}

	w.Write(bytes)
}

func startServer(port string) {
	router := mux.NewRouter()

	router.Handle("/debug", debugAPIHandler())

	http.ListenAndServe(port, router)
}

func debugAPIHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/run/{article}", handleDebugRun).Methods("POST")

	return router
}
