package action

import (
	"io/ioutil"
	"path/filepath"

	"github.com/helm/helm/chart"
	"gopkg.in/yaml.v2"
)

const (
	warningLevel = 0
	errorLevel   = 1
)

// Validation represents a specific type of validation against a specific directory
type Validation struct {
	chartfile *chart.Chartfile
	children  []*Validation
	Path      string
	validator validator
	Message   string
	Level     int
}

// ChartYamlPath - path to Chart.yaml
func (v *Validation) ChartYamlPath() string {
	return filepath.Join(v.Path, "Chart.yaml")
}

// Chartfile - return chartfile, error in reading
func (v *Validation) Chartfile() (*chart.Chartfile, error) {
	if v.chartfile != nil {
		return v.chartfile, nil
	}

	var y *chart.Chartfile

	b, err := ioutil.ReadFile(v.ChartYamlPath())

	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(b, &y); err != nil {
		return nil, err
	}

	v.chartfile = y

	return y, nil
}

type validator func(v *Validation) (result bool)

func (v *Validation) addValidator(pv *Validation) {
	v.children = append(v.children, pv)
}

// AddError - add error level validation
func (v *Validation) AddError(message string, fn validator) *Validation {
	pv := &Validation{Message: message, validator: fn, Level: errorLevel, Path: v.Path}

	v.addValidator(pv)

	return pv
}

// AddWarning - add warning level validation
func (v *Validation) AddWarning(message string, fn validator) *Validation {
	pv := &Validation{Message: message, validator: fn, Level: warningLevel, Path: v.Path}

	v.addValidator(pv)

	return pv
}

// ChartName - return chart name as reported by the path
func (v *Validation) ChartName() string {
	return filepath.Base(v.Path)
}

func (v *Validation) valid() bool {
	return v.validator == nil || v.validator(v)
}

func (v *Validation) walk(talker func(_ *Validation)) {
	if v.validator != nil {
		talker(v)
	}

	if v.valid() {
		for _, pv := range v.children {
			pv.walk(talker)
		}
	}
}

// Validate - true if every validation passes, yeild function to report results
func (v *Validation) Validate(fn func(_ bool, _ *Validation)) bool {
	valid := true

	v.walk(func(cv *Validation) {
		vv := cv.valid()
		valid = valid && vv
		fn(valid, cv)
	})

	return valid
}
