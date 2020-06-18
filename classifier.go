package align

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	//"github.com/juliangruber/go-intersect"
	"github.com/labstack/echo/v4"
	"github.com/nsip/curriculum-align/bayesian"
	"gopkg.in/fatih/set.v0"
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
	return response
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
}

func Align(c echo.Context) error {
	var learning_area, text string
	learning_area = c.QueryParam("area")
	text = c.QueryParam("text")
	log.Printf("Area: %s\nText: %s\n", learning_area, text)
	if text == "" {
		err := fmt.Errorf("text parameter not supplied")
		c.String(http.StatusBadRequest, err.Error())
		return err
	}
	if learning_area != "Literacy" && learning_area != "Numeracy" {
		err := fmt.Errorf("area parameter must be Literacy or Numeracy")
		c.String(http.StatusBadRequest, err.Error())
		return err
	}
	response := classify_text(classifiers[learning_area], curriculum[learning_area][granularity], text)
	return c.JSONPretty(http.StatusOK, response, "  ")
}
