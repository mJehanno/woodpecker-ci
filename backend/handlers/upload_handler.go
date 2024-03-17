package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"woodpecker-ci/logger"

	"github.com/labstack/echo"
	"github.com/muesli/termenv"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/linter"
)

type PipelineFileInput struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func lintFile(filepath string) error {
	logger := logger.GetLogger()
	output := termenv.NewOutput(os.Stdout)

	logger.Info("opening file")
	fi, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer fi.Close()

	logger.Info("reading file")
	buf, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	rawConfig := string(buf)

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
			logger.Error("config has error")
			return errors.New("config has errors")
		}

		return nil
	}

	logger.Info("✅ Config is valid")

	return nil
}

func Upload(c echo.Context) error {
	logger := logger.GetLogger()
	input := PipelineFileInput{}
	err := c.Bind(&input)
	if err != nil {
		logger.WithError(err).Error("failed to bind body params")
		return c.String(http.StatusBadRequest, err.Error())
	}

	if !strings.Contains(input.Type, "yaml") {
		logger.Error("file is not a yaml file")
		return c.String(http.StatusBadRequest, errors.New("need to provide an yaml file").Error())
	}

	if err = lintFile(input.Path); err != nil {
		logger.WithError(err).Error("file is not a valid woodpecker pipeline")
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, input)
}
