package handlers

import (
	"woodpecker-ci/linter"
	"woodpecker-ci/pipeline"

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
