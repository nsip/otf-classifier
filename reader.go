package align

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

/* Learning Area; Indicator vs DevLevel; Indicator/DevLevel => { Text; DevLevel } */

type Curriculum = map[string]map[string]map[string]*CurricContent
type CurricContent struct {
	Text     []string
	DevLevel string
}

func read_curriculum(path string) (Curriculum, error) {
	var lp []map[string]interface{}
	files, _ := filepath.Glob(path + "/*.json")
	if len(files) == 0 {
		log.Fatalln("No *.json curriculum files found in input folder" + path)
	}

	ret := make(map[string]map[string]map[string]*CurricContent)
	for _, filename := range files {
		dat, err := ioutil.ReadFile(filename)
		if err != nil {
			return ret, err
		}
		json.Unmarshal([]byte(dat), &lp)
		// fmt.Printf("%+v\n", lp)
		for _, r := range lp {
			result := make(map[string]*CurricContent)
			result = parse_lp(r, result, "", true)
			for k, v := range result {
				fmt.Printf("%s\t%s\n", k, strings.Join(v.Text, "; "))
			}
			key := r["text"].(string)
			ret[key] = make(map[string]map[string]*CurricContent)
			ret[key]["Indicator"] = result
			result = make(map[string]*CurricContent)
			result = parse_lp(r, result, "", false)
			for k, v := range result {
				fmt.Printf("%s\t%s\n", k, strings.Join(v.Text, "; "))
			}
			ret[key]["Devlevel"] = result
		}
	}
	return ret, nil
}

func parse_lp(r map[string]interface{}, result map[string]*CurricContent, devlevel string, indicator bool) map[string]*CurricContent {
	l, err := dig(r, "asn_statementLabel", "literal")
	if err != nil {
		return result
	}

	if l == "Progression Level" {
		devlevel, err = dig(r, "asn_statementNotation", "literal")
		if err != nil {
			panic(err)
		}
	}

	if l == "Indicator" {
		var key string
		if indicator {
			key = r["id"].(string)
		} else {
			key = devlevel
		}
		if _, ok := result[key]; !ok {
			result[key] = &CurricContent{Text: make([]string, 0), DevLevel: devlevel}
		}
		result[key].Text = append(result[key].Text, r["text"].(string))
	}

	c, ok := r["children"]
	if ok {
		for _, r1 := range c.([]interface{}) {
			result = parse_lp(r1.(map[string]interface{}), result, devlevel, indicator)
		}
	}
	return result
}

func dig(r map[string]interface{}, key1 string, key2 string) (string, error) {
	l, ok := r[key1]
	if !ok {
		fmt.Printf("%+v", r)
		fmt.Println("Fail 1 " + key1)
		return "", errors.New("missing")
	}
	m, ok := l.(map[string]interface{})[key2]
	if !ok {
		fmt.Printf("%+v", l)
		fmt.Printf("Fail 2 " + key2)
		return "", errors.New("missing")
	}
	return m.(string), nil
}
