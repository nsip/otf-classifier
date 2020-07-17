package otfclassifier

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nsip/otf-classifier/bayesian"
	set "gopkg.in/fatih/set.v0"
)

var granularity = "Indicator"

type ClassifierType struct {
	Classifier *bayesian.Classifier
	Classes    []bayesian.Class
}

var classifiers map[string]ClassifierType

var curriculum Curriculum

// create a classifier specific to components of the curriculum
func train_curriculum(curriculum map[string]*CurricContent) (ClassifierType, error) {
	classes := make([]bayesian.Class, 0)
	class_set := set.New(set.ThreadSafe)
	for key, _ := range curriculum {
		classes = append(classes, bayesian.Class(key))
		class_set.Add(key)
	}
	if len(classes) < 2 {
		return ClassifierType{}, fmt.Errorf("Not enough matching curriculum statements for classification")
	}
	// log.Printf("Training on: %+v\n", classes)
	classifier := bayesian.NewClassifierTfIdf(classes...)
	for key, record := range curriculum {
		if !class_set.Has(key) {
			continue
		}
		train := strings.Join(record.Text, " ")
		classifier.Learn(Tokenise(key, train, record.Text), bayesian.Class(key))
	}
	classifier.ConvertTermsFreqToTfIdf()
	ret := ClassifierType{Classifier: classifier, Classes: classes}
	return ret, nil
}

type AlignmentType struct {
	Item     string
	Text     string
	DevLevel string
	Path     []*Keyval
	Score    float64
	Matches  []bayesian.MatchStruct
}

//
// query params for classifier alignment
// supports query-string, form and json payload inputs
//
// Area: LP Capabilty, currently Literacy or Numeracy
// Text: the input to send to the classifier, such as observation or question text
//
type AlignmentQuery struct {
	Area string `json:"area" form:"area" query:"area"`
	Text string `json:"text" form:"text" query:"text"`
}

func keyval2path(path []*Keyval) string {
	b, _ := json.Marshal(path)
	return string(b)
}

func classify_text(classif ClassifierType, curriculum_map map[string]*CurricContent, input string) []AlignmentType {
	scores1, matches, _, _ := classif.Classifier.LogScores(Tokenise("", input, nil))
	response := make([]AlignmentType, 0)
	for i := 0; i < len(scores1); i++ {
		response = append(response, AlignmentType{
			Item:     string(classif.Classes[i]),
			Text:     strings.Join(curriculum_map[string(classif.Classes[i])].Text, " "),
			DevLevel: curriculum_map[string(classif.Classes[i])].DevLevel,
			Path:     curriculum_map[string(classif.Classes[i])].Path,
			Score:    scores1[i],
			Matches:  matches[i]})
	}
	sort.Slice(response, func(i, j int) bool { return response[i].Score > response[j].Score })
	return response[:5]
}

func Init() {
	var err error
	if err = InitTokeniser(); err != nil {
		log.Fatalln(err)
	}
	classifiers = make(map[string]ClassifierType)
	curriculum, err = read_curriculum("./curricula")
	if err != nil {
		log.Fatalln(err)
	}
	for k, _ := range curriculum {
		cl, err := train_curriculum(curriculum[k][granularity])
		if err != nil {
			log.Fatalln(err)
		} else {
			classifiers[k] = cl
		}
	}
	// log.Printf("model training complete")
}

func Align(c echo.Context) error {

	// //
	// // TODO: disable for production/release
	// // show the full inboud request
	// //
	// requestDump, err := httputil.DumpRequest(c.Request(), true)
	// if err != nil {
	// 	fmt.Println("req-dump error: ", err)
	// }
	// fmt.Println(string(requestDump))

	var learning_area, text string

	// check required params are in input
	aq := &AlignmentQuery{}
	if err := c.Bind(aq); err != nil {
		fmt.Println("align query binding failed")
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if aq.Area == "" || aq.Text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "must supply values for area and text")
	}

	learning_area = strings.Title(aq.Area) // in case query param in all upper/lower
	text = aq.Text
	if learning_area != "Literacy" && learning_area != "Numeracy" {
		err := fmt.Errorf("area parameter must be Literacy or Numeracy")
		c.String(http.StatusBadRequest, err.Error())
		return err
	}
	response := classify_text(classifiers[learning_area], curriculum[learning_area][granularity], text)
	return c.JSON(http.StatusOK, response)
}

func Keys(m map[string]*CurricContent) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Lookup(query string) (interface{}, error) {
	for k, _ := range curriculum {
		for level, _ := range curriculum[k] {
			//log.Printf("%s\t%s\t%+v\n", k, level, curriculum[k][level])
			if ret, ok := curriculum[k][level][query]; ok {
				return ret.Path, nil
			}
		}
	}
	return nil, fmt.Errorf("No such string found")
}
