package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func httpResponse(method string, url string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, nil)

	debugAPIHandler().ServeHTTP(recorder, req)

	return recorder
}

// TestRunArticle jsut makes sure it works
func TestRunArticle(t *testing.T) {
	article := "The Horror in San Bernardino"
	escedArticle := url.QueryEscape(article)
	response := httpResponse("POST", "/run/"+escedArticle)

	assert.Equal(t, 200, response.Code)

	var data map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &data)
	assert.Nil(t, err)
}
