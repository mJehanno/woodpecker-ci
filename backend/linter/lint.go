package linter

import (
	"errors"
	"fmt"
	"os"
	"path"
	"woodpecker-ci/logger"

	"github.com/muesli/termenv"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/linter"
)

func LintFile(filepath, rawConfig string) error {
	logger := logger.GetLogger()
	output := termenv.NewOutput(os.Stdout)
	errs := []error{}

	logger.Info("parsing file")
	c, err := yaml.ParseString(rawConfig)
	if err != nil {
		return err
	}

	config := &linter.WorkflowConfig{
		File:      path.Base(filepath),
		RawConfig: rawConfig,
		Workflow:  c,
	}

	logger.Info("linting file")
	err = linter.New(linter.WithTrusted(true)).Lint([]*linter.WorkflowConfig{config})
	if err != nil {
		logger.Infof("🔥 %s has warnings / errors:\n", output.String(config.File).Underline())

		hasErrors := false
		for _, err := range pipeline_errors.GetPipelineErrors(err) {
			line := "  "

			if err.IsWarning {
				line = fmt.Sprintf("%s ⚠️ ", line)
			} else {
				line = fmt.Sprintf("%s ❌", line)
				errs = append(errs, err)
				hasErrors = true
			}

			if data := err.GetLinterData(); data != nil {
				line = fmt.Sprintf("%s %s\t%s", line, output.String(data.Field).Bold(), err.Message)
			} else {
				line = fmt.Sprintf("%s %s", line, err.Message)
			}

			// TODO: use table output
			logger.Infof("%s\n", line)
		}

		if hasErrors {
			err = errors.Join(errs...)
			logger.Error("config has error")
			return fmt.Errorf("config has errors -> %w", err)
		}

		return nil
	}

	logger.Info("✅ Config is valid")

	return nil
}
