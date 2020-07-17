package otfclassifier

import (
	"encoding/json"
	"log"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

var m *mapping.IndexMappingImpl
var i bleve.Index

func InitTokeniser() error {
	var err error
	m = bleve.NewIndexMapping()
	//i, err = bleve.New("bleve.index", m)
	i, err = bleve.NewMemOnly(m)
	return err
}

// both tokenise a string and index a record (which contains the string)
// if id is not empty, index the string under id as an identifier
func Tokenise(id string, txt string, record interface{}) []string {
	//log.Printf("id %s txt %s\n", id, txt)
	if len(id) > 0 {
		//log.Println("Indexed " + id)
		index(id, record)
	}
	// expose the tokenisation of txt; this has already been done in indexing it
	tokenstream, err := m.AnalyzeText(m.DefaultAnalyzer, []byte(txt))
	if err != nil {
		return []string{}
	}
	ret := make([]string, 0)
	for _, t := range tokenstream {
		ret = append(ret, string(t.Term))
	}
	return ret
	//return normalise_tokens(tokenize.TextToWords(normalise_text(txt)))
}

func index(id string, data interface{}) error {
	return i.Index(id, data)
}

// simple Bleve search. Return JSON of search results
func Search(query string) ([]byte, error) {
	//log.Println("Searching for " + query)
	s := bleve.NewSearchRequest(bleve.NewMatchQuery(query))
	ret, err := i.Search(s)
	if err == nil {
		//log.Println(ret)
		json, err := json.Marshal(ret)
		//log.Println(string(json))
		return json, err
	} else {
		log.Println("ERROR")
		log.Println(err)
		return []byte{}, err
	}
}
