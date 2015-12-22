package action

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/helm/helm/log"
	"github.com/helm/helm/util"
)

// Owner is default Helm repository owner or organization.
var Owner = "helm"

// Project is the default Charts repository name.
var Project = "charts"

// RepoService is a GitHub client instance.
var RepoService GHRepoService

// GHRepoService is a restricted interface to GitHub client operations.
type GHRepoService interface {
	DownloadContents(string, string, string, *github.RepositoryContentGetOptions) (io.ReadCloser, error)
}

// LintAll vlaidates all charts are well-formed
//
// - homedir is the home directory for the user
func LintAll(homedir string) {
	md := util.WorkspaceChartDirectory(homedir, "*")
	chartPaths, err := filepath.Glob(md)
	if err != nil {
		log.Warn("Could not find any charts in %q: %s", md, err)
	}

	if len(chartPaths) == 0 {
		log.Warn("Could not find any charts in %q", md)
	} else {
		for _, chartPath := range chartPaths {
			Lint(chartPath)
		}
	}
}

// Lint validates that a chart is well-formed
//
// - chartPath path to chart directory
func Lint(chartPath string) {
	v := Validation{Path: chartPath}

	directoryValidation := v.AddError("Directory exists", func(v *Validation) bool {
		stat, err := os.Stat(v.Path)

		return err == nil && stat.IsDir()
	})

	chartYamlValidation := directoryValidation.AddError("Chart.yaml is present", func(v *Validation) bool {
		stat, err := os.Stat(v.ChartYamlPath())

		return err == nil && stat.Mode().IsRegular()
	})

	chartYamlValidation.AddError("Has name", func(v *Validation) bool {
		chartfile, err := v.Chartfile()

		return err == nil && chartfile.Name != ""
	})

	chartYamlValidation.AddWarning("Has description", func(v *Validation) bool {
		chartfile, err := v.Chartfile()

		return err == nil && chartfile.Description != ""
	})

	chartYamlValidation.AddWarning("Has maintainers", func(v *Validation) bool {
		chartfile, err := v.Chartfile()

		return err == nil && chartfile.Maintainers != nil
	})

	chartYamlValidation.AddError("Has version", func(v *Validation) bool {
		chartfile, err := v.Chartfile()

		return err == nil && chartfile.Version != ""
	})

	valid := v.Validate(func(valid bool, v *Validation) {
		message := fmt.Sprintf(v.Message+" : %v", valid)

		if valid {
			log.Info(message)
		} else if v.Level == errorLevel {
			log.Err(message)
		} else {
			log.Warn(message)
		}
	})

	if valid {
		log.Info("Chart [%s] has passed all necessary checks", v.ChartName())
	} else {
		log.Info("Chart [%s] is not completely valid", v.ChartName())
	}
}
