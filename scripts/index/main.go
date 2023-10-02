//nolint:all
package main

import (
	"encoding/json"
	"fmt"

	"github.com/blevesearch/bleve/v2"
)

func main() {
	index, err := bleve.Open("example.bleve")
	if err != nil {
		panic(err)
	}

	_, err = index.Document("ash")

	query := bleve.NewMatchQuery("bleve")

	res, err := index.Search(bleve.NewSearchRequest(query))
	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
