package handlers

import (
	"woodpecker-ci/linter"
	"woodpecker-ci/pipeline"
	"woodpecker-ci/runner"

	"github.com/labstack/echo"
)

func GetPipeline(c echo.Context) error {
	pipes, err := pipeline.GetPipeline(c.Request().Context())

	if err != nil {
		return c.String(500, err.Error())
	}

	return c.JSON(200, pipes)
}

func LintPipeline(c echo.Context) error {
	file := PipelineFileInput{}
	c.Bind(&file)
	err := linter.LintFile(file.Path, file.Content)
	if err != nil {
		c.String(400, err.Error())
	}
	return c.JSON(200, nil)
}

type StartPipelineInput struct {
	runner.RunnerConfigOptions `json:"config"`
	Path                       string `json:"path"`
	File                       string `json:"file"`
}

func StartPipeline(c echo.Context) error {
	config := StartPipelineInput{}
	c.Bind(&config)
	err := runner.RunExec(nil, config.RunnerConfigOptions, config.File, config.Path)
	if err != nil {
		c.String(400, err.Error())
	}
	return c.JSON(200, nil)
}
