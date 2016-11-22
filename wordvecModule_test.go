package pipeline_test

import (
	"fmt"
	"github.com/opinionated/word2vec"
	"testing"
)

func TestSimpleWordVec(t *testing.T) {
	fmt.Println("hey hey")

	client := word2vec.Client{Addr: "localhost:1234"}

	a := word2vec.Expr{}
	a.Add(1, "Obama")

	b := word2vec.Expr{}
	b.Add(1, "Clinton")

	score, err := client.Cos(a, b)

	if err != nil {
		panic(err)
	}

	fmt.Println("score:", score)
}
