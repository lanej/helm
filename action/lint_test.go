package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/helm/helm/test"
	"github.com/helm/helm/util"

	"gopkg.in/yaml.v2"
)

func TestLintSuccess(t *testing.T) {
	tmpHome := test.CreateTmpHome()

	chartName := "goodChart"

	Create(chartName, tmpHome)

	output := test.CaptureOutput(func() {
		Lint(util.WorkspaceChartDirectory(tmpHome, chartName))
	})

	expected := "Chart [goodChart] has passed all necessary checks"

	test.ExpectContains(t, output, expected)
}

func TestLintMissingReadme(t *testing.T) {
	tmpHome := test.CreateTmpHome()

	chartName := "badChart"

	Create(chartName, tmpHome)

	os.Remove(filepath.Join(util.WorkspaceChartDirectory(tmpHome, chartName), "README.md"))

	output := test.CaptureOutput(func() {
		Lint(util.WorkspaceChartDirectory(tmpHome, chartName))
	})

	test.ExpectContains(t, output, "A README file was not found")
}

func TestLintMismatchNameAndDir(t *testing.T) {

}

func TestLintMissingChartYaml(t *testing.T) {
	tmpHome := test.CreateTmpHome()

	chartName := "badChart"

	Create(chartName, tmpHome)

	os.Remove(filepath.Join(util.WorkspaceChartDirectory(tmpHome, chartName), "Chart.yaml"))

	output := test.CaptureOutput(func() {
		Lint(util.WorkspaceChartDirectory(tmpHome, chartName))
	})

	test.ExpectContains(t, output, "A Chart.yaml file was not found")
	test.ExpectContains(t, output, "Chart [badChart] failed some checks")
}

func TestLintMissingManifestDirectory(t *testing.T) {
	tmpHome := test.CreateTmpHome()

	chartName := "brokeChart"

	Create(chartName, tmpHome)

	os.RemoveAll(filepath.Join(util.WorkspaceChartDirectory(tmpHome, chartName), "manifests"))

	output := test.CaptureOutput(func() {
		Lint(util.WorkspaceChartDirectory(tmpHome, chartName))
	})

	test.ExpectMatches(t, output, fmt.Sprintf("A manifests directory was not found.*%s", chartName))
	test.ExpectContains(t, output, fmt.Sprintf("Chart [%s] failed some checks", chartName))
}

func TestLintEmptyChartYaml(t *testing.T) {
	tmpHome := test.CreateTmpHome()

	chartName := "badChart"

	Create(chartName, tmpHome)

	badChartYaml, _ := yaml.Marshal(make(map[string]string))

	chartYaml := util.WorkspaceChartDirectory(tmpHome, chartName, "Chart.yaml")

	os.Remove(chartYaml)
	ioutil.WriteFile(chartYaml, badChartYaml, 0644)

	output := test.CaptureOutput(func() {
		Lint(util.WorkspaceChartDirectory(tmpHome, chartName))
	})

	test.ExpectContains(t, output, "Missing Name specification in Chart.yaml file")
	test.ExpectContains(t, output, "Missing Version specification in Chart.yaml file")
	test.ExpectContains(t, output, "Missing description in Chart.yaml file")
	test.ExpectContains(t, output, "Missing maintainers information in Chart.yaml file")
	test.ExpectContains(t, output, fmt.Sprintf("Chart [%s] failed some checks", chartName))
}

func TestLintBadPath(t *testing.T) {
	tmpHome := test.CreateTmpHome()
	chartName := "badChart"

	output := test.CaptureOutput(func() {
		Lint(util.WorkspaceChartDirectory(tmpHome, chartName))
	})

	test.ExpectContains(t, output, chartName+" not found in workspace")
}
