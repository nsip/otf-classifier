package otfclassifier

import (
	"encoding/json"
	"errors"

	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/mitchellh/copystructure"

	strip "github.com/grokify/html-strip-tags-go"
)

/* Learning Area; Indicator vs DevLevel; Indicator/DevLevel => { Text; DevLevel } */

type Curriculum = map[string]map[string]map[string]*CurricContent
type CurricContent struct {
	Text     []string
	DevLevel string
	Path     []*Keyval
}
type Keyval struct {
	Key string
	Val string
}

func read_curriculum(path string) (Curriculum, error) {
	var r map[string]interface{}
	files, _ := filepath.Glob(path + "/*.json")
	if len(files) == 0 {
		log.Fatalln("No *.json curriculum files found in input folder = " + path)
	}

	ret := make(map[string]map[string]map[string]*CurricContent)
	for _, filename := range files {
		dat, err := ioutil.ReadFile(filename)
		if err != nil {
			return ret, err
		}
		json.Unmarshal(dat, &r)
		key := r["text"].(string)
		if key == "National Literacy Learning Progression" {
			key = "Literacy"
		} else if key == "National Numeracy Learning Progression" {
			key = "Numeracy"
		}
		path := make([]*Keyval, 0)
		result := make(map[string]*CurricContent)
		result = parse_lp(r, result, "", true, path)
		ret[key] = make(map[string]map[string]*CurricContent)
		ret[key]["Indicator"] = result
		result = make(map[string]*CurricContent)
		path = make([]*Keyval, 0)
		result = parse_lp(r, result, "", false, path) // needed for lookup
		ret[key]["Devlevel"] = result
	}
	return ret, nil
}

func parse_lp(r map[string]interface{}, result map[string]*CurricContent, devlevel string, indicator bool, path_input []*Keyval) map[string]*CurricContent {
	l, err := dig(r, "asn_statementLabel", "literal")
	if err != nil {
		// root does not have a label
		l = "General Capability"
	}

	name, err := dig(r, "asn_statementNotation", "literal")
	ok := true
	if err != nil {
		name, ok = r["id"].(string)
		if !ok {
			name, ok = r["text"].(string)
		}
	}

	raw, err := copystructure.Copy(path_input)
	if err != nil {
		panic(err)
	}
	path := raw.([]*Keyval)
	path = append(path, &Keyval{Key: l, Val: name})

	if l == "Progression level" {
		devlevel = name
	}

	if l == "Indicator" {
		var key string
		if indicator {
			key = name
		} else {
			key = devlevel
		}
		if _, ok := result[key]; !ok {
			result[key] = &CurricContent{Text: make([]string, 0), DevLevel: devlevel, Path: path}
		}
		result[key].Text = append(result[key].Text, strings.TrimSpace(strip.StripTags(r["text"].(string))))
	}

	if l == "Indicator" && indicator || l == "Progression level" && !indicator {
		id := name
		result[id] = &CurricContent{Text: make([]string, 0), DevLevel: devlevel, Path: path}
		result[id].Text = append(result[id].Text, strings.TrimSpace(strip.StripTags(r["text"].(string))))
	}

	c, ok := r["children"]
	if ok {
		raw, err := copystructure.Copy(path)
		if err != nil {
			panic(err)
		}
		orig_path := raw.([]*Keyval)
		for _, r1 := range c.([]interface{}) {
			result = parse_lp(r1.(map[string]interface{}), result, devlevel, indicator, path)
		}
		if l == "Progression level" && !indicator {
			id := name
			result[id].Path = orig_path
			result[devlevel].Path = orig_path
		}
	}
	return result
}

func dig(r map[string]interface{}, key1 string, key2 string) (string, error) {
	l, ok := r[key1]
	if !ok {
		// fmt.Printf("%+v", r)
		// fmt.Println("Fail 1 " + key1)
		return "", errors.New("missing")
	}
	m, ok := l.(map[string]interface{})[key2]
	if !ok {
		// fmt.Printf("%+v", l)
		// fmt.Printf("Fail 2 " + key2)
		return "", errors.New("missing")
	}
	return m.(string), nil
}
