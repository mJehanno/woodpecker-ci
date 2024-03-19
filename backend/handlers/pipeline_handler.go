package handlers

import (
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
