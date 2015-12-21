package action

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/helm/helm/chart"
	"gopkg.in/yaml.v2"
)

// ChartValidation represents a specific instance of validation against a specific directory
type ChartValidation struct {
	Path        string
	Validations []*Validation
}

const (
	warningLevel = 0
	errorLevel   = 1
)

// Validation represents a specific type of validation against a specific directory
type Validation struct {
	children  []*Validation
	path      string
	validator validator
	Message   string
	level     int
}

// ChartYamlPath - path to Chart.yaml
func (v *Validation) ChartYamlPath() string {
	return filepath.Join(v.path, "Chart.yaml")
}

// Chartfile - return chartfile, error in reading
func (v *Validation) Chartfile() (*chart.Chartfile, error) {
	var y *chart.Chartfile

	b, err := ioutil.ReadFile(v.ChartYamlPath())

	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(b, &y); err != nil {
		return nil, err
	}

	return y, nil
}

type validator func(path string, v *Validation) (result bool)

func (cv *ChartValidation) addValidator(v *Validation) {
	cv.Validations = append(cv.Validations, v)
}

func (v *Validation) addValidator(pv *Validation) {
	v.children = append(v.children, pv)
}

// AddError - add error level validation
func (cv *ChartValidation) AddError(message string, fn validator) *Validation {
	v := new(Validation)
	v.Message = message
	v.validator = fn
	v.level = errorLevel
	v.path = cv.Path

	cv.addValidator(v)

	return v
}

// AddWarning - add warning level validation
func (cv *ChartValidation) AddWarning(message string, fn validator) *Validation {
	v := new(Validation)
	v.Message = message
	v.validator = fn
	v.level = warningLevel
	v.path = cv.Path

	cv.addValidator(v)

	return v
}

// AddError - add error level validation
func (v *Validation) AddError(message string, fn validator) *Validation {
	pv := new(Validation)
	pv.Message = message
	pv.validator = fn
	pv.level = errorLevel
	pv.path = v.path

	v.addValidator(pv)

	return pv
}

// AddWarning - add warning level validation
func (v *Validation) AddWarning(message string, fn validator) *Validation {
	pv := new(Validation)
	pv.Message = message
	pv.validator = fn
	pv.level = warningLevel
	pv.path = v.path

	v.addValidator(pv)

	return pv
}

// ChartName - return chart name as reported by the path
func (cv *ChartValidation) ChartName() string {
	return filepath.Base(cv.Path)
}

func (v *Validation) valid() bool {
	return v.validator(v.path, v)
}

func (v *Validation) walk(talker func(_ *Validation)) {
	talker(v)

	if v.valid() {
		for _, pv := range v.children {
			pv.walk(talker)
		}
	}
}

func (cv *ChartValidation) walk(talker func(v *Validation)) {
	for _, v := range cv.Validations {
		v.walk(talker)
	}
}

// Valid - true if every validation passes
func (cv *ChartValidation) Valid() bool {
	var valid bool

	cv.walk(func(v *Validation) {
		vv := v.valid()
		fmt.Println(fmt.Sprintf(v.Message+" : %v", vv))
		valid = valid && vv
	})

	return valid
}
