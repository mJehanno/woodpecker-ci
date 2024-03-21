package handlers

import (
	"errors"
	"net/http"
	"strings"
	"woodpecker-ci/linter"
	"woodpecker-ci/logger"
	"woodpecker-ci/pipeline"

	"github.com/labstack/echo"
)

type PipelineFileInput struct {
	pipeline.File
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (p PipelineFileInput) ToDbModel() pipeline.Pipeline {
	return pipeline.Pipeline{
		File: p.File,
	}
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

	if err = linter.LintFile(input.Path, input.Content); err != nil {
		logger.WithError(err).Error("file is not a valid woodpecker pipeline")
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = pipeline.SavePipeline(c.Request().Context(), input.ToDbModel())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, input)
}
